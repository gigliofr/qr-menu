package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qr-menu/pkg/cache"
)

// TestResponseCachingMiddlewareWithGET tests caching of GET requests
func TestResponseCachingMiddlewareWithGET(t *testing.T) {
	// Create response cache
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	// Create middleware
	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	// Create a test handler that tracks calls
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	wrapped := middleware.Middleware()(handler)

	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request - should hit handler
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("Expected X-Cache: MISS, got %s", w1.Header().Get("X-Cache"))
	}

	// Second request - should hit cache
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req)

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once, was called %d times", callCount)
	}

	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Expected X-Cache: HIT, got %s", w2.Header().Get("X-Cache"))
	}

	if w2.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got %s", w2.Body.String())
	}
}

// TestResponseCachingMiddlewareWithPOST tests that POST is not cached
func TestResponseCachingMiddlewareWithPOST(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	wrapped := middleware.Middleware()(handler)

	req, err := http.NewRequest("POST", "/api/users", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req)

	// Second request - should not be cached since it's POST
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice (not cached), was called %d times", callCount)
	}
}

// TestResponseCachingMiddlewareNonCacheableStatus tests that non-cacheable status codes are not cached
func TestResponseCachingMiddlewareNonCacheableStatus(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})

	wrapped := middleware.Middleware()(handler)

	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request returns error
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req)

	// Second request should also return error (not cached)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice, was called %d times", callCount)
	}
}

// TestResponseCachingMiddlewareWithQueryString tests caching with query parameters
func TestResponseCachingMiddlewareWithQueryString(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("page: " + r.URL.Query().Get("page")))
	})

	wrapped := middleware.Middleware()(handler)

	// First URL with page=1
	req1, _ := http.NewRequest("GET", "/api/users?page=1", nil)
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req1)

	// Same URL - should hit cache
	req2, _ := http.NewRequest("GET", "/api/users?page=1", nil)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once for same query, was called %d times", callCount)
	}

	// Different query - should miss cache
	req3, _ := http.NewRequest("GET", "/api/users?page=2", nil)
	w3 := httptest.NewRecorder()
	wrapped.ServeHTTP(w3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called twice for different query, was called %d times", callCount)
	}
}

// TestCacheInvalidationMiddlewarePOST tests cache invalidation on POST
func TestCacheInvalidationMiddlewarePOST(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	queryCache := cache.NewQueryResultCache(baseCache)

	// Set up invalidation middleware
	invalidation := NewCacheInvalidationMiddleware(respCache, queryCache)
	invalidation.RegisterPattern("/api/users", "users")

	// First, populate cache
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a GET request and cache it
	cacheKey := cache.GenerateResponseCacheKey("GET", "/api/users", "")
	respCache.SetCachedResponse(cacheKey, &cache.CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("users"),
	}, 1*time.Hour)

	// Verify it's cached
	if respCache.Size() != 1 {
		t.Errorf("Expected 1 item in cache, got %d", respCache.Size())
	}

	wrapped := invalidation.Middleware()(handler)

	// POST to /api/users should invalidate cache
	postReq, _ := http.NewRequest("POST", "/api/users", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, postReq)

	// Cache should be cleared
	if respCache.Size() != 0 {
		t.Errorf("Expected cache to be cleared, got %d items", respCache.Size())
	}
}

// TestCacheInvalidationMiddlewareGET tests that GET requests don't invalidate
func TestCacheInvalidationMiddlewareGET(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)
	queryCache := cache.NewQueryResultCache(baseCache)

	invalidation := NewCacheInvalidationMiddleware(respCache, queryCache)
	invalidation.RegisterPattern("/api/users", "users")

	// Populate cache
	cacheKey := cache.GenerateResponseCacheKey("GET", "/api/users", "")
	respCache.SetCachedResponse(cacheKey, &cache.CachedResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("users"),
	}, 1*time.Hour)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := invalidation.Middleware()(handler)

	// GET should not invalidate
	getReq, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, getReq)

	// Cache should still exist
	if respCache.Size() != 1 {
		t.Errorf("Expected cache to remain, got %d items", respCache.Size())
	}
}

// TestIsMutationMethod tests mutation detection
func TestIsMutationMethod(t *testing.T) {
	tests := []struct {
		method     string
		isMutation bool
	}{
		{"GET", false},
		{"HEAD", false},
		{"OPTIONS", false},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", true},
	}

	for _, test := range tests {
		result := isMutationMethod(test.method)
		if result != test.isMutation {
			t.Errorf("isMutationMethod(%s) = %v, expected %v", test.method, result, test.isMutation)
		}
	}
}

// TestMatchesPattern tests pattern matching
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		matches bool
	}{
		{"/api/users", "users", true},
		{"/api/users/123", "users", true},
		{"/api/teams", "teams", true},
		{"/api/users", "teams", false},
		{"/admin/users", "users", true},
	}

	for _, test := range tests {
		result := matchesPattern(test.path, test.pattern)
		if result != test.matches {
			t.Errorf("matchesPattern(%s, %s) = %v, expected %v", test.path, test.pattern, result, test.matches)
		}
	}
}

// TestResponseCachingMiddlewareWithHeaders tests that cached headers are preserved
func TestResponseCachingMiddlewareWithHeaders(t *testing.T) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})

	wrapped := middleware.Middleware()(handler)

	req, _ := http.NewRequest("GET", "/api/data", nil)

	// First request
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req)

	contentType1 := w1.Header().Get("Content-Type")
	customHeader1 := w1.Header().Get("X-Custom-Header")

	// Second request from cache
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req)

	contentType2 := w2.Header().Get("Content-Type")
	customHeader2 := w2.Header().Get("X-Custom-Header")

	if contentType1 != contentType2 {
		t.Errorf("Content-Type mismatch: %s vs %s", contentType1, contentType2)
	}

	if customHeader1 != customHeader2 {
		t.Errorf("Custom header mismatch: %s vs %s", customHeader1, customHeader2)
	}

	if customHeader2 != "custom-value" {
		t.Errorf("Expected custom header 'custom-value', got %s", customHeader2)
	}
}

// BenchmarkResponseCachingMiddlewareHit benchmarks cache hit
func BenchmarkResponseCachingMiddlewareHit(b *testing.B) {
	baseCache := cache.NewInMemoryCache()
	respCache := cache.NewResponseCache(baseCache)

	middleware := NewResponseCachingMiddleware(respCache, 1*time.Hour)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	wrapped := middleware.Middleware()(handler)

	req, _ := http.NewRequest("GET", "/api/test", nil)

	// Pre-populate cache
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}
