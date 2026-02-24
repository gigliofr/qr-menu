package cache

import (
	"sync"
	"time"
)

// Cache defines the interface for caching
type Cache interface {
	// Get retrieves a value from cache
	Get(key string) (interface{}, bool)

	// Set stores a value in cache with TTL
	Set(key string, value interface{}, ttl time.Duration)

	// Delete removes a value from cache
	Delete(key string)

	// Clear removes all values from cache
	Clear()

	// Size returns the number of items in cache
	Size() int

	// Exists checks if a key exists
	Exists(key string) bool
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessedAt time.Time
	HitCount   int
}

// InMemoryCache is a simple in-memory cache implementation
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
	stats CacheStats
}

// CacheStats tracks cache performance
type CacheStats struct {
	Hits      int
	Misses    int
	Evictions int
	Total     int
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*CacheEntry),
	}

	// Start cleanup goroutine for expired items
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.items[key]
	if !exists {
		c.mu.RUnlock()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.items, key)
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	// Update access stats
	entry.AccessedAt = time.Now()
	entry.HitCount++
	c.stats.Hits++
	c.mu.RUnlock()

	return entry.Value, true
}

// Set stores a value in cache with TTL
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		HitCount:   0,
	}
	c.stats.Total++
}

// Delete removes a value from cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	c.stats.Evictions++
}

// Clear removes all values from cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheEntry)
}

// Size returns the number of items in cache
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Exists checks if a key exists and is not expired
func (c *InMemoryCache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return false
	}

	if time.Now().After(entry.ExpiresAt) {
		return false
	}

	return true
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// GetEntry returns the full cache entry (for admin/debugging)
func (c *InMemoryCache) GetEntry(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	return entry, exists
}

// cleanupExpired periodically removes expired items
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		now := time.Now()
		expired := []string{}

		for key, entry := range c.items {
			if now.After(entry.ExpiresAt) {
				expired = append(expired, key)
			}
		}

		for _, key := range expired {
			delete(c.items, key)
			c.stats.Evictions++
		}

		c.mu.Unlock()
	}
}

// CacheWithTTL creates a cache with default TTL
type CacheWithTTL struct {
	cache      Cache
	defaultTTL time.Duration
}

// NewCacheWithTTL creates a cache with default TTL
func NewCacheWithTTL(cache Cache, defaultTTL time.Duration) *CacheWithTTL {
	return &CacheWithTTL{
		cache:      cache,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves from underlying cache
func (c *CacheWithTTL) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Set uses default TTL
func (c *CacheWithTTL) Set(key string, value interface{}) {
	c.cache.Set(key, value, c.defaultTTL)
}

// SetWithTTL uses specified TTL
func (c *CacheWithTTL) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

// Delete removes from cache
func (c *CacheWithTTL) Delete(key string) {
	c.cache.Delete(key)
}

// Clear clears the cache
func (c *CacheWithTTL) Clear() {
	c.cache.Clear()
}
