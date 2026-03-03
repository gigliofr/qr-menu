package api

import (
	"encoding/json"
	"net/http"
	"qr-menu/security"
	"time"

	"github.com/gorilla/mux"
)

// SetupAPIRoutes configura tutte le route API
func SetupAPIRoutes(r *mux.Router) {
	// Seed test data (una sola volta)
	if len(apiRestaurants) == 0 {
		SeedTestData()
	}

	// Sottoruter per le API con prefisso /api/v1
	api := r.PathPrefix("/api/v1").Subrouter()

	// Rate limiting per tutte le API (100 richieste per minuto)
	rateLimiter := RateLimitMiddleware(100)

	// Authentication endpoints (non richiedono autenticazione)
	api.HandleFunc("/auth/login", rateLimiter(APILoginHandler)).Methods("POST")
	api.HandleFunc("/auth/register", rateLimiter(APIRegisterHandler)).Methods("POST")
	api.HandleFunc("/auth/refresh", rateLimiter(APIRefreshTokenHandler)).Methods("POST")

	// Protected routes (richiedono autenticazione JWT)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return rateLimiter(func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	})

	// Authentication protected endpoints
	protected.HandleFunc("/auth/logout", APILogoutHandler).Methods("POST")
	protected.HandleFunc("/auth/change-password", ChangePasswordHandler).Methods("POST")

	// Restaurant endpoints
	protected.HandleFunc("/restaurant/profile", RequirePermissions(PermRestaurantRead)(GetRestaurantProfileHandler)).Methods("GET")
	protected.HandleFunc("/restaurant/profile", RequirePermissions(PermRestaurantWrite)(UpdateRestaurantProfileHandler)).Methods("PUT")

	// Menu endpoints
	protected.HandleFunc("/menus", RequirePermissions(PermMenusRead)(GetMenusHandler)).Methods("GET")
	protected.HandleFunc("/menus", RequirePermissions(PermMenusWrite)(CreateMenuHandler)).Methods("POST")
	protected.HandleFunc("/menus/{id}", RequirePermissions(PermMenusRead)(GetMenuHandler)).Methods("GET")
	protected.HandleFunc("/menus/{id}", RequirePermissions(PermMenusWrite)(UpdateMenuHandler)).Methods("PUT")
	protected.HandleFunc("/menus/{id}", RequirePermissions(PermMenusDelete)(DeleteMenuHandler)).Methods("DELETE")
	protected.HandleFunc("/menus/{id}/activate", RequirePermissions(PermMenusActivate)(SetActiveMenuHandler)).Methods("POST")

	// Category endpoints
	protected.HandleFunc("/menus/{id}/categories", RequirePermissions(PermMenusWrite)(AddCategoryHandler)).Methods("POST")

	// Item endpoints
	protected.HandleFunc("/menus/{menu_id}/categories/{category_id}/items", RequirePermissions(PermMenusWrite)(AddItemHandler)).Methods("POST")

	// Billing endpoints
	protected.HandleFunc("/billing/plans", RequirePermissions(PermBillingRead)(GetBillingPlansHandler)).Methods("GET")
	protected.HandleFunc("/billing/subscription", RequirePermissions(PermBillingRead)(GetSubscriptionHandler)).Methods("GET")
	protected.HandleFunc("/billing/subscription", RequirePermissions(PermBillingWrite)(CreateSubscriptionHandler)).Methods("POST")
	protected.HandleFunc("/billing/subscription/cancel", RequirePermissions(PermBillingWrite)(CancelSubscriptionHandler)).Methods("POST")
	protected.HandleFunc("/billing/portal", RequirePermissions(PermBillingWrite)(CreateBillingPortalHandler)).Methods("POST")

	// Webhook endpoints
	protected.HandleFunc("/webhooks", RequirePermissions(PermWebhooksRead)(ListWebhooksHandler)).Methods("GET")
	protected.HandleFunc("/webhooks", RequirePermissions(PermWebhooksWrite)(CreateWebhookHandler)).Methods("POST")
	protected.HandleFunc("/webhooks/{id}", RequirePermissions(PermWebhooksWrite)(DeleteWebhookHandler)).Methods("DELETE")
	protected.HandleFunc("/webhooks/{id}/test", RequirePermissions(PermWebhooksDeliver)(TestWebhookHandler)).Methods("POST")
	protected.HandleFunc("/webhooks/deliveries", RequirePermissions(PermWebhooksRead)(ListWebhookDeliveriesHandler)).Methods("GET")

	// ML & Analytics endpoints
	protected.HandleFunc("/ml/recommendations", GetRecommendationsHandler).Methods("GET")
	protected.HandleFunc("/ml/items/{id}/similar", GetSimilarItemsHandler).Methods("GET")
	protected.HandleFunc("/ml/items/trending", GetTrendingItemsHandler).Methods("GET")
	protected.HandleFunc("/ml/interactions", TrackInteractionHandler).Methods("POST")
	protected.HandleFunc("/ml/recommendations/train", TrainRecommendationsHandler).Methods("POST")

	protected.HandleFunc("/ml/forecast", ForecastDemandHandler).Methods("GET")
	protected.HandleFunc("/ml/seasonality", DetectSeasonalityHandler).Methods("GET")
	protected.HandleFunc("/ml/trend", AnalyzeTrendHandler).Methods("GET")
	protected.HandleFunc("/ml/peak-times", PredictPeakTimesHandler).Methods("GET")
	protected.HandleFunc("/ml/inventory/{item_id}/optimize", OptimizeInventoryHandler).Methods("GET")
	protected.HandleFunc("/ml/data-points", AddDataPointHandler).Methods("POST")

	protected.HandleFunc("/ml/experiments", CreateExperimentHandler).Methods("POST")
	protected.HandleFunc("/ml/experiments", ListExperimentsHandler).Methods("GET")
	protected.HandleFunc("/ml/experiments/{id}/start", StartExperimentHandler).Methods("POST")
	protected.HandleFunc("/ml/experiments/{id}/stop", StopExperimentHandler).Methods("POST")
	protected.HandleFunc("/ml/experiments/{id}/results", GetExperimentResultsHandler).Methods("GET")
	protected.HandleFunc("/ml/experiments/{id}/assign", AssignVariantHandler).Methods("POST")
	protected.HandleFunc("/ml/experiments/conversions", TrackConversionHandler).Methods("POST")

	protected.HandleFunc("/ml/stats", GetMLStatsHandler).Methods("GET")

	// Billing webhook (no auth)
	api.HandleFunc("/billing/webhook", BillingWebhookHandler).Methods("POST")

	// API Documentation endpoint
	api.HandleFunc("/docs", APIDocsHandler).Methods("GET")
	api.HandleFunc("/docs/ui", SwaggerUIHandler).Methods("GET")      // Interactive Swagger UI
	api.HandleFunc("/docs/spec.json", APIDocsHandler).Methods("GET") // OpenAPI spec (same as /docs)
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// Public health check alias
	r.HandleFunc("/health", HealthCheckHandler).Methods("GET")
}

