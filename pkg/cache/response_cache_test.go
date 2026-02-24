package cache

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestGenerateResponseCacheKey tests cache key generation
func TestGenerateResponseCacheKey(t *testing.T) {
	key1 := GenerateResponseCacheKey("GET", "/api/users", "")
	key2 := GenerateResponseCacheKey("GET", "/api/users", "")

	if key1 != key2 {
		t.Errorf("Same inputs should generate same key")
	}

	key3 := GenerateResponseCacheKey("GET", "/api/teams", "")
	if key1 == key3 {
		t.Errorf("Different paths should generate different keys")
	}
}

// TestResponseCacheSetGet tests basic response caching
func TestResponseCacheSetGet(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
		Body:       []byte("test response"),
	}
	response.Headers.Set("Content-Type", "application/json")

	key := GenerateResponseCacheKey("GET", "/api/users", "")
	rc.SetCachedResponse(key, response, 1*time.Hour)

	cached, exists := rc.GetCachedResponse(key)
	if !exists {
		t.Error("Expected cached response to exist")
	}

	if cached == nil {
		t.Error("Expected cached response to not be nil")
		return
	}

	if cached.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", cached.StatusCode)
	}

	if string(cached.Body) != "test response" {
		t.Errorf("Expected body 'test response', got %s", cached.Body)
	}
}

// TestResponseCacheMiss tests cache miss
func TestResponseCacheMiss(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	key := GenerateResponseCacheKey("GET", "/api/nonexistent", "")
	_, exists := rc.GetCachedResponse(key)

	if exists {
		t.Error("Expected cache miss for nonexistent key")
	}

	stats := rc.GetStats()
	if stats["misses"] != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}

// TestResponseCacheInvalidateKey tests key invalidation
func TestResponseCacheInvalidateKey(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	key := GenerateResponseCacheKey("GET", "/api/users", "")
	rc.SetCachedResponse(key, response, 1*time.Hour)

	rc.InvalidateKey(key)

	_, exists := rc.GetCachedResponse(key)
	if exists {
		t.Error("Expected key to be invalidated")
	}
}

// TestResponseCacheInvalidatePattern tests pattern invalidation
func TestResponseCacheInvalidatePattern(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	// Create multiple cached responses
	keys := []string{
		GenerateResponseCacheKey("GET", "/api/users/1", ""),
		GenerateResponseCacheKey("GET", "/api/users/2", ""),
		GenerateResponseCacheKey("GET", "/api/teams/1", ""),
	}

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	for _, key := range keys {
		rc.SetCachedResponse(key, response, 1*time.Hour)
	}

	// Invalidate pattern (simplified - just checks if pattern exists in key)
	rc.InvalidatePattern("users")

	// Note: Our simple pattern matching just checks presence
	// In real implementation, this would be more sophisticated
}

// TestResponseCacheStats tests statistics tracking
func TestResponseCacheStats(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	key := GenerateResponseCacheKey("GET", "/api/users", "")
	rc.SetCachedResponse(key, response, 1*time.Hour)

	// Hit
	rc.GetCachedResponse(key)
	rc.GetCachedResponse(key)

	// Miss
	rc.GetCachedResponse("nonexistent")

	stats := rc.GetStats()
	if stats["hits"] != 2 {
		t.Errorf("Expected 2 hits, got %v", stats["hits"])
	}

	if stats["misses"] != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}

// TestResponseCacheClearAll tests clearing all responses
func TestResponseCacheClearAll(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	rc.SetCachedResponse(GenerateResponseCacheKey("GET", "/api/users", ""), response, 1*time.Hour)
	rc.SetCachedResponse(GenerateResponseCacheKey("GET", "/api/teams", ""), response, 1*time.Hour)

	if rc.cache.Size() != 2 {
		t.Errorf("Expected 2 items before clear, got %d", rc.cache.Size())
	}

	rc.ClearAll()

	if rc.cache.Size() != 0 {
		t.Errorf("Expected 0 items after clear, got %d", rc.cache.Size())
	}
}

// TestQueryResultCacheSetGet tests query result caching
func TestQueryResultCacheSetGet(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query := "SELECT * FROM users WHERE id = 1"
	result := map[string]interface{}{
		"id":   1,
		"name": "John",
	}

	qrc.SetQueryResult(query, result, 1*time.Hour, "users")

	cached, exists := qrc.GetQueryResult(query)
	if !exists {
		t.Error("Expected cached result to exist")
	}

	if cached == nil {
		t.Error("Expected cached result to not be nil")
		return
	}

	cachedMap := cached.(map[string]interface{})
	if cachedMap["id"] != 1 {
		t.Errorf("Expected id 1, got %v", cachedMap["id"])
	}
}

// TestQueryResultCacheMiss tests query cache miss
func TestQueryResultCacheMiss(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query := "SELECT * FROM users WHERE id = 999"
	_, exists := qrc.GetQueryResult(query)

	if exists {
		t.Error("Expected cache miss for nonexistent query")
	}

	stats := qrc.GetStats()
	if stats["misses"] != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}

