package storage

import (
	"context"
	"fmt"
	"time"
)

// Provider represents a storage provider type
type Provider string

const (
	ProviderR2 Provider = "r2"
)

// MaxDownloadSize is the maximum allowed size for a single downloaded object (100MB).
// This prevents memory exhaustion from oversized or malicious remote files.
const MaxDownloadSize = 100 * 1024 * 1024

// ObjectInfo contains metadata about a stored object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

// Storage defines the interface for cloud storage operations
type Storage interface {
	// Upload stores data with the given key
	Upload(ctx context.Context, key string, data []byte) error

	// Download retrieves data for the given key
	Download(ctx context.Context, key string) ([]byte, error)

	// Delete removes the object with the given key
	Delete(ctx context.Context, key string) error

	// DeleteBatch removes multiple objects in a single operation
	DeleteBatch(ctx context.Context, keys []string) error

	// List returns all objects with the given prefix
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)

	// Head returns metadata for the given key without downloading content
	Head(ctx context.Context, key string) (*ObjectInfo, error)

	// BucketExists checks if the configured bucket exists
	BucketExists(ctx context.Context) (bool, error)
}

// New creates a new Storage instance based on the provided configuration
func New(cfg *StorageConfig) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	return NewR2(cfg)
}
