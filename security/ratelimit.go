package security

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	cleanup  time.Duration
	stopChan chan struct{}
}

type bucket struct {
	tokens    float64
	capacity  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu        sync.Mutex
}

// RateLimitConfig configures rate limiting per endpoint
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

var defaultConfig = RateLimitConfig{
	RequestsPerSecond: 10,
	BurstSize:         20,
}

var endpointConfigs = map[string]RateLimitConfig{
	"/api/auth/login": {
		RequestsPerSecond: 3,
		BurstSize:         5,
	},
	"/api/auth/register": {
		RequestsPerSecond: 2,
		BurstSize:         3,
	},
	"/api/webhooks": {
		RequestsPerSecond: 100,
		BurstSize:         200,
	},
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		cleanup:  time.Minute * 5,
		stopChan: make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.removeOldBuckets()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *RateLimiter) removeOldBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, b := range rl.buckets {
		b.mu.Lock()
		if now.Sub(b.lastRefill) > rl.cleanup {
			delete(rl.buckets, key)
		}
		b.mu.Unlock()
	}
}

func (rl *RateLimiter) getBucket(key string, config RateLimitConfig) *bucket {
	rl.mu.RLock()
	b, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return b
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if b, exists := rl.buckets[key]; exists {
		return b
	}

	b = &bucket{
		tokens:     float64(config.BurstSize),
		capacity:   float64(config.BurstSize),
		refillRate: config.RequestsPerSecond,
		lastRefill: time.Now(),
	}
	rl.buckets[key] = b
	return b
}

func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	
	// Refill tokens
	b.tokens = min(b.capacity, b.tokens+elapsed*b.refillRate)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware applies rate limiting per user and per endpoint
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user identifier (IP or user ID from JWT)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = r.RemoteAddr
		}

		// Get endpoint pattern
		route := mux.CurrentRoute(r)
		var endpoint string
		if route != nil {
			pathTemplate, _ := route.GetPathTemplate()
			endpoint = pathTemplate
		} else {
			endpoint = r.URL.Path
		}

		// Get config for this endpoint
		config, exists := endpointConfigs[endpoint]
		if !exists {
			config = defaultConfig
		}

		// Create unique key for user+endpoint
		key := userID + ":" + endpoint
		bucket := rl.getBucket(key, config)

		if !bucket.allow() {
			w.Header().Set("X-RateLimit-Limit", formatFloat(config.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		bucket.mu.Lock()
		remaining := int(bucket.tokens)
		bucket.mu.Unlock()

		w.Header().Set("X-RateLimit-Limit", formatFloat(config.RequestsPerSecond))
		w.Header().Set("X-RateLimit-Remaining", formatInt(remaining))

		next.ServeHTTP(w, r)
	})
}

func formatFloat(f float64) string {
	return formatInt(int(f))
}

func formatInt(i int) string {
	return string(rune(i + '0'))
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}
