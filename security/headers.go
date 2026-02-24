package security

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to all responses
type SecurityHeadersMiddleware struct {
	config SecurityHeadersConfig
}

// SecurityHeadersConfig configures security headers
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSP string
	
	// Strict-Transport-Security (HSTS)
	HSTS string
	
	// X-Frame-Options
	FrameOptions string
	
	// X-Content-Type-Options
	ContentTypeOptions string
	
	// X-XSS-Protection
	XSSProtection string
	
	// Referrer-Policy
	ReferrerPolicy string
	
	// Permissions-Policy
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns secure default headers
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		CSP: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://js.stripe.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' https://api.stripe.com; " +
			"frame-src https://js.stripe.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests",
		HSTS:               "max-age=31536000; includeSubDomains; preload",
		FrameOptions:       "DENY",
		ContentTypeOptions: "nosniff",
		XSSProtection:      "1; mode=block",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		PermissionsPolicy:  "geolocation=(), microphone=(), camera=()",
	}
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(config SecurityHeadersConfig) *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{
		config: config,
	}
}

// Middleware returns the HTTP middleware
func (shm *SecurityHeadersMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		if shm.config.CSP != "" {
			w.Header().Set("Content-Security-Policy", shm.config.CSP)
		}
		
		// HSTS - only on HTTPS
		if shm.config.HSTS != "" && r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", shm.config.HSTS)
		}
		
		// Frame Options
		if shm.config.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", shm.config.FrameOptions)
		}
		
		// Content Type Options
		if shm.config.ContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", shm.config.ContentTypeOptions)
		}
		
		// XSS Protection
		if shm.config.XSSProtection != "" {
			w.Header().Set("X-XSS-Protection", shm.config.XSSProtection)
		}
		
		// Referrer Policy
		if shm.config.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", shm.config.ReferrerPolicy)
		}
		
		// Permissions Policy
		if shm.config.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", shm.config.PermissionsPolicy)
		}
		
		// Additional security headers
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("X-Download-Options", "noopen")
		
		next.ServeHTTP(w, r)
	})
}

// CORSConfig configures CORS settings
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // seconds
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

// CORSMiddleware handles CORS preflight and headers
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// Middleware returns the HTTP middleware
func (cm *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cm.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		if cm.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", joinStrings(cm.config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", joinStrings(cm.config.AllowedHeaders, ", "))
			
			if cm.config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", formatInt(cm.config.MaxAge))
			}
			
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		// Set exposed headers
		if len(cm.config.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", joinStrings(cm.config.ExposedHeaders, ", "))
		}
		
		next.ServeHTTP(w, r)
	})
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
