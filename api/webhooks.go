package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"qr-menu/logger"
	"qr-menu/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	EventBillingSubscriptionUpdated  = "billing.subscription.updated"
	EventBillingSubscriptionCanceled = "billing.subscription.canceled"
	EventWebhookTest                 = "webhook.test"
)

var (
	webhookMu         sync.RWMutex
	webhookEndpoints  = map[string]*models.WebhookEndpoint{}
	webhookDeliveries = map[string]*models.WebhookDelivery{}
)

type createWebhookRequest struct {
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Secret  string   `json:"secret"`
	Active  *bool    `json:"active"`
}

type webhookPayload struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	CreatedAt string      `json:"created_at"`
}

// CreateWebhookHandler registers a new webhook.
func CreateWebhookHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "JSON non valido", err.Error())
		return
	}

	endpointURL := strings.TrimSpace(req.URL)
	if err := validateWebhookURL(endpointURL); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_URL", "URL non valido", err.Error())
		return
	}

	events := normalizeEvents(req.Events)
	if len(events) == 0 {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_EVENTS", "Eventi non validi", "Specificare almeno un evento")
		return
	}

	secret := strings.TrimSpace(req.Secret)
	if secret == "" {
		secret = generateWebhookSecret()
	}

	isActive := true
	if req.Active != nil {
		isActive = *req.Active
	}

	now := time.Now()
	endpoint := &models.WebhookEndpoint{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		URL:          endpointURL,
		Events:       events,
		Secret:       secret,
		IsActive:     isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	webhookMu.Lock()
	webhookEndpoints[endpoint.ID] = endpoint
	webhookMu.Unlock()

	CreatedResponse(w, endpoint)
}

// ListWebhooksHandler returns all webhooks for the restaurant.
func ListWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	webhookMu.RLock()
	defer webhookMu.RUnlock()

	endpoints := make([]*models.WebhookEndpoint, 0)
	for _, endpoint := range webhookEndpoints {
		if endpoint.RestaurantID == restaurantID {
			endpoints = append(endpoints, endpoint)
		}
	}

	SuccessResponse(w, endpoints, nil)
}

// DeleteWebhookHandler removes a webhook.
func DeleteWebhookHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	webhookID := mux.Vars(r)["id"]

	webhookMu.Lock()
	defer webhookMu.Unlock()

	endpoint, ok := webhookEndpoints[webhookID]
	if !ok || endpoint.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusNotFound, "WEBHOOK_NOT_FOUND", "Webhook non trovato", "")
		return
	}

	delete(webhookEndpoints, webhookID)
	SuccessResponse(w, map[string]string{"status": "deleted"}, nil)
}

// TestWebhookHandler triggers a test event.
func TestWebhookHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	webhookID := mux.Vars(r)["id"]

	webhookMu.RLock()
	endpoint, ok := webhookEndpoints[webhookID]
	webhookMu.RUnlock()

	if !ok || endpoint.RestaurantID != restaurantID {
		ErrorResponse(w, http.StatusNotFound, "WEBHOOK_NOT_FOUND", "Webhook non trovato", "")
		return
	}

	EmitEvent(restaurantID, EventWebhookTest, map[string]interface{}{
		"message": "Webhook test event",
		"webhook_id": webhookID,
	})

	SuccessResponse(w, map[string]string{"status": "queued"}, nil)
}

// ListWebhookDeliveriesHandler returns deliveries for the restaurant.
func ListWebhookDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	limit := 50
	if val := r.URL.Query().Get("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	webhookMu.RLock()
	defer webhookMu.RUnlock()

	items := make([]*models.WebhookDelivery, 0)
	for _, delivery := range webhookDeliveries {
		if delivery.RestaurantID == restaurantID {
			items = append(items, delivery)
		}
	}

	if len(items) > limit {
		items = items[:limit]
	}

	SuccessResponse(w, items, nil)
}

