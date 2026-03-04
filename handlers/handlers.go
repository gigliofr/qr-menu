package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"qr-menu/analytics"
	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/draw"
)

var (
	templates         *template.Template
	menus             = make(map[string]*models.Menu) // Storage in memoria (temporaneo)
	csrfTokens        = make(map[string]time.Time)    // CSRF protection
	maxFileSize       = int64(5 << 20)                // 5MB max file size
	allowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
)

// SetTemplates imposta i template dall'esterno (chiamato da main)
func SetTemplates(t *template.Template) {
	templates = t
	log.Printf("✅ Templates impostati in handlers package")
}

func init() {
	// Crea le directory necessarie se non esistono
	createDirectories()
	// Templates sono ora caricati da main.InitTemplates()
	// Nota: loadMenusFromStorage() rimosso - i menu sono ora caricati direttamente da MongoDB
	// Pulisci i token CSRF scaduti periodicamente
	go cleanupCSRFTokens()
}

// generateCSRFToken genera un token CSRF sicuro
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := base64.URLEncoding.EncodeToString(bytes)
	csrfTokens[token] = time.Now().Add(1 * time.Hour)
	return token
}

// validateCSRFToken valida un token CSRF
func validateCSRFToken(token string) bool {
	expiry, exists := csrfTokens[token]
	if !exists || time.Now().After(expiry) {
		delete(csrfTokens, token)
		return false
	}
	delete(csrfTokens, token) // Usa il token una sola volta
	return true
}

// cleanupCSRFTokens pulisce i token scaduti
func cleanupCSRFTokens() {
	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		now := time.Now()
		for token, expiry := range csrfTokens {
			if now.After(expiry) {
				delete(csrfTokens, token)
			}
		}
	}
}

// sanitizeInput pulisce e valida l'input utente
func sanitizeInput(input string) string {
	// Rimuove tag HTML pericolosi
	input = html.EscapeString(input)
	// Rimuove caratteri di controllo
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")
	return strings.TrimSpace(input)
}

// setSecurityHeaders imposta gli header di sicurezza
func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; script-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' https://fonts.gstatic.com; connect-src 'self'")
}

// createDirectories creates necessary directories
func createDirectories() {
	dirs := []string{"storage", "static", "static/qrcodes", "static/images", "static/images/dishes"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Errore nella creazione della directory %s: %v", dir, err)
		}
	}
}

func loadTemplates() {
	// Non fa nulla - i template sono caricati da main.InitTemplates()
	// Questo viene mantenuto per compatibilità
	log.Printf("ℹ️  loadTemplates() chiamato (templates già caricati in main)")
}

func createFallbackTemplates() {
	templates = template.New("fallback")
}

// HomeHandler gestisce la homepage - redirect al login se non autenticato
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	// Controlla se l'utente è già loggato
	_, err := getCurrentRestaurant(r)
	if err != nil {
		// Non loggato, vai al login
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		// Già loggato, vai all'admin
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// AdminHandler mostra l'interfaccia di amministrazione
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	// Verifica autenticazione e selezione ristorante
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	// Carica i menu dal database MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	menusFromDB, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurant.ID)
	if err != nil {
		log.Printf("❌ Errore nel recupero menu dal database: %v", err)
		// Continua con array vuoto in caso di errore
		menusFromDB = []*models.Menu{}
	} else {
		log.Printf("✅ AdminHandler: Recuperati %d menu per restaurant %s", len(menusFromDB), restaurant.ID)
		for i, m := range menusFromDB {
			log.Printf("   Menu %d: ID=%s, Name=%s, Categories=%d", i+1, m.ID, m.Name, len(m.Categories))
		}
	}

	// Converti slice in map per compatibilità con il template
	restaurantMenus := make(map[string]*models.Menu)
	for _, menu := range menusFromDB {
		restaurantMenus[menu.ID] = menu
	}
	
	log.Printf("✅ AdminHandler: Mappa menu creata con %d elementi", len(restaurantMenus))

	// Controlla messaggi dalla query string
	welcome := r.URL.Query().Get("welcome")
	success := r.URL.Query().Get("success")

	// Calcola statistiche e trova menu attivo
	stats := struct {
		CompletedCount  int
		TotalCategories int
	}{}

	var activeMenuID string
	for id, menu := range restaurantMenus {
		if menu.IsCompleted {
			stats.CompletedCount++
		}
		if menu.IsActive {
			activeMenuID = id
		}
		stats.TotalCategories += len(menu.Categories)
	}

	data := struct {
		Restaurant   *models.Restaurant
		Menus        map[string]*models.Menu
		Welcome      bool
		Success      string
		Stats        interface{}
		ActiveMenuID string
	}{
		Restaurant:   restaurant,
		Menus:        restaurantMenus,
		Welcome:      welcome == "1",
		Success:      success,
		Stats:        stats,
		ActiveMenuID: activeMenuID,
	}
	
	log.Printf("✅ AdminHandler: Rendering template 'admin' con %d menu, ActiveMenuID=%s", len(data.Menus), data.ActiveMenuID)
	renderTemplate(w, "admin", data)
}

