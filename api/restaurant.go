package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qr-menu/logger"
	"qr-menu/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Authentication API Endpoints

// LoginRequest rappresenta una richiesta di login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse rappresenta una risposta di login
type LoginResponse struct {
	Token      string             `json:"token"`
	ExpiresAt  string             `json:"expires_at"`
	Restaurant *models.Restaurant `json:"restaurant"`
}

// RegisterRequest rappresenta una richiesta di registrazione
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=50"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	RestaurantName  string `json:"restaurant_name" validate:"required,min=2,max=100"`
	Description     string `json:"description" validate:"max=500"`
	Address         string `json:"address" validate:"max=200"`
	Phone           string `json:"phone" validate:"max=20"`
}

// RefreshTokenRequest rappresenta una richiesta di refresh token
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// ChangePasswordRequest rappresenta una richiesta di cambio password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// UpdateRestaurantRequest rappresenta una richiesta di aggiornamento ristorante
type UpdateRestaurantRequest struct {
	RestaurantName string `json:"restaurant_name" validate:"required,min=2,max=100"`
	Description    string `json:"description" validate:"max=500"`
	Address        string `json:"address" validate:"max=200"`
	Phone          string `json:"phone" validate:"max=20"`
}

// APILoginHandler gestisce il login API
func APILoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione input
	if req.Username == "" || req.Password == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Username e password richiesti", "")
		return
	}

	ip := getClientIP(r)
	userAgent := r.UserAgent()
	username := strings.TrimSpace(req.Username)

	// Log tentativo di login
	logger.AuditLog("API_LOGIN_ATTEMPT", "authentication",
		"Tentativo di login API", "", ip, userAgent,
		map[string]interface{}{
			"username": username,
		})

	// Trova il ristorante
	var restaurant *models.Restaurant
	for _, rest := range apiRestaurants {
		if (rest.Username == username || rest.Email == username) && rest.IsActive {
			restaurant = rest
			break
		}
	}

	if restaurant == nil {
		logger.SecurityEvent("API_LOGIN_FAILED", "Ristorante non trovato",
			"", ip, userAgent,
			map[string]interface{}{
				"username": username,
				"reason":   "restaurant_not_found",
			})

		ErrorResponse(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
			"Credenziali non valide", "")
		return
	}

	if restaurant.Role == "" {
		restaurant.Role = defaultRole
	}

	// Verifica password
	if err := bcrypt.CompareHashAndPassword([]byte(restaurant.PasswordHash), []byte(req.Password)); err != nil {
		logger.SecurityEvent("API_LOGIN_FAILED", "Password errata",
			"", ip, userAgent,
			map[string]interface{}{
				"username":      username,
				"restaurant_id": restaurant.ID,
				"reason":        "invalid_password",
			})

		ErrorResponse(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
			"Credenziali non valide", "")
		return
	}

	// Genera JWT token
	token, err := GenerateJWT(restaurant)
	if err != nil {
		logger.Error("Errore nella generazione del token JWT", map[string]interface{}{
			"error":         err.Error(),
			"restaurant_id": restaurant.ID,
		})

		ErrorResponse(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED",
			"Errore nella generazione del token", "")
		return
	}

	// Aggiorna ultimo login
	restaurant.LastLogin = time.Now()

	// Log login riuscito
	logger.AuditLog("API_LOGIN_SUCCESS", "authentication",
		"Login API completato con successo", restaurant.ID, ip, userAgent,
		map[string]interface{}{
			"username":        username,
			"restaurant_name": restaurant.Name,
		})

	response := LoginResponse{
		Token:      token,
		ExpiresAt:  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		Restaurant: restaurant,
	}

	SuccessResponse(w, response, nil)
}

// APIRegisterHandler gestisce la registrazione API
func APIRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	ip := getClientIP(r)
	userAgent := r.UserAgent()

	// Log tentativo di registrazione
	logger.AuditLog("API_REGISTER_ATTEMPT", "authentication",
		"Tentativo di registrazione API", "", ip, userAgent,
		map[string]interface{}{
			"username":        req.Username,
			"email":           req.Email,
			"restaurant_name": req.RestaurantName,
		})

	// Validazione input
	if err := validateRegisterRequest(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Dati non validi", err.Error())
		return
	}

	// Verifica unicità username ed email
	if err := checkUniqueCredentials(req.Username, req.Email); err != nil {
		ErrorResponse(w, http.StatusConflict, "DUPLICATE_CREDENTIALS",
			"Credenziali già esistenti", err.Error())
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "PASSWORD_HASH_FAILED",
			"Errore nella crittografia della password", "")
		return
	}

	// Crea ristorante
	restaurant := &models.Restaurant{
		ID:           uuid.New().String(),
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(strings.ToLower(req.Email)),
		PasswordHash: string(passwordHash),
		Role:         defaultRole,
		Name:         strings.TrimSpace(req.RestaurantName),
		Description:  strings.TrimSpace(req.Description),
		Address:      strings.TrimSpace(req.Address),
		Phone:        strings.TrimSpace(req.Phone),
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	// Salva ristorante
	apiRestaurants[restaurant.ID] = restaurant

	// Genera JWT token
	token, err := GenerateJWT(restaurant)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED",
			"Errore nella generazione del token", "")
		return
	}

	logger.AuditLog("API_REGISTER_SUCCESS", "authentication",
		"Registrazione API completata con successo", restaurant.ID, ip, userAgent,
		map[string]interface{}{
			"username":        restaurant.Username,
			"email":           restaurant.Email,
			"restaurant_name": restaurant.Name,
		})

	response := LoginResponse{
		Token:      token,
		ExpiresAt:  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		Restaurant: restaurant,
	}

	CreatedResponse(w, response)
}

