package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper to create a simple test handler
func testHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// TestLoggingMiddleware tests logging functionality
func TestLoggingMiddleware(t *testing.T) {
	handler := Logging()(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %s", w.Body.String())
	}
}

// TestErrorRecoveryMiddleware tests panic recovery
func TestErrorRecoveryMiddleware(t *testing.T) {
	// Handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := ErrorRecovery()(panicHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handler panicked: %v", r)
			}
		}()
		handler.ServeHTTP(w, req)
	}()

	// Should return error response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check response contains error message
	body := w.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("Expected error in response, got: %s", body)
	}
}

// TestErrorRecoveryMiddlewareNoPanic tests normal operation
func TestErrorRecoveryMiddlewareNoPanic(t *testing.T) {
	handler := ErrorRecovery()(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAuthenticationMiddlewareValid tests valid token
func TestAuthenticationMiddlewareValid(t *testing.T) {
	validateToken := func(token string) (map[string]interface{}, error) {
		if token == "valid_token" {
			return map[string]interface{}{
				"sub":   "123",
				"email": "user@example.com",
			}, nil
		}
		return nil, fmt.Errorf("invalid token")
	}

	handler := Authentication(validateToken)(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer valid_token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAuthenticationMiddlewareInvalid tests invalid token
func TestAuthenticationMiddlewareInvalid(t *testing.T) {
	validateToken := func(token string) (map[string]interface{}, error) {
		if token == "valid_token" {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("invalid token")
	}

	handler := Authentication(validateToken)(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer invalid_token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestAuthenticationMiddlewareMissing tests missing auth header
func TestAuthenticationMiddlewareMissing(t *testing.T) {
	validateToken := func(token string) (map[string]interface{}, error) {
		if token == "valid_token" {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("invalid token")
	}

	handler := Authentication(validateToken)(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// No Authorization header

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestCORSMiddlewareAllowedOrigin tests allowed origin
func TestCORSMiddlewareAllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:3000", "https://example.com"})(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "http://localhost:3000" {
		t.Errorf("Expected CORS header for allowed origin, got: %s", corsHeader)
	}
}

// TestCORSMiddlewareDisallowedOrigin tests disallowed origin
func TestCORSMiddlewareDisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:3000"})(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "http://attacker.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "" {
		t.Errorf("Expected no CORS header for disallowed origin, got: %s", corsHeader)
	}
}

// TestCORSMiddlewarePreflightRequest tests OPTIONS preflight
func TestCORSMiddlewarePreflightRequest(t *testing.T) {
	handler := CORS([]string{"http://localhost:3000"})(testHandler(t))

	req, err := http.NewRequest("OPTIONS", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for preflight, got %d", w.Code)
	}

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "http://localhost:3000" {
		t.Errorf("Expected CORS header in preflight response")
	}
}

// TestRateLimitingMiddleware tests rate limiting
func TestRateLimitingMiddleware(t *testing.T) {
	handler := RateLimiting(2)(testHandler(t)) // 2 requests per second

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request should succeed
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("First request should succeed, got %d", w1.Code)
	}

	// Second request should succeed
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Errorf("Second request should succeed, got %d", w2.Code)
	}

	// Third request should be rate limited
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got %d", w3.Code)
	}
}

// TestRateLimitingMiddlewareBucketRefill tests token bucket refill
func TestRateLimitingMiddlewareBucketRefill(t *testing.T) {
	handler := RateLimiting(1)(testHandler(t)) // 1 request per second

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request succeeds
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("First request should succeed")
	}

	// Second request is rate limited
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request should be rate limited")
	}

	// Wait for bucket refill
	time.Sleep(1100 * time.Millisecond)

	// Third request should succeed after refill
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	if w3.Code != http.StatusOK {
		t.Errorf("Third request should succeed after refill")
	}
}

// TestRequestMetricsMiddleware tests metrics collection
func TestRequestMetricsMiddleware(t *testing.T) {
	handler := RequestMetrics()(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that response time header was set
	responseTime := w.Header().Get("X-Response-Time")
	if responseTime == "" {
		t.Error("Expected X-Response-Time header")
	}
}

// TestSecurityHeadersMiddleware tests security headers
func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeaders()(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("Expected header %s to be set", header)
		}
	}

	// Verify specific values
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected X-Content-Type-Options to be 'nosniff'")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Expected X-Frame-Options to be 'DENY'")
	}
}

// TestMiddlewareChain tests chaining multiple middleware
func TestMiddlewareChain(t *testing.T) {
	// Create a handler
	handler := testHandler(t)

	// Create middlewares
	middlewares := []Middleware{
		SecurityHeaders(),
		ErrorRecovery(),
		RequestMetrics(),
	}

	// Chain them
	chainedHandler := Chain(handler, middlewares...)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	chainedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that all middleware effects are present
	if w.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected security headers from SecurityHeaders middleware")
	}

	if w.Header().Get("X-Response-Time") == "" {
		t.Error("Expected X-Response-Time from RequestMetrics middleware")
	}
}

// TestMiddlewareChainOrder tests that middleware is applied in correct order
func TestMiddlewareChainOrder(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// ErrorRecovery should catch the panic
	chainedHandler := Chain(panicHandler, ErrorRecovery(), RequestMetrics())

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handler panicked: %v", r)
			}
		}()
		chainedHandler.ServeHTTP(w, req)
	}()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// TestGetClientIP tests IP extraction
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		forwardedFor  string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:       "X-Forwarded-For",
			forwardedFor: "192.168.1.1, 10.0.0.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP",
			xRealIP:    "192.168.1.2",
			expectedIP: "192.168.1.2",
		},
		{
			name:       "Remote-Addr",
			remoteAddr: "192.168.1.3:8080",
			expectedIP: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := getClientIP(req)
			if !strings.HasPrefix(ip, tt.expectedIP) {
				t.Errorf("Expected IP starting with %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

// TestResponseWriterWrapper tests status code capturing
func TestResponseWriterWrapper(t *testing.T) {
	w := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: w, statusCode: 0}

	wrapped.WriteHeader(http.StatusOK)
	if wrapped.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, wrapped.statusCode)
	}

	wrapped.Write([]byte("test"))
	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

// TestLoggingMiddlewareWithRequestBody tests logging with request body
func TestLoggingMiddlewareWithRequestBody(t *testing.T) {
	handler := Logging()(testHandler(t))

	body := bytes.NewBufferString("test body")
	req, err := http.NewRequest("POST", "/test", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAuthenticationMiddlewareWithErrorInValidation tests validation error
func TestAuthenticationMiddlewareWithErrorInValidation(t *testing.T) {
	validateToken := func(token string) (map[string]interface{}, error) {
		return nil, fmt.Errorf("validation error")
	}

	handler := Authentication(validateToken)(testHandler(t))

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer some_token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 on validation error, got %d", w.Code)
	}
}

// BenchmarkMiddlewareChain benchmarks middleware chain performance
func BenchmarkMiddlewareChain(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middlewares := []Middleware{
		SecurityHeaders(),
		ErrorRecovery(),
		RequestMetrics(),
		Logging(),
	}

	chainedHandler := Chain(handler, middlewares...)

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		chainedHandler.ServeHTTP(w, req)
	}
}

// BenchmarkLoggingMiddleware benchmarks logging middleware
func BenchmarkLoggingMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := Logging()
	wrappedHandler := middleware(handler)

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}

// BenchmarkSecurityHeadersMiddleware benchmarks security headers middleware
func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SecurityHeaders()
	wrappedHandler := middleware(handler)

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}