// SelectRestaurantHandler mostra la pagina di selezione ristorante (GET)
func SelectRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	
	// Verifica che l'utente sia autenticato
	session, err := getSessionFromRequest(r)
	if err != nil || session == nil || session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// Recupera tutti i ristoranti dell'utente
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	restaurants, err := db.MongoInstance.GetRestaurantsByOwnerID(ctx, session.UserID)
	if err != nil {
		log.Printf("Errore nel recupero ristoranti: %v", err)
		http.Error(w, "Errore nel recupero dei ristoranti", http.StatusInternalServerError)
		return
	}
	
	// Se l'utente ha un solo ristorante, selezionalo automaticamente
	if len(restaurants) == 1 {
		// Aggiorna la sessione con il ristorante selezionato
		session.RestaurantID = restaurants[0].ID
		updateSessionInMemory(session)
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	
	// Mostra la pagina di selezione
	data := struct {
		Restaurants     []models.Restaurant
		RestaurantCount int
	}{
		Restaurants:     restaurants,
		RestaurantCount: len(restaurants),
	}
	
	renderTemplate(w, "select_restaurant", data)
}

// SelectRestaurantPostHandler gestisce la selezione del ristorante (POST)
func SelectRestaurantPostHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}
	
	restaurantID := r.FormValue("restaurant_id")
	if restaurantID == "" {
		http.Error(w, "ID ristorante mancante", http.StatusBadRequest)
		return
	}
	
	// Verifica che l'utente sia autenticato
	session, err := getSessionFromRequest(r)
	if err != nil || session == nil || session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// Verifica che il ristorante appartenga all'utente
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, restaurantID)
	if err != nil {
		logger.Error("Errore nel recupero del ristorante", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"user_id":       session.UserID,
		})
		http.Error(w, "Errore nel recupero del ristorante", http.StatusInternalServerError)
		return
	}
	
	if restaurant == nil {
		logger.Warn("Ristorante non trovato", map[string]interface{}{
			"restaurant_id": restaurantID,
			"user_id":       session.UserID,
		})
		http.Error(w, "Ristorante non trovato", http.StatusNotFound)
		return
	}
	
	logger.Debug("Verifica ownership ristorante", map[string]interface{}{
		"restaurant_id":      restaurantID,
		"restaurant_name":    restaurant.Name,
		"restaurant_ownerid": restaurant.OwnerID,
		"session_userid":     session.UserID,
		"match":              restaurant.OwnerID == session.UserID,
	})
	
	if restaurant.OwnerID != session.UserID {
		logger.Warn("Tentativo di accesso non autorizzato al ristorante", map[string]interface{}{
			"restaurant_id":      restaurantID,
			"restaurant_ownerid": restaurant.OwnerID,
			"user_id":            session.UserID,
		})
		http.Error(w, "Accesso non autorizzato al ristorante", http.StatusForbidden)
		return
	}
	
	// Aggiorna la sessione con il ristorante selezionato
	session.RestaurantID = restaurantID
	updateSessionInMemory(session)
	
	// Log della selezione
	ip := getClientIP(r)
	log.Printf("Utente %s ha selezionato il ristorante %s (%s) da IP %s", 
		session.UserID, restaurantID, restaurant.Name, ip)
	
	// Redirect all'admin
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// AddRestaurantHandler mostra il form per aggiungere un nuovo ristorante (GET)
func AddRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	
	// Verifica che l'utente sia autenticato
	session, err := getSessionFromRequest(r)
	if err != nil || session == nil || session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	data := struct {
		Errors   []string
		FormData struct {
			Name        string
			Description string
			Address     string
			Phone       string
		}
	}{}
	
	renderTemplate(w, "add_restaurant", data)
}

