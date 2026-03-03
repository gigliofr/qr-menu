package api

import (
	"context"
	"encoding/json"
	"net/http"
	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/models"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Menu API Endpoints

// MenuCreateRequest rappresenta una richiesta di creazione menu
type MenuCreateRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=100"`
	Description string                `json:"description" validate:"max=500"`
	MealType    string                `json:"meal_type" validate:"required,oneof=breakfast lunch dinner generic"`
	Categories  []models.MenuCategory `json:"categories" validate:"dive"`
}

// MenuUpdateRequest rappresenta una richiesta di aggiornamento menu
type MenuUpdateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	MealType    string `json:"meal_type" validate:"required,oneof=breakfast lunch dinner generic"`
}

// ItemCreateRequest rappresenta una richiesta di creazione piatto
type ItemCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,min=0"`
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
}

// ItemUpdateRequest rappresenta una richiesta di aggiornamento piatto
type ItemUpdateRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Available   bool    `json:"available"`
}

// CategoryCreateRequest rappresenta una richiesta di creazione categoria
type CategoryCreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// GetMenusHandler restituisce tutti i menu del ristorante autenticato
func GetMenusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	restaurantID := GetRestaurantIDFromRequest(r)
	
	// DEBUG: Log per vedere il restaurant_id ricevuto
	logger.Info("GetMenusHandler chiamato", map[string]interface{}{
		"restaurant_id": restaurantID,
		"method":        r.Method,
		"path":          r.URL.Path,
	})

	// Query parameters per paginazione
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Ottieni menu da MongoDB
	menus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurantID)
	if err != nil {
		logger.Error("Errore nel recupero menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "GET_MENUS_FAILED",
			"Errore nel recupero menu", "")
		return
	}
	
	// DEBUG: Log risultati
	logger.Info("Query completata", map[string]interface{}{
		"restaurant_id": restaurantID,
		"menus_found":   len(menus),
	})

	// Paginazione
	total := len(menus)
	start_idx := (page - 1) * perPage
	end_idx := start_idx + perPage

	if start_idx >= total {
		menus = []*models.Menu{}
	} else {
		if end_idx > total {
			end_idx = total
		}
		menus = menus[start_idx:end_idx]
	}

	totalPages := (total + perPage - 1) / perPage

	metadata := &Metadata{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	logger.PerformanceLog("API GetMenus", time.Since(start), map[string]interface{}{
		"restaurant_id": restaurantID,
		"total_menus":   total,
		"page":          page,
	})

	SuccessResponse(w, menus, metadata)
}

// GetMenuHandler restituisce un singolo menu
func GetMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	// Verifica ownership
	if menu.RestaurantID != restaurantID {
		logger.SecurityEvent("UNAUTHORIZED_ACCESS", "Tentativo di accesso a menu di altro ristorante",
			restaurantID, getClientIP(r), r.UserAgent(),
			map[string]interface{}{
				"menu_id":  menuID,
				"owner_id": menu.RestaurantID,
			})

		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "Non hai i permessi per accedere a questo menu")
		return
	}

	SuccessResponse(w, menu, nil)
}

// CreateMenuHandler crea un nuovo menu
func CreateMenuHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	var req MenuCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione input
	if req.Name == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Nome menu richiesto", "")
		return
	}

	// Crea menu
	menu := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         req.Name,
		Description:  req.Description,
		MealType:     req.MealType, // lunch, dinner, breakfast, generic
		Categories:   req.Categories,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsCompleted:  len(req.Categories) > 0,
		IsActive:     false,
	}

	// Genera ID per categorie e piatti se non presenti
	for i := range menu.Categories {
		if menu.Categories[i].ID == "" {
			menu.Categories[i].ID = uuid.New().String()
		}
		for j := range menu.Categories[i].Items {
			if menu.Categories[i].Items[j].ID == "" {
				menu.Categories[i].Items[j].ID = uuid.New().String()
			}
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Salva menu in MongoDB
	if err := db.MongoInstance.CreateMenu(ctx, menu); err != nil {
		logger.Error("Errore nel salvataggio del menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_name":     menu.Name,
		})
		ErrorResponse(w, http.StatusInternalServerError, "CREATE_MENU_FAILED",
			"Errore nella creazione del menu", "")
		return
	}

	// TODO: Log audit to MongoDB audit_logs collection

	logger.AuditLog("MENU_CREATED", "menu",
		"Menu creato via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":          menu.ID,
			"menu_name":        menu.Name,
			"categories_count": len(menu.Categories),
		})

	CreatedResponse(w, menu)
}

// UpdateMenuHandler aggiorna un menu esistente
func UpdateMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	if menu.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "")
		return
	}

	var req MenuUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione
	if req.Name == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Nome menu richiesto", "")
		return
	}

	// Aggiorna menu
	oldName := menu.Name
	menu.Name = req.Name
	menu.Description = req.Description
	menu.MealType = req.MealType
	menu.UpdatedAt = time.Now()

	// Aggiorna in MongoDB
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		logger.Error("Errore nell'aggiornamento del menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_id":       menuID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "UPDATE_MENU_FAILED",
			"Errore nell'aggiornamento del menu", "")
		return
	}

	// TODO: Log audit to MongoDB audit_logs collection

	logger.AuditLog("MENU_UPDATED", "menu",
		"Menu aggiornato via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":  menuID,
			"old_name": oldName,
			"new_name": req.Name,
		})

	SuccessResponse(w, menu, nil)
}

