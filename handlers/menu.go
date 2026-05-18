package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"

	"qr-menu/analytics"
	"qr-menu/db"
	"qr-menu/models"
)

// GetMenusHandler restituisce tutti i menu del ristorante corrente in formato JSON
func GetMenusHandler(w http.ResponseWriter, r *http.Request) {
	restaurant, err := getCurrentRestaurant(r)
	if err != nil {
		http.Error(w, "Ristorante non trovato", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurant.ID)
	if err != nil {
		http.Error(w, "Errore nel recupero dei menu", http.StatusInternalServerError)
		return
	}

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
		RestaurantID: restaurant.ID,
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

// SetActiveMenuHandler imposta un menu come attivo
func SetActiveMenuHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || menu == nil || menu.RestaurantID != restaurant.ID || !menu.IsCompleted {
		http.NotFound(w, r)
		return
	}

	// Disattiva tutti i menu del ristorante con una singola batch operation (risolve N+1 query)
	err = db.MongoInstance.UpdateManyMenus(ctx,
		bson.M{"restaurant_id": restaurant.ID, "is_active": true},
		bson.M{"is_active": false},
	)
	if err != nil {
		log.Printf("Errore nel disattivare menu: %v", err)
		http.Error(w, "Errore nell'operazione", http.StatusInternalServerError)
		return
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

// DeleteMenuHandler elimina un menu
func DeleteMenuHandler(w http.ResponseWriter, r *http.Request) {
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

	username, err := ensureRestaurantUsername(ctx, restaurant)
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

	// Genera l'URL pubblico del ristorante (permanente)
	baseURL := getBaseURL(r)
	restaurantURL := fmt.Sprintf("%s/r/%s", baseURL, username)

	// Genera il QR code del ristorante
	qrCodePath := fmt.Sprintf("static/qrcodes/restaurant_%s.png", restaurant.ID)
	err = qrcode.WriteFile(restaurantURL, qrcode.Medium, 256, qrCodePath)
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
	menu.PublicURL = restaurantURL
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

	qrCodeURL := fmt.Sprintf("%s/qr/restaurant_%s.png", baseURL, restaurant.ID)
	response := models.QRCodeResponse{
		Success:   true,
		Message:   "QR code generato con successo",
		QRCodeURL: qrCodeURL,
		MenuURL:   restaurantURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		Available:   true,
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
	if !requireValidCSRF(w, r) {
		return
	}
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
		IsCompleted:  false,
		IsActive:     false,
	}

	// Duplica tutte le categorie e i piatti
	for i, category := range originalMenu.Categories {
		newCategory := models.MenuCategory{
			ID:    uuid.New().String(),
			Name:  category.Name,
			Items: make([]models.MenuItem, len(category.Items)),
		}
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

	if err := db.MongoInstance.CreateMenu(ctx, duplicatedMenu); err != nil {
		log.Printf("Errore nella duplicazione del menu: %v", err)
		http.Error(w, "Errore nella duplicazione del menu", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin?success=menu_duplicated&id=%s", duplicatedMenu.ID), http.StatusFound)
}