// APIDocsHandler fornisce la documentazione API in formato OpenAPI 3.0.0
func APIDocsHandler(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "QR Menu System API",
			"description": "API REST per la gestione di menu digitali con QR code",
			"version":     "2.0.0",
			"contact": map[string]string{
				"name": "QR Menu Support",
			},
		},
		"servers": []map[string]string{
			{
				"url":         "http://localhost:8080/api/v1",
				"description": "Local development server",
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "Usa l'endpoint /auth/login per ottenere il token",
				},
			},
		},
		"security": []map[string]interface{}{
			{
				"BearerAuth": []interface{}{},
			},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Health check",
					"description": "Verifica lo stato della API",
					"security":    []interface{}{},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API è online",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"status":    map[string]interface{}{"type": "string"},
													"version":   map[string]interface{}{"type": "string"},
													"database":  map[string]interface{}{"type": "string"},
													"services":  map[string]interface{}{"type": "object"},
													"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
												},
											},
											"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Accesso",
					"description": "Autentica un ristorante e restituisce un JWT token",
					"security":    []interface{}{},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"username": map[string]interface{}{"type": "string"},
										"password": map[string]interface{}{"type": "string"},
									},
									"required": []string{"username", "password"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
						},
						"400": map[string]interface{}{
							"description": "Invalid credentials",
						},
					},
				},
			},
			"/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Registrazione",
					"description": "Registra un nuovo ristorante",
					"security":    []interface{}{},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"username":        map[string]interface{}{"type": "string"},
										"email":           map[string]interface{}{"type": "string", "format": "email"},
										"password":        map[string]interface{}{"type": "string"},
										"restaurant_name": map[string]interface{}{"type": "string"},
									},
									"required": []string{"username", "email", "password", "restaurant_name"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Registration successful",
						},
					},
				},
			},
			"/menus": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Menus"},
					"summary":     "Lista menu",
					"description": "Ottieni tutti i menu del ristorante",
					"parameters": []map[string]interface{}{
						{
							"name":     "page",
							"in":       "query",
							"required": false,
							"schema":   map[string]interface{}{"type": "integer", "default": 1},
						},
						{
							"name":     "per_page",
							"in":       "query",
							"required": false,
							"schema":   map[string]interface{}{"type": "integer", "default": 20},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Menu list retrieved successfully",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{"Menus"},
					"summary":     "Crea menu",
					"description": "Crea un nuovo menu",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"name":        map[string]interface{}{"type": "string"},
										"description": map[string]interface{}{"type": "string"},
									},
									"required": []string{"name"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Menu created successfully",
						},
					},
				},
			},
			"/menus/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Menus"},
					"summary":     "Ottieni menu",
					"description": "Ottieni un menu specifico",
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Menu ID",
							"schema":      map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Menu retrieved successfully",
						},
						"404": map[string]interface{}{
							"description": "Menu not found",
						},
					},
				},
				"put": map[string]interface{}{
					"tags":        []string{"Menus"},
					"summary":     "Aggiorna menu",
					"description": "Aggiorna un menu esistente",
					"parameters": []map[string]interface{}{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema":   map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Menu updated successfully",
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":        []string{"Menus"},
					"summary":     "Elimina menu",
					"description": "Elimina un menu",
					"parameters": []map[string]interface{}{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema":   map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"204": map[string]interface{}{
							"description": "Menu deleted successfully",
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(spec)
}

// HealthCheckHandler verifica lo stato dell'API
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.0.0",
		"uptime":    "N/A",       // Implementare se necessario
		"database":  "in-memory", // Cambiare quando si usa database reale
		"services": map[string]string{
			"authentication": "running",
			"logging":        "running",
			"rate_limiting":  "running",
		},
		"stats": map[string]interface{}{
			"restaurants":   len(apiRestaurants),
			"menus":         len(apiMenus),
			"active_tokens": len(revokedTokens), // Numero token revocati
		},
	}

	SuccessResponse(w, health, nil)
}

