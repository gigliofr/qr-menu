package models

import "time"

// BillingPlan represents a subscription plan.
type BillingPlan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	PriceCents  int64    `json:"price_cents"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"` // monthly, yearly
	Features    []string `json:"features"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// BillingSubscription represents a restaurant subscription.
type BillingSubscription struct {
	ID                     string    `json:"id"`
	RestaurantID           string    `json:"restaurant_id"`
	PlanID                 string    `json:"plan_id"`
	Status                 string    `json:"status"` // active, canceled, past_due
	Provider               string    `json:"provider"` // stripe, mock
	ProviderSubscriptionID string    `json:"provider_subscription_id,omitempty"`
	ProviderCustomerID     string    `json:"provider_customer_id,omitempty"`
	CurrentPeriodEnd       time.Time `json:"current_period_end"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// BillingPortalSession represents a customer portal session.
type BillingPortalSession struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BillingCheckoutSession represents a checkout session.
type BillingCheckoutSession struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Provider  string    `json:"provider"`
	ExpiresAt time.Time `json:"expires_at"`
}