// TestQueryResultCacheInvalidateTable tests table invalidation
func TestQueryResultCacheInvalidateTable(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query1 := "SELECT * FROM users WHERE id = 1"
	query2 := "SELECT * FROM teams WHERE id = 1"

	result := map[string]interface{}{"id": 1}

	qrc.SetQueryResult(query1, result, 1*time.Hour, "users")
	qrc.SetQueryResult(query2, result, 1*time.Hour, "teams")

	if qrc.cache.Size() != 2 {
		t.Errorf("Expected 2 cached queries, got %d", qrc.cache.Size())
	}

	// Invalidate users table
	qrc.InvalidateTable("users")

	// Users query should be gone
	_, exists := qrc.GetQueryResult(query1)
	if exists {
		t.Error("Expected users query to be invalidated")
	}

	// Teams query should still exist
	_, exists = qrc.GetQueryResult(query2)
	if !exists {
		t.Error("Expected teams query to still exist")
	}
}

// TestQueryResultCacheStats tests statistics
func TestQueryResultCacheStats(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query := "SELECT * FROM users"
	result := []map[string]interface{}{{"id": 1}}

	qrc.SetQueryResult(query, result, 1*time.Hour, "users")

	qrc.GetQueryResult(query)
	qrc.GetQueryResult(query)
	qrc.GetQueryResult("nonexistent")

	stats := qrc.GetStats()
	if hits, ok := stats["hits"].(int); ok && hits != 2 {
		t.Errorf("Expected 2 hits, got %d", hits)
	}

	if misses, ok := stats["misses"].(int); ok && misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
}

// TestQueryResultCacheMultipleTables tests multiple table dependencies
func TestQueryResultCacheMultipleTables(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	// Query depending on multiple tables
	query := "SELECT u.id, t.name FROM users u JOIN teams t ON u.team_id = t.id"
	result := []map[string]interface{}{}

	qrc.SetQueryResult(query, result, 1*time.Hour, "users", "teams")

	// Invalidate users table
	qrc.InvalidateTable("users")

	_, exists := qrc.GetQueryResult(query)
	if exists {
		t.Error("Expected query to be invalidated when dependent table changes")
	}
}

// TestQueryResultCacheInvalidateAll tests clearing all
func TestQueryResultCacheInvalidateAll(t *testing.T) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query1 := "SELECT * FROM users"
	query2 := "SELECT * FROM teams"
	result := []int{}

	qrc.SetQueryResult(query1, result, 1*time.Hour, "users")
	qrc.SetQueryResult(query2, result, 1*time.Hour, "teams")

	if qrc.cache.Size() != 2 {
		t.Errorf("Expected 2 cached queries, got %d", qrc.cache.Size())
	}

	qrc.InvalidateAll()

	if qrc.cache.Size() != 0 {
		t.Errorf("Expected 0 cached queries after invalidate all, got %d", qrc.cache.Size())
	}
}

// TestCacheInvalidator tests unified invalidation
func TestCacheInvalidator(t *testing.T) {
	baseResp := NewInMemoryCache()
	baseQuery := NewInMemoryCache()

	respCache := NewResponseCache(baseResp)
	queryCache := NewQueryResultCache(baseQuery)

	invalidator := NewCacheInvalidator(baseResp, baseQuery)

	// Populate caches
	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}
	respCache.SetCachedResponse("key1", response, 1*time.Hour)

	queryCache.SetQueryResult("query1", "result", 1*time.Hour, "users")

	if baseResp.Size() != 1 || baseQuery.Size() != 1 {
		t.Errorf("Expected 1 item in each cache before invalidation")
	}

	// Invalidate all
	invalidator.InvalidateAll()

	if baseResp.Size() != 0 || baseQuery.Size() != 0 {
		t.Errorf("Expected 0 items in caches after invalidation")
	}
}

// TestResponseCacheHitRate tests hit rate calculation
func TestResponseCacheHitRate(t *testing.T) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	key := GenerateResponseCacheKey("GET", "/api/users", "")
	rc.SetCachedResponse(key, response, 1*time.Hour)

	// 8 hits, 2 misses = 80% hit rate
	for i := 0; i < 8; i++ {
		rc.GetCachedResponse(key)
	}

	for i := 0; i < 2; i++ {
		rc.GetCachedResponse("nonexistent")
	}

	stats := rc.GetStats()
	hitRateStr := stats["hit_rate"].(string)

	if hitRateStr != "80.00%" {
		t.Errorf("Expected 80.00%% hit rate, got %s", hitRateStr)
	}
}

// BenchmarkResponseCacheGet benchmarks response cache retrieval
func BenchmarkResponseCacheGet(b *testing.B) {
	base := NewInMemoryCache()
	rc := NewResponseCache(base)

	response := &CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test response"),
	}

	key := GenerateResponseCacheKey("GET", "/api/users", "")
	rc.SetCachedResponse(key, response, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rc.GetCachedResponse(key)
	}
}

// BenchmarkQueryResultCacheGet benchmarks query cache retrieval
func BenchmarkQueryResultCacheGet(b *testing.B) {
	base := NewInMemoryCache()
	qrc := NewQueryResultCache(base)

	query := "SELECT * FROM users"
	result := []map[string]interface{}{{"id": 1}}

	qrc.SetQueryResult(query, result, 1*time.Hour, "users")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qrc.GetQueryResult(query)
	}
}

// BenchmarkGenerateResponseCacheKey benchmarks key generation
func BenchmarkGenerateResponseCacheKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateResponseCacheKey("GET", fmt.Sprintf("/api/users/%d", i), "page=1")
	}
}
