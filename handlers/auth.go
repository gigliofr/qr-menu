package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/models"
	"qr-menu/security"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Store per le sessioni (usa cookie sicuri)
	store *sessions.CookieStore
	// Storage locale per backwards compatibility (in fase di migrazione a MongoDB)
	restaurants = make(map[string]*models.Restaurant)
)

const defaultRestaurantRole = "owner"

func seedTestUsers() {
	// Create test users with credentials from TESTING_GUIDE.md
	testUsers := []struct {
		username string
		password string
		email    string
		name     string
		role     string
	}{
		{
			username: "admin",
			password: "admin123",
			email:    "admin@qrmenu.com",
			name:     "Admin User",
			role:     "admin",
		},
		{
			username: "owner1",
			password: "pass123",
			email:    "owner1@qrmenu.com",
			name:     "Owner Restaurant 1",
			role:     "owner",
		},
		{
			username: "staff1",
			password: "pass123",
			email:    "staff1@qrmenu.com",
			name:     "Staff Member 1",
			role:     "staff",
		},
	}

	for _, user := range testUsers {
		// Hash password
		hashedPassword, err := security.HashPassword(user.password)
		if err != nil {
			logger.Error("Errore durante l'hash della password per utente di test", map[string]interface{}{
				"username": user.username,
				"error":    err.Error(),
			})
			continue
		}

		// NOTA: Seed disabilitato per nuova architettura User/Restaurant
		// TODO: Implementare seed MongoDB con createUser + createRestaurant
		_ = hashedPassword // evita warning unused

		logger.Info("Utente di test creato", map[string]interface{}{
			"username": user.username,
			"role":     user.role,
			//"id": restaurant.ID, // COMMENTATO
		})
	}

	logger.Info("Seeding utenti di test completato", map[string]interface{}{
		"total_users": len(testUsers),
		//"restaurants": len(restaurants), // COMMENTATO
	})
}

func init() {
	// Inizializza il session store con una chiave segreta
	sessionKey := getOrCreateSessionKey()
	store = sessions.NewCookieStore([]byte(sessionKey))
	
	// Determina se siamo in produzione (Railway usa PORT env var)
	isProduction := os.Getenv("PORT") != ""
	
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 giorni
		HttpOnly: true,
		Secure:   isProduction, // ⭐ Secure=true su Railway (HTTPS), false in locale
		SameSite: http.SameSiteLaxMode,
	}

	// Seed test data se necessario (MongoDB-only, no file storage)
	if len(restaurants) == 0 {
		seedTestUsers()
	}

	logger.Info("Sistema di autenticazione inizializzato", map[string]interface{}{
		"session_max_age":    86400 * 7,
		"restaurants_count": len(restaurants),
		"secure_cookies":    isProduction,
	})
}

// getOrCreateSessionKey genera o recupera una chiave segreta per le sessioni
func getOrCreateSessionKey() string {
	keyPath := "storage/session_key.txt"

	if data, err := os.ReadFile(keyPath); err == nil {
		return string(data)
	}

	// Genera nuova chiave se non esiste
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Errore nella generazione della chiave di sessione:", err)
	}

	keyStr := hex.EncodeToString(key)

	// Salva la chiave
	os.MkdirAll("storage", 0755)
	if err := os.WriteFile(keyPath, []byte(keyStr), 0600); err != nil {
		log.Printf("Attenzione: impossibile salvare la chiave di sessione: %v", err)
	}

	return keyStr
}

