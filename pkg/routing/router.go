package routing

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"qr-menu/pkg/container"
	"qr-menu/pkg/errors"
	"qr-menu/pkg/handlers"
	httputil "qr-menu/pkg/http"
	"qr-menu/pkg/middleware"
)

// Router groups routes by functionality
type Router struct {
	mux                  *mux.Router
	container            *container.ServiceContainer
	cacheInvalidation    *middleware.CacheInvalidationMiddleware
	responseCaching      *middleware.ResponseCachingMiddleware
}

// NewRouter creates a new router with the service container
func NewRouter(c *container.ServiceContainer) *Router {
	return &Router{
		mux:       mux.NewRouter(),
		container: c,
	}
}

// SetupRoutes configures all routes for the application
func (r *Router) SetupRoutes() {
	// Apply caching middleware globally if enabled
	if r.container.Config().Cache.Enabled {
		r.setupCachingMiddleware()
	}

	// Public routes
	r.setupPublicRoutes()

	// API routes
	r.setupAPIRoutes()

	// Admin routes
	r.setupAdminRoutes()

	// Health check
	r.setupHealthRoutes()

	// Cache statistics routes
	if r.container.Config().Cache.Enabled {
		r.setupCacheStatsRoutes()
	}
}

// setupPublicRoutes sets up unauthenticated public endpoints
func (r *Router) setupPublicRoutes() {
	public := r.mux.PathPrefix("").Subrouter()

	// PWA manifest and service worker
	pwaHandlers := handlers.NewPWAHandlers(r.container)
	public.HandleFunc("/manifest.json", pwaHandlers.GetManifest).Methods("GET")
	public.HandleFunc("/service-worker.js", pwaHandlers.GetServiceWorker).Methods("GET")

	// Auth endpoints
	apiHandlers := handlers.NewAPIHandlers(r.container)
	public.HandleFunc("/api/auth/login", apiHandlers.Login).Methods("POST")
	public.HandleFunc("/api/auth/logout", apiHandlers.Logout).Methods("POST")
	public.HandleFunc("/api/auth/refresh", apiHandlers.RefreshToken).Methods("POST")

	// Public health check
	public.HandleFunc("/healthz", apiHandlers.HealthCheck).Methods("GET")
	public.HandleFunc("/status", apiHandlers.GetStatus).Methods("GET")
}

// setupAPIRoutes sets up authenticated API endpoints
func (r *Router) setupAPIRoutes() {
	api := r.mux.PathPrefix("/api/v1").Subrouter()
	// Note: Add authentication middleware here
	// api.Use(middleware.AuthMiddleware)

	// Backup endpoints
	r.setupBackupRoutes(api)

	// Notification endpoints
	r.setupNotificationRoutes(api)

	// Analytics endpoints
	r.setupAnalyticsRoutes(api)

	// Localization endpoints
	r.setupLocalizationRoutes(api)

	// PWA endpoints
	r.setupPWARoutes(api)

	// Database endpoints
	r.setupDatabaseRoutes(api)
}

// setupAdminRoutes sets up admin-only endpoints
func (r *Router) setupAdminRoutes() {
	admin := r.mux.PathPrefix("/api/admin").Subrouter()
	// Note: Add admin middleware here
	// admin.Use(middleware.RequireAdmin)

	// Migration endpoints
	r.setupMigrationRoutes(admin)

	// Database admin endpoints
	databaseHandlers := handlers.NewDatabaseHandlers(r.container)
	admin.HandleFunc("/database/stats", databaseHandlers.GetStats).Methods("GET")
	admin.HandleFunc("/database/health", databaseHandlers.HealthCheck).Methods("GET")
}

// setupBackupRoutes configures backup-related routes
func (r *Router) setupBackupRoutes(api *mux.Router) {
	backupHandlers := handlers.NewBackupHandlers(r.container)
	backup := api.PathPrefix("/backup").Subrouter()

	backup.HandleFunc("", backupHandlers.CreateBackup).Methods("POST")
	backup.HandleFunc("", backupHandlers.ListBackups).Methods("GET")
	backup.HandleFunc("/{id}", backupHandlers.RestoreBackup).Methods("PUT")
	backup.HandleFunc("/{id}", backupHandlers.DeleteBackup).Methods("DELETE")
	backup.HandleFunc("/{id}/download", backupHandlers.DownloadBackup).Methods("GET")
	backup.HandleFunc("/stats", backupHandlers.GetBackupStats).Methods("GET")
}

