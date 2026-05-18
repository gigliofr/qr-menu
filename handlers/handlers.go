package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
)

var (
	csrfTokens        = make(map[string]time.Time)    // CSRF protection
	maxFileSize       = int64(5 << 20)                // 5MB max file size
	allowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
)

// Template helpers moved to handlers/utils.go (SetTemplates/GetTemplates)

func init() {
	// Crea le directory necessarie se non esistono
	createDirectories()
	// Templates sono ora caricati da main.InitTemplates()
	// Nota: loadMenusFromStorage() rimosso - i menu sono ora caricati direttamente da MongoDB
	// Pulisci i token CSRF scaduti periodicamente
	go cleanupCSRFTokens()
}

// Menu handlers moved to handlers/menu.go

// AdminHandler mostra la dashboard admin con la lista dei menu
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	session, err := getSessionFromRequest(r)
	if err != nil || session == nil || session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if session.RestaurantID == "" {
		http.Redirect(w, r, "/select-restaurant", http.StatusFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, session.RestaurantID)
	if err != nil {
		log.Printf("Errore nel recupero del ristorante: %v", err)
		http.Error(w, "Errore interno", http.StatusInternalServerError)
		return
	}

	restaurantMenus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurant.ID)
	if err != nil {
		log.Printf("Errore nel recupero dei menu: %v", err)
		http.Error(w, "Errore interno", http.StatusInternalServerError)
		return
	}

	stats := map[string]int{"menus_count": len(restaurantMenus)}
	activeMenuID := restaurant.ActiveMenuID
	welcome := r.URL.Query().Get("welcome")
	success := r.URL.Query().Get("success")

	data := struct {
		Restaurant   *models.Restaurant
		Menus        []*models.Menu
		Welcome      bool
		Success      string
		Stats        map[string]int
		ActiveMenuID string
		BaseURL      string
		CSRFToken    string
	}{
		Restaurant:   restaurant,
		Menus:        restaurantMenus,
		Welcome:      welcome == "1",
		Success:      success,
		Stats:        stats,
		ActiveMenuID: activeMenuID,
		BaseURL:      getBaseURL(r),
		CSRFToken:    generateCSRFToken(),
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

	restaurantUsername, err := generateUniqueRestaurantUsername(ctx, name)
	if err != nil {
		log.Printf("Errore nella generazione username ristorante: %v", err)
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
	
	restaurant := &models.Restaurant{
		ID:          uuid.New().String(),
		Username:    restaurantUsername,
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
	renderTemplate(w, "create_menu", struct{ CSRFToken string }{CSRFToken: generateCSRFToken()})
}

// CreateMenuPostHandler gestisce la creazione di un nuovo menu
func CreateMenuPostHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	if !requireValidCSRF(w, r) {
		return
	}
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
		CSRFToken  string
	}{
		Menu:       menu,
		Restaurant: restaurant,
		CSRFToken:  generateCSRFToken(),
	}

	renderTemplate(w, "edit_menu", data)
}

// UpdateMenuHandler aggiorna un menu esistente
func UpdateMenuHandler(w http.ResponseWriter, r *http.Request) {
	if !requireValidCSRF(w, r) {
		return
	}
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
	if !requireValidCSRF(w, r) {
		return
	}
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

	username, err := ensureRestaurantUsername(ctx, restaurant)
	if err != nil {
		log.Printf("Errore nella gestione username ristorante: %v", err)
		http.Error(w, "Errore nella generazione del QR code", http.StatusInternalServerError)
		return
	}

	// Genera l'URL pubblico del ristorante (non del menu specifico)
	// Il QR code punta al ristorante, che mostrerà sempre il menu attivo
	baseURL := getBaseURL(r)
	restaurantURL := fmt.Sprintf("%s/r/%s", baseURL, username)

	// Genera il QR code che punta al ristorante (permanente)
	qrCodePath := fmt.Sprintf("static/qrcodes/restaurant_%s.png", restaurant.ID)
	err = qrcode.WriteFile(restaurantURL, qrcode.Medium, 256, qrCodePath)
	if err != nil {
		http.Error(w, "Errore nella generazione del QR code", http.StatusInternalServerError)
		return
	}

	// Aggiorna il menu
	menu.IsCompleted = true
	menu.QRCodePath = qrCodePath
	menu.PublicURL = restaurantURL // URL del ristorante, non del menu specifico
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

// DeleteMenuHandler moved to handlers/menu.go


// SetActiveMenuHandler moved to handlers/menu.go


// GetActiveMenuHandler moved to handlers/menu.go


// PublicMenuHandler moved to handlers/menu.go

// API Handlers

// Menu handlers moved to handlers/menu.go

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

// DownloadQRHandler gestisce il download del QR code di un menu
func DownloadQRHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verifica che il menu esista
	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		http.Error(w, "Menu non trovato", http.StatusNotFound)
		return
	}

	// Verifica che il QR code esista
	qrCodePath := fmt.Sprintf("static/qrcodes/menu_%s.png", menuID)
	if _, err := os.Stat(qrCodePath); os.IsNotExist(err) {
		http.Error(w, "QR Code non trovato", http.StatusNotFound)
		return
	}

	// Leggi il file
	fileData, err := os.ReadFile(qrCodePath)
	if err != nil {
		log.Printf("Errore lettura QR code: %v", err)
		http.Error(w, "Errore nel caricamento del QR code", http.StatusInternalServerError)
		return
	}

	// Imposta gli headers per il download
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=qrcode_%s.png", menu.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// Scrivi il file
	w.Write(fileData)
}

// CacheStatsHandler ritorna statistiche del cache
func CacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := map[string]interface{}{
		"hits":      0, // TODO: implementare stats tracking
		"misses":    0,
		"size":      0,
		"status":    "ok",
	}

	json.NewEncoder(w).Encode(stats)
}
