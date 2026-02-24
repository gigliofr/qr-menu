package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"qr-menu/security"
)

// GDPR API Handlers

// GetMyDataHandler handles GDPR data export requests
func GetMyDataHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user data from storage (restaurants are users in this system)
		user := apiRestaurants[userID]
		if user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Get user's restaurants
		restaurants := make([]interface{}, 0)
		for _, rest := range apiRestaurants {
			if rest.ID == userID {
				restaurants = append(restaurants, rest)
			}
		}

		// Get user's menus
		menus := make([]interface{}, 0)
		for _, menu := range apiMenus {
			if menu.RestaurantID == userID {
				menus = append(menus, menu)
			}
		}

		// Export as JSON
		exportData, err := gdprMgr.ExportUserDataJSON(userID, user, restaurants, menus, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"my-data.json\"")
		w.Write(exportData)
	}
}

// RequestDataDeletionHandler handles GDPR deletion requests
func RequestDataDeletionHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		deletionReq, err := gdprMgr.RequestDataDeletion(userID, req.Reason)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deletionReq)
	}
}

// CancelDataDeletionHandler handles cancellation of deletion requests
func CancelDataDeletionHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := gdprMgr.CancelDataDeletion(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetDeletionRequestHandler retrieves deletion request status
func GetDeletionRequestHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := gdprMgr.GetDeletionRequest(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(req)
	}
}

// RecordConsentHandler handles consent recording
func RecordConsentHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ConsentType string `json:"consent_type"`
			Granted     bool   `json:"granted"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		record := security.ConsentRecord{
			UserID:      userID,
			ConsentType: req.ConsentType,
			Granted:     req.Granted,
			IPAddress:   r.RemoteAddr,
			UserAgent:   r.UserAgent(),
		}

		err := gdprMgr.RecordConsent(record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)
	}
}

// GetConsentsHandler retrieves user consents
func GetConsentsHandler(gdprMgr *security.GDPRManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		consents := gdprMgr.GetConsents(userID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(consents)
	}
}

// Audit Log API Handlers

// GetAuditLogsHandler retrieves audit logs (admin only)
func GetAuditLogsHandler(auditLogger *security.AuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		action := r.URL.Query().Get("action")
		userID := r.URL.Query().Get("user_id")

		var events []security.AuditEvent
		if action != "" {
			events = auditLogger.GetEventsByAction(action, limit)
		} else if userID != "" {
			events = auditLogger.GetEventsByUser(userID, limit)
		} else {
			events = auditLogger.GetEvents(nil, limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

// GetMyAuditLogsHandler retrieves user's own audit logs
func GetMyAuditLogsHandler(auditLogger *security.AuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		events := auditLogger.GetEventsByUser(userID, limit)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

// ExportAuditLogsHandler exports audit logs as JSON (admin only)
func ExportAuditLogsHandler(auditLogger *security.AuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exportData, err := auditLogger.ExportJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		timestamp := time.Now().Format("2006-01-02")
		filename := "audit-logs-" + timestamp + ".json"

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.Write(exportData)
	}
}
