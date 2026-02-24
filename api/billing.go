package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"qr-menu/logger"
	"qr-menu/models"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
	billingPortalSession "github.com/stripe/stripe-go/v79/billingportal/session"
	checkoutSession "github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/webhook"
)

const (
	stripeProvider = "stripe"
	mockProvider   = "mock"
)

var (
	billingPlans = map[string]*models.BillingPlan{}
	subscriptions = map[string]*models.BillingSubscription{}
)

func init() {
	seedBillingPlans()
}

func seedBillingPlans() {
	now := time.Now()
	billingPlans["free"] = &models.BillingPlan{
		ID:         "free",
		Name:       "Free",
		PriceCents: 0,
		Currency:   "usd",
		Interval:   "monthly",
		Features:   []string{"Up to 1 menu", "Basic analytics", "Email support"},
		IsActive:   true,
		CreatedAt:  now,
	}
	billingPlans["pro"] = &models.BillingPlan{
		ID:         "pro",
		Name:       "Pro",
		PriceCents: 4900,
		Currency:   "usd",
		Interval:   "monthly",
		Features:   []string{"Unlimited menus", "Advanced analytics", "Priority support"},
		IsActive:   true,
		CreatedAt:  now,
	}
	billingPlans["enterprise"] = &models.BillingPlan{
		ID:         "enterprise",
		Name:       "Enterprise",
		PriceCents: 19900,
		Currency:   "usd",
		Interval:   "monthly",
		Features:   []string{"Custom branding", "SLAs", "Dedicated support"},
		IsActive:   true,
		CreatedAt:  now,
	}
}