// AddRestaurantPostHandler gestisce la creazione di un nuovo ristorante (POST)
func AddRestaurantPostHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}
	
	// Verifica che l'utente sia autenticato
	session, err := getSessionFromRequest(r)
	if err != nil || session == nil || session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// Valida input
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	address := strings.TrimSpace(r.FormValue("address"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	
	var errors []string
	
	if name == "" {
		errors = append(errors, "Il nome del ristorante è obbligatorio")
	} else if len(name) < 2 {
		errors = append(errors, "Il nome del ristorante deve essere almeno 2 caratteri")
	} else if len(name) > 100 {
		errors = append(errors, "Il nome del ristorante non può superare 100 caratteri")
	}
	
	if len(description) > 500 {
		errors = append(errors, "La descrizione non può superare 500 caratteri")
	}
	
	if len(address) > 200 {
		errors = append(errors, "L'indirizzo non può superare 200 caratteri")
	}
	
	if len(phone) > 20 {
		errors = append(errors, "Il telefono non può superare 20 caratteri")
	}
	
	// Se ci sono errori, mostra il form con i dati inseriti
	if len(errors) > 0 {
		data := struct {
			Errors   []string
			FormData struct {
				Name        string
				Description string
				Address     string
				Phone       string
			}
		}{
			Errors: errors,
		}
		data.FormData.Name = name
		data.FormData.Description = description
		data.FormData.Address = address
		data.FormData.Phone = phone
		
		renderTemplate(w, "add_restaurant", data)
		return
	}
	
	// Crea nuovo ristorante
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	restaurant := &models.Restaurant{
		ID:          uuid.New().String(),
		OwnerID:     session.UserID, // ⭐ Collega al user loggato
		Name:        name,
		Description: description,
		Address:     address,
		Phone:       phone,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	
	if err := db.MongoInstance.CreateRestaurant(ctx, restaurant); err != nil {
		log.Printf("Errore nella creazione del ristorante: %v", err)
		errors = append(errors, "Errore durante la creazione del ristorante. Riprova.")
		
		data := struct {
			Errors   []string
			FormData struct {
				Name        string
				Description string
				Address     string
				Phone       string
			}
		}{
			Errors: errors,
		}
		data.FormData.Name = name
		data.FormData.Description = description
		data.FormData.Address = address
		data.FormData.Phone = phone
		
		renderTemplate(w, "add_restaurant", data)
		return
	}
	
	// Log creazione
	ip := getClientIP(r)
	log.Printf("Nuovo ristorante creato: %s (ID: %s) da user %s da IP %s", 
		restaurant.Name, restaurant.ID, session.UserID, ip)
	
	// Aggiorna sessione per selezionare automaticamente il nuovo ristorante
	session.RestaurantID = restaurant.ID
	updateSessionInMemory(session)
	
	// Redirect all'admin con messaggio di successo
	http.Redirect(w, r, "/admin?success=restaurant_created", http.StatusFound)
}

// updateSessionInMemory aggiorna la sessione in MongoDB
func updateSessionInMemory(session *models.Session) {
	session.LastAccessed = time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.MongoInstance.UpdateSession(ctx, session); err != nil {
		logger.Error("Errore nell'aggiornamento della sessione in MongoDB", map[string]interface{}{
			"error":      err.Error(),
			"session_id": session.ID,
		})
	}
}


// CreateMenuHandler mostra il form per creare un nuovo menu
func CreateMenuHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	renderTemplate(w, "create_menu", nil)
}

// CreateMenuPostHandler gestisce la creazione di un nuovo menu
func CreateMenuPostHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	menu := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurant.ID, // Associa al ristorante loggato
		Name:         r.FormValue("name"),
		Description:  r.FormValue("description"),
		Categories:   []models.MenuCategory{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsCompleted:  false,
		IsActive:     false, // Non attivo inizialmente
	}

	// Aggiungi categorie e items dal form
	categoryNames := r.Form["category_name[]"]
	categoryDescriptions := r.Form["category_description[]"]

	for i, catName := range categoryNames {
		if catName != "" {
			category := models.MenuCategory{
				ID:          uuid.New().String(),
				Name:        catName,
				Description: "",
				Items:       []models.MenuItem{},
			}

			if i < len(categoryDescriptions) {
				category.Description = categoryDescriptions[i]
			}

			// Aggiungi i piatti per questa categoria
			categoryIndex := i + 1
			itemNames := r.Form[fmt.Sprintf("item_name_%d[]", categoryIndex)]
			itemDescriptions := r.Form[fmt.Sprintf("item_description_%d[]", categoryIndex)]
			itemPricesStr := r.Form[fmt.Sprintf("item_price_%d[]", categoryIndex)]

			for j, itemName := range itemNames {
				if itemName != "" {
					var price float64 = 0
					if j < len(itemPricesStr) && itemPricesStr[j] != "" {
						if parsedPrice, err := strconv.ParseFloat(itemPricesStr[j], 64); err == nil {
							price = parsedPrice
						}
					}

					var description string
					if j < len(itemDescriptions) {
						description = itemDescriptions[j]
					}

					item := models.MenuItem{
						ID:          uuid.New().String(),
						Name:        itemName,
						Description: description,
						Price:       price,
						Category:    catName,
						Available:   true,
					}

					category.Items = append(category.Items, item)
				}
			}

			menu.Categories = append(menu.Categories, category)
		}
	}

	// Salva il menu in MongoDB
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := db.MongoInstance.CreateMenu(ctx, menu); err != nil {
		log.Printf("Errore nel salvataggio del menu: %v", err)
		http.Error(w, "Errore nel salvataggio del menu", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menu.ID), http.StatusFound)
}

// EditMenuHandler mostra il form per modificare un menu esistente
func EditMenuHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		// Usa il template 404 personalizzato per menu non trovati
		data := struct {
			Title   string
			Message string
		}{
			Title:   "Menu Non Trovato",
			Message: "Il menu richiesto non esiste o non hai i permessi per modificarlo.",
		}
		w.WriteHeader(http.StatusNotFound)
		renderTemplate(w, "404", data)
		return
	}

	// Genera URL pubblico se non esiste
	if menu.PublicURL == "" {
		baseURL := getBaseURL(r)
		menu.PublicURL = fmt.Sprintf("%s/menu/%s", baseURL, menuID)
		if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
			log.Printf("Errore nell'aggiornamento URL pubblico: %v", err)
		}
	}

	data := struct {
		Menu       *models.Menu
		Restaurant *models.Restaurant
	}{
		Menu:       menu,
		Restaurant: restaurant,
	}

	renderTemplate(w, "edit_menu", data)
}

