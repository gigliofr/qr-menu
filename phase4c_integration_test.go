package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qr-menu/pkg/cache"
	"qr-menu/pkg/middleware"
)

// Integration Tests for Complete Caching + Middleware Stack

// TestFullCacheMiddlewareIntegration tests response caching with middleware chain
func TestFullCacheMiddlewareIntegration(t *testing.T) {
	// Setup cache
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	// Setup middleware
	cacheMW := middleware.NewResponseCachingMiddleware(respCache, 1*time.Hour)

	// Track handler calls
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Test"}`))
	})

	// Wrap with middleware chain
	wrappedHandler := middleware.Chain(
		handler,
		middleware.SecurityHeaders(),
		cacheMW.Middleware(),
		middleware.Logging(),
	)

	// First request - cache miss
	req1, _ := http.NewRequest("GET", "/api/test", nil)
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req1)

	if callCount != 1 {
		t.Errorf("Expected 1 handler call, got %d", callCount)
	}

	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("Expected X-Cache: MISS, got %v", w1.Header().Get("X-Cache"))
	}

	// Second request - cache hit
	req2, _ := http.NewRequest("GET", "/api/test", nil)
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req2)

	if callCount != 1 {
		t.Errorf("Expected still 1 handler call (cache hit), got %d", callCount)
	}

	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Expected X-Cache: HIT, got %v", w2.Header().Get("X-Cache"))
	}

	// Verify security headers still applied
	if w2.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected security headers to be applied")
	}
}

// TestCacheInvalidationWorkflow tests complete cache invalidation workflow
func TestCacheInvalidationWorkflow(t *testing.T) {
	// Setup caches
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	queryCache := cache.NewQueryResultCache(baseCache)

	// Setup middlewares
	cacheMW := middleware.NewResponseCachingMiddleware(respCache, 1*time.Hour)
	invalidationMW := middleware.NewCacheInvalidationMiddleware(respCache, queryCache)
	invalidationMW.RegisterPattern("/api/users", "users")

	// Track calls
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("users"))
	})

	// Wrap handler
	wrappedHandler := middleware.Chain(
		handler,
		invalidationMW.Middleware(),
		cacheMW.Middleware(),
	)

	// GET - should cache
	getReq, _ := http.NewRequest("GET", "/api/users", nil)
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, getReq)

	if callCount != 1 {
		t.Errorf("GET: expected 1 call, got %d", callCount)
	}

	// Verify cachedResponse
	if respCache.Size() != 1 {
		t.Errorf("Expected 1 cached response, got %d", respCache.Size())
	}

	// Second GET - should hit cache
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, getReq)

	if callCount != 1 {
		t.Errorf("GET 2: expected still 1 call (cache hit), got %d", callCount)
	}

	// POST - should invalidate cache
	postReq, _ := http.NewRequest("POST", "/api/users", nil)
	w3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w3, postReq)

	if callCount != 2 {
		t.Errorf("POST: expected 2 calls, got %d", callCount)
	}

	// Cache should be cleared
	if respCache.Size() != 0 {
		t.Errorf("Expected cache cleared after POST, got %d items", respCache.Size())
	}

	// Next GET - should miss cache (cache cleared)
	w4 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w4, getReq)

	if callCount != 3 {
		t.Errorf("GET 3: expected 3 calls (cache miss), got %d", callCount)
	}
}

// TestQueryCachingWithDependencies tests query caching with table dependencies
func TestQueryCachingWithDependencies(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	queryCache := cache.NewQueryResultCache(baseCache)

	// Simulate query execution
	query1 := "SELECT * FROM users WHERE id = 1"
	query2 := "SELECT * FROM users JOIN teams ON users.team_id = teams.id"
	query3 := "SELECT * FROM teams WHERE id = 1"

	result1 := map[string]interface{}{"id": 1, "name": "John"}
	result2 := []interface{}{result1, map[string]interface{}{"team": "Engineering"}}
	result3 := map[string]interface{}{"id": 1, "name": "Engineering"}

	// Cache queries with dependencies
	queryCache.SetQueryResult(query1, result1, 1*time.Hour, "users")
	queryCache.SetQueryResult(query2, result2, 1*time.Hour, "users", "teams")
	queryCache.SetQueryResult(query3, result3, 1*time.Hour, "teams")

	if queryCache.Size() != 3 {
		t.Errorf("Expected 3 cached queries, got %d", queryCache.Size())
	}

	// Invalidate users table
	queryCache.InvalidateTable("users")

	// Check dependencies:
	// query1 (depends on users) - should be gone
	_, exists := queryCache.GetQueryResult(query1)
	if exists {
		t.Error("Query1 should be invalidated (depends on users)")
	}

	// query2 (depends on users, teams) - should be gone
	_, exists = queryCache.GetQueryResult(query2)
	if exists {
		t.Error("Query2 should be invalidated (depends on users)")
	}

	// query3 (depends only on teams) - should still exist
	_, exists = queryCache.GetQueryResult(query3)
	if !exists {
		t.Error("Query3 should still exist (doesn't depend on users)")
	}

	// Invalidate teams table
	queryCache.InvalidateTable("teams")

	// Now query3 should be gone
	_, exists = queryCache.GetQueryResult(query3)
	if exists {
		t.Error("Query3 should be invalidated (depends on teams)")
	}
}

