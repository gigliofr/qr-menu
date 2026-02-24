package models

import "time"

// WebhookEndpoint represents a configured webhook endpoint.
type WebhookEndpoint struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	URL          string    `json:"url"`
	Events       []string  `json:"events"`
	Secret       string    `json:"secret"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WebhookEvent represents an event emitted by the system.
type WebhookEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
}

// WebhookDelivery tracks a delivery attempt.
type WebhookDelivery struct {
	ID           string    `json:"id"`
	WebhookID    string    `json:"webhook_id"`
	RestaurantID string    `json:"restaurant_id"`
	EventType    string    `json:"event_type"`
	Status       string    `json:"status"` // success, failed, retrying
	Attempt      int       `json:"attempt"`
	LastError    string    `json:"last_error,omitempty"`
	NextRetryAt  time.Time `json:"next_retry_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
