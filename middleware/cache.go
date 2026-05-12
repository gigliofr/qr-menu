package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"qr-menu/pkg/cache"
)

// CacheMiddleware cachea le response HTTP GET
func CacheMiddleware(ttl time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Solo cachea GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Genera cache key basato su URL e query params
			cacheKey := generateCacheKey(r.URL.Path, r.URL.RawQuery)

			// Prova a recuperare dal cache
			if cachedResponse, found := cache.DefaultInMemoryCache.Get(cacheKey); found {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, cachedResponse)
				return
			}

			// Se non in cache, wrappa il writer per captire la response
			wrappedWriter := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer([]byte{}),
			}

			next.ServeHTTP(wrappedWriter, r)

			// Salva nel cache se status code è 2xx
			if wrappedWriter.statusCode >= 200 && wrappedWriter.statusCode < 300 {
				w.Header().Set("X-Cache", "MISS")
				cache.DefaultInMemoryCache.Set(cacheKey, wrappedWriter.body.String(), ttl)
			}
		})
	}
}

// responseWriterWrapper cattura la response body
type responseWriterWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateCacheKey crea una chiave di cache basata su path e query
func generateCacheKey(path, query string) string {
	key := fmt.Sprintf("http:%s:%s", path, query)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("cache:%x", hash)
}

// InvalidateCacheByPattern invalida tutte le cache entries che matchano il pattern
func InvalidateCacheByPattern(pattern string) {
	// Simple pattern matching: se contiene la stringa
	cache.DefaultInMemoryCache.Clear() // Per ora clear tutto (TODO: implementare pattern matching)
}
