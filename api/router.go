package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// SetupAPIRoutes configura tutte le route API
func SetupAPIRoutes(r *mux.Router) {
	// Sottoruter per le API con prefisso /api/v1
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Rate limiting per tutte le API (100 richieste per minuto)
	rateLimiter := RateLimitMiddleware(100)
	
	// Authentication endpoints (non richiedono autenticazione)
	api.HandleFunc("/auth/login", rateLimiter(APILoginHandler)).Methods("POST")
	api.HandleFunc("/auth/register", rateLimiter(APIRegisterHandler)).Methods("POST")
	api.HandleFunc("/auth/refresh", rateLimiter(APIRefreshTokenHandler)).Methods("POST")
	
	// Protected routes (richiedono autenticazione JWT)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return rateLimiter(func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	})
	
	// Authentication protected endpoints
	protected.HandleFunc("/auth/logout", APILogoutHandler).Methods("POST")
	protected.HandleFunc("/auth/change-password", ChangePasswordHandler).Methods("POST")
	
	// Restaurant endpoints
	protected.HandleFunc("/restaurant/profile", GetRestaurantProfileHandler).Methods("GET")
	protected.HandleFunc("/restaurant/profile", UpdateRestaurantProfileHandler).Methods("PUT")
	
	// Menu endpoints
	protected.HandleFunc("/menus", GetMenusHandler).Methods("GET")
	protected.HandleFunc("/menus", CreateMenuHandler).Methods("POST")
	protected.HandleFunc("/menus/{id}", GetMenuHandler).Methods("GET")
	protected.HandleFunc("/menus/{id}", UpdateMenuHandler).Methods("PUT")
	protected.HandleFunc("/menus/{id}", DeleteMenuHandler).Methods("DELETE")
	protected.HandleFunc("/menus/{id}/activate", SetActiveMenuHandler).Methods("POST")
	
	// Category endpoints
	protected.HandleFunc("/menus/{id}/categories", AddCategoryHandler).Methods("POST")
	
	// Item endpoints  
	protected.HandleFunc("/menus/{menu_id}/categories/{category_id}/items", AddItemHandler).Methods("POST")
	
	// API Documentation endpoint
	api.HandleFunc("/docs", APIDocsHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
}

// APIDocsHandler fornisce la documentazione API
func APIDocsHandler(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"title":       "QR Menu System API",
		"version":     "2.0.0",
		"description": "API REST per la gestione di menu digitali con QR code",
		"base_url":    "/api/v1",
		"authentication": map[string]interface{}{
			"type":        "JWT Bearer Token",
			"header":      "Authorization: Bearer <token>",
			"description": "Usare l'endpoint /api/v1/auth/login per ottenere il token",
		},
		"rate_limiting": map[string]interface{}{
			"requests_per_minute": 100,
			"headers": []string{
				"X-RateLimit-Limit",
				"X-RateLimit-Remaining", 
				"X-RateLimit-Reset",
			},
		},
		"endpoints": map[string]interface{}{
			"authentication": map[string]interface{}{
				"POST /auth/login": map[string]interface{}{
					"description": "Autentica un ristorante e restituisce un JWT token",
					"body": map[string]string{
						"username": "Username o email del ristorante",
						"password": "Password del ristorante",
					},
					"response": "LoginResponse con token JWT",
				},
				"POST /auth/register": map[string]interface{}{
					"description": "Registra un nuovo ristorante",
					"body": map[string]string{
						"username":         "Username univoco (min 3 caratteri)",
						"email":            "Email univoca",
						"password":         "Password (min 8 caratteri)",
						"confirm_password": "Conferma password",
						"restaurant_name":  "Nome del ristorante",
						"description":      "Descrizione (opzionale)",
						"address":          "Indirizzo (opzionale)",
						"phone":            "Telefono (opzionale)",
					},
					"response": "LoginResponse con token JWT",
				},
				"POST /auth/refresh": map[string]interface{}{
					"description": "Rinnova un token JWT",
					"body": map[string]string{
						"token": "Token JWT esistente da rinnovare",
					},
					"response": "Nuovo token JWT",
				},
				"POST /auth/logout": map[string]interface{}{
					"description": "Effettua logout e revoca il token",
					"auth_required": true,
					"response": "Messaggio di conferma",
				},
				"POST /auth/change-password": map[string]interface{}{
					"description": "Cambia la password del ristorante",
					"auth_required": true,
					"body": map[string]string{
						"current_password": "Password attuale",
						"new_password":     "Nuova password (min 8 caratteri)",
						"confirm_password": "Conferma nuova password",
					},
					"response": "Messaggio di conferma",
				},
			},
			"restaurant": map[string]interface{}{
				"GET /restaurant/profile": map[string]interface{}{
					"description":   "Ottieni profilo del ristorante",
					"auth_required": true,
					"response":      "Oggetto Restaurant (senza password)",
				},
				"PUT /restaurant/profile": map[string]interface{}{
					"description":   "Aggiorna profilo del ristorante",
					"auth_required": true,
					"body": map[string]string{
						"restaurant_name": "Nome del ristorante",
						"description":     "Descrizione",
						"address":         "Indirizzo",
						"phone":           "Telefono",
					},
					"response": "Oggetto Restaurant aggiornato",
				},
			},
			"menus": map[string]interface{}{
				"GET /menus": map[string]interface{}{
					"description":   "Lista tutti i menu del ristorante",
					"auth_required": true,
					"query_params": map[string]string{
						"page":     "Numero pagina (default: 1)",
						"per_page": "Elementi per pagina (default: 20, max: 100)",
					},
					"response": "Array di menu con metadata paginazione",
				},
				"POST /menus": map[string]interface{}{
					"description":   "Crea un nuovo menu",
					"auth_required": true,
					"body": map[string]interface{}{
						"name":        "Nome del menu",
						"description": "Descrizione (opzionale)",
						"categories":  "Array di categorie (opzionale)",
					},
					"response": "Oggetto Menu creato",
				},
				"GET /menus/{id}": map[string]interface{}{
					"description":   "Ottieni un menu specifico",
					"auth_required": true,
					"path_params": map[string]string{
						"id": "ID univoco del menu",
					},
					"response": "Oggetto Menu completo",
				},
				"PUT /menus/{id}": map[string]interface{}{
					"description":   "Aggiorna un menu esistente",
					"auth_required": true,
					"path_params": map[string]string{
						"id": "ID univoco del menu", 
					},
					"body": map[string]string{
						"name":        "Nome del menu",
						"description": "Descrizione",
					},
					"response": "Oggetto Menu aggiornato",
				},
				"DELETE /menus/{id}": map[string]interface{}{
					"description":   "Elimina un menu (non può essere attivo)",
					"auth_required": true,
					"path_params": map[string]string{
						"id": "ID univoco del menu",
					},
					"response": "Messaggio di conferma",
				},
				"POST /menus/{id}/activate": map[string]interface{}{
					"description":   "Attiva un menu (disattiva gli altri)",
					"auth_required": true,
					"path_params": map[string]string{
						"id": "ID univoco del menu",
					},
					"response": "Oggetto Menu attivato",
				},
				"POST /menus/{id}/categories": map[string]interface{}{
					"description":   "Aggiungi categoria a un menu",
					"auth_required": true,
					"path_params": map[string]string{
						"id": "ID univoco del menu",
					},
					"body": map[string]string{
						"name":        "Nome della categoria",
						"description": "Descrizione (opzionale)",
					},
					"response": "Oggetto Category creato",
				},
				"POST /menus/{menu_id}/categories/{category_id}/items": map[string]interface{}{
					"description":   "Aggiungi piatto a una categoria",
					"auth_required": true,
					"path_params": map[string]string{
						"menu_id":     "ID univoco del menu",
						"category_id": "ID univoco della categoria",
					},
					"body": map[string]string{
						"name":        "Nome del piatto",
						"description": "Descrizione (opzionale)",
						"price":       "Prezzo (numero)",
					},
					"response": "Oggetto MenuItem creato",
				},
			},
		},
		"response_format": map[string]interface{}{
			"success_response": map[string]interface{}{
				"success":   true,
				"data":      "Dati richiesti",
				"metadata":  "Informazioni aggiuntive (paginazione, ecc.)",
				"timestamp": "Timestamp ISO 8601",
			},
			"error_response": map[string]interface{}{
				"success": false,
				"error": map[string]string{
					"code":    "Codice errore",
					"message": "Messaggio errore",
					"details": "Dettagli aggiuntivi (opzionale)",
				},
				"timestamp": "Timestamp ISO 8601",
			},
		},
		"status_codes": map[string]string{
			"200": "OK - Richiesta completata con successo",
			"201": "Created - Risorsa creata con successo",
			"400": "Bad Request - Dati della richiesta non validi",
			"401": "Unauthorized - Autenticazione richiesta o non valida",
			"403": "Forbidden - Accesso negato alla risorsa",
			"404": "Not Found - Risorsa non trovata",
			"409": "Conflict - Risorsa già esistente o conflitto",
			"429": "Too Many Requests - Rate limit superato",
			"500": "Internal Server Error - Errore interno del server",
		},
	}
	
	SuccessResponse(w, docs, nil)
}

// HealthCheckHandler verifica lo stato dell'API
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.0.0",
		"uptime":    "N/A", // Implementare se necessario
		"database":  "in-memory", // Cambiare quando si usa database reale
		"services": map[string]string{
			"authentication": "running",
			"logging":        "running",
			"rate_limiting":  "running",
		},
		"stats": map[string]interface{}{
			"restaurants": len(apiRestaurants),
			"menus":       len(apiMenus),
			"active_tokens": len(revokedTokens), // Numero token revocati
		},
	}
	
	SuccessResponse(w, health, nil)
}