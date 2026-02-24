package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"qr-menu/models"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Store per le sessioni (usa cookie sicuri)
	store *sessions.CookieStore
	// Storage in memoria per utenti e sessioni (in produzione usare database)
	restaurants = make(map[string]*models.Restaurant)
	sessions_map = make(map[string]*models.Session)
)

func init() {
	// Inizializza il session store con una chiave segreta
	sessionKey := getOrCreateSessionKey()
	store = sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 giorni
		HttpOnly: true,
		Secure:   false, // In produzione impostare a true con HTTPS
	}

	// Carica restaurants esistenti
	loadRestaurantsFromStorage()
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

// createSession crea una nuova sessione per un ristorante
func createSession(restaurantID string, r *http.Request) (*models.Session, error) {
	session := &models.Session{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	sessions_map[session.ID] = session
	saveSessionToStorage(session)
	
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

	userSession, exists := sessions_map[sessionID]
	if !exists {
		return nil, fmt.Errorf("sessione non valida")
	}

	// Aggiorna il timestamp dell'ultimo accesso
	userSession.LastAccessed = time.Now()
	saveSessionToStorage(userSession)

	return userSession, nil
}

// getCurrentRestaurant recupera il ristorante attualmente loggato
func getCurrentRestaurant(r *http.Request) (*models.Restaurant, error) {
	userSession, err := getSessionFromRequest(r)
	if err != nil {
		return nil, err
	}

	restaurant, exists := restaurants[userSession.RestaurantID]
	if !exists || !restaurant.IsActive {
		return nil, fmt.Errorf("ristorante non trovato o disattivato")
	}

	return restaurant, nil
}

// LoginHandler gestisce il login
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

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Trova il ristorante per username o email
	var restaurant *models.Restaurant
	for _, rest := range restaurants {
		if (rest.Username == username || rest.Email == username) && rest.IsActive {
			restaurant = rest
			break
		}
	}

	if restaurant == nil || !checkPassword(restaurant.PasswordHash, password) {
		data := struct {
			Error string
		}{
			Error: "Username o password non validi",
		}
		renderTemplate(w, "login", data)
		return
	}

	// Crea sessione
	userSession, err := createSession(restaurant.ID, r)
	if err != nil {
		http.Error(w, "Errore nella creazione della sessione", http.StatusInternalServerError)
		return
	}

	// Imposta il cookie di sessione
	session, _ := store.Get(r, "qr-menu-session")
	session.Values["session_id"] = userSession.ID
	session.Save(r, w)

	// Aggiorna ultimo login
	restaurant.LastLogin = time.Now()
	saveRestaurantToStorage(restaurant)

	// Redirect all'admin
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// RegisterHandler gestisce la registrazione
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

	// Controlla unicità username ed email
	for _, rest := range restaurants {
		if rest.Username == username {
			errors = append(errors, "Username già esistente")
		}
		if rest.Email == email {
			errors = append(errors, "Email già registrata")
		}
	}

	if len(errors) > 0 {
		data := struct {
			Errors   []string
			Username string
			Email    string
			RestaurantName string
			Description string
			Address string
			Phone string
		}{
			Errors: errors,
			Username: username,
			Email: email,
			RestaurantName: restaurantName,
			Description: description,
			Address: address,
			Phone: phone,
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

	// Crea nuovo ristorante
	restaurant := &models.Restaurant{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         restaurantName,
		Description:  description,
		Address:      address,
		Phone:        phone,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	restaurants[restaurant.ID] = restaurant
	saveRestaurantToStorage(restaurant)

	// Auto-login dopo registrazione
	userSession, err := createSession(restaurant.ID, r)
	if err != nil {
		http.Error(w, "Errore nella creazione della sessione", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "qr-menu-session")
	session.Values["session_id"] = userSession.ID
	session.Save(r, w)

	// Redirect all'admin con messaggio di benvenuto
	http.Redirect(w, r, "/admin?welcome=1", http.StatusFound)
}

// LogoutHandler gestisce il logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "qr-menu-session")
	if err == nil {
		// Rimuovi la sessione dal server
		if sessionID, ok := session.Values["session_id"].(string); ok {
			delete(sessions_map, sessionID)
			deleteSessionFromStorage(sessionID)
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

		restaurants[restaurant.ID] = &restaurant
		file.Close()
	}

	log.Printf("Caricati %d ristoranti dallo storage", len(restaurants))

	// Carica anche le sessioni
	loadSessionsFromStorage()
}

func loadSessionsFromStorage() {
	files, err := filepath.Glob("storage/session_*.json")
	if err != nil {
		log.Printf("Errore nella lettura dei file session: %v", err)
		return
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Errore nell'apertura del file %s: %v", filename, err)
			continue
		}

		var session models.Session
		if err := json.NewDecoder(file).Decode(&session); err != nil {
			log.Printf("Errore nel decode della session da %s: %v", filename, err)
			file.Close()
			continue
		}

		// Mantieni solo le sessioni recenti (ultime 24h)
		if time.Since(session.LastAccessed) < 24*time.Hour {
			sessions_map[session.ID] = &session
		} else {
			// Rimuovi sessioni scadute
			os.Remove(filename)
		}
		file.Close()
	}

	log.Printf("Caricate %d sessioni attive dallo storage", len(sessions_map))
}