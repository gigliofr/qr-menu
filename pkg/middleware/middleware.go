package middleware

import (
	"fmt"
	"net/http"
	"time"

	"qr-menu/logger"
	"qr-menu/pkg/errors"
	httputil "qr-menu/pkg/http"
)

// Middleware is a function that wraps an HTTP handler
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middleware in sequence
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Reverse order so first middleware in list wraps last
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ChainHandlerFunc chains middleware for a handler function
func ChainHandlerFunc(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Chain(h, middlewares...)
}

// Logging middleware logs HTTP requests and responses
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			method := r.Method
			path := r.RequestURI
			ip := getClientIP(r)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log the request
			duration := time.Since(start).Milliseconds()
			logger.InfoWithContext(
				"HTTP Request",
				map[string]interface{}{
					"method":      method,
					"path":        path,
					"status":      wrapped.statusCode,
					"duration_ms": duration,
					"user_agent":  r.UserAgent(),
				},
				"", // user_id - would be set by auth middleware
				ip,
				r.UserAgent(),
			)
		})
	}
}

// ErrorRecovery middleware recovers from panics
func ErrorRecovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered", map[string]interface{}{
						"error":      err,
						"method":     r.Method,
						"path":       r.RequestURI,
						"user_agent": r.UserAgent(),
					})

					appErr := errors.InternalError("Internal server error").
						WithDetails("An unexpected error occurred")
					httputil.Error(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Authentication middleware validates JWT tokens
func Authentication(validateToken func(token string) (map[string]interface{}, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.Unauthorized(w, "Missing authorization header")
				return
			}

			// Extract bearer token
			const bearerScheme = "Bearer "
			if len(authHeader) < len(bearerScheme) || authHeader[:len(bearerScheme)] != bearerScheme {
				httputil.Unauthorized(w, "Invalid authorization header format")
				return
			}

			token := authHeader[len(bearerScheme):]

			// Validate token
			claims, err := validateToken(token)
			if err != nil {
				httputil.Unauthorized(w, "Invalid or expired token")
				return
			}

			// Store claims in request context (simplified - use context package in production)
			r.Header.Set("X-User-ID", claims["sub"].(string))
			if email, ok := claims["email"]; ok {
				r.Header.Set("X-User-Email", email.(string))
			}

			logger.Info("Authenticated request", map[string]interface{}{
				"user_id": claims["sub"],
				"path":    r.RequestURI,
			})

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware adds CORS headers
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := isOriginAllowed(origin, allowedOrigins)

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiting middleware limits requests per IP
func RateLimiting(requestsPerSecond int) Middleware {
	// Simple in-memory rate limiter (use Redis in production)
	limiters := make(map[string]*rateLimiter)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			// Get or create limiter for this IP
			if _, exists := limiters[ip]; !exists {
				limiters[ip] = newRateLimiter(requestsPerSecond)
			}

			limiter := limiters[ip]

			// Check if request is allowed
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				httputil.Error(w, errors.RateLimited("Too many requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestMetrics middleware collects request metrics
func RequestMetrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Record metrics (in production, send to monitoring system)
			duration := time.Since(start).Milliseconds()
			wrapped.Header().Set("X-Response-Time", fmt.Sprintf("%dms", duration))

			logger.Debug("Metrics", map[string]interface{}{
				"method":      r.Method,
				"path":        r.RequestURI,
				"status":      wrapped.statusCode,
				"duration_ms": duration,
			})
		})
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			next.ServeHTTP(w, r)
		})
	}
}

// Helper types

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// rateLimiter implements simple token bucket algorithm
type rateLimiter struct {
	tokens    int
	maxTokens int
	lastRefil time.Time
	refillRate int // tokens per second
}

func newRateLimiter(requestsPerSecond int) *rateLimiter {
	return &rateLimiter{
		tokens:     requestsPerSecond,
		maxTokens:  requestsPerSecond,
		refillRate: requestsPerSecond,
		lastRefil:  time.Now(),
	}
}

func (rl *rateLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefil).Seconds()

	// Refill tokens
	tokensToAdd := int(elapsed * float64(rl.refillRate))
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefil = now
	}

	// Check if we have tokens
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP (nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed || allowed == "*" {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