// UpdateMenuHandler aggiorna un menu esistente
func UpdateMenuHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	// Aggiorna i dettagli base del menu
	menu.Name = r.FormValue("name")
	menu.Description = r.FormValue("description")
	menu.UpdatedAt = time.Now()

	// Salva le modifiche in MongoDB
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		log.Printf("Errore nell'aggiornamento del menu: %v", err)
		http.Error(w, "Errore nell'aggiornamento del menu", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menu.ID), http.StatusFound)
}

// CompleteMenuHandler marca un menu come completato e genera il QR code
func CompleteMenuHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Genera l'URL pubblico del menu
	baseURL := getBaseURL(r)
	menuURL := fmt.Sprintf("%s/menu/%s", baseURL, menu.ID)

	// Genera il QR code
	qrCodePath := fmt.Sprintf("static/qrcodes/menu_%s.png", menu.ID)
	err = qrcode.WriteFile(menuURL, qrcode.Medium, 256, qrCodePath)
	if err != nil {
		http.Error(w, "Errore nella generazione del QR code", http.StatusInternalServerError)
		return
	}

	// Aggiorna il menu
	menu.IsCompleted = true
	menu.QRCodePath = qrCodePath
	menu.PublicURL = menuURL
	menu.UpdatedAt = time.Now()

	// Salva le modifiche in MongoDB
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		log.Printf("Errore nell'aggiornamento del menu: %v", err)
		http.Error(w, "Errore nell'aggiornamento del menu", http.StatusInternalServerError)
		return
	}

	// Redirect all'admin con messaggio di successo
	http.Redirect(w, r, "/admin?success=menu_completed", http.StatusFound)
}

// DeleteMenuHandler elimina un menu
func DeleteMenuHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Se era il menu attivo, rimuovi il riferimento
	if restaurant.ActiveMenuID == menuID {
		restaurant.ActiveMenuID = ""
		if err := db.MongoInstance.UpdateRestaurant(ctx, restaurant); err != nil {
			log.Printf("Errore nell'aggiornamento ristorante: %v", err)
		}
	}

	// Elimina il file QR se esiste
	if menu.QRCodePath != "" {
		os.Remove(menu.QRCodePath)
	}

	// Elimina il menu da MongoDB
	if err := db.MongoInstance.DeleteMenu(ctx, menuID); err != nil {
		log.Printf("Errore nell'eliminazione del menu: %v", err)
		http.Error(w, "Errore nell'eliminazione del menu", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?success=menu_deleted", http.StatusFound)
}

// SetActiveMenuHandler imposta un menu come attivo
func SetActiveMenuHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if handleAuthError(w, r, err) {
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID || !menu.IsCompleted {
		http.NotFound(w, r)
		return
	}

	// Disattiva tutti i menu del ristorante
	allMenus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurant.ID)
	if err != nil {
		log.Printf("Errore nel recupero menu: %v", err)
		http.Error(w, "Errore nell'operazione", http.StatusInternalServerError)
		return
	}

	for _, m := range allMenus {
		if m.IsActive {
			m.IsActive = false
			if err := db.MongoInstance.UpdateMenu(ctx, m); err != nil {
				log.Printf("Errore nell'aggiornamento menu: %v", err)
			}
		}
	}

	// Attiva il menu selezionato
	menu.IsActive = true
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		log.Printf("Errore nell'attivazione del menu: %v", err)
		http.Error(w, "Errore nell'attivazione del menu", http.StatusInternalServerError)
		return
	}

	// Aggiorna il ristorante
	restaurant.ActiveMenuID = menuID
	if err := db.MongoInstance.UpdateRestaurant(ctx, restaurant); err != nil {
		log.Printf("Errore nell'aggiornamento ristorante: %v", err)
	}

	http.Redirect(w, r, "/admin?success=menu_activated", http.StatusFound)
}

// GetActiveMenuHandler restituisce il menu attivo del ristorante (per QR code)
func GetActiveMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantUsername := vars["username"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Trova il ristorante per username da MongoDB
	restaurant, err := db.MongoInstance.GetRestaurantByUsername(ctx, restaurantUsername)
	if err != nil || restaurant == nil || !restaurant.IsActive {
		http.NotFound(w, r)
		return
	}

	// Track della scansione QR code
	go func() {
		userAgent := r.Header.Get("User-Agent")
		clientIP := getClientIP(r)
		event := analytics.QRScanEvent{
			RestaurantID: restaurant.ID,
			MenuID:       restaurant.ActiveMenuID,
			Timestamp:    time.Now(),
			UserIP:       clientIP,
			UserAgent:    userAgent,
		}
		analytics.GetAnalytics().TrackQRScan(event)
	}()

	// Redirect al menu attivo
	http.Redirect(w, r, fmt.Sprintf("/menu/%s", restaurant.ActiveMenuID), http.StatusFound)
}