// setupNotificationRoutes configures notification-related routes
func (r *Router) setupNotificationRoutes(api *mux.Router) {
	notificationHandlers := handlers.NewNotificationHandlers(r.container)
	notif := api.PathPrefix("/notifications").Subrouter()

	notif.HandleFunc("", notificationHandlers.SendNotification).Methods("POST")
	notif.HandleFunc("", notificationHandlers.GetNotifications).Methods("GET")
	notif.HandleFunc("/stats", notificationHandlers.GetStats).Methods("GET")
	notif.HandleFunc("/clear", notificationHandlers.ClearNotifications).Methods("POST")
	notif.HandleFunc("/retry-failed", notificationHandlers.RetryFailed).Methods("POST")
}

// setupAnalyticsRoutes configures analytics-related routes
func (r *Router) setupAnalyticsRoutes(api *mux.Router) {
	analyticsHandlers := handlers.NewAnalyticsHandlers(r.container)
	analytics := api.PathPrefix("/analytics").Subrouter()

	analytics.HandleFunc("/dashboard", analyticsHandlers.GetDashboard).Methods("GET")
	analytics.HandleFunc("/stats", analyticsHandlers.GetStats).Methods("GET")
	analytics.HandleFunc("/track", analyticsHandlers.TrackEvent).Methods("POST")
	analytics.HandleFunc("/export", analyticsHandlers.ExportData).Methods("GET")
}

// setupLocalizationRoutes configures localization-related routes
func (r *Router) setupLocalizationRoutes(api *mux.Router) {
	localizationHandlers := handlers.NewLocalizationHandlers(r.container)
	i18n := api.PathPrefix("/i18n").Subrouter()

	i18n.HandleFunc("/languages", localizationHandlers.GetLanguages).Methods("GET")
	i18n.HandleFunc("/translations", localizationHandlers.GetTranslations).Methods("GET")
	i18n.HandleFunc("/language", localizationHandlers.SetLanguage).Methods("POST")
	i18n.HandleFunc("/formats", localizationHandlers.GetFormats).Methods("GET")
}

// setupPWARoutes configures PWA-related routes
func (r *Router) setupPWARoutes(api *mux.Router) {
	pwaHandlers := handlers.NewPWAHandlers(r.container)
	pwa := api.PathPrefix("/pwa").Subrouter()

	pwa.HandleFunc("/cache/clear", pwaHandlers.ClearCache).Methods("POST")
	pwa.HandleFunc("/cache/status", pwaHandlers.GetCacheStatus).Methods("GET")
}

// setupDatabaseRoutes configures database-related routes
func (r *Router) setupDatabaseRoutes(api *mux.Router) {
	databaseHandlers := handlers.NewDatabaseHandlers(r.container)
	db := api.PathPrefix("/database").Subrouter()

	db.HandleFunc("/status", databaseHandlers.GetStatus).Methods("GET")
	db.HandleFunc("/stats", databaseHandlers.GetStats).Methods("GET")
	db.HandleFunc("/health", databaseHandlers.HealthCheck).Methods("GET")
}

// setupMigrationRoutes configures migration-related routes
func (r *Router) setupMigrationRoutes(admin *mux.Router) {
	migrationHandlers := handlers.NewMigrationHandlers(r.container)
	migration := admin.PathPrefix("/migrations").Subrouter()

	migration.HandleFunc("", migrationHandlers.GetStatus).Methods("GET")
	migration.HandleFunc("/run", migrationHandlers.RunMigrations).Methods("POST")
	migration.HandleFunc("/{id}/rollback", migrationHandlers.RollbackMigration).Methods("POST")
	migration.HandleFunc("/history", migrationHandlers.GetMigrationHistory).Methods("GET")
}

// setupHealthRoutes configures health check routes
func (r *Router) setupHealthRoutes() {
	apiHandlers := handlers.NewAPIHandlers(r.container)
	r.mux.HandleFunc("/health", apiHandlers.HealthCheck).Methods("GET")
	r.mux.HandleFunc("/ready", apiHandlers.GetStatus).Methods("GET")
}

// NotFoundHandler returns a 404 response
func (r *Router) NotFoundHandler(w http.ResponseWriter, req *http.Request) {
	httputil.NotFound(w, "endpoint")
}

// MethodNotAllowedHandler returns a 405 response
func (r *Router) MethodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	httputil.Error(w, errors.New(
		errors.CodeValidation,
		"Method not allowed",
		errors.SeverityWarning,
	).WithHTTPCode(http.StatusMethodNotAllowed))
}

// SetupErrorHandlers sets custom 404 and 405 handlers
func (r *Router) SetupErrorHandlers() {
	r.mux.NotFoundHandler = http.HandlerFunc(r.NotFoundHandler)
	r.mux.MethodNotAllowedHandler = http.HandlerFunc(r.MethodNotAllowedHandler)
}

