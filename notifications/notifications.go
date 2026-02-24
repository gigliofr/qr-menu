package notifications

import (
	"fmt"
	"sync"
	"time"

	"qr-menu/logger"
)

// NotificationManager gestisce le notifiche push del sistema
type NotificationManager struct {
	mu                 sync.Mutex
	notificationQueue  chan *Notification
	queueSize          int
	isRunning          bool
	maxRetries         int
	retryDelay         time.Duration
	userPreferences    map[string]NotificationPreferences // restaurantID -> preferences
	notificationHistory map[string][]NotificationRecord   // restaurantID -> history
}

// Notification rappresenta una singola notificazione
type Notification struct {
	ID            string                 `json:"id"`
	RestaurantID  string                 `json:"restaurant_id"`
	Type          string                 `json:"type"` // order, reservation, promotion, alert, system
	Title         string                 `json:"title"`
	Body          string                 `json:"body"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Priority      string                 `json:"priority"` // high, normal, low
	CreatedAt     time.Time              `json:"created_at"`
	ExpiresAt     time.Time              `json:"expires_at,omitempty"`
	RetryCount    int                    `json:"-"`
	Status        string                 `json:"status"` // pending, sent, failed, expired
	FCMToken      string                 `json:"-"` // Device token per Firebase
	ImageURL      string                 `json:"image_url,omitempty"`
	ActionURL     string                 `json:"action_url,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
}

// NotificationRecord è una notificazione salvata in history
type NotificationRecord struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	Status    string                 `json:"status"`
}

// NotificationPreferences contiene le preferenze di notificazione dell'utente
type NotificationPreferences struct {
	RestaurantID        string    `json:"restaurant_id"`
	EnablePush          bool      `json:"enable_push"`
	EnableEmail         bool      `json:"enable_email"`
	EnableSMS           bool      `json:"enable_sms"`
	OrderNotifications  bool      `json:"order_notifications"`
	ReservationNotif    bool      `json:"reservation_notifications"`
	PromoNotifications  bool      `json:"promo_notifications"`
	AlertNotifications  bool      `json:"alert_notifications"`
	SystemNotifications bool      `json:"system_notifications"`
	QuietHoursStart     string    `json:"quiet_hours_start"` // HH:MM
	QuietHoursEnd       string    `json:"quiet_hours_end"`   // HH:MM
	LastUpdated         time.Time `json:"last_updated"`
	FCMTokens           []string  `json:"fcm_tokens"` // Multiple devices
}

// NotificationTemplate è un template per le notifiche
type NotificationTemplate struct {
	Type        string
	Title       string
	BodyTemplate string // Supporta {{placeholder}}
	Priority    string
	ImageURL    string
}