// PublicMenuHandler mostra il menu pubblico
func PublicMenuHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		// Usa il template 404 personalizzato
		data := struct {
			Title   string
			Message string
		}{
			Title:   "Menu Non Trovato",
			Message: "Il menu che stai cercando non esiste più o è stato rimosso dal ristorante.",
		}
		w.WriteHeader(http.StatusNotFound)
		renderTemplate(w, "404", data)
		return
	}

	// Track della visualizzazione del menu
	go func() {
		userAgent := r.Header.Get("User-Agent")
		clientIP := getClientIP(r)
		event := analytics.ViewEvent{
			RestaurantID: menu.RestaurantID,
			MenuID:       menuID,
			Timestamp:    time.Now(),
			UserIP:       clientIP,
			UserAgent:    userAgent,
			Referrer:     r.Header.Get("Referer"),
		}
		analytics.GetAnalytics().TrackView(event)
	}()

	// Ottieni i dati del ristorante da MongoDB
	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, menu.RestaurantID)
	if err != nil || restaurant == nil {
		log.Printf("Ristorante non trovato per menu pubblico: %s", menu.RestaurantID)
		// Continua anche se non troviamo il ristorante
		restaurant = &models.Restaurant{
			ID:   menu.RestaurantID,
			Name: "Ristorante",
		}
	}

	data := struct {
		Menu       *models.Menu
		Restaurant *models.Restaurant
	}{
		Menu:       menu,
		Restaurant: restaurant,
	}

	renderTemplate(w, "public_menu", data)
}

// API Handlers

// GetMenusHandler restituisce tutti i menu in formato JSON
func GetMenusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menus)
}

// GetMenuHandler restituisce un singolo menu in formato JSON
func GetMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		http.Error(w, "Menu non trovato", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

// CreateMenuAPIHandler crea un nuovo menu tramite API JSON
func CreateMenuAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione per API
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Autenticazione richiesta"})
		return
	}

	var menuReq models.MenuRequest
	if err := json.NewDecoder(r.Body).Decode(&menuReq); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		return
	}

	menu := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurant.ID, // Forza l'ID del ristorante autenticato
		Name:         menuReq.Name,
		Description:  menuReq.Description,
		Categories:   menuReq.Categories,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsCompleted:  false,
		IsActive:     false,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = db.MongoInstance.CreateMenu(ctx, menu)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Errore nella creazione del menu"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menu)
}