type createSubscriptionRequest struct {
	PlanID     string `json:"plan_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// GetBillingPlansHandler returns available plans.
func GetBillingPlansHandler(w http.ResponseWriter, r *http.Request) {
	plans := make([]*models.BillingPlan, 0, len(billingPlans))
	for _, plan := range billingPlans {
		if plan.IsActive {
			plans = append(plans, plan)
		}
	}
	SuccessResponse(w, plans, nil)
}

// GetSubscriptionHandler returns the current subscription for the restaurant.
func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	if sub, ok := subscriptions[restaurantID]; ok {
		SuccessResponse(w, sub, nil)
		return
	}

	// Default free subscription
	free := &models.BillingSubscription{
		ID:               uuid.New().String(),
		RestaurantID:     restaurantID,
		PlanID:           "free",
		Status:           "active",
		Provider:         mockProvider,
		CurrentPeriodEnd: time.Now().AddDate(1, 0, 0),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	SuccessResponse(w, free, nil)
}

// CreateSubscriptionHandler creates a checkout session for a plan.
func CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)

	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "JSON non valido", err.Error())
		return
	}

	planID := strings.TrimSpace(req.PlanID)
	plan, ok := billingPlans[planID]
	if !ok || !plan.IsActive {
		ErrorResponse(w, http.StatusBadRequest, "PLAN_NOT_FOUND", "Piano non valido", "Seleziona un piano valido")
		return
	}

	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey == "" {
		checkout := createMockCheckout(restaurantID, planID)
		SuccessResponse(w, checkout, nil)
		return
	}

	stripe.Key = stripeKey
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(plan.Currency),
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String(plan.Interval),
					},
					UnitAmount: stripe.Int64(plan.PriceCents),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(plan.Name),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(restaurantID),
		SuccessURL:        stripe.String(defaultSuccessURL(req.SuccessURL)),
		CancelURL:         stripe.String(defaultCancelURL(req.CancelURL)),
	}

	cs, err := checkoutSession.New(params)
	if err != nil {
		logger.Error("Stripe checkout error", map[string]interface{}{"error": err.Error()})
		ErrorResponse(w, http.StatusBadRequest, "CHECKOUT_FAILED", "Errore nella creazione checkout", err.Error())
		return
	}

	checkout := &models.BillingCheckoutSession{
		ID:        cs.ID,
		URL:       cs.URL,
		Provider:  stripeProvider,
		ExpiresAt: time.Unix(cs.ExpiresAt, 0),
	}

	SuccessResponse(w, checkout, nil)
}

// CancelSubscriptionHandler cancels the active subscription.
func CancelSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	sub, ok := subscriptions[restaurantID]
	if !ok {
		ErrorResponse(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Nessun abbonamento attivo", "")
		return
	}

	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey != "" && sub.Provider == stripeProvider && sub.ProviderSubscriptionID != "" {
		stripe.Key = stripeKey
		_, err := subscription.Cancel(sub.ProviderSubscriptionID, nil)
		if err != nil {
			logger.Error("Stripe cancel error", map[string]interface{}{"error": err.Error()})
			ErrorResponse(w, http.StatusBadRequest, "CANCEL_FAILED", "Errore nella cancellazione", err.Error())
			return
		}
	}

	sub.Status = "canceled"
	sub.UpdatedAt = time.Now()
	subscriptions[restaurantID] = sub
	SuccessResponse(w, sub, nil)
}

// CreateBillingPortalHandler creates a customer portal session.
func CreateBillingPortalHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	sub, ok := subscriptions[restaurantID]
	if !ok || sub.ProviderCustomerID == "" {
		portal := &models.BillingPortalSession{
			URL:       defaultPortalURL(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		SuccessResponse(w, portal, nil)
		return
	}

	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey == "" {
		portal := &models.BillingPortalSession{
			URL:       defaultPortalURL(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		SuccessResponse(w, portal, nil)
		return
	}

	stripe.Key = stripeKey
	params := &stripe.BillingPortalSessionParams{
		Customer: stripe.String(sub.ProviderCustomerID),
		ReturnURL: stripe.String(defaultPortalURL()),
	}

	portalSession, err := billingPortalSession.New(params)
	if err != nil {
		logger.Error("Stripe portal error", map[string]interface{}{"error": err.Error()})
		ErrorResponse(w, http.StatusBadRequest, "PORTAL_FAILED", "Errore nella creazione portale", err.Error())
		return
	}

	portal := &models.BillingPortalSession{
		URL:       portalSession.URL,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	SuccessResponse(w, portal, nil)
}

// BillingWebhookHandler processes Stripe webhooks.
func BillingWebhookHandler(w http.ResponseWriter, r *http.Request) {
	secret := strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_BODY", "Payload non valido", err.Error())
		return
	}

	if secret == "" {
		SuccessResponse(w, map[string]string{"status": "skipped"}, nil)
		return
	}

	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sig, secret)
	if err != nil {
		logger.SecurityEvent("STRIPE_WEBHOOK_INVALID", "Firma webhook non valida", "", getClientIP(r), r.UserAgent(), nil)
		ErrorResponse(w, http.StatusBadRequest, "WEBHOOK_INVALID", "Firma non valida", err.Error())
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var checkout stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &checkout); err == nil {
			restaurantID := checkout.ClientReferenceID
			planID := "pro"
			subscriptionID := ""
			if checkout.Subscription != nil {
				subscriptionID = checkout.Subscription.ID
			}
			customerID := ""
			if checkout.Customer != nil {
				customerID = checkout.Customer.ID
			}
			createOrUpdateSubscription(restaurantID, planID, stripeProvider, subscriptionID, customerID)
		}
	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
			markSubscriptionCanceled(sub.ID)
		}
	}

	SuccessResponse(w, map[string]string{"status": "ok"}, nil)
}

func createOrUpdateSubscription(restaurantID, planID, provider, providerID, customerID string) {
	sub := &models.BillingSubscription{
		ID:                     uuid.New().String(),
		RestaurantID:           restaurantID,
		PlanID:                 planID,
		Status:                 "active",
		Provider:               provider,
		ProviderSubscriptionID: providerID,
		ProviderCustomerID:     customerID,
		CurrentPeriodEnd:       time.Now().AddDate(0, 1, 0),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
	subscriptions[restaurantID] = sub
	EmitEvent(restaurantID, EventBillingSubscriptionUpdated, sub)
}

func markSubscriptionCanceled(providerID string) {
	for restaurantID, sub := range subscriptions {
		if sub.ProviderSubscriptionID == providerID {
			sub.Status = "canceled"
			sub.UpdatedAt = time.Now()
			subscriptions[restaurantID] = sub
			EmitEvent(restaurantID, EventBillingSubscriptionCanceled, sub)
			return
		}
	}
}

func createMockCheckout(restaurantID, planID string) *models.BillingCheckoutSession {
	createOrUpdateSubscription(restaurantID, planID, mockProvider, "mock", "mock")
	return &models.BillingCheckoutSession{
		ID:        uuid.New().String(),
		URL:       fmt.Sprintf("/admin/billing?success=1&plan=%s", planID),
		Provider:  mockProvider,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
}

func defaultSuccessURL(value string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return "http://localhost:3000/admin/billing?success=1"
}

func defaultCancelURL(value string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return "http://localhost:3000/admin/billing?canceled=1"
}

func defaultPortalURL() string {
	return "http://localhost:3000/admin/billing"
}
