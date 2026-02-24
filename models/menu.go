package models

import (
	"time"
)

// MenuItem rappresenta un singolo elemento del menu
type MenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Available   bool    `json:"available"`
	ImageURL    string  `json:"image_url,omitempty"`
}

// MenuCategory rappresenta una categoria del menu
type MenuCategory struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Items       []MenuItem `json:"items"`
}

// Menu rappresenta il menu completo
type Menu struct {
	ID           string         `json:"id"`
	RestaurantID string         `json:"restaurant_id"` // Ora è l'ID del ristorante proprietario
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Categories   []MenuCategory `json:"categories"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	IsCompleted  bool           `json:"is_completed"`
	IsActive     bool           `json:"is_active"` // Se è il menu attivo per il QR code
	QRCodePath   string         `json:"qr_code_path,omitempty"`
	PublicURL    string         `json:"public_url,omitempty"`
}

// Restaurant rappresenta le informazioni del ristorante con autenticazione
type Restaurant struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`        // Username unico per login
	Email           string    `json:"email"`           // Email unica
	PasswordHash    string    `json:"-"`               // Password hash (non serializzata in JSON)
	Name            string    `json:"name"`            // Nome del ristorante
	Description     string    `json:"description"`
	Address         string    `json:"address"`
	Phone           string    `json:"phone"`
	Logo            string    `json:"logo,omitempty"`
	ActiveMenuID    string    `json:"active_menu_id,omitempty"` // ID del menu attivo per QR code
	CreatedAt       time.Time `json:"created_at"`
	LastLogin       time.Time `json:"last_login,omitempty"`
	IsActive        bool      `json:"is_active"`       // Account attivo
}

// MenuRequest rappresenta i dati per creare/modificare un menu
type MenuRequest struct {
	RestaurantID string         `json:"restaurant_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Categories   []MenuCategory `json:"categories"`
}

// QRCodeResponse rappresenta la risposta con il QR code generato
type QRCodeResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	QRCodeURL string `json:"qr_code_url,omitempty"`
	MenuURL   string `json:"menu_url,omitempty"`
}

// User Authentication Models

// RegisterRequest rappresenta i dati per la registrazione
type RegisterRequest struct {
	Username        string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	RestaurantName string `json:"restaurant_name"`
	Description    string `json:"description,omitempty"`
	Address        string `json:"address,omitempty"`
	Phone          string `json:"phone,omitempty"`
}

// LoginRequest rappresenta i dati per il login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse rappresenta la risposta di autenticazione
type AuthResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Restaurant *Restaurant `json:"restaurant,omitempty"`
}

// Session rappresenta una sessione utente
type Session struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

// SetActiveMenuRequest rappresenta la richiesta per impostare un menu come attivo
type SetActiveMenuRequest struct {
	MenuID string `json:"menu_id"`
}