// hashPassword genera l'hash della password
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// checkPassword verifica se la password è corretta
func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// createSession crea una nuova sessione per un utente (con ristorante opzionale)
func createSession(userID string, restaurantID string, r *http.Request) (*models.Session, error) {
	session := &models.Session{
		ID:           uuid.New().String(),
		UserID:       userID,       // ⭐ Utente loggato
		RestaurantID: restaurantID, // ⭐ Ristorante selezionato (può essere vuoto)
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	// ⭐ Salva sessione in MongoDB invece che in memoria
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.MongoInstance.CreateSession(ctx, session); err != nil {
		logger.Error("Errore nel salvataggio della sessione in MongoDB", map[string]interface{}{
			"error":      err.Error(),
			"session_id": session.ID,
			"user_id":    userID,
		})
		return nil, fmt.Errorf("errore salvataggio sessione: %v", err)
	}
	
	logger.Info("Sessione creata in MongoDB", map[string]interface{}{
		"session_id":    session.ID,
		"user_id":       userID,
		"restaurant_id": restaurantID,
	})

	return session, nil
}

// getSessionFromRequest recupera la sessione dalla richiesta HTTP
func getSessionFromRequest(r *http.Request) (*models.Session, error) {
	session, err := store.Get(r, "qr-menu-session")
	if err != nil {
		return nil, err
	}

	sessionID, ok := session.Values["session_id"].(string)
	if !ok || sessionID == "" {
		return nil, fmt.Errorf("nessuna sessione trovata")
	}

	// ⭐ Recupera sessione da MongoDB invece che da memoria
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	userSession, err := db.MongoInstance.GetSessionByID(ctx, sessionID)
	if err != nil {
		logger.Error("Errore nel recupero della sessione da MongoDB", map[string]interface{}{
			"error":      err.Error(),
			"session_id": sessionID,
		})
		return nil, fmt.Errorf("sessione non valida")
	}
	
	if userSession == nil {
		return nil, fmt.Errorf("sessione non trovata")
	}

	// Aggiorna il timestamp dell'ultimo accesso
	userSession.LastAccessed = time.Now()
	if err := db.MongoInstance.UpdateSession(ctx, userSession); err != nil {
		logger.Warn("Errore nell'aggiornamento LastAccessed della sessione", map[string]interface{}{
			"error":      err.Error(),
			"session_id": sessionID,
		})
	}

	return userSession, nil
}

// getCurrentRestaurant recupera il ristorante attualmente loggato
func getCurrentRestaurant(r *http.Request) (*models.Restaurant, error) {
	userSession, err := getSessionFromRequest(r)
	if err != nil {
		return nil, err
	}

	// ⭐ Verifica che un ristorante sia stato selezionato
	if userSession.RestaurantID == "" {
		return nil, fmt.Errorf("nessun ristorante selezionato")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, userSession.RestaurantID)
	if err != nil || restaurant == nil || !restaurant.IsActive {
		return nil, fmt.Errorf("ristorante non trovato o disattivato")
	}

	return restaurant, nil
}

// handleAuthError gestisce gli errori di autenticazione in modo centralizzato
// Ritorna true se ha gestito l'errore (con redirect), false altrimenti
func handleAuthError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	
	// Se l'errore è "nessun ristorante selezionato", redirect a selezione
	if err.Error() == "nessun ristorante selezionato" {
		http.Redirect(w, r, "/select-restaurant", http.StatusFound)
		return true
	}
	
	// Per tutti gli altri errori, redirect al login
	http.Redirect(w, r, "/login", http.StatusFound)
	return true
}


