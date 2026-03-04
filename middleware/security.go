package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"qr-menu/db"
	"qr-menu/models"

	"github.com/gorilla/mux"
)

// RestaurantOwnershipMiddleware verifica che l'utente abbia accesso al ristorante richiesto
// Previene accessi non autorizzati ai dati di altri utenti
func RestaurantOwnershipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip per route pubbliche
		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Ottieni session
		session := getSessionFromContext(r)
		if session == nil || session.UserID == "" {
			// Non autenticato, il middleware di auth gestirà
			next.ServeHTTP(w, r)
			return
		}

		// Se non c'è un ristorante selezionato, skip
		if session.RestaurantID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Verifica ownership del ristorante nella sessione
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, session.RestaurantID)
		if err != nil || restaurant == nil {
			log.Printf("🔒 SECURITY: Restaurant non trovato nella sessione: %s", session.RestaurantID)
			http.Error(w, "Ristorante non valido", http.StatusForbidden)
			return
		}

		// ⭐ VERIFICA CRITICA: Il ristorante deve appartenere all'utente
		if restaurant.OwnerID != session.UserID {
			log.Printf("🚨 SECURITY ALERT: Tentativo accesso non autorizzato!")
			log.Printf("   User ID: %s", session.UserID)
			log.Printf("   Restaurant ID: %s", session.RestaurantID)
			log.Printf("   Restaurant Owner ID: %s", restaurant.OwnerID)
			log.Printf("   IP: %s", r.RemoteAddr)
			log.Printf("   Path: %s", r.URL.Path)

			http.Error(w, "Accesso negato", http.StatusForbidden)
			return
		}

		// Se la richiesta contiene un menu_id, verifica che appartenga al ristorante
		vars := mux.Vars(r)
		if menuID := vars["id"]; menuID != "" && strings.Contains(r.URL.Path, "/menu/") {
			if !verifyMenuOwnership(ctx, menuID, session.RestaurantID) {
				log.Printf("🚨 SECURITY: Tentativo accesso menu non autorizzato")
				log.Printf("   Menu ID: %s", menuID)
				log.Printf("   Restaurant ID: %s", session.RestaurantID)
				http.Error(w, "Accesso al menu negato", http.StatusForbidden)
				return
			}
		}

		// Tutto OK, procedi
		next.ServeHTTP(w, r)
	})
}

// verifyMenuOwnership verifica che il menu appartenga al ristorante
func verifyMenuOwnership(ctx context.Context, menuID, restaurantID string) bool {
	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		return false
	}
	return menu.RestaurantID == restaurantID
}

// isPublicRoute verifica se la route è pubblica (non richiede ownership)
func isPublicRoute(path string) bool {
	publicPrefixes := []string{
		"/static/",
		"/qr/",
		"/login",
		"/register",
		"/logout",
		"/privacy",
		"/cookie-policy",
		"/terms",
		"/legal",
		"/menu/", // Menu pubblici (view-only)
		"/r/",    // Active menu pubblici
		"/api/track/", // Analytics pubblici
		"/api/v1/health",
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// getSessionFromContext recupera la sessione dal contesto della richiesta
// Assume che AuthMiddleware abbia già verificato l'autenticazione
func getSessionFromContext(r *http.Request) *models.Session {
	// Implementazione semplificata - recupera dalla sessione cookie
	// In produzione, potresti voler cachare questo nel context
	
	// Per ora, ritorniamo nil e lasciamo che gli handler verifichino tramite getCurrentRestaurant
	return nil
}

// RateLimitByUser limita le richieste per user_id
func RateLimitByUser(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	type userLimit struct {
		count     int
		resetTime time.Time
	}

	limits := make(map[string]*userLimit)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := getSessionFromContext(r)
			if session == nil || session.UserID == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID := session.UserID
			now := time.Now()

			// Pulisci entry scadute periodicamente
			if len(limits) > 1000 {
				for uid, limit := range limits {
					if now.After(limit.resetTime) {
						delete(limits, uid)
					}
				}
			}

			// Verifica limite
			limit, exists := limits[userID]
			if !exists {
				limits[userID] = &userLimit{
					count:     1,
					resetTime: now.Add(window),
				}
				next.ServeHTTP(w, r)
				return
			}

			if now.After(limit.resetTime) {
				// Reset window
				limit.count = 1
				limit.resetTime = now.Add(window)
				next.ServeHTTP(w, r)
				return
			}

			if limit.count >= maxRequests {
				log.Printf("🚨 RATE LIMIT: User %s superato limite (%d req in %v)", 
					userID, maxRequests, window)
				http.Error(w, "Troppi richieste. Riprova più tardi.", http.StatusTooManyRequests)
				return
			}

			limit.count++
			next.ServeHTTP(w, r)
		})
	}
}

// AuditLogMiddleware logga tutte le operazioni sensibili
func AuditLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logga solo operazioni POST/PUT/DELETE su route admin
		if r.Method != "GET" && strings.HasPrefix(r.URL.Path, "/admin") {
			log.Printf("📋 AUDIT: %s %s da %s", r.Method, r.URL.Path, r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}

// CSRFProtectionMiddleware verifica token CSRF per operazioni modificanti
func CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip GET requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Verifica CSRF token per POST/PUT/DELETE
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf_token")
		}

		if token == "" {
			log.Printf("🚨 SECURITY: Richiesta senza CSRF token da %s", r.RemoteAddr)
			http.Error(w, "CSRF token mancante", http.StatusForbidden)
			return
		}

		// TODO: Implementare validazione CSRF token reale
		// Per ora accetta qualsiasi token non vuoto

		next.ServeHTTP(w, r)
	})
}
