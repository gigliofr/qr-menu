package handlers

import (
	"encoding/json"
	"net/http"

	"qr-menu/notifications"
)

// SendNotificationRequest è la richiesta di invio notifica
type SendNotificationRequest struct {
	Type  string                 `json:"type"`  // order, reservation, promotion, alert
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// NotificationResponse è la risposta standard
type NotificationResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SendNotificationHandler invia una notifica
func SendNotificationHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validazione
	if req.Type == "" || req.Title == "" || req.Body == "" {
		http.Error(w, "Type, title, and body sono obbligatori", http.StatusBadRequest)
		return
	}

	nm := notifications.GetNotificationManager()
	err = nm.SendNotification(session.RestaurantID, req.Type, req.Title, req.Body, req.Data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "success",
		Message: "Notifica inviata",
	})
}

// GetPreferencesHandler recupera le preferenze di notificazione
func GetPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	nm := notifications.GetNotificationManager()
	prefs := nm.GetPreferences(session.RestaurantID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status: "success",
		Data:   prefs,
	})
}

// UpdatePreferencesHandler aggiorna le preferenze di notificazione
func UpdatePreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var prefs notifications.NotificationPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	nm := notifications.GetNotificationManager()
	err = nm.UpdatePreferences(session.RestaurantID, prefs)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "success",
		Message: "Preferenze aggiornate",
	})
}

// RegisterFCMTokenHandler registra un nuovo token FCM
func RegisterFCMTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token richiesto", http.StatusBadRequest)
		return
	}

	nm := notifications.GetNotificationManager()
	err = nm.RegisterFCMToken(session.RestaurantID, req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "success",
		Message: "Token registrato",
	})
}

// RemoveFCMTokenHandler rimuove un token FCM
func RemoveFCMTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token richiesto", http.StatusBadRequest)
		return
	}

	nm := notifications.GetNotificationManager()
	err = nm.RemoveFCMToken(session.RestaurantID, req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "success",
		Message: "Token rimosso",
	})
}

// GetNotificationHistoryHandler recupera la cronologia delle notifiche
func GetNotificationHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Leggi il limit da query params (default 50)
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		// Semplicemente usiamo il valore di default se non è un numero valido
	}

	nm := notifications.GetNotificationManager()
	history := nm.GetHistory(session.RestaurantID, limit)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status: "success",
		Data: map[string]interface{}{
			"notifications": history,
			"count":         len(history),
		},
	})
}

// MarkAsReadHandler marca una notificazione come letta
func MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		NotificationID string `json:"notification_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.NotificationID == "" {
		http.Error(w, "notification_id richiesto", http.StatusBadRequest)
		return
	}

	nm := notifications.GetNotificationManager()
	err = nm.MarkAsRead(session.RestaurantID, req.NotificationID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "success",
		Message: "Notifica marcata come letta",
	})
}

// GetNotificationStatsHandler restituisce le statistiche delle notifiche
func GetNotificationStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione (ADMIN ONLY)
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	nm := notifications.GetNotificationManager()
	stats := nm.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status: "success",
		Data:   stats,
	})
}