// SwaggerUIHandler fornisce l'interfaccia interattiva per testare le API
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `<!DOCTYPE html>
<html>
<head>
    <title>QR Menu System API - Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 20px;
            text-align: center;
        }
        header h1 {
            margin: 0;
            font-size: 32px;
        }
        header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
            font-size: 16px;
        }
        .content {
            padding: 30px;
        }
        .endpoint {
            margin-bottom: 30px;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 4px;
            border-left: 4px solid #667eea;
        }
        .method {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 12px;
            margin-right: 10px;
        }
        .get { background: #61affe; color: white; }
        .post { background: #49cc90; color: white; }
        .put { background: #fca130; color: white; }
        .delete { background: #f93e3e; color: white; }
        .path {
            font-family: monospace;
            font-weight: bold;
            font-size: 14px;
            color: #333;
        }
        .description {
            margin: 10px 0;
            color: #666;
        }
        .button {
            display: inline-block;
            margin-top: 15px;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-size: 14px;
        }
        .button:hover {
            background: #764ba2;
        }
        code {
            background: #f0f0f0;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🍽️ QR Menu System API</h1>
            <p>v2.0.0 - Digital Menu Management with QR Codes</p>
        </header>
        <div class="content">
            <h2>📚 API Endpoints</h2>
            <p>Base URL: <code>http://localhost:8080/api/v1</code></p>
            
            <h3>🔐 Authentication</h3>
            <p>Most endpoints require a JWT token. Send it in the header: <code>Authorization: Bearer &lt;token&gt;</code></p>

            <h3>Endpoints</h3>
            
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/health</span>
                <p class="description">Check API health status</p>
            </div>

            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/auth/login</span>
                <p class="description">Authenticate and get JWT token</p>
                <p><strong>Body:</strong></p>
                <pre>{ "username": "user", "password": "pass" }</pre>
            </div>

            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/auth/register</span>
                <p class="description">Register a new restaurant</p>
                <p><strong>Body:</strong></p>
                <pre>{ "username": "user", "email": "user@example.com", "password": "pass", "restaurant_name": "My Restaurant" }</pre>
            </div>

            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/menus</span>
                <p class="description">Get all menus for the restaurant</p>
                <p><strong>Query params:</strong> page, per_page (requires auth)</p>
            </div>

            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/menus</span>
                <p class="description">Create a new menu</p>
                <p><strong>Body:</strong> { "name": "Menu Name", "description": "..." } (requires auth)</p>
            </div>

            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/menus/{id}</span>
                <p class="description">Get a specific menu by ID (requires auth)</p>
            </div>

            <div class="endpoint">
                <span class="method put">PUT</span>
                <span class="path">/menus/{id}</span>
                <p class="description">Update a menu (requires auth)</p>
            </div>

            <div class="endpoint">
                <span class="method delete">DELETE</span>
                <span class="path">/menus/{id}</span>
                <p class="description">Delete a menu (requires auth)</p>
            </div>

            <h3>📋 Full OpenAPI Spec</h3>
            <p>Download or view the complete OpenAPI specification:</p>
            <a href="/api/v1/docs/spec.json" class="button">📥 Download OpenAPI Spec (JSON)</a>
            
            <h3>ℹ️ Response Format</h3>
            <p>All responses follow this format:</p>
            <pre>{
  "success": true,
  "data": { /* response data */ },
  "timestamp": "2026-02-24T13:33:14Z"
}</pre>

            <h3>📊 Rate Limiting</h3>
            <p>100 requests per minute per IP address. Check response headers:</p>
            <ul>
                <li><code>X-RateLimit-Limit</code></li>
                <li><code>X-RateLimit-Remaining</code></li>
                <li><code>X-RateLimit-Reset</code></li>
            </ul>

            <h3>✅ Status Codes</h3>
            <ul>
                <li><strong>200</strong> - OK</li>
                <li><strong>201</strong> - Created</li>
                <li><strong>400</strong> - Bad Request</li>
                <li><strong>401</strong> - Unauthorized</li>
                <li><strong>403</strong> - Forbidden</li>
                <li><strong>404</strong> - Not Found</li>
                <li><strong>429</strong> - Too Many Requests (Rate Limit)</li>
                <li><strong>500</strong> - Server Error</li>
            </ul>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML))
}

// SetupSecurityRoutes configura le route per sicurezza e compliance
func SetupSecurityRoutes(r *mux.Router, auditLogger *security.AuditLogger, gdprMgr *security.GDPRManager) {
	api := r.PathPrefix("/api/v1").Subrouter()

	// GDPR endpoints (richiedono autenticazione)
	api.HandleFunc("/gdpr/my-data", GetMyDataHandler(gdprMgr)).Methods("GET")
	api.HandleFunc("/gdpr/request-deletion", RequestDataDeletionHandler(gdprMgr)).Methods("POST")
	api.HandleFunc("/gdpr/cancel-deletion", CancelDataDeletionHandler(gdprMgr)).Methods("POST")
	api.HandleFunc("/gdpr/deletion-request", GetDeletionRequestHandler(gdprMgr)).Methods("GET")
	api.HandleFunc("/gdpr/consent", RecordConsentHandler(gdprMgr)).Methods("POST")
	api.HandleFunc("/gdpr/consents", GetConsentsHandler(gdprMgr)).Methods("GET")

	// Audit log endpoints
	api.HandleFunc("/audit/logs", GetAuditLogsHandler(auditLogger)).Methods("GET") // Admin only
	api.HandleFunc("/audit/my-logs", GetMyAuditLogsHandler(auditLogger)).Methods("GET")
	api.HandleFunc("/audit/export", ExportAuditLogsHandler(auditLogger)).Methods("GET") // Admin only
}
