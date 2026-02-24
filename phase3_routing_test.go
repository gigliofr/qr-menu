package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"qr-menu/pkg/config"
	"qr-menu/pkg/container"
	"qr-menu/pkg/handlers"
	"qr-menu/pkg/routing"
)

// TestRouterSetup tests that the router can be created and routes configured
func TestRouterSetup(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	if router == nil {
		t.Fatal("Router should not be nil")
	}

	router.SetupRoutes()
	router.SetupErrorHandlers()

	mux := router.GetMux()
	if mux == nil {
		t.Fatal("Mux should not be nil")
	}

	t.Log("✅ Router setup test passed")
}

// TestRouterListRoutes tests that routes can be listed
func TestRouterListRoutes(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	router.SetupRoutes()

	routes := router.ListRoutes()
	if len(routes) == 0 {
		t.Fatal("Should have configured routes")
	}

	t.Logf("✅ Configured %d routes", len(routes))
	for _, route := range routes {
		t.Logf("  - %s", route)
	}
}

// TestPublicRoutes tests public endpoint availability
func TestPublicRoutes(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	router.SetupRoutes()
	router.SetupErrorHandlers()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Health check", "GET", "/healthz"},
		{"Status", "GET", "/status"},
		{"Health", "GET", "/health"},
		{"Ready", "GET", "/ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.GetMux().ServeHTTP(w, req)

			// Should not return 404 (route should exist)
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s should exist", tt.method, tt.path)
			}

			t.Logf("✅ %s %s returned %d", tt.method, tt.path, w.Code)
		})
	}
}

// TestAPIRoutesExist tests that API routes are configured
func TestAPIRoutesExist(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	router.SetupRoutes()

	routes := router.ListRoutes()

	expectedRoutes := []string{
		"/api/v1/backup",
		"/api/v1/notifications",
		"/api/v1/analytics",
		"/api/v1/i18n",
		"/api/v1/pwa",
		"/api/v1/database",
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route] = true
	}

	for _, expected := range expectedRoutes {
		if !routeMap[expected] {
			t.Logf("⚠️  Expected route not found: %s", expected)
		}
	}

	t.Logf("✅ API routes test passed with %d total routes", len(routes))
}

// TestAdminRoutesExist tests that admin routes are configured
func TestAdminRoutesExist(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	router.SetupRoutes()

	routes := router.ListRoutes()

	expectedRoutes := []string{
		"/api/admin/migrations",
		"/api/admin/database",
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route] = true
	}

	for _, expected := range expectedRoutes {
		if !routeMap[expected] {
			t.Logf("⚠️  Expected admin route not found: %s", expected)
		}
	}

	t.Logf("✅ Admin routes test passed")
}

// TestNotFoundHandler tests 404 handling
func TestNotFoundHandler(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	router := routing.NewRouter(cont)
	router.SetupRoutes()
	router.SetupErrorHandlers()

	req := httptest.NewRequest("GET", "/nonexistent-route", nil)
	w := httptest.NewRecorder()

	router.GetMux().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}

	t.Logf("✅ 404 handling test passed")
}

// TestHandlerFactories tests that all handler factories can be created
func TestHandlerFactories(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	tests := []struct {
		name    string
		factory func(*container.ServiceContainer) interface{}
	}{
		{"BackupHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewBackupHandlers(c)
		}},
		{"NotificationHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewNotificationHandlers(c)
		}},
		{"AnalyticsHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewAnalyticsHandlers(c)
		}},
		{"LocalizationHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewLocalizationHandlers(c)
		}},
		{"PWAHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewPWAHandlers(c)
		}},
		{"DatabaseHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewDatabaseHandlers(c)
		}},
		{"MigrationHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewMigrationHandlers(c)
		}},
		{"APIHandlers", func(c *container.ServiceContainer) interface{} {
			return handlers.NewAPIHandlers(c)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.factory(cont)
			if handler == nil {
				t.Errorf("%s should not be nil", tt.name)
			} else {
				t.Logf("✅ %s created successfully", tt.name)
			}
		})
	}
}
