package middleware

import (
	"net/http"
	"qr-menu/logger"
	"strings"
	"time"
)

// ResponseWriter wrapper per catturare status code e response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggingMiddleware intercetta tutte le richieste HTTP e le logga
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapper per catturare response info
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default
		}

		// Estrae informazioni della richiesta
		ip := getClientIP(r)
		userAgent := r.UserAgent()

		// Log della richiesta in arrivo
		logger.InfoWithContext("HTTP Request", map[string]interface{}{
			"method":     r.Method,
			"url":        r.URL.String(),
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"referer":    r.Referer(),
			"proto":      r.Proto,
			"host":       r.Host,
			"request_id": generateRequestID(),
		}, "", ip, userAgent)

		// Esegue la richiesta
		next.ServeHTTP(wrapped, r)

		// Calcola durata
		duration := time.Since(start)

		// Determina il livello di log basato sullo status code
		logLevel := "info"
		if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
			logLevel = "warn"
		} else if wrapped.statusCode >= 500 {
			logLevel = "error"
		}

		// Log della risposta
		logData := map[string]interface{}{
			"method":        r.Method,
			"url":           r.URL.String(),
			"status_code":   wrapped.statusCode,
			"response_size": wrapped.size,
			"duration_ms":   duration.Milliseconds(),
		}

		message := "HTTP Response"

		switch logLevel {
		case "warn":
			logger.WarnWithContext(message, logData, "", ip, userAgent)
		case "error":
			logger.ErrorWithContext(message, logData, "", ip, userAgent)
		default:
			logger.InfoWithContext(message, logData, "", ip, userAgent)
		}

		// Log performance se la richiesta è lenta
		if duration > time.Second {
			logger.PerformanceLog("HTTP Request", duration, map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		}
	})
}

// SecurityMiddleware logga eventi di sicurezza sospetti
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		userAgent := r.UserAgent()

		// Controlla per pattern sospetti nell'URL
		suspiciousPatterns := []string{
			"../", "..\\", // Directory traversal
			"<script", "</script>", // XSS attempts
			"SELECT ", "INSERT ", "UPDATE ", "DELETE ", // SQL injection
			"eval(", "javascript:", // Code injection
		}

		url := r.URL.String()
		for _, pattern := range suspiciousPatterns {
			if containsCaseInsensitive(url, pattern) {
				logger.SecurityEvent("SUSPICIOUS_URL",
					"Pattern sospetto rilevato nell'URL",
					"", ip, userAgent,
					map[string]interface{}{
						"url":     url,
						"pattern": pattern,
						"method":  r.Method,
					})
				break
			}
		}

		// Controlla User-Agent sospetti
		suspiciousAgents := []string{
			"sqlmap", "nikto", "nmap", "masscan", "zap",
			"burp", "grabber", "w3af", "havij",
		}

		for _, agent := range suspiciousAgents {
			if containsCaseInsensitive(userAgent, agent) {
				logger.SecurityEvent("SUSPICIOUS_USER_AGENT",
					"User-Agent sospetto rilevato",
					"", ip, userAgent,
					map[string]interface{}{
						"detected_tool": agent,
						"url":           url,
					})
				break
			}
		}

		// Rate limiting check (implementazione base)
		if isRateLimitExceeded(ip) {
			logger.SecurityEvent("RATE_LIMIT_EXCEEDED",
				"Troppe richieste dal stesso IP",
				"", ip, userAgent,
				map[string]interface{}{
					"url": url,
				})
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware logga eventi di autenticazione
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		userAgent := r.UserAgent()

		// Log tentativo di accesso a pagine protette
		if isProtectedRoute(r.URL.Path) {
			logger.AuditLog("ACCESS_ATTEMPT", "protected_route",
				"Tentativo di accesso a risorsa protetta",
				"", ip, userAgent,
				map[string]interface{}{
					"path":   r.URL.Path,
					"method": r.Method,
				})
		}

		next.ServeHTTP(w, r)
	})
}

// Funzioni helper

func getClientIP(r *http.Request) string {
	// Cerca in vari header per l'IP reale
	headers := []string{"X-Forwarded-For", "X-Real-Ip", "X-Client-Ip"}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			return ip
		}
	}

	return r.RemoteAddr
}

func generateRequestID() string {
	// Genera un ID univoco per la richiesta
	return time.Now().Format("20060102150405") + "-" +
		string(rune('A'+time.Now().Nanosecond()%26))
}

func containsCaseInsensitive(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				containsCaseInsensitiveHelper(s, substr))
}

func containsCaseInsensitiveHelper(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// Rate limiting semplice (in produzione usare Redis)
var ipRequestCount = make(map[string]int)
var ipLastReset = make(map[string]time.Time)

func isRateLimitExceeded(ip string) bool {
	const maxRequests = 100 // max richieste per minuto
	const resetInterval = time.Minute

	now := time.Now()

	// Reset contatore se è passato troppo tempo
	if lastReset, exists := ipLastReset[ip]; !exists || now.Sub(lastReset) > resetInterval {
		ipRequestCount[ip] = 0
		ipLastReset[ip] = now
	}

	ipRequestCount[ip]++

	return ipRequestCount[ip] > maxRequests
}

func isProtectedRoute(path string) bool {
	protectedPaths := []string{
		"/admin",
		"/api/",
	}

	for _, protectedPath := range protectedPaths {
		if strings.HasPrefix(path, protectedPath) {
			return true
		}
	}

	return false
}