// TestCachePerformanceWithDifferentSizes tests cache performance with different payload sizes
func TestCachePerformanceWithDifferentSizes(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	cacheMW := middleware.NewResponseCachingMiddleware(respCache, 1*time.Hour)

	// Test with different payload sizes
	payloads := []struct {
		name string
		size int
	}{
		{"small", 100},
		{"medium", 10000},
		{"large", 100000},
	}

	for _, p := range payloads {
		t.Run(p.name, func(t *testing.T) {
			// Create handler with specific payload
			body := make([]byte, p.size)
			for i := range body {
				body[i] = byte('a' + i%26)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(body)
			})

			wrappedHandler := cacheMW.Middleware()(handler)

			// First request - cache miss
			req, _ := http.NewRequest("GET", "/api/data/"+p.name, nil)
			w1 := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w1, req)

			if w1.Header().Get("X-Cache") != "MISS" {
				t.Error("Expected MISS on first request")
			}

			// Second request - cache hit
			w2 := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w2, req)

			if w2.Header().Get("X-Cache") != "HIT" {
				t.Error("Expected HIT on second request")
			}

			// Verify payload matches
			if len(w2.Body.Bytes()) != p.size {
				t.Errorf("Expected %d bytes, got %d", p.size, len(w2.Body.Bytes()))
			}
		})
	}
}

// TestCacheStatisticsAccuracy tests that cache statistics are accurate
func TestCacheStatisticsAccuracy(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	// Create responses
	response := &cache.CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("test"),
	}

	// Cache multiple responses
	keys := []string{
		cache.GenerateResponseCacheKey("GET", "/api/users", ""),
		cache.GenerateResponseCacheKey("GET", "/api/teams", ""),
		cache.GenerateResponseCacheKey("GET", "/api/projects", ""),
	}

	for _, key := range keys {
		respCache.SetCachedResponse(key, response, 1*time.Hour)
	}

	// Perform hits and misses
	respCache.GetCachedResponse(keys[0]) // hit
	respCache.GetCachedResponse(keys[0]) // hit
	respCache.GetCachedResponse(keys[1]) // hit
	respCache.GetCachedResponse("missing") // miss

	stats := respCache.GetStats()

	if stats["hits"] != 3 {
		t.Errorf("Expected 3 hits, got %v", stats["hits"])
	}

	if stats["misses"] != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}

	if stats["total"] != 3 {
		t.Errorf("Expected 3 total, got %v", stats["total"])
	}

	// Hit rate should be 75%
	hitRateStr := stats["hit_rate"].(string)
	if hitRateStr != "75.00%" {
		t.Errorf("Expected 75.00%% hit rate, got %s", hitRateStr)
	}
}

// TestCacheExpirationWithMiddleware tests cache TTL expiration with middleware
func TestCacheExpirationWithMiddleware(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	cacheMW := middleware.NewResponseCachingMiddleware(respCache, 100*time.Millisecond)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	wrappedHandler := cacheMW.Middleware()(handler)

	// First request
	req, _ := http.NewRequest("GET", "/api/test", nil)
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req)

	// Immediate second request - should hit cache
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)

	if callCount != 1 {
		t.Errorf("Expected 1 call (cache hit), got %d", callCount)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Request after expiration - should miss cache
	w3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w3, req)

	if callCount != 2 {
		t.Errorf("Expected 2 calls (cache miss after expiration), got %d", callCount)
	}
}

// TestMultipleStatusCodesCache tests that different status codes are handled correctly
func TestMultipleStatusCodesCache(t *testing.T) {
	testCases := []struct {
		status     int
		shouldCache bool
	}{
		{http.StatusOK, true},
		{http.StatusCreated, true},
		{http.StatusNoContent, true},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusNotFound, false},
		{http.StatusInternalServerError, false},
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			// Create fresh cache for each test case
			baseCache := cache.NewInMemoryCache()
			respCache := cache.NewResponseCache(baseCache)
			cacheMW := middleware.NewResponseCachingMiddleware(respCache, 1*time.Hour)

			callCount := 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(tc.status)
			})

			wrappedHandler := cacheMW.Middleware()(handler)

			// First request
			req, _ := http.NewRequest("GET", "/api/test", nil)
			w1 := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w1, req)

			// Second request
			w2 := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w2, req)

			if tc.shouldCache && callCount != 1 {
				t.Errorf("Status %d should be cached, but handler called %d times", tc.status, callCount)
			}

			if !tc.shouldCache && callCount != 2 {
				t.Errorf("Status %d should NOT be cached, but handler called %d times", tc.status, callCount)
			}
		})
	}
}

// TestConcurrentCacheAccess tests concurrent access to caches
func TestConcurrentCacheAccess(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	queryCache := cache.NewQueryResultCache(baseCache)

	done := make(chan bool, 20)

	// Concurrent response cache writes
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := cache.GenerateResponseCacheKey("GET", "/api/test", "")
			resp := &cache.CachedResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("response"),
			}
			respCache.SetCachedResponse(key, resp, 1*time.Hour)
			done <- true
		}(i)
	}

	// Concurrent query cache writes
	for i := 0; i < 10; i++ {
		go func(index int) {
			query := "SELECT * FROM test"
			queryCache.SetQueryResult(query, map[string]interface{}{"id": index}, 1*time.Hour, "test")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Caches should have entries
	if respCache.Size() == 0 {
		t.Error("Response cache should have items")
	}

	if queryCache.Size() == 0 {
		t.Error("Query cache should have items")
	}
}

// BenchmarkCacheHitVsMiss benchmarks difference between cache hit and miss
func BenchmarkCacheHitVsMiss(b *testing.B) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	cacheMW := middleware.NewResponseCachingMiddleware(respCache, 1*time.Hour)

	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate slow handler
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	wrappedHandler := cacheMW.Middleware()(slowHandler)

	// Pre-populate cache
	preReq, _ := http.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, preReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, preReq)
	}
}
