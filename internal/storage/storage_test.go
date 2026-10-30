package storage

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// MockStorage implements Storage interface for testing
type MockStorage struct {
	UploadFunc       func(ctx context.Context, key string, data []byte) error
	DownloadFunc     func(ctx context.Context, key string) ([]byte, error)
	DeleteFunc       func(ctx context.Context, key string) error
	DeleteBatchFunc  func(ctx context.Context, keys []string) error
	ListFunc         func(ctx context.Context, prefix string) ([]ObjectInfo, error)
	HeadFunc         func(ctx context.Context, key string) (*ObjectInfo, error)
	BucketExistsFunc func(ctx context.Context) (bool, error)
}

func (m *MockStorage) Upload(ctx context.Context, key string, data []byte) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, key, data)
	}
	return nil
}

func (m *MockStorage) Download(ctx context.Context, key string) ([]byte, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockStorage) DeleteBatch(ctx context.Context, keys []string) error {
	if m.DeleteBatchFunc != nil {
		return m.DeleteBatchFunc(ctx, keys)
	}
	return nil
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, prefix)
	}
	return nil, nil
}

func (m *MockStorage) Head(ctx context.Context, key string) (*ObjectInfo, error) {
	if m.HeadFunc != nil {
		return m.HeadFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockStorage) BucketExists(ctx context.Context) (bool, error) {
	if m.BucketExistsFunc != nil {
		return m.BucketExistsFunc(ctx)
	}
	return true, nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMockStorage_BucketExists_NotFound(t *testing.T) {
	ctx := context.Background()
	mock := &MockStorage{
		BucketExistsFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}

	exists, err := mock.BucketExists(ctx)
	if err != nil {
		t.Errorf("BucketExists() error = %v, want nil", err)
	}
	if exists {
		t.Errorf("BucketExists() = true, want false")
	}
}

func TestMockStorage_BucketExists_Error(t *testing.T) {
	ctx := context.Background()
	mock := &MockStorage{
		BucketExistsFunc: func(ctx context.Context) (bool, error) {
			return false, fmt.Errorf("failed to check bucket: access denied")
		},
	}

	exists, err := mock.BucketExists(ctx)
	if err == nil {
		t.Errorf("BucketExists() error = nil, want error")
	}
	if exists {
		t.Errorf("BucketExists() = true, want false")
	}
	if !contains(err.Error(), "access denied") {
		t.Errorf("BucketExists() error = %v, want error containing 'access denied'", err)
	}
}

func TestMockStorage_BucketExists_Success(t *testing.T) {
	ctx := context.Background()
	mock := &MockStorage{
		BucketExistsFunc: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	}

	exists, err := mock.BucketExists(ctx)
	if err != nil {
		t.Errorf("BucketExists() error = %v, want nil", err)
	}
	if !exists {
		t.Errorf("BucketExists() = false, want true")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *StorageConfig
		errMsg string
	}{
		{
			name:   "nil config",
			config: nil,
			errMsg: "storage config is required",
		},
		{
			name: "missing bucket",
			config: &StorageConfig{
				AccountID:       "acc123",
				AccessKeyID:     "key123",
				SecretAccessKey: "secret123",
			},
			errMsg: "bucket is required",
		},
		{
			name: "invalid provider",
			config: &StorageConfig{
				Provider:        "s3",
				Bucket:          "test",
				AccountID:       "acc123",
				AccessKeyID:     "key123",
				SecretAccessKey: "secret123",
			},
			errMsg: "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if err == nil {
				t.Errorf("New() expected error, got nil")
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("New() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestNew_ValidR2Config(t *testing.T) {
	cfg := &StorageConfig{
		Provider:        ProviderR2,
		Bucket:          "my-bucket",
		AccountID:       "acc123",
		AccessKeyID:     "key123",
		SecretAccessKey: "secret123",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() unexpected error = %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil client")
	}
}

func TestProviderConstants(t *testing.T) {
	if ProviderR2 != "r2" {
		t.Errorf("ProviderR2 = %q, want %q", ProviderR2, "r2")
	}
}

func TestMockStorage_Interface(t *testing.T) {
	var _ Storage = (*MockStorage)(nil)

	ctx := context.Background()
	mock := &MockStorage{
		UploadFunc: func(ctx context.Context, key string, data []byte) error {
			if key != "test-key" {
				t.Errorf("Upload key = %q, want %q", key, "test-key")
			}
			if string(data) != "test-data" {
				t.Errorf("Upload data = %q, want %q", string(data), "test-data")
			}
			return nil
		},
		DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return []byte("downloaded-data"), nil
		},
		BucketExistsFunc: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	}

	// Test Upload
	if err := mock.Upload(ctx, "test-key", []byte("test-data")); err != nil {
		t.Errorf("Upload() error = %v", err)
	}

	// Test Download
	data, err := mock.Download(ctx, "test-key")
	if err != nil {
		t.Errorf("Download() error = %v", err)
	}
	if string(data) != "downloaded-data" {
		t.Errorf("Download() = %q, want %q", string(data), "downloaded-data")
	}

	// Test BucketExists
	exists, err := mock.BucketExists(ctx)
	if err != nil {
		t.Errorf("BucketExists() error = %v", err)
	}
	if !exists {
		t.Errorf("BucketExists() = %v, want true", exists)
	}
}
