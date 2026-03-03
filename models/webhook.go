package models

import "time"

// WebhookEndpoint represents a configured webhook endpoint.
type WebhookEndpoint struct {
	ID           string    `json:"id" bson:"id"`
	RestaurantID string    `json:"restaurant_id" bson:"restaurant_id"`
	URL          string    `json:"url" bson:"url"`
	Events       []string  `json:"events" bson:"events"`
	Secret       string    `json:"secret" bson:"secret"`
	IsActive     bool      `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

// WebhookEvent represents an event emitted by the system.
type WebhookEvent struct {
	ID        string      `json:"id" bson:"id"`
	Type      string      `json:"type" bson:"type"`
	Data      interface{} `json:"data" bson:"data"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
}

// WebhookDelivery tracks a delivery attempt.
type WebhookDelivery struct {
	ID           string    `json:"id" bson:"id"`
	WebhookID    string    `json:"webhook_id" bson:"webhook_id"`
	RestaurantID string    `json:"restaurant_id" bson:"restaurant_id"`
	EventType    string    `json:"event_type" bson:"event_type"`
	Status       string    `json:"status" bson:"status"` // success, failed, retrying
	Attempt      int       `json:"attempt" bson:"attempt"`
	LastError    string    `json:"last_error,omitempty" bson:"last_error,omitempty"`
	NextRetryAt  time.Time `json:"next_retry_at,omitempty" bson:"next_retry_at,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}
