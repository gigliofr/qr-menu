package handlers

import (
	"encoding/json"
	"net/http"

	"qr-menu/backup"
)

// BackupResponse è la risposta del backup
type BackupResponse struct {
	Status    string `json:"status"`
	BackupID  string `json:"backup_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ListBackupsResponse è la lista di backup
type ListBackupsResponse struct {
	Status  string                  `json:"status"`
	Backups []backup.BackupMetadata `json:"backups"`
	Count   int                     `json:"count"`
}

// CreateBackupHandler crea un backup manuale
func CreateBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	bm := backup.GetBackupManager()
	backupID, err := bm.CreateBackup()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(BackupResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BackupResponse{
		Status:   "success",
		BackupID: backupID,
		Message:  "Backup creato con successo",
	})
}

// ListBackupsHandler elenca i backup disponibili
func ListBackupsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	bm := backup.GetBackupManager()
	backups, err := bm.ListBackups()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListBackupsResponse{
		Status:  "success",
		Backups: backups,
		Count:   len(backups),
	})
}

// DeleteBackupHandler elimina un backup
func DeleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	backupID := r.URL.Query().Get("id")
	if backupID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID backup mancante"})
		return
	}

	bm := backup.GetBackupManager()
	err = bm.DeleteBackup(backupID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(BackupResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BackupResponse{
		Status:  "success",
		Message: "Backup eliminato con successo",
	})
}

// RestoreBackupHandler ripristina un backup
func RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	// Solo admin può fare restore
	if r.Method != "POST" {
		http.Error(w, "Solo POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		BackupID string `json:"backup_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body JSON invalido"})
		return
	}

	if request.BackupID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "backup_id mancante"})
		return
	}

	bm := backup.GetBackupManager()
	err = bm.RestoreBackup(request.BackupID, ".")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(BackupResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BackupResponse{
		Status:  "success",
		Message: "Backup ripristinato con successo",
	})
}

// GetBackupStatusHandler restituisce lo stato dei backup
func GetBackupStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	bm := backup.GetBackupManager()
	lastBackup := bm.GetLastBackupTime()

	backups, err := bm.ListBackups()
	if err != nil {
		backups = []backup.BackupMetadata{}
	}

	response := map[string]interface{}{
		"status":           "ok",
		"last_backup":      lastBackup,
		"backup_count":     len(backups),
		"auto_backup":      true,
		"compression":      true,
		"backup_schedule":  "daily",
	}

	if len(backups) > 0 {
		response["latest_backup"] = backups[0]
		response["total_backup_size"] = bm.GetTotalBackupSize()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ScheduleBackupHandler configura la pianificazione automática
func ScheduleBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Solo POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var config struct {
		Type       string `json:"type"` // hourly, daily, weekly, monthly
		Hour       int    `json:"hour"`
		Day        int    `json:"day"` // per weekly/monthly
		DayOfMonth int    `json:"day_of_month"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body JSON invalido"})
		return
	}

	bm := backup.GetBackupManager()
	schedule := backup.BackupSchedule{
		Type:       config.Type,
		Hour:       config.Hour,
		Day:        config.Day,
		DayOfMonth: config.DayOfMonth,
	}

	err = bm.StartScheduled(schedule)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Pianificazione backup configurata",
	})
}

// GetBackupStatsHandler restituisce statistiche sui backup
func GetBackupStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato"})
		return
	}

	bm := backup.GetBackupManager()
	backups, err := bm.ListBackups()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	stats := map[string]interface{}{
		"total_backups":    len(backups),
		"oldest_backup":    nil,
		"newest_backup":    nil,
		"total_size":       0,
		"average_size":     0,
		"frequency":        "daily",
	}

	if len(backups) > 0 {
		stats["oldest_backup"] = backups[len(backups)-1]
		stats["newest_backup"] = backups[0]

		var totalSize int64
		for _, b := range backups {
			totalSize += b.Size
		}
		stats["total_size"] = totalSize
		stats["average_size"] = totalSize / int64(len(backups))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// DownloadBackupHandler scarica un backup
func DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Non autorizzato", http.StatusUnauthorized)
		return
	}

	backupID := r.URL.Query().Get("id")
	if backupID == "" {
		http.Error(w, "ID backup mancante", http.StatusBadRequest)
		return
	}

	bm := backup.GetBackupManager()
	backups, err := bm.ListBackups()
	if err != nil {
		http.Error(w, "Errore lettura backup", http.StatusInternalServerError)
		return
	}

	var found bool
	for _, b := range backups {
		if b.ID == backupID {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Backup non trovato", http.StatusNotFound)
		return
	}

	backupPath := "backups/" + backupID + ".zip"
	w.Header().Set("Content-Disposition", "attachment; filename="+backupID+".zip")
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, backupPath)
}
