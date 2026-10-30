package storage

import (
	"testing"
)

func TestStorageConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config",
			config:  StorageConfig{},
			wantErr: true,
			errMsg:  "bucket is required",
		},
		{
			name: "invalid provider",
			config: StorageConfig{
				Provider:        "s3",
				Bucket:          "test-bucket",
				AccountID:       "acc123",
				AccessKeyID:     "key123",
				SecretAccessKey: "secret123",
			},
			wantErr: true,
			errMsg:  "unsupported provider",
		},
		{
			name: "valid R2 config with explicit provider",
			config: StorageConfig{
				Provider:        ProviderR2,
				Bucket:          "test-bucket",
				AccountID:       "account123",
				AccessKeyID:     "access123",
				SecretAccessKey: "secret123",
			},
			wantErr: false,
		},
		{
			name: "valid R2 config with default provider",
			config: StorageConfig{
				Bucket:          "test-bucket",
				AccountID:       "account123",
				AccessKeyID:     "access123",
				SecretAccessKey: "secret123",
			},
			wantErr: false,
		},
		{
			name: "R2 missing account ID",
			config: StorageConfig{
				Provider:        ProviderR2,
				Bucket:          "test-bucket",
				AccessKeyID:     "access123",
				SecretAccessKey: "secret123",
			},
			wantErr: true,
			errMsg:  "account_id is required",
		},
		{
			name: "R2 missing access key",
			config: StorageConfig{
				Provider:        ProviderR2,
				Bucket:          "test-bucket",
				AccountID:       "account123",
				SecretAccessKey: "secret123",
			},
			wantErr: true,
			errMsg:  "access_key_id is required",
		},
		{
			name: "R2 missing secret key",
			config: StorageConfig{
				Provider:    ProviderR2,
				Bucket:      "test-bucket",
				AccountID:   "account123",
				AccessKeyID: "access123",
			},
			wantErr: true,
			errMsg:  "secret_access_key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("StorageConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("StorageConfig.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestStorageConfig_GetEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		config   StorageConfig
		expected string
	}{
		{
			name: "R2 standard endpoint",
			config: StorageConfig{
				AccountID: "my-account-123",
			},
			expected: "https://my-account-123.r2.cloudflarestorage.com",
		},
		{
			name: "R2 custom endpoint override",
			config: StorageConfig{
				AccountID: "my-account-123",
				Endpoint:  "https://custom-r2.example.com",
			},
			expected: "https://custom-r2.example.com",
		},
		{
			name:     "empty account ID",
			config:   StorageConfig{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetEndpoint()
			if got != tt.expected {
				t.Errorf("StorageConfig.GetEndpoint() = %q, want %q", got, tt.expected)
			}
		})
	}
}