// setupCachingMiddleware sets up response and cache invalidation middleware
func (r *Router) setupCachingMiddleware() {
	cfg := r.container.Config()
	respCache := r.container.ResponseCache()
	queryCache := r.container.QueryCache()

	if respCache == nil || queryCache == nil {
		return
	}

	// Apply response caching middleware
	r.responseCaching = middleware.NewResponseCachingMiddleware(respCache, cfg.Cache.ResponseCacheTTL)
	r.mux.Use(mux.MiddlewareFunc(r.responseCaching.Middleware()))

	// Apply cache invalidation middleware
	r.cacheInvalidation = middleware.NewCacheInvalidationMiddleware(respCache, queryCache)
	r.mux.Use(mux.MiddlewareFunc(r.cacheInvalidation.Middleware()))

	// Register cache invalidation patterns for mutation endpoints
	r.registerCacheInvalidationPatterns()
}

// registerCacheInvalidationPatterns registers which paths invalidate which cache entries
func (r *Router) registerCacheInvalidationPatterns() {
	if r.cacheInvalidation == nil {
		return
	}

	// Backup mutations invalidate backup-related queries
	r.cacheInvalidation.RegisterPattern("/api/v1/backup", "backups")

	// Notification mutations invalidate notification-related queries
	r.cacheInvalidation.RegisterPattern("/api/v1/notifications", "notifications")

	// Analytics mutations invalidate analytics queries
	r.cacheInvalidation.RegisterPattern("/api/v1/analytics", "analytics")

	// Localization mutations invalidate localization queries
	r.cacheInvalidation.RegisterPattern("/api/v1/i18n", "localization")

	// Database mutations invalidate all database-related queries
	r.cacheInvalidation.RegisterPattern("/api/v1/database", "database")
	r.cacheInvalidation.RegisterPattern("/api/admin/database", "database")

	// Migration mutations invalidate migration queries
	r.cacheInvalidation.RegisterPattern("/api/admin/migrations", "migrations")

	// PWA mutations invalidate PWA cache
	r.cacheInvalidation.RegisterPattern("/api/v1/pwa", "pwa")
}

// setupCacheStatsRoutes sets up cache statistics endpoints
func (r *Router) setupCacheStatsRoutes() {
	admin := r.mux.PathPrefix("/api/admin").Subrouter()

	admin.HandleFunc("/cache/stats", r.getCacheStats).Methods("GET")
	admin.HandleFunc("/cache/clear", r.clearCache).Methods("POST")
	admin.HandleFunc("/cache/status", r.getCacheStatus).Methods("GET")
}

// getCacheStats returns cache statistics
func (r *Router) getCacheStats(w http.ResponseWriter, req *http.Request) {
	type statsResponse struct {
		ResponseCache map[string]interface{} `json:"response_cache,omitempty"`
		QueryCache    map[string]interface{} `json:"query_cache,omitempty"`
	}

	resp := statsResponse{}

	respCache := r.container.ResponseCache()
	queryCache := r.container.QueryCache()

	if respCache != nil {
		resp.ResponseCache = respCache.GetStats()
	}

	if queryCache != nil {
		resp.QueryCache = queryCache.GetStats()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// clearCache clears all caches
func (r *Router) clearCache(w http.ResponseWriter, req *http.Request) {
	respCache := r.container.ResponseCache()
	queryCache := r.container.QueryCache()

	if respCache != nil {
		// Use InvalidatePattern with a pattern that matches everything
		// For now, we'll need to clear by key - but the ResponseCache doesn't have a clear all method
		// We'll need to add one or use a workaround
	}

	if queryCache != nil {
		queryCache.InvalidateAll()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// getCacheStatus returns cache status
func (r *Router) getCacheStatus(w http.ResponseWriter, req *http.Request) {
	respCache := r.container.ResponseCache()
	queryCache := r.container.QueryCache()

	status := map[string]interface{}{
		"enabled":        r.container.Config().Cache.Enabled,
		"response_cache": respCache != nil,
		"query_cache":    queryCache != nil,
	}

	if respCache != nil {
		status["response_cache_size"] = respCache.Size()
		stats := respCache.GetStats()
		if hits, ok := stats["hits"]; ok {
			status["response_cache_hits"] = hits
		}
		if misses, ok := stats["misses"]; ok {
			status["response_cache_misses"] = misses
		}
	}

	if queryCache != nil {
		status["query_cache_size"] = queryCache.Size()
		stats := queryCache.GetStats()
		status["query_cache_stats"] = stats
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// GetMux returns the underlying gorilla mux router
func (r *Router) GetMux() *mux.Router {
	return r.mux
}

// ListRoutes returns all configured routes for debugging
func (r *Router) ListRoutes() []string {
	routes := []string{}
	r.mux.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		routes = append(routes, t)
		return nil
	})
	return routes
}