// GenerateQRHandler genera il QR code per un menu
func GenerateQRHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione per API
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		response := models.QRCodeResponse{
			Success: false,
			Message: "Autenticazione richiesta",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		response := models.QRCodeResponse{
			Success: false,
			Message: "Menu non trovato",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Genera l'URL pubblico del menu
	baseURL := getBaseURL(r)
	menuURL := fmt.Sprintf("%s/menu/%s", baseURL, menu.ID)

	// Genera il QR code
	qrCodePath := fmt.Sprintf("static/qrcodes/menu_%s.png", menu.ID)
	err = qrcode.WriteFile(menuURL, qrcode.Medium, 256, qrCodePath)
	if err != nil {
		response := models.QRCodeResponse{
			Success: false,
			Message: "Errore nella generazione del QR code",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Aggiorna il menu
	menu.IsCompleted = true
	menu.QRCodePath = qrCodePath
	menu.PublicURL = menuURL
	menu.UpdatedAt = time.Now()

	err = db.MongoInstance.UpdateMenu(ctx, menu)
	if err != nil {
		response := models.QRCodeResponse{
			Success: false,
			Message: "Errore nell'aggiornamento del menu",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	qrCodeURL := fmt.Sprintf("%s/qr/menu_%s.png", baseURL, menu.ID)
	response := models.QRCodeResponse{
		Success:   true,
		Message:   "QR code generato con successo",
		QRCodeURL: qrCodeURL,
		MenuURL:   menuURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Utility functions

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if templates == nil {
		renderFallbackTemplate(w, tmpl, data)
		return
	}

	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Printf("Errore nel rendering del template %s: %v", tmpl, err)
		renderFallbackTemplate(w, tmpl, data)
	}
}

func renderFallbackTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	switch tmpl {
	case "admin":
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head><title>QR Menu Admin</title></head>
		<body>
		<h1>Amministrazione Menu</h1>
		<a href="/admin/menu/create">Crea Nuovo Menu</a>
		<p>Template non caricati - utilizzando fallback</p>
		</body>
		</html>`)
	case "create_menu":
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head><title>Crea Menu</title></head>
		<body>
		<h1>Crea Nuovo Menu</h1>
		<form method="POST">
		<p><label>Nome: <input type="text" name="name" required></label></p>
		<p><label>Descrizione: <textarea name="description"></textarea></label></p>
		<p><label>ID Ristorante: <input type="text" name="restaurant_id" required></label></p>
		<p><input type="submit" value="Crea Menu"></p>
		</form>
		<a href="/admin">Torna all'admin</a>
		</body>
		</html>`)
	default:
		fmt.Fprintf(w, "<h1>Template %s non disponibile</h1>", tmpl)
	}
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func saveMenuToStorage(menu *models.Menu) {
	filename := filepath.Join("storage", fmt.Sprintf("menu_%s.json", menu.ID))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Errore nella creazione del file %s: %v", filename, err)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(menu)
}

func deleteMenuFromStorage(menuID string) {
	filename := filepath.Join("storage", fmt.Sprintf("menu_%s.json", menuID))
	os.Remove(filename)
}

func loadMenusFromStorage() {
	files, err := filepath.Glob("storage/menu_*.json")
	if err != nil {
		log.Printf("Errore nella lettura dei file di storage: %v", err)
		return
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Errore nell'apertura del file %s: %v", filename, err)
			continue
		}

		var menu models.Menu
		if err := json.NewDecoder(file).Decode(&menu); err != nil {
			log.Printf("Errore nel decode del menu da %s: %v", filename, err)
			file.Close()
			continue
		}

		menus[menu.ID] = &menu
		file.Close()
	}
}

// DuplicateItemHandler duplica un piatto esistente
func DuplicateItemHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["menuId"]
	categoryID := vars["categoryId"]
	itemID := vars["itemId"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Trova la categoria e il piatto
	var targetCategory *models.MenuCategory
	var targetItem *models.MenuItem

	for i, category := range menu.Categories {
		if category.ID == categoryID {
			targetCategory = &menu.Categories[i]
			for _, item := range category.Items {
				if item.ID == itemID {
					targetItem = &item
					break
				}
			}
			break
		}
	}

	if targetCategory == nil || targetItem == nil {
		http.Error(w, "Categoria o piatto non trovati", http.StatusNotFound)
		return
	}

	// Crea una copia del piatto
	duplicatedItem := models.MenuItem{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("%s (Copia)", targetItem.Name),
		Description: targetItem.Description,
		Price:       targetItem.Price,
		Category:    targetItem.Category,
		Available:   true, // Assicura che il piatto duplicato sia disponibile
		ImageURL:    targetItem.ImageURL,
	}

	// Aggiungi il piatto duplicato alla categoria
	targetCategory.Items = append(targetCategory.Items, duplicatedItem)

	// Aggiorna timestamp
	menu.UpdatedAt = time.Now()

	// Salva le modifiche in MongoDB
	err = db.MongoInstance.UpdateMenu(ctx, menu)
	if err != nil {
		log.Printf("Errore nell'aggiornamento del menu: %v", err)
		http.Error(w, "Errore nell'aggiornamento", http.StatusInternalServerError)
		return
	}

	// Redirect back to edit menu
	http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menuID), http.StatusSeeOther)
}

// DuplicateMenuHandler duplica un menu completo
func DuplicateMenuHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	originalMenu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || originalMenu == nil || originalMenu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Crea una copia del menu
	duplicatedMenu := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurant.ID,
		Name:         fmt.Sprintf("%s (Copia)", originalMenu.Name),
		Description:  originalMenu.Description,
		Categories:   make([]models.MenuCategory, len(originalMenu.Categories)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsCompleted:  false, // Il menu duplicato inizia come bozza
		IsActive:     false,
	}

	// Duplica tutte le categorie e i piatti
	for i, category := range originalMenu.Categories {
		newCategory := models.MenuCategory{
			ID:          uuid.New().String(),
			Name:        category.Name,
			Description: category.Description,
			Items:       make([]models.MenuItem, len(category.Items)),
		}

		// Duplica tutti i piatti della categoria
		for j, item := range category.Items {
			newItem := models.MenuItem{
				ID:          uuid.New().String(),
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				Category:    item.Category,
				Available:   item.Available,
				ImageURL:    item.ImageURL,
			}
			newCategory.Items[j] = newItem
		}

		duplicatedMenu.Categories[i] = newCategory
	}

	// Salva il menu duplicato in MongoDB
	err = db.MongoInstance.CreateMenu(ctx, duplicatedMenu)
	if err != nil {
		log.Printf("Errore nella creazione del menu duplicato: %v", err)
		http.Error(w, "Errore nella duplicazione del menu", http.StatusInternalServerError)
		return
	}

	// Redirect alla modifica del menu duplicato
	http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", duplicatedMenu.ID), http.StatusSeeOther)
}

// EditItemHandler modifica un piatto esistente
func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["menuId"]
	categoryID := vars["categoryId"]
	itemID := vars["itemId"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	// Trova e modifica il piatto
	for i, category := range menu.Categories {
		if category.ID == categoryID {
			for j, item := range category.Items {
				if item.ID == itemID {
					// Aggiorna i dati del piatto
					menu.Categories[i].Items[j].Name = r.FormValue("name")
					menu.Categories[i].Items[j].Description = r.FormValue("description")

					if priceStr := r.FormValue("price"); priceStr != "" {
						if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
							menu.Categories[i].Items[j].Price = price
						}
					}

					// Aggiorna timestamp
					menu.UpdatedAt = time.Now()

					// Salva le modifiche in MongoDB
					err = db.MongoInstance.UpdateMenu(ctx, menu)
					if err != nil {
						log.Printf("Errore nell'aggiornamento del menu: %v", err)
						http.Error(w, "Errore nell'aggiornamento", http.StatusInternalServerError)
						return
					}

					// Redirect back to edit menu
					http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menuID), http.StatusSeeOther)
					return
				}
			}
		}
	}

	http.Error(w, "Piatto non trovato", http.StatusNotFound)
}

