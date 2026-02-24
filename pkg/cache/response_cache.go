package cache

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ResponseCache caches HTTP responses (GET requests)
type ResponseCache struct {
	cache  Cache
	mu     sync.RWMutex
	keys   map[string]time.Time // Track key expiration times
	stats  CacheStats
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	CachedAt   time.Time
	ExpiresAt  time.Time
}

// NewResponseCache creates a response cache wrapper
func NewResponseCache(cache Cache) *ResponseCache {
	return &ResponseCache{
		cache: cache,
		keys:  make(map[string]time.Time),
	}
}

// GetCachedResponse retrieves a cached HTTP response
func (rc *ResponseCache) GetCachedResponse(key string) (*CachedResponse, bool) {
	value, exists := rc.cache.Get(key)
	if !exists {
		rc.mu.Lock()
		rc.stats.Misses++
		rc.mu.Unlock()
		return nil, false
	}

	response, ok := value.(*CachedResponse)
	if !ok {
		return nil, false
	}

	rc.mu.Lock()
	rc.stats.Hits++
	rc.mu.Unlock()

	return response, true
}

// SetCachedResponse stores an HTTP response in cache
func (rc *ResponseCache) SetCachedResponse(key string, response *CachedResponse, ttl time.Duration) {
	response.CachedAt = time.Now()
	response.ExpiresAt = time.Now().Add(ttl)

	rc.cache.Set(key, response, ttl)

	rc.mu.Lock()
	rc.keys[key] = response.ExpiresAt
	rc.stats.Total++
	rc.mu.Unlock()
}

// InvalidateKey removes a cached response (called on data updates)
func (rc *ResponseCache) InvalidateKey(key string) {
	rc.cache.Delete(key)

	rc.mu.Lock()
	delete(rc.keys, key)
	rc.stats.Evictions++
	rc.mu.Unlock()
}

// InvalidatePattern removes all cached responses matching a pattern
func (rc *ResponseCache) InvalidatePattern(pattern string) {
	rc.mu.Lock()
	keysToDelete := []string{}

	for key := range rc.keys {
		// Simple pattern matching (contains)
		if contains(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		rc.cache.Delete(key)
		delete(rc.keys, key)
		rc.stats.Evictions++
	}
	rc.mu.Unlock()
}

// ClearAll clears all cached responses
func (rc *ResponseCache) ClearAll() {
	rc.cache.Clear()

	rc.mu.Lock()
	rc.keys = make(map[string]time.Time)
	rc.mu.Unlock()
}

// GetStats returns cache statistics
func (rc *ResponseCache) GetStats() map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	var hitRate float64
	total := rc.stats.Hits + rc.stats.Misses
	if total > 0 {
		hitRate = float64(rc.stats.Hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":       rc.stats.Hits,
		"misses":     rc.stats.Misses,
		"evictions":  rc.stats.Evictions,
		"total":      rc.stats.Total,
		"hit_rate":   fmt.Sprintf("%.2f%%", hitRate),
		"size":       rc.cache.Size(),
	}
}

// Size returns the number of items in cache
func (rc *ResponseCache) Size() int {
	return rc.cache.Size()
}

// GenerateResponseCacheKey generates a cache key from method, path, and query
func GenerateResponseCacheKey(method, path, query string) string {
	h := md5.New()
	fmt.Fprintf(h, "%s:%s?%s", method, path, query)
	return fmt.Sprintf("resp:%x", h.Sum(nil))
}

// QueryResultCache caches database query results
type QueryResultCache struct {
	cache  Cache
	mu     sync.RWMutex
	keys   map[string]time.Time
	stats  CacheStats
	deps   map[string][]string // Dependencies: table -> cached queries
}

// NewQueryResultCache creates a query result cache
func NewQueryResultCache(cache Cache) *QueryResultCache {
	return &QueryResultCache{
		cache: cache,
		keys:  make(map[string]time.Time),
		deps:  make(map[string][]string),
	}
}

// GetQueryResult retrieves a cached query result
func (qrc *QueryResultCache) GetQueryResult(query string) (interface{}, bool) {
	key := generateQueryCacheKey(query)
	value, exists := qrc.cache.Get(key)

	if !exists {
		qrc.mu.Lock()
		qrc.stats.Misses++
		qrc.mu.Unlock()
		return nil, false
	}

	qrc.mu.Lock()
	qrc.stats.Hits++
	qrc.mu.Unlock()

	return value, true
}

// SetQueryResult caches a query result
func (qrc *QueryResultCache) SetQueryResult(query string, result interface{}, ttl time.Duration, dependsOnTables ...string) {
	key := generateQueryCacheKey(query)
	qrc.cache.Set(key, result, ttl)

	qrc.mu.Lock()
	qrc.keys[key] = time.Now().Add(ttl)
	qrc.stats.Total++

	// Track dependencies
	for _, table := range dependsOnTables {
		if _, exists := qrc.deps[table]; !exists {
			qrc.deps[table] = []string{}
		}
		qrc.deps[table] = append(qrc.deps[table], key)
	}
	qrc.mu.Unlock()
}

// InvalidateTable invalidates all cached queries depending on a table
func (qrc *QueryResultCache) InvalidateTable(tableName string) {
	qrc.mu.Lock()
	if keys, exists := qrc.deps[tableName]; exists {
		for _, key := range keys {
			qrc.cache.Delete(key)
			delete(qrc.keys, key)
			qrc.stats.Evictions++
		}
		qrc.deps[tableName] = []string{}
	}
	qrc.mu.Unlock()
}

// InvalidateAll clears all cached query results
func (qrc *QueryResultCache) InvalidateAll() {
	qrc.cache.Clear()

	qrc.mu.Lock()
	qrc.keys = make(map[string]time.Time)
	qrc.deps = make(map[string][]string)
	qrc.mu.Unlock()
}

// GetStats returns query cache statistics
func (qrc *QueryResultCache) GetStats() map[string]interface{} {
	qrc.mu.RLock()
	defer qrc.mu.RUnlock()

	var hitRate float64
	total := qrc.stats.Hits + qrc.stats.Misses
	if total > 0 {
		hitRate = float64(qrc.stats.Hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":       qrc.stats.Hits,
		"misses":     qrc.stats.Misses,
		"evictions":  qrc.stats.Evictions,
		"total":      qrc.stats.Total,
		"hit_rate":   fmt.Sprintf("%.2f%%", hitRate),
		"size":       qrc.cache.Size(),
		"tables":     len(qrc.deps),
	}
}

// Size returns the number of cached query results
func (qrc *QueryResultCache) Size() int {
	return qrc.cache.Size()
}

// generateQueryCacheKey creates a cache key from query string
func generateQueryCacheKey(query string) string {
	h := md5.New()
	fmt.Fprintf(h, "query:%s", query)
	return fmt.Sprintf("q:%x", h.Sum(nil))
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > 0)
}

// CacheInvalidator provides unified cache invalidation
type CacheInvalidator struct {
	responseCache Cache
	queryCache    Cache
}

// NewCacheInvalidator creates a cache invalidator
func NewCacheInvalidator(responseCache, queryCache Cache) *CacheInvalidator {
	return &CacheInvalidator{
		responseCache: responseCache,
		queryCache:    queryCache,
	}
}

// InvalidateAll clears all caches
func (ci *CacheInvalidator) InvalidateAll() {
	if ci.responseCache != nil {
		ci.responseCache.Clear()
	}
	if ci.queryCache != nil {
		ci.queryCache.Clear()
	}
}