// APIRefreshTokenHandler rinnova un token JWT
func APIRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Valida il token esistente
	claims, err := ValidateJWT(req.Token)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN",
			"Token non valido", err.Error())
		return
	}

	// Trova il ristorante
	restaurant, exists := apiRestaurants[claims.RestaurantID]
	if !exists || !restaurant.IsActive {
		ErrorResponse(w, http.StatusUnauthorized, "RESTAURANT_NOT_FOUND",
			"Ristorante non trovato o disattivato", "")
		return
	}

	// Revoca il token vecchio
	RevokeJWT(req.Token)

	// Genera nuovo token
	newToken, err := GenerateJWT(restaurant)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED",
			"Errore nella generazione del nuovo token", "")
		return
	}

	logger.AuditLog("TOKEN_REFRESHED", "authentication",
		"Token JWT rinnovato", restaurant.ID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"old_token_hash": hashToken(req.Token),
			"new_token_hash": hashToken(newToken),
		})

	response := map[string]interface{}{
		"token":      newToken,
		"expires_at": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}

	SuccessResponse(w, response, nil)
}

// APILogoutHandler gestisce il logout (revoca token)
func APILogoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 {
		tokenString := authHeader[7:]
		RevokeJWT(tokenString)

		logger.AuditLog("API_LOGOUT", "authentication",
			"Logout API", GetRestaurantIDFromRequest(r), getClientIP(r), r.UserAgent(),
			map[string]interface{}{
				"token_hash": hashToken(tokenString),
			})
	}

	SuccessResponse(w, map[string]string{"message": "Logout completato"}, nil)
}

// GetRestaurantProfileHandler restituisce il profilo del ristorante
func GetRestaurantProfileHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	restaurant, exists := apiRestaurants[restaurantID]
	if !exists {
		ErrorResponse(w, http.StatusNotFound, "RESTAURANT_NOT_FOUND",
			"Ristorante non trovato", "")
		return
	}

	// Rimuovi password hash dalla risposta
	restaurantCopy := *restaurant
	restaurantCopy.PasswordHash = ""

	SuccessResponse(w, &restaurantCopy, nil)
}

// UpdateRestaurantProfileHandler aggiorna il profilo del ristorante
func UpdateRestaurantProfileHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	restaurant, exists := apiRestaurants[restaurantID]
	if !exists {
		ErrorResponse(w, http.StatusNotFound, "RESTAURANT_NOT_FOUND",
			"Ristorante non trovato", "")
		return
	}

	var req UpdateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione
	if req.RestaurantName == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR",
			"Nome ristorante richiesto", "")
		return
	}

	// Aggiorna dati
	oldName := restaurant.Name
	restaurant.Name = strings.TrimSpace(req.RestaurantName)
	restaurant.Description = strings.TrimSpace(req.Description)
	restaurant.Address = strings.TrimSpace(req.Address)
	restaurant.Phone = strings.TrimSpace(req.Phone)

	logger.AuditLog("RESTAURANT_UPDATED", "restaurant",
		"Profilo ristorante aggiornato via API", restaurantID, getClientIP(r), r.UserAgent(),
		map[string]interface{}{
			"old_name": oldName,
			"new_name": restaurant.Name,
		})

	// Rimuovi password hash dalla risposta
	restaurantCopy := *restaurant
	restaurantCopy.PasswordHash = ""

	SuccessResponse(w, &restaurantCopy, nil)
}

// ChangePasswordHandler cambia la password del ristorante
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	restaurant, exists := apiRestaurants[restaurantID]
	if !exists {
		ErrorResponse(w, http.StatusNotFound, "RESTAURANT_NOT_FOUND",
			"Ristorante non trovato", "")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON",
			"JSON non valido", err.Error())
		return
	}

	// Validazione
	if req.NewPassword != req.ConfirmPassword {
		ErrorResponse(w, http.StatusBadRequest, "PASSWORD_MISMATCH",
			"Le password non coincidono", "")
		return
	}

	if len(req.NewPassword) < 8 {
		ErrorResponse(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT",
			"Password troppo corta", "Minimo 8 caratteri")
		return
	}

	// Verifica password attuale
	if err := bcrypt.CompareHashAndPassword([]byte(restaurant.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		logger.SecurityEvent("PASSWORD_CHANGE_FAILED", "Password attuale errata",
			restaurantID, getClientIP(r), r.UserAgent(), nil)

		ErrorResponse(w, http.StatusUnauthorized, "INVALID_CURRENT_PASSWORD",
			"Password attuale non valida", "")
		return
	}

	// Hash nuova password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "PASSWORD_HASH_FAILED",
			"Errore nella crittografia della nuova password", "")
		return
	}

	restaurant.PasswordHash = string(newPasswordHash)

	logger.AuditLog("PASSWORD_CHANGED", "authentication",
		"Password cambiata via API", restaurantID, getClientIP(r), r.UserAgent(), nil)

	SuccessResponse(w, map[string]string{"message": "Password cambiata con successo"}, nil)
}

// Funzioni helper per validazione

func validateRegisterRequest(req *RegisterRequest) error {
	if len(req.Username) < 3 {
		return fmt.Errorf("username deve avere almeno 3 caratteri")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password deve avere almeno 8 caratteri")
	}

	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("le password non coincidono")
	}

	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("email non valida")
	}

	if len(req.RestaurantName) < 2 {
		return fmt.Errorf("nome ristorante deve avere almeno 2 caratteri")
	}

	return nil
}

func checkUniqueCredentials(username, email string) error {
	for _, restaurant := range apiRestaurants {
		if restaurant.Username == username {
			return fmt.Errorf("username già esistente")
		}
		if restaurant.Email == strings.ToLower(email) {
			return fmt.Errorf("email già esistente")
		}
	}
	return nil
}