// DeleteItemHandler elimina un piatto
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["menuId"]
	categoryID := vars["categoryId"]
	itemID := vars["itemId"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Trova ed elimina il piatto
	for i, category := range menu.Categories {
		if category.ID == categoryID {
			for j, item := range category.Items {
				if item.ID == itemID {
					// Rimuovi il piatto dalla lista
					menu.Categories[i].Items = append(
						menu.Categories[i].Items[:j],
						menu.Categories[i].Items[j+1:]...)

					// Aggiorna timestamp
					menu.UpdatedAt = time.Now()

					// Salva le modifiche in MongoDB
					err = db.MongoInstance.UpdateMenu(ctx, menu)
					if err != nil {
						log.Printf("Errore nell'aggiornamento del menu: %v", err)
						http.Error(w, "Errore nell'aggiornamento", http.StatusInternalServerError)
						return
					}

					// Redirect back to edit menu
					http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menuID), http.StatusSeeOther)
					return
				}
			}
		}
	}

	http.Error(w, "Piatto non trovato", http.StatusNotFound)
}

// AddItemHandler aggiunge un nuovo piatto a una categoria esistente
func AddItemHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	categoryID := r.FormValue("category_id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")

	if name == "" || categoryID == "" {
		http.Error(w, "Nome piatto e categoria sono obbligatori", http.StatusBadRequest)
		return
	}

	var price float64 = 0
	if priceStr != "" {
		if parsedPrice, err := strconv.ParseFloat(priceStr, 64); err == nil {
			price = parsedPrice
		}
	}

	// Trova la categoria e aggiungi il piatto
	for i, category := range menu.Categories {
		if category.ID == categoryID {
			newItem := models.MenuItem{
				ID:          uuid.New().String(),
				Name:        name,
				Description: description,
				Price:       price,
				Category:    category.Name,
				Available:   true,
			}

			menu.Categories[i].Items = append(menu.Categories[i].Items, newItem)

			// Aggiorna timestamp
			menu.UpdatedAt = time.Now()

			// Salva le modifiche in MongoDB
			err = db.MongoInstance.UpdateMenu(ctx, menu)
			if err != nil {
				log.Printf("Errore nell'aggiornamento del menu: %v", err)
				http.Error(w, "Errore nell'aggiornamento", http.StatusInternalServerError)
				return
			}

			// Redirect back to edit menu
			http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menuID), http.StatusSeeOther)
			return
		}
	}

	http.Error(w, "Categoria non trovata", http.StatusNotFound)
}

// processImageUpload gestisce l'upload e l'ottimizzazione delle immagini
func processImageUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Verifica dimensione file
	if header.Size > maxFileSize {
		return "", fmt.Errorf("file troppo grande: max 5MB")
	}

	// Verifica tipo di file
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return "", fmt.Errorf("tipo di file non supportato: %s", contentType)
	}

	// Genera nome file unico
	fileExt := filepath.Ext(header.Filename)
	if fileExt == "" {
		fileExt = ".jpg"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)
	filepath := filepath.Join("static", "images", "dishes", filename)

	// Decodifica l'immagine
	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("errore nel decoding dell'immagine: %v", err)
	}

	// Ridimensiona l'immagine per ottimizzazione (max 800x600)
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	if width > 800 || height > 600 {
		ratio := float64(width) / float64(height)
		if width > height {
			width = 800
			height = int(800 / ratio)
		} else {
			height = 600
			width = int(600 * ratio)
		}

		// Crea nuova immagine ridimensionata
		resized := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.BiLinear.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
		img = resized
	}

	// Salva l'immagine ottimizzata
	outFile, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("errore nella creazione del file: %v", err)
	}
	defer outFile.Close()

	// Encoding basato sul formato originale o come JPEG per ottimizzazione
	if format == "png" {
		err = png.Encode(outFile, img)
	} else {
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return "", fmt.Errorf("errore nell'encoding dell'immagine: %v", err)
	}

	return fmt.Sprintf("images/dishes/%s", filename), nil
}

// UploadItemImageHandler gestisce l'upload di immagini per i piatti
func UploadItemImageHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	// Verifica autenticazione
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	menuID := vars["menuId"]
	categoryID := vars["categoryId"]
	itemID := vars["itemId"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID {
		http.NotFound(w, r)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, "Errore nel parsing del form", http.StatusBadRequest)
		return
	}

	// Ottieni il file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Nessuna immagine caricata", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Processa l'upload
	imagePath, err := processImageUpload(file, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Aggiorna il piatto con l'immagine
	for i, category := range menu.Categories {
		if category.ID == categoryID {
			for j, item := range category.Items {
				if item.ID == itemID {
					// Rimuovi immagine precedente se esiste
					if item.ImageURL != "" {
						oldPath := filepath.Join("static", item.ImageURL)
						os.Remove(oldPath)
					}

					// Aggiorna con nuova immagine
					menu.Categories[i].Items[j].ImageURL = imagePath
					menu.UpdatedAt = time.Now()

					// Salva le modifiche in MongoDB
					err = db.MongoInstance.UpdateMenu(ctx, menu)
					if err != nil {
						log.Printf("Errore nell'aggiornamento del menu: %v", err)
						http.Error(w, "Errore nell'aggiornamento", http.StatusInternalServerError)
						return
					}

					// Redirect back to edit menu
					http.Redirect(w, r, fmt.Sprintf("/admin/menu/%s", menuID), http.StatusSeeOther)
					return
				}
			}
		}
	}

	http.Error(w, "Piatto non trovato", http.StatusNotFound)
}