// EmitEvent sends an event to matching webhook endpoints.
func EmitEvent(restaurantID, eventType string, data interface{}) {
	event := &models.WebhookEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		CreatedAt: time.Now(),
	}

	endpoints := collectEndpointsForEvent(restaurantID, eventType)
	for _, endpoint := range endpoints {
		go deliverEvent(endpoint, event)
	}
}

func collectEndpointsForEvent(restaurantID, eventType string) []*models.WebhookEndpoint {
	webhookMu.RLock()
	defer webhookMu.RUnlock()

	var endpoints []*models.WebhookEndpoint
	for _, endpoint := range webhookEndpoints {
		if endpoint.RestaurantID != restaurantID || !endpoint.IsActive {
			continue
		}
		if endpointHasEvent(endpoint, eventType) {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func endpointHasEvent(endpoint *models.WebhookEndpoint, eventType string) bool {
	for _, event := range endpoint.Events {
		if event == "*" || event == eventType {
			return true
		}
	}
	return false
}

func deliverEvent(endpoint *models.WebhookEndpoint, event *models.WebhookEvent) {
	payload := webhookPayload{
		ID:        event.ID,
		Type:      event.Type,
		Data:      event.Data,
		CreatedAt: event.CreatedAt.UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Webhook payload marshal failed", map[string]interface{}{"error": err.Error()})
		return
	}

	attempt := 1
	deliverWithRetry(endpoint, event.Type, body, attempt)
}

func deliverWithRetry(endpoint *models.WebhookEndpoint, eventType string, body []byte, attempt int) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := signWebhookPayload(endpoint.Secret, timestamp, body)

	req, err := http.NewRequest(http.MethodPost, endpoint.URL, bytes.NewReader(body))
	if err != nil {
		recordDelivery(endpoint, eventType, "failed", attempt, err.Error(), time.Time{})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Id", endpoint.ID)
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Webhook-Signature", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		scheduleRetry(endpoint, eventType, body, attempt, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		recordDelivery(endpoint, eventType, "success", attempt, "", time.Time{})
		return
	}

	scheduleRetry(endpoint, eventType, body, attempt, fmt.Sprintf("status %d", resp.StatusCode))
}

func scheduleRetry(endpoint *models.WebhookEndpoint, eventType string, body []byte, attempt int, errMsg string) {
	backoff := []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute}
	if attempt > len(backoff) {
		recordDelivery(endpoint, eventType, "failed", attempt, errMsg, time.Time{})
		return
	}

	nextRetry := time.Now().Add(backoff[attempt-1])
	recordDelivery(endpoint, eventType, "retrying", attempt, errMsg, nextRetry)

	time.AfterFunc(backoff[attempt-1], func() {
		deliverWithRetry(endpoint, eventType, body, attempt+1)
	})
}

func recordDelivery(endpoint *models.WebhookEndpoint, eventType, status string, attempt int, errMsg string, nextRetry time.Time) {
	delivery := &models.WebhookDelivery{
		ID:           uuid.New().String(),
		WebhookID:    endpoint.ID,
		RestaurantID: endpoint.RestaurantID,
		EventType:    eventType,
		Status:       status,
		Attempt:      attempt,
		LastError:    errMsg,
		NextRetryAt:  nextRetry,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	webhookMu.Lock()
	webhookDeliveries[delivery.ID] = delivery
	webhookMu.Unlock()

	if status == "failed" {
		logger.SecurityEvent("WEBHOOK_DELIVERY_FAILED", "Webhook delivery failed", endpoint.RestaurantID, "", "", map[string]interface{}{
			"webhook_id": endpoint.ID,
			"event":      eventType,
			"attempt":    attempt,
			"error":      errMsg,
		})
	}
}

func validateWebhookURL(value string) error {
	if value == "" {
		return errors.New("URL richiesta")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("schema non valido")
	}
	if parsed.Host == "" {
		return errors.New("host mancante")
	}
	return nil
}

func normalizeEvents(events []string) []string {
	result := make([]string, 0)
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event == "" {
			continue
		}
		result = append(result, event)
	}
	return result
}

func generateWebhookSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return uuid.New().String()
	}
	return hex.EncodeToString(buf)
}

func signWebhookPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
