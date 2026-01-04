package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations.
type Cache interface {
	// Get retrieves a value from the cache.
	// Returns the value and true if found, nil and false if not found or expired.
	Get(ctx context.Context, key string) ([]byte, bool)

	// Set stores a value in the cache with the given TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache.
	Clear(ctx context.Context) error
}
