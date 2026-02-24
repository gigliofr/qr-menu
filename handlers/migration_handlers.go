package handlers

import (
	"encoding/json"
	"net/http"

	"qr-menu/db"
)

// MigrationResponse è la risposta delle richieste di migrazione
type MigrationResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GetMigrationStatusHandler recupera lo status delle migrazioni
func GetMigrationStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mm := db.GetMigrationManager()
	status := mm.GetMigrationStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status: "success",
		Data:   status,
	})
}

// ListMigrationsHandler elenca tutte le migrazioni
func ListMigrationsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mm := db.GetMigrationManager()
	migrations := mm.GetMigrations()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status: "success",
		Data:   migrations,
	})
}

// CreateMigrationFilesHandler crea i file di migrazione di default
func CreateMigrationFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mm := db.GetMigrationManager()
	if err := mm.CreateDefaultMigrations(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MigrationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status:  "success",
		Message: "Migration files creati",
	})
}

// GetDatabaseHealthHandler recupera lo stato di salute del database
func GetDatabaseHealthHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dm := db.GetDatabaseManager()
	health := dm.GetHealth()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status: "success",
		Data:   health,
	})
}

// GetAppliedMigrationsHandler recupera le migrazioni applicate
func GetAppliedMigrationsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mm := db.GetMigrationManager()
	migrations := mm.GetAppliedMigrations()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status: "success",
		Data: map[string]interface{}{
			"migrations": migrations,
			"count":      len(migrations),
		},
	})
}

// GetPendingMigrationsHandler recupera le migrazioni in attesa
func GetPendingMigrationsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mm := db.GetMigrationManager()
	migrations := mm.GetPendingMigrations()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MigrationResponse{
		Status: "success",
		Data: map[string]interface{}{
			"migrations": migrations,
			"count":      len(migrations),
		},
	})
}