// DeleteMenuHandler elimina un menu
func DeleteMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	if menu.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "")
		return
	}

	// Verifica se è il menu attivo
	if menu.IsActive {
		ErrorResponse(w, http.StatusConflict, "MENU_ACTIVE",
			"Impossibile eliminare menu attivo", "Disattiva il menu prima di eliminarlo")
		return
	}

	// Elimina da MongoDB
	if err := db.MongoInstance.DeleteMenu(ctx, menuID); err != nil {
		logger.Error("Errore nell'eliminazione del menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_id":       menuID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "DELETE_MENU_FAILED",
			"Errore nell'eliminazione del menu", "")
		return
	}

	// TODO: Log audit to MongoDB audit_logs collection

	logger.AuditLog("MENU_DELETED", "menu",
		"Menu eliminato via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":   menuID,
			"menu_name": menu.Name,
		})

	SuccessResponse(w, map[string]string{"message": "Menu eliminato con successo"}, nil)
}

// SetActiveMenuHandler imposta un menu come attivo
func SetActiveMenuHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	if menu.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "")
		return
	}

	if !menu.IsCompleted {
		ErrorResponse(w, http.StatusBadRequest, "MENU_INCOMPLETE",
			"Menu non completato", "Il menu deve essere completo prima di essere attivato")
		return
	}

	// Disattiva altri menu del ristorante
	allMenus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurantID)
	if err != nil {
		logger.Error("Errore nel recupero menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "OPERATION_FAILED",
			"Errore nell'operazione", "")
		return
	}

	for _, m := range allMenus {
		if m.IsActive && m.ID != menuID {
			m.IsActive = false
			m.UpdatedAt = time.Now()
			if err := db.MongoInstance.UpdateMenu(ctx, m); err != nil {
				logger.Error("Errore nell'aggiornamento menu", map[string]interface{}{
					"error":         err.Error(),
					"menu_id":       m.ID,
					"restaurant_id": restaurantID,
				})
			}
		}
	}

	// Attiva questo menu
	menu.IsActive = true
	menu.UpdatedAt = time.Now()

	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		logger.Error("Errore nell'attivazione del menu", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_id":       menuID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "UPDATE_MENU_FAILED",
			"Errore nell'attivazione del menu", "")
		return
	}

	logger.AuditLog("MENU_ACTIVATED", "menu",
		"Menu attivato via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":   menuID,
			"menu_name": menu.Name,
		})

	SuccessResponse(w, menu, nil)
}

// AddCategoryHandler aggiunge una categoria a un menu
func AddCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	if menu.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "")
		return
	}

	var req CategoryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	if req.Name == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Nome categoria richiesto", "")
		return
	}

	category := models.MenuCategory{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Items:       []models.MenuItem{},
	}

	menu.Categories = append(menu.Categories, category)
	menu.UpdatedAt = time.Now()

	// Aggiorna in MongoDB
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		logger.Error("Errore nell'aggiunta categoria", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_id":       menuID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "UPDATE_MENU_FAILED",
			"Errore nell'aggiunta della categoria", "")
		return
	}

	logger.AuditLog("CATEGORY_ADDED", "menu",
		"Categoria aggiunta via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":       menuID,
			"category_id":   category.ID,
			"category_name": category.Name,
		})

	CreatedResponse(w, category)
}

// AddItemHandler aggiunge un piatto a una categoria
func AddItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	menuID := vars["menu_id"]
	categoryID := vars["category_id"]
	restaurantID := GetRestaurantIDFromRequest(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil || menu == nil {
		ErrorResponse(w, http.StatusNotFound, "MENU_NOT_FOUND",
			"Menu non trovato", "")
		return
	}

	if menu.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusForbidden, "ACCESS_DENIED",
			"Accesso negato", "")
		return
	}

	var req ItemCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione
	if req.Name == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Nome piatto richiesto", "")
		return
	}
	if req.Price < 0 {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Prezzo non valido", "")
		return
	}

	// Trova la categoria
	var categoryIndex = -1
	for i, cat := range menu.Categories {
		if cat.ID == categoryID {
			categoryIndex = i
			break
		}
	}

	if categoryIndex == -1 {
		ErrorResponse(w, http.StatusNotFound, "CATEGORY_NOT_FOUND",
			"Categoria non trovata", "")
		return
	}

	item := models.MenuItem{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Available:   true,
	}

	menu.Categories[categoryIndex].Items = append(menu.Categories[categoryIndex].Items, item)
	menu.UpdatedAt = time.Now()

	// Aggiorna in MongoDB
	if err := db.MongoInstance.UpdateMenu(ctx, menu); err != nil {
		logger.Error("Errore nell'aggiunta piatto", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurantID,
			"menu_id":       menuID,
			"category_id":   categoryID,
		})
		ErrorResponse(w, http.StatusInternalServerError, "UPDATE_MENU_FAILED",
			"Errore nell'aggiunta del piatto", "")
		return
	}

	logger.AuditLog("ITEM_ADDED", "menu",
		"Piatto aggiunto via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"menu_id":     menuID,
			"category_id": categoryID,
			"item_id":     item.ID,
			"item_name":   item.Name,
			"price":       item.Price,
		})

	CreatedResponse(w, item)
}
