package storage

import (
	"fmt"
)

// StorageConfig holds configuration for Cloudflare R2 storage
type StorageConfig struct {
	Provider        Provider `yaml:"provider"`
	Bucket          string   `yaml:"bucket"`
	AccountID       string   `yaml:"account_id,omitempty"`
	AccessKeyID     string   `yaml:"access_key_id,omitempty"`
	SecretAccessKey string   `yaml:"secret_access_key,omitempty"`
	Endpoint        string   `yaml:"endpoint,omitempty"`
}

// Validate checks if the configuration is valid for Cloudflare R2
func (c *StorageConfig) Validate() error {
	if c.Provider != "" && c.Provider != ProviderR2 {
		return fmt.Errorf("unsupported provider: %s (only Cloudflare R2 is supported)", c.Provider)
	}
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if c.AccountID == "" {
		return fmt.Errorf("account_id is required")
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("secret_access_key is required")
	}
	return nil
}

// GetEndpoint returns the endpoint URL for Cloudflare R2
func (c *StorageConfig) GetEndpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	if c.AccountID != "" {
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.AccountID)
	}
	return ""
}
