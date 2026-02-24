package middleware

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"qr-menu/logger"
	"qr-menu/pkg/cache"
)

// ResponseCachingMiddleware creates a middleware that caches HTTP responses
type ResponseCachingMiddleware struct {
	cache           *cache.ResponseCache
	cacheTTL        time.Duration
	cacheableStatus map[int]bool
	cacheableMethods map[string]bool
}

// NewResponseCachingMiddleware creates a new response caching middleware
func NewResponseCachingMiddleware(c *cache.ResponseCache, ttl time.Duration) *ResponseCachingMiddleware {
	return &ResponseCachingMiddleware{
		cache:    c,
		cacheTTL: ttl,
		cacheableStatus: map[int]bool{
			http.StatusOK:                  true,
			http.StatusCreated:             true,
			http.StatusNoContent:           true,
			http.StatusPartialContent:      true,
		},
		cacheableMethods: map[string]bool{
			"GET":  true,
			"HEAD": true,
		},
	}
}

// Middleware returns the middleware function
func (rcm *ResponseCachingMiddleware) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET/HEAD requests
			if !rcm.cacheableMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}

			// Check for cache hit
			cacheKey := cache.GenerateResponseCacheKey(r.Method, r.URL.Path, r.URL.RawQuery)
			if cachedResp, exists := rcm.cache.GetCachedResponse(cacheKey); exists {
				// Write cached response
				for key, values := range cachedResp.Headers {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("X-Cached-At", cachedResp.CachedAt.Format(time.RFC3339))

				w.WriteHeader(cachedResp.StatusCode)
				w.Write(cachedResp.Body)

				logger.Debug("Response cache hit", map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"key":    cacheKey,
				})
				return
			}

			// Wrap response writer to capture response
			wrapped := &responseCapture{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(wrapped, r)

			// Cache the response if cacheable
			if rcm.cacheableStatus[wrapped.statusCode] {
				cachedResp := &cache.CachedResponse{
					StatusCode: wrapped.statusCode,
					Headers:    wrapped.Header(),
					Body:       wrapped.body.Bytes(),
				}

				rcm.cache.SetCachedResponse(cacheKey, cachedResp, rcm.cacheTTL)

				w.Header().Set("X-Cache", "MISS")
				logger.Debug("Response cached", map[string]interface{}{
					"method":  r.Method,
					"path":    r.URL.Path,
					"status":  wrapped.statusCode,
					"size":    len(wrapped.body.Bytes()),
					"ttl_sec": int(rcm.cacheTTL.Seconds()),
				})
			}
		})
	}
}

// responseCapture captures HTTP response for caching
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    http.Header
}

// WriteHeader captures status code
func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
	rc.headers = rc.ResponseWriter.Header()
}

// Write captures response body
func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body.Write(data)
	return rc.ResponseWriter.Write(data)
}

// CacheInvalidationMiddleware helps with cache invalidation on mutations
type CacheInvalidationMiddleware struct {
	cache           *cache.ResponseCache
	queryCache      *cache.QueryResultCache
	pathPatterns    map[string][]string // path pattern -> tables to invalidate
}

// NewCacheInvalidationMiddleware creates invalidation middleware
func NewCacheInvalidationMiddleware(
	respCache *cache.ResponseCache,
	queryCache *cache.QueryResultCache,
) *CacheInvalidationMiddleware {
	return &CacheInvalidationMiddleware{
		cache:        respCache,
		queryCache:   queryCache,
		pathPatterns: make(map[string][]string),
	}
}

// RegisterPattern registers a path pattern and tables to invalidate
func (cim *CacheInvalidationMiddleware) RegisterPattern(pathPattern string, tables ...string) {
	cim.pathPatterns[pathPattern] = tables
}

// Middleware returns the middleware function
func (cim *CacheInvalidationMiddleware) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pass request through
			next.ServeHTTP(w, r)

			// Check if this is a mutation (POST, PUT, DELETE, PATCH)
			if isMutationMethod(r.Method) {
				// Find matching patterns and invalidate
				for pattern, tables := range cim.pathPatterns {
					if matchesPattern(r.URL.Path, pattern) {
						for _, table := range tables {
							if cim.queryCache != nil {
								cim.queryCache.InvalidateTable(table)
							}
							if cim.cache != nil {
								cim.cache.InvalidatePattern(table)
							}

							logger.Debug("Cache invalidated", map[string]interface{}{
								"method": r.Method,
								"path":   r.URL.Path,
								"table":  table,
							})
						}
					}
				}
			}
		})
	}
}

// isMutationMethod checks if method modifies data
func isMutationMethod(method string) bool {
	switch method {
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

// matchesPattern checks if path matches a pattern
func matchesPattern(path, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex
	return strings.Contains(path, pattern)
}
