package earthengine

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Cache is an interface for caching Earth Engine query results.
type Cache interface {
	// Get retrieves a cached value.
	Get(ctx context.Context, key string) (interface{}, bool, error)

	// Set stores a value in the cache.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache.
	Clear(ctx context.Context) error
}

// MemoryCache is an in-memory implementation of Cache.
type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]*cacheEntry
	maxSize int
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache.
//
// maxSize limits the number of entries (0 = unlimited).
func NewMemoryCache(maxSize int) *MemoryCache {
	cache := &MemoryCache{
		data:    make(map[string]*cacheEntry),
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a cached value.
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return nil, false, nil
	}

	// Check expiration
	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		return nil, false, nil
	}

	return entry.value, true, nil
}

// Set stores a value in the cache.
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check size limit
	if mc.maxSize > 0 && len(mc.data) >= mc.maxSize {
		// Evict oldest entry (simple FIFO)
		for k := range mc.data {
			delete(mc.data, k)
			break
		}
	}

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	mc.data[key] = &cacheEntry{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache.
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.data, key)
	return nil
}

// Clear removes all values from the cache.
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*cacheEntry)
	return nil
}

// Size returns the number of entries in the cache.
func (mc *MemoryCache) Size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return len(mc.data)
}

// cleanupExpired removes expired entries periodically.
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()
		for key, entry := range mc.data {
			if !entry.expiration.IsZero() && now.After(entry.expiration) {
				delete(mc.data, key)
			}
		}
		mc.mu.Unlock()
	}
}

// CacheKey generates a cache key from query parameters.
func CacheKey(params ...interface{}) string {
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// WithCache wraps a query function with caching.
//
// Example:
//
//	cache := earthengine.NewMemoryCache(1000)
//	cachedNDVI := earthengine.WithCache(cache, 1*time.Hour, func(ctx context.Context) (interface{}, error) {
//	    return helpers.NDVI(client, lat, lon, date)
//	})
//	result, err := cachedNDVI(ctx, lat, lon, date)
func WithCache(cache Cache, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error)) func(context.Context, ...interface{}) (interface{}, error) {
	return func(ctx context.Context, params ...interface{}) (interface{}, error) {
		// Generate cache key
		key := CacheKey(params...)

		// Try to get from cache
		if cached, found, err := cache.Get(ctx, key); err == nil && found {
			return cached, nil
		}

		// Execute function
		result, err := fn(ctx, params...)
		if err != nil {
			return nil, err
		}

		// Store in cache
		if err := cache.Set(ctx, key, result, ttl); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to cache result: %v\n", err)
		}

		return result, nil
	}
}

// CachedClient wraps a Client with caching.
type CachedClient struct {
	*Client
	cache Cache
	ttl   time.Duration
}

// NewCachedClient creates a client with built-in caching.
//
// Example:
//
//	client, _ := earthengine.NewClient(ctx, "credentials.json")
//	cache := earthengine.NewMemoryCache(1000)
//	cachedClient := earthengine.NewCachedClient(client, cache, 1*time.Hour)
func NewCachedClient(client *Client, cache Cache, ttl time.Duration) *CachedClient {
	return &CachedClient{
		Client: client,
		cache:  cache,
		ttl:    ttl,
	}
}

// Example: ComputeWithCache for ReduceRegionOperation
//
// To cache a reduce region operation:
//
//	cache := earthengine.NewMemoryCache(1000)
//	cacheKey := earthengine.CacheKey(lat, lon, date, "ndvi")
//
//	// Try cache
//	if cached, found, _ := cache.Get(ctx, cacheKey); found {
//	    result = cached.(map[string]interface{})
//	} else {
//	    result, err = operation.Compute(ctx)
//	    cache.Set(ctx, cacheKey, result, 1*time.Hour)
//	}

// ClearCache clears all cached data.
func (cc *CachedClient) ClearCache(ctx context.Context) error {
	return cc.cache.Clear(ctx)
}

// CacheStats represents cache statistics.
type CacheStats struct {
	Size    int
	MaxSize int
	Hits    int64
	Misses  int64
	HitRate float64
}

// Stats returns cache statistics (if supported by the cache implementation).
func (mc *MemoryCache) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return CacheStats{
		Size:    len(mc.data),
		MaxSize: mc.maxSize,
	}
}
