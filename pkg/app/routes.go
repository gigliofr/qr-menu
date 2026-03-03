package app

import (
	"net/http"
	"qr-menu/api"
	"qr-menu/handlers"
	"qr-menu/middleware"
	"qr-menu/security"

	"github.com/gorilla/mux"
)

// RouteDefinition definisce una singola route
type RouteDefinition struct {
	Path    string
	Handler http.HandlerFunc
	Methods []string
}

// SetupRouter configura tutte le route dell'applicazione
func SetupRouter(services *Services) *mux.Router {
	r := mux.NewRouter()

	// File statici
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.PathPrefix("/qr/").Handler(http.StripPrefix("/qr/", http.FileServer(http.Dir("./static/qrcodes/"))))

	// Middleware stack (ordine importante!)
	r.Use(services.CORSMiddleware.Middleware)
	r.Use(services.SecurityHeaders.Middleware)
	r.Use(services.RateLimiter.RateLimitMiddleware)
	r.Use(security.NewAuditMiddleware(services.AuditLogger).Middleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.SecurityMiddleware)
	r.Use(middleware.AuthMiddleware)

	// Route pubbliche
	setupPublicRoutes(r)

	// Route protette
	setupProtectedRoutes(r)

	// API REST v2
	api.SetupAPIRoutes(r)
	api.SetupSecurityRoutes(r, services.AuditLogger, services.GDPRManager)

	// Route amministrative
	setupAdminRoutes(r)

	return r
}

// registerProtectedRoutes è un helper per registrare route protette con autenticazione
func registerProtectedRoutes(r *mux.Router, routes []RouteDefinition) {
	for _, route := range routes {
		r.HandleFunc(route.Path, handlers.RequireAuth(route.Handler)).Methods(route.Methods...)
	}
}

func setupPublicRoutes(r *mux.Router) {
	// Pagine pubbliche
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")

	// Menu pubblici
	r.HandleFunc("/menu/{id}", handlers.PublicMenuHandler).Methods("GET")
	r.HandleFunc("/r/{username}", handlers.GetActiveMenuHandler).Methods("GET")
	r.HandleFunc("/menu/{id}/share", handlers.ShareMenuHandler).Methods("GET")

	// Analytics tracking
	r.HandleFunc("/api/track/share", handlers.TrackShareHandler).Methods("POST")
}

func setupProtectedRoutes(r *mux.Router) {
	// Dashboard e admin base
	r.HandleFunc("/admin", handlers.RequireAuth(handlers.AdminHandler)).Methods("GET")
	r.HandleFunc("/admin/analytics", handlers.RequireAuth(handlers.AnalyticsDashboardHandler)).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET", "POST")

	// Gestione menu
	menuRoutes := []RouteDefinition{
		{"/admin/menu/create", handlers.CreateMenuHandler, []string{"GET"}},
		{"/admin/menu/create", handlers.CreateMenuPostHandler, []string{"POST"}},
		{"/admin/menu/{id}", handlers.EditMenuHandler, []string{"GET"}},
		{"/admin/menu/{id}/update", handlers.UpdateMenuHandler, []string{"POST"}},
		{"/admin/menu/{id}/complete", handlers.CompleteMenuHandler, []string{"POST"}},
		{"/admin/menu/{id}/activate", handlers.SetActiveMenuHandler, []string{"POST"}},
		{"/admin/menu/{id}/delete", handlers.DeleteMenuHandler, []string{"POST"}},
		{"/admin/menu/{id}/duplicate", handlers.DuplicateMenuHandler, []string{"POST"}},
		{"/admin/menu/{id}/add-item", handlers.AddItemHandler, []string{"POST"}},
	}
	registerProtectedRoutes(r, menuRoutes)

	// Gestione item menu
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/duplicate",
		handlers.RequireAuth(handlers.DuplicateItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/edit",
		handlers.RequireAuth(handlers.EditItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/delete",
		handlers.RequireAuth(handlers.DeleteItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/upload-image",
		handlers.RequireAuth(handlers.UploadItemImageHandler)).Methods("POST")

	// API JSON
	r.HandleFunc("/api/analytics", handlers.RequireAuth(handlers.AnalyticsAPIHandler)).Methods("GET")
	r.HandleFunc("/api/menus", handlers.RequireAuth(handlers.GetMenusHandler)).Methods("GET")
	r.HandleFunc("/api/menu/{id}", handlers.GetMenuHandler).Methods("GET")
	r.HandleFunc("/api/menu", handlers.RequireAuth(handlers.CreateMenuAPIHandler)).Methods("POST")
	r.HandleFunc("/api/menu/{id}/generate-qr", handlers.RequireAuth(handlers.GenerateQRHandler)).Methods("POST")
}

func setupAdminRoutes(r *mux.Router) {
	// Admin routes reserved for future features
}
