package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"qr-menu/logger"
	"qr-menu/models"

	"github.com/golang-jwt/jwt/v5"
)

// APIResponse rappresenta una risposta API standardizzata
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Metadata  *Metadata   `json:"metadata,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// APIError rappresenta un errore API strutturato
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Metadata contiene informazioni aggiuntive sulla risposta
type Metadata struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// JWT Claims personalizzate
type Claims struct {
	RestaurantID string `json:"restaurant_id"`
	Username     string `json:"username"`
	jwt.RegisteredClaims
}

var (
	// Chiave segreta JWT (in produzione deve essere configurabile)
	jwtSecret = []byte("qr-menu-jwt-secret-2024")
	// Cache per token revocati (in produzione usare Redis)
	revokedTokens = make(map[string]bool)
)

// Funzioni helper per risposte API

// SuccessResponse crea una risposta di successo
func SuccessResponse(w http.ResponseWriter, data interface{}, metadata *Metadata) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse crea una risposta di errore
func ErrorResponse(w http.ResponseWriter, statusCode int, code, message, details string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// CreatedResponse crea una risposta per risorsa creata
func CreatedResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// JWT Authentication

// GenerateJWT genera un token JWT per un ristorante
func GenerateJWT(restaurant *models.Restaurant) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		RestaurantID: restaurant.ID,
		Username:     restaurant.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   restaurant.ID,
			Issuer:    "qr-menu-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT valida un token JWT
func ValidateJWT(tokenString string) (*Claims, error) {
	// Controlla se il token è stato revocato
	if revokedTokens[tokenString] {
		return nil, fmt.Errorf("token revocato")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo di firma non valido")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token non valido")
	}

	return claims, nil
}

// RevokeJWT revoca un token JWT
func RevokeJWT(tokenString string) {
	revokedTokens[tokenString] = true

	logger.AuditLog("TOKEN_REVOKED", "authentication",
		"Token JWT revocato", "", "", "",
		map[string]interface{}{
			"token_hash": hashToken(tokenString),
		})
}

// Middleware di autenticazione API
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ErrorResponse(w, http.StatusUnauthorized, "MISSING_AUTH_HEADER",
				"Header Authorization mancante", "Includere 'Authorization: Bearer <token>'")
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			ErrorResponse(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT",
				"Formato Authorization non valido", "Usare 'Bearer <token>'")
			return
		}

		tokenString := authHeader[7:]
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			logger.SecurityEvent("INVALID_JWT", "Token JWT non valido",
				"", getClientIP(r), r.UserAgent(),
				map[string]interface{}{
					"error":      err.Error(),
					"token_hash": hashToken(tokenString),
				})

			ErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN",
				"Token non valido", err.Error())
			return
		}

		// Aggiungi le claims al contesto della richiesta
		r.Header.Set("X-Restaurant-ID", claims.RestaurantID)
		r.Header.Set("X-Username", claims.Username)

		logger.AuditLog("API_ACCESS", "api",
			"Accesso API autorizzato", claims.RestaurantID, getClientIP(r), r.UserAgent(),
			map[string]interface{}{
				"endpoint": r.URL.Path,
				"method":   r.Method,
			})

		next.ServeHTTP(w, r)
	}
}

// Rate limiting middleware per API
func RateLimitMiddleware(requestsPerMinute int) func(http.HandlerFunc) http.HandlerFunc {
	type clientInfo struct {
		requests  int
		resetTime time.Time
	}

	clients := make(map[string]*clientInfo)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			now := time.Now()

			client, exists := clients[ip]
			if !exists {
				client = &clientInfo{
					requests:  0,
					resetTime: now.Add(time.Minute),
				}
				clients[ip] = client
			}

			// Reset contatore se è passato il tempo
			if now.After(client.resetTime) {
				client.requests = 0
				client.resetTime = now.Add(time.Minute)
			}

			client.requests++

			if client.requests > requestsPerMinute {
				logger.SecurityEvent("RATE_LIMIT_EXCEEDED",
					fmt.Sprintf("Rate limit superato (%d req/min)", requestsPerMinute),
					"", ip, r.UserAgent(),
					map[string]interface{}{
						"requests": client.requests,
						"endpoint": r.URL.Path,
					})

				ErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					"Troppe richieste",
					fmt.Sprintf("Maximum %d richieste per minuto", requestsPerMinute))
				return
			}

			// Aggiungi header di rate limiting
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(requestsPerMinute-client.requests))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(client.resetTime.Unix(), 10))

			next.ServeHTTP(w, r)
		}
	}
}

// Funzioni helper

func getClientIP(r *http.Request) string {
	headers := []string{"X-Forwarded-For", "X-Real-Ip", "X-Client-Ip"}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			return ip
		}
	}

	return r.RemoteAddr
}

func hashToken(token string) string {
	if len(token) > 10 {
		return token[:5] + "..." + token[len(token)-5:]
	}
	return "***"
}

// Funzione per ottenere l'ID del ristorante dalla richiesta
func GetRestaurantIDFromRequest(r *http.Request) string {
	return r.Header.Get("X-Restaurant-ID")
}

// Funzione per ottenere l'username dalla richiesta
func GetUsernameFromRequest(r *http.Request) string {
	return r.Header.Get("X-Username")
}