// LoginHandler gestisce il login con supporto multi-ristorante
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	if r.Method == "GET" {
		renderTemplate(w, "login", nil)
		return
	}

	// POST: elabora il login
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	ip := getClientIP(r)
	userAgent := r.UserAgent()

	// Log tentativo di login
	logger.AuditLog("LOGIN_ATTEMPT", "authentication",
		"Tentativo di login", "", ip, userAgent,
		map[string]interface{}{
			"username": username,
		})

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// ⭐ STEP 1: Trova User (non Restaurant) per username o email
	user, err := db.MongoInstance.GetUserByUsername(ctx, username)
	if err != nil || user == nil {
		// Prova con email
		user, err = db.MongoInstance.GetUserByEmail(ctx, strings.ToLower(username))
	}

	// ⭐ STEP 2: Verifica credenziali su User
	if user == nil || !user.IsActive || !checkPassword(user.PasswordHash, password) {
		// Log login fallito
		logger.SecurityEvent("LOGIN_FAILED", "Credenziali non valide",
			"", ip, userAgent,
			map[string]interface{}{
				"username": username,
				"reason":   "invalid_credentials",
			})

		data := struct {
			Error    string
			Username string
		}{
			Error:    "Username o password non validi",
			Username: username,
		}
		renderTemplate(w, "login", data)
		return
	}

	// ⭐ STEP 3: Ottieni tutti i ristoranti dell'utente
	restaurants, err := db.MongoInstance.GetRestaurantsByOwnerID(ctx, user.ID)
	if err != nil {
		logger.Error("Errore nel recupero ristoranti", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
		http.Error(w, "Errore nel recupero dei ristoranti", http.StatusInternalServerError)
		return
	}

	// ⭐ STEP 4: Gestisci multi-ristorante
	var userSession *models.Session
	var redirectURL string

	if len(restaurants) == 0 {
		// Caso edge: utente senza ristoranti → crea il primo
		userSession, err = createSession(user.ID, "", r)
		redirectURL = "/add-restaurant"

		logger.Warn("Utente senza ristoranti", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
		})

	} else if len(restaurants) == 1 {
		// Un solo ristorante → seleziona automaticamente
		userSession, err = createSession(user.ID, restaurants[0].ID, r)
		redirectURL = "/admin"

		logger.Info("Login con ristorante singolo", map[string]interface{}{
			"user_id":       user.ID,
			"restaurant_id": restaurants[0].ID,
		})

	} else {
		// Più ristoranti → mostra pagina di selezione
		userSession, err = createSession(user.ID, "", r) // ⭐ RestaurantID vuoto
		redirectURL = "/select-restaurant"

		logger.Info("Login multi-ristorante", map[string]interface{}{
			"user_id":          user.ID,
			"restaurant_count": len(restaurants),
		})
	}

	if err != nil {
		http.Error(w, "Errore nella creazione della sessione", http.StatusInternalServerError)
		return
	}

	// Imposta il cookie di sessione
	session, err := store.Get(r, "qr-menu-session")
	if err != nil {
		logger.Error("Errore nel recupero della sessione cookie", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
		http.Error(w, "Errore nella gestione della sessione", http.StatusInternalServerError)
		return
	}
	
	session.Values["session_id"] = userSession.ID
	
	// ⚠️ IMPORTANTE: Salva la sessione PRIMA del redirect
	if err := session.Save(r, w); err != nil {
		logger.Error("Errore nel salvataggio della sessione cookie", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
		http.Error(w, "Errore nel salvataggio della sessione", http.StatusInternalServerError)
		return
	}
	
	logger.Info("Sessione cookie salvata con successo", map[string]interface{}{
		"session_id": userSession.ID,
		"user_id":    user.ID,
	})

	// ⭐ STEP 5: Aggiorna ultimo login su User (non Restaurant)
	if err := db.MongoInstance.UpdateUserLastLogin(ctx, user.ID); err != nil {
		logger.Error("Errore nell'aggiornamento LastLogin", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
	}

	// Log login riuscito
	logger.AuditLog("LOGIN_SUCCESS", "authentication",
		"Login completato con successo", user.ID, ip, userAgent,
		map[string]interface{}{
			"user_id":         user.ID,
			"username":        user.Username,
			"restaurant_count": len(restaurants),
		})

	// Redirect appropriato
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// RegisterHandler gestisce la registrazione (User + Restaurant separati + GDPR)
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	if r.Method == "GET" {
		renderTemplate(w, "register", nil)
		return
	}

	// POST: elabora la registrazione
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	restaurantName := r.FormValue("restaurant_name")
	description := r.FormValue("description")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	
	// ⭐ GDPR: leggi consensi dal form
	privacyConsent := r.FormValue("privacy_consent") == "on"
	marketingConsent := r.FormValue("marketing_consent") == "on"

	// Validazioni
	var errors []string

	if username == "" || len(username) < 3 {
		errors = append(errors, "Username deve essere di almeno 3 caratteri")
	}

	if email == "" {
		errors = append(errors, "Email è richiesta")
	}

	if password == "" || len(password) < 6 {
		errors = append(errors, "Password deve essere di almeno 6 caratteri")
	}

	if password != confirmPassword {
		errors = append(errors, "Le password non coincidono")
	}

	if restaurantName == "" {
		errors = append(errors, "Nome ristorante è richiesto")
	}

	// ⭐ GDPR: validazione consenso obbligatorio
	if !privacyConsent {
		errors = append(errors, "Devi accettare la Privacy Policy per continuare (GDPR Art. 7)")
	}

	// Controlla unicità username ed email su MongoDB (nella nuova collection users)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existingUser, _ := db.MongoInstance.GetUserByUsername(ctx, username)
	if existingUser != nil {
		errors = append(errors, "Username già esistente")
	}

	existingEmail, _ := db.MongoInstance.GetUserByEmail(ctx, strings.ToLower(email))
	if existingEmail != nil {
		errors = append(errors, "Email già registrata")
	}

	if len(errors) > 0 {
		data := struct {
			Errors         []string
			Username       string
			Email          string
			RestaurantName string
			Description    string
			Address        string
			Phone          string
		}{
			Errors:         errors,
			Username:       username,
			Email:          email,
			RestaurantName: restaurantName,
			Description:    description,
			Address:        address,
			Phone:          phone,
		}
		renderTemplate(w, "register", data)
		return
	}

	// Hash della password
	passwordHash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Errore nella creazione dell'account", http.StatusInternalServerError)
		return
	}

	// ⭐ STEP 1: Crea nuovo User (autenticazione)
	userID := uuid.New().String()
	user := &models.User{
		ID:               userID,
		Username:         username,
		Email:            strings.ToLower(email),
		PasswordHash:     passwordHash,
		PrivacyConsent:   privacyConsent,
		MarketingConsent: marketingConsent,
		ConsentDate:      time.Now(),
		CreatedAt:        time.Now(),
		IsActive:         true,
	}

	// Salva User in MongoDB
	if err := db.MongoInstance.CreateUser(ctx, user); err != nil {
		logger.Error("Errore nel salvataggio dell'utente", map[string]interface{}{
			"error":    err.Error(),
			"username": username,
		})
		http.Error(w, "Errore nella creazione dell'account", http.StatusInternalServerError)
		return
	}

	// ⭐ STEP 2: Crea primo Restaurant dell'utente
	restaurantID := uuid.New().String()
	restaurant := &models.Restaurant{
		ID:          restaurantID,
		OwnerID:     userID, // ⭐ Link a User
		Name:        restaurantName,
		Description: description,
		Address:     address,
		Phone:       phone,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	// Salva Restaurant in MongoDB
	if err := db.MongoInstance.CreateRestaurant(ctx, restaurant); err != nil {
		logger.Error("Errore nel salvataggio del ristorante", map[string]interface{}{
			"error":    err.Error(),
			"username": username,
		})
		http.Error(w, "Errore nella creazione dell'account", http.StatusInternalServerError)
		return
	}

	// ⭐ STEP 3: Auto-login dopo registrazione (crea session con user_id)
	userSession, err := createSession(userID, restaurantID, r)
	if err != nil {
		logger.Error("Errore nella creazione della sessione dopo registrazione", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		http.Error(w, "Errore nella creazione della sessione", http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "qr-menu-session")
	if err != nil {
		logger.Error("Errore nel recupero del cookie store dopo registrazione", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		http.Error(w, "Errore nella gestione della sessione", http.StatusInternalServerError)
		return
	}
	
	session.Values["session_id"] = userSession.ID
	
	if err := session.Save(r, w); err != nil {
		logger.Error("Errore nel salvataggio del cookie dopo registrazione", map[string]interface{}{
			"error":      err.Error(),
			"user_id":    userID,
			"session_id": userSession.ID,
		})
		http.Error(w, "Errore nel salvataggio della sessione", http.StatusInternalServerError)
		return
	}

	// Log successo registrazione GDPR
	logger.Info("Nuova registrazione completata", map[string]interface{}{
		"user_id":           userID,
		"username":          username,
		"email":             email,
		"restaurant_id":     restaurantID,
		"privacy_consent":   privacyConsent,
		"marketing_consent": marketingConsent,
	})

	// Redirect all'admin con messaggio di benvenuto
	http.Redirect(w, r, "/admin?welcome=1", http.StatusFound)
}

// LogoutHandler gestisce il logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "qr-menu-session")
	if err == nil {
		// Rimuovi la sessione da MongoDB
		if sessionID, ok := session.Values["session_id"].(string); ok {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			if err := db.MongoInstance.DeleteSession(ctx, sessionID); err != nil {
				logger.Error("Errore nella cancellazione della sessione da MongoDB", map[string]interface{}{
					"error":      err.Error(),
					"session_id": sessionID,
				})
			} else {
				logger.Info("Sessione eliminata da MongoDB", map[string]interface{}{
					"session_id": sessionID,
				})
			}
		}

		// Cancella il cookie
		session.Values["session_id"] = ""
		session.Options.MaxAge = -1
		session.Save(r, w)
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

// RequireAuth middleware per proteggere le route
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getCurrentRestaurant(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

// Storage functions per persistenza

func saveRestaurantToStorage(restaurant *models.Restaurant) {
	filename := filepath.Join("storage", fmt.Sprintf("restaurant_%s.json", restaurant.ID))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Errore nella creazione del file restaurant %s: %v", filename, err)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(restaurant)
}

func saveSessionToStorage(session *models.Session) {
	filename := filepath.Join("storage", fmt.Sprintf("session_%s.json", session.ID))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Errore nella creazione del file session %s: %v", filename, err)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(session)
}

func deleteSessionFromStorage(sessionID string) {
	filename := filepath.Join("storage", fmt.Sprintf("session_%s.json", sessionID))
	os.Remove(filename)
}

func loadRestaurantsFromStorage() {
	files, err := filepath.Glob("storage/restaurant_*.json")
	if err != nil {
		log.Printf("Errore nella lettura dei file restaurant: %v", err)
		return
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Errore nell'apertura del file %s: %v", filename, err)
			continue
		}

		var restaurant models.Restaurant
		if err := json.NewDecoder(file).Decode(&restaurant); err != nil {
			log.Printf("Errore nel decode del restaurant da %s: %v", filename, err)
			file.Close()
			continue
		}

		// NOTA: Role rimosso dalla struttura Restaurant
		// TODO: Aggiornare load per usare MongoDB

		restaurants[restaurant.ID] = &restaurant
		file.Close()
	}

	log.Printf("Caricati %d ristoranti dallo storage", len(restaurants))

	// ⭐ Sessioni ora gestite direttamente da MongoDB - non serve più caricare da file
	// loadSessionsFromStorage() - DEPRECATO
}

// ⭐ DEPRECATA - Le sessioni ora sono in MongoDB
// func loadSessionsFromStorage() {
// 	Le sessioni vengono ora recuperate dinamicamente da MongoDB
// 	tramite GetSessionByID quando necessario
// }

// getClientIP estrae l'IP reale del client considerando proxy e load balancer
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
