package storage

import (
	"testing"
)

// TestInterfaceCompliance ensures R2Client implements the Storage interface at compile time.
func TestInterfaceCompliance(t *testing.T) {
	var _ Storage = (*R2Client)(nil)
	t.Log("R2Client implements the Storage interface (verified at compile time)")
}