// FCMResponse è la risposta di Firebase Cloud Messaging
type FCMResponse struct {
	Name             string `json:"name"`
	Error            string `json:"error,omitempty"`
	ErrorCode        int    `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

var (
	defaultManager *NotificationManager
	once           sync.Once

	// Notification templates
	templates = map[string]NotificationTemplate{
		"order_created": {
			Type:        "order",
			Title:       "Nuovo Ordine",
			BodyTemplate: "Hai ricevuto un nuovo ordine da {{customer}}",
			Priority:    "high",
		},
		"order_ready": {
			Type:        "order",
			Title:       "Ordine Pronto",
			BodyTemplate: "L'ordine #{{order_id}} è pronto per il ritiro",
			Priority:    "high",
		},
		"reservation_confirmed": {
			Type:        "reservation",
			Title:       "Prenotazione Confermata",
			BodyTemplate: "Prenotazione per {{party_size}} persone il {{date}} alle {{time}}",
			Priority:    "normal",
		},
		"promotion_active": {
			Type:        "promotion",
			Title:       "Promozione Attiva",
			BodyTemplate: "{{promotion_name}} - {{discount}}% di sconto",
			Priority:    "normal",
		},
		"system_alert": {
			Type:        "alert",
			Title:       "Avviso di Sistema",
			BodyTemplate: "{{message}}",
			Priority:    "high",
		},
	}
)

// GetNotificationManager restituisce il singleton NotificationManager
func GetNotificationManager() *NotificationManager {
	once.Do(func() {
		defaultManager = &NotificationManager{
			queueSize:           100,
			maxRetries:          3,
			retryDelay:          5 * time.Second,
			userPreferences:     make(map[string]NotificationPreferences),
			notificationHistory: make(map[string][]NotificationRecord),
		}
	})
	return defaultManager
}

// Init inizializza il notification manager
func (nm *NotificationManager) Init(queueSize int) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.queueSize = queueSize
	nm.notificationQueue = make(chan *Notification, queueSize)

	logger.Info("Notification manager inizializzato", map[string]interface{}{
		"queue_size": queueSize,
		"max_retries": nm.maxRetries,
	})

	return nil
}

// Start avvia il notification worker
func (nm *NotificationManager) Start(numWorkers int) error {
	nm.mu.Lock()
	if nm.isRunning {
		nm.mu.Unlock()
		return fmt.Errorf("notification manager già in esecuzione")
	}
	nm.isRunning = true
	nm.mu.Unlock()

	// Avvia worker goroutine
	for i := 0; i < numWorkers; i++ {
		go nm.notificationWorker(i)
	}

	logger.Info("Notification workers avviati", map[string]interface{}{
		"workers": numWorkers,
	})

	return nil
}

// notificationWorker processa le notifiche dalla queue
func (nm *NotificationManager) notificationWorker(id int) {
	for notification := range nm.notificationQueue {
		err := nm.sendNotification(notification)
		if err != nil && notification.RetryCount < nm.maxRetries {
			notification.RetryCount++
			logger.Warn("Retry notifica", map[string]interface{}{
				"notification_id": notification.ID,
				"retry_count":     notification.RetryCount,
				"error":           err.Error(),
			})
			time.Sleep(nm.retryDelay)
			// Reinserisci in queue
			nm.QueueNotification(notification)
		} else if err != nil {
			notification.Status = "failed"
			logger.Error("Notifica fallita", map[string]interface{}{
				"notification_id": notification.ID,
				"error":           err.Error(),
			})
		} else {
			notification.Status = "sent"
			logger.Info("Notifica inviata", map[string]interface{}{
				"notification_id": notification.ID,
				"restaurant_id":   notification.RestaurantID,
				"type":            notification.Type,
			})
		}

		// Salva in history
		nm.addToHistory(notification)
	}
}

// sendNotification invia una singola notificazione via FCM
func (nm *NotificationManager) sendNotification(notif *Notification) error {
	if notif.FCMToken == "" {
		return fmt.Errorf("FCM token non disponibile per il restaurante %s", notif.RestaurantID)
	}

	// Verifica preferenze utente
	prefs, exists := nm.userPreferences[notif.RestaurantID]
	if !exists {
		prefs = nm.getDefaultPreferences(notif.RestaurantID)
	}

	// Controlla se in quiet hours
	if nm.isInQuietHours(prefs) && notif.Priority != "high" {
		logger.Warn("Notifica silenziata (quiet hours)", map[string]interface{}{
			"notification_id": notif.ID,
		})
		return nil
	}

	// Controlla preferenze per tipo notifica
	if !nm.isNotificationTypeEnabled(notif.Type, prefs) {
		logger.Warn("Tipo di notifica disabilitato", map[string]interface{}{
			"notification_id": notif.ID,
			"type":            notif.Type,
		})
		return nil
	}

	// TODO: Implementare invio reale via Firebase Cloud Messaging
	// Per ora, simuliamo l'invio
	logger.Info("Invio notifica (simulato)", map[string]interface{}{
		"fcm_token": notif.FCMToken[:20] + "...",
		"title":     notif.Title,
		"body":      notif.Body,
	})

	return nil
}

// QueueNotification aggiunge una notificazione alla queue
func (nm *NotificationManager) QueueNotification(notif *Notification) error {
	if notif.ID == "" {
		notif.ID = fmt.Sprintf("notif-%d", time.Now().UnixNano())
	}
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = time.Now()
	}
	if notif.Status == "" {
		notif.Status = "pending"
	}

	select {
	case nm.notificationQueue <- notif:
		return nil
	default:
		return fmt.Errorf("notification queue piena")
	}
}

// SendNotification crea e queua una notificazione
func (nm *NotificationManager) SendNotification(restaurantID string, notificationType string, title string, body string, data map[string]interface{}) error {
	if restaurantID == "" {
		return fmt.Errorf("restaurant_id richiesto")
	}

	// Prendi i token FCM dell'utente
	prefs, exists := nm.userPreferences[restaurantID]
	if !exists || len(prefs.FCMTokens) == 0 {
		logger.Warn("Nessun FCM token disponibile", map[string]interface{}{
			"restaurant_id": restaurantID,
		})
		return nil
	}

	// Crea una notificazione per ogni token
	for _, token := range prefs.FCMTokens {
		notif := &Notification{
			ID:           fmt.Sprintf("notif-%d", time.Now().UnixNano()),
			RestaurantID: restaurantID,
			Type:         notificationType,
			Title:        title,
			Body:         body,
			Data:         data,
			FCMToken:     token,
			Priority:     "normal",
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
			Status:       "pending",
		}

		if err := nm.QueueNotification(notif); err != nil {
			logger.Error("Errore queueing notifica", map[string]interface{}{
				"restaurant_id": restaurantID,
				"error":         err.Error(),
			})
			return err
		}
	}

	return nil
}

// UpdatePreferences aggiorna le preferenze di notificazione
func (nm *NotificationManager) UpdatePreferences(restaurantID string, prefs NotificationPreferences) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	prefs.RestaurantID = restaurantID
	prefs.LastUpdated = time.Now()
	nm.userPreferences[restaurantID] = prefs

	logger.Info("Preferenze notifiche aggiornate", map[string]interface{}{
		"restaurant_id":      restaurantID,
		"enable_push":        prefs.EnablePush,
		"fcm_tokens_count":   len(prefs.FCMTokens),
	})

	return nil
}

// GetPreferences recupera le preferenze dell'utente
func (nm *NotificationManager) GetPreferences(restaurantID string) NotificationPreferences {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if prefs, exists := nm.userPreferences[restaurantID]; exists {
		return prefs
	}

	return nm.getDefaultPreferences(restaurantID)
}

// getDefaultPreferences restituisce le preferenze predefinite
func (nm *NotificationManager) getDefaultPreferences(restaurantID string) NotificationPreferences {
	return NotificationPreferences{
		RestaurantID:        restaurantID,
		EnablePush:          true,
		EnableEmail:         true,
		EnableSMS:           false,
		OrderNotifications:  true,
		ReservationNotif:    true,
		PromoNotifications:  true,
		AlertNotifications:  true,
		SystemNotifications: true,
		QuietHoursStart:     "22:00",
		QuietHoursEnd:       "08:00",
		LastUpdated:         time.Now(),
		FCMTokens:           []string{},
	}
}

// RegisterFCMToken registra un nuovo token FCM
func (nm *NotificationManager) RegisterFCMToken(restaurantID string, token string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	prefs, exists := nm.userPreferences[restaurantID]
	if !exists {
		prefs = nm.getDefaultPreferences(restaurantID)
	}

	// Verifica se il token esiste già
	for _, t := range prefs.FCMTokens {
		if t == token {
			return nil
		}
	}

	prefs.FCMTokens = append(prefs.FCMTokens, token)
	prefs.LastUpdated = time.Now()
	nm.userPreferences[restaurantID] = prefs

	logger.Info("Token FCM registrato", map[string]interface{}{
		"restaurant_id": restaurantID,
		"token":         token[:20] + "...",
	})

	return nil
}

// RemoveFCMToken rimuove un token FCM
func (nm *NotificationManager) RemoveFCMToken(restaurantID string, token string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	prefs, exists := nm.userPreferences[restaurantID]
	if !exists {
		return fmt.Errorf("preferenze non trovate")
	}

	// Rimuovi il token
	newTokens := []string{}
	for _, t := range prefs.FCMTokens {
		if t != token {
			newTokens = append(newTokens, t)
		}
	}

	prefs.FCMTokens = newTokens
	prefs.LastUpdated = time.Now()
	nm.userPreferences[restaurantID] = prefs

	return nil
}

// GetHistory recupera la cronologia delle notifiche
func (nm *NotificationManager) GetHistory(restaurantID string, limit int) []NotificationRecord {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	records, exists := nm.notificationHistory[restaurantID]
	if !exists || len(records) == 0 {
		return []NotificationRecord{}
	}

	// Ordina per data decrescente (più recenti prima)
	if len(records) > limit {
		return records[len(records)-limit:]
	}

	return records
}

// MarkAsRead marca una notificazione come letta
func (nm *NotificationManager) MarkAsRead(restaurantID string, notificationID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	records, exists := nm.notificationHistory[restaurantID]
	if !exists {
		return fmt.Errorf("cronologia non trovata")
	}

	for i := range records {
		if records[i].ID == notificationID {
			now := time.Now()
			records[i].ReadAt = &now
			break
		}
	}

	nm.notificationHistory[restaurantID] = records
	return nil
}

// addToHistory aggiunge una notificazione alla cronologia
func (nm *NotificationManager) addToHistory(notif *Notification) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	record := NotificationRecord{
		ID:        notif.ID,
		Type:      notif.Type,
		Title:     notif.Title,
		Body:      notif.Body,
		Data:      notif.Data,
		CreatedAt: notif.CreatedAt,
		Status:    notif.Status,
	}

	nm.notificationHistory[notif.RestaurantID] = append(nm.notificationHistory[notif.RestaurantID], record)

	// Mantieni solo gli ultimi 500 record per restaurante
	history := nm.notificationHistory[notif.RestaurantID]
	if len(history) > 500 {
		nm.notificationHistory[notif.RestaurantID] = history[len(history)-500:]
	}
}

// isInQuietHours verifica se siamo nelle quiet hours
func (nm *NotificationManager) isInQuietHours(prefs NotificationPreferences) bool {
	if prefs.QuietHoursStart == "" || prefs.QuietHoursEnd == "" {
		return false
	}

	now := time.Now()
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	// Se end < start, significa che le quiet hours attraversano mezzanotte
	if prefs.QuietHoursEnd < prefs.QuietHoursStart {
		return currentTime >= prefs.QuietHoursStart || currentTime < prefs.QuietHoursEnd
	}

	return currentTime >= prefs.QuietHoursStart && currentTime < prefs.QuietHoursEnd
}

// isNotificationTypeEnabled verifica se il tipo di notifica è abilitato
func (nm *NotificationManager) isNotificationTypeEnabled(notifType string, prefs NotificationPreferences) bool {
	switch notifType {
	case "order":
		return prefs.OrderNotifications
	case "reservation":
		return prefs.ReservationNotif
	case "promotion":
		return prefs.PromoNotifications
	case "alert":
		return prefs.AlertNotifications
	case "system":
		return prefs.SystemNotifications
	default:
		return true
	}
}

// Stop ferma il notification manager
func (nm *NotificationManager) Stop() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.isRunning {
		nm.isRunning = false
		close(nm.notificationQueue)
		logger.Info("Notification manager fermato", nil)
	}
}

// GetStats restituisce statistiche del sistema di notifiche
func (nm *NotificationManager) GetStats() map[string]interface{} {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	totalUsers := len(nm.userPreferences)
	totalHistory := 0
	for _, records := range nm.notificationHistory {
		totalHistory += len(records)
	}

	return map[string]interface{}{
		"total_users":     totalUsers,
		"total_history":   totalHistory,
		"queue_size":      cap(nm.notificationQueue),
		"queue_used":      len(nm.notificationQueue),
		"is_running":      nm.isRunning,
	}
}