// ShareMenuHandler gestisce le richieste di condivisione del menu
func ShareMenuHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		http.NotFound(w, r)
		return
	}

	// Track dell'accesso alla pagina di condivisione
	go func() {
		userAgent := r.Header.Get("User-Agent")
		clientIP := getClientIP(r)
		event := analytics.ShareEvent{
			RestaurantID: menu.RestaurantID,
			MenuID:       menuID,
			Platform:     "share_page",
			Timestamp:    time.Now(),
			UserIP:       clientIP,
			UserAgent:    userAgent,
		}
		analytics.GetAnalytics().TrackShare(event)
	}()

	// Ottieni dati del ristorante da MongoDB
	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, menu.RestaurantID)
	if err != nil || restaurant == nil {
		restaurant = &models.Restaurant{Name: "Ristorante"}
	}

	baseURL := getBaseURL(r)
	menuURL := fmt.Sprintf("%s/menu/%s", baseURL, menuID)
	shareText := fmt.Sprintf("Scopri il menu di %s! 🍽️", restaurant.Name)

	data := struct {
		Menu        *models.Menu
		Restaurant  *models.Restaurant
		MenuURL     string
		ShareText   string
		WhatsAppURL string
		TelegramURL string
		FacebookURL string
		TwitterURL  string
	}{
		Menu:        menu,
		Restaurant:  restaurant,
		MenuURL:     menuURL,
		ShareText:   shareText,
		WhatsAppURL: fmt.Sprintf("https://wa.me/?text=%s%%20%s", strings.ReplaceAll(shareText, " ", "%%20"), menuURL),
		TelegramURL: fmt.Sprintf("https://t.me/share/url?url=%s&text=%s", menuURL, strings.ReplaceAll(shareText, " ", "%%20")),
		FacebookURL: fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", menuURL),
		TwitterURL:  fmt.Sprintf("https://twitter.com/intent/tweet?text=%s%%20%s", strings.ReplaceAll(shareText, " ", "%%20"), menuURL),
	}

	renderTemplate(w, "share_menu", data)
}

// AnalyticsDashboardHandler gestisce la dashboard analytics
func AnalyticsDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parametri per filtrare i dati
	days := 7 // default 7 giorni
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsed, err := strconv.Atoi(daysParam); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	// Ottieni dati analytics
	dashboardData := analytics.GetAnalytics().GetDashboardData(session.RestaurantID, days)

	// Ottieni informazioni ristorante da MongoDB
	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, session.RestaurantID)
	if err != nil || restaurant == nil {
		// Crea un restaurant di default se non esiste
		restaurant = &models.Restaurant{
			Name:    "Il Tuo Ristorante",
			Address: "Indirizzo non specificato",
			Phone:   "Telefono non specificato",
		}
	}

	// Prepara i dati per il template
	data := struct {
		Restaurant *models.Restaurant
		Analytics  map[string]interface{}
	}{
		Restaurant: restaurant,
		Analytics:  dashboardData,
	}

	// Render del template
	renderTemplate(w, "analytics_dashboard", data)
}

// AnalyticsAPIHandler gestisce le richieste API per gli analytics
func AnalyticsAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	// Parametri
	days := 7
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsed, err := strconv.Atoi(daysParam); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	// Ottieni dati analytics
	dashboardData := analytics.GetAnalytics().GetDashboardData(session.RestaurantID, days)

	// Restituisci JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboardData)
}

// TrackShareHandler tracka le condivisioni specifiche per piattaforma
func TrackShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Solo POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		MenuID   string `json:"menu_id"`
		Platform string `json:"platform"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Track della condivisione
	go func() {
		userAgent := r.Header.Get("User-Agent")
		clientIP := getClientIP(r)

		// Trova il menu per ottenere il restaurantID
		var restaurantID string

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if menu, err := db.MongoInstance.GetMenuByID(ctx, requestData.MenuID); err == nil && menu != nil {
			restaurantID = menu.RestaurantID
		}

		event := analytics.ShareEvent{
			RestaurantID: restaurantID,
			MenuID:       requestData.MenuID,
			Platform:     requestData.Platform,
			Timestamp:    time.Now(),
			UserIP:       clientIP,
			UserAgent:    userAgent,
		}
		analytics.GetAnalytics().TrackShare(event)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
// ==========================================
// LEGAL PAGES HANDLERS
// ==========================================

// PrivacyPolicyHandler serves the privacy policy page
func PrivacyPolicyHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	tmpl := template.Must(template.ParseFiles("templates/privacy_policy.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering privacy policy: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
	}
}

// CookiePolicyHandler serves the cookie policy page
func CookiePolicyHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	tmpl := template.Must(template.ParseFiles("templates/cookie_policy.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering cookie policy: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
	}
}

// TermsOfServiceHandler serves the terms of service page
func TermsOfServiceHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	tmpl := template.Must(template.ParseFiles("templates/terms_of_service.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering terms of service: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
	}
}

// LegalNotesHandler serves the legal notes page (Italian specific)
func LegalNotesHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	tmpl := template.Must(template.ParseFiles("templates/legal_notes.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering legal notes: %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
	}
}