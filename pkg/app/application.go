package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"qr-menu/logger"
	"qr-menu/pkg/config"
	"qr-menu/pkg/container"
	"qr-menu/pkg/errors"
	"qr-menu/pkg/routing"
)

// Application represents the main application
type Application struct {
	config    *config.Config
	container *container.ServiceContainer
	router    *routing.Router
	server    *http.Server
}

// NewApplication creates and initializes a new application
func NewApplication(cfg *config.Config) (*Application, error) {
	if cfg == nil {
		return nil, errors.New(
			errors.CodeValidation,
			"Configuration cannot be nil",
			errors.SeverityFatal,
		)
	}

	// Create service container
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		return nil, errors.InitializationError("container", err)
	}

	// Create router
	rtr := routing.NewRouter(cont)
	rtr.SetupRoutes()
	rtr.SetupErrorHandlers()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      rtr.GetMux(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Application{
		config:    cfg,
		container: cont,
		router:    rtr,
		server:    srv,
	}, nil
}

// Start starts the HTTP server
func (a *Application) Start() error {
	logger.Info("Starting QR Menu application", map[string]interface{}{
		"host": a.server.Addr,
		"env":  a.config.Server.Environment,
	})

	// Print configured routes in development
	if a.config.IsDevelopment() {
		routes := a.router.ListRoutes()
		logger.Info("Configured routes", map[string]interface{}{
			"count": len(routes),
		})
	}

	if a.config.Security.EnableHTTPS {
		logger.Info("Starting HTTPS server", nil)
		return a.server.ListenAndServeTLS(
			a.config.Security.CertFile,
			a.config.Security.KeyFile,
		)
	}

	logger.Info("Starting HTTP server", nil)
	return a.server.ListenAndServe()
}

// Stop gracefully shuts down the application
func (a *Application) Stop(ctx context.Context) error {
	logger.Info("Shutting down application", nil)

	// Close HTTP server
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Warn("Error shutting down HTTP server", map[string]interface{}{"error": err.Error()})
	}

	// Close service container
	if err := a.container.Shutdown(ctx); err != nil {
		logger.Warn("Error shutting down services", map[string]interface{}{"error": err.Error()})
	}

	logger.Info("Application stopped", nil)
	return nil
}

// Container returns the service container
func (a *Application) Container() *container.ServiceContainer {
	return a.container
}

// Router returns the router
func (a *Application) Router() *routing.Router {
	return a.router
}

// Config returns the configuration
func (a *Application) Config() *config.Config {
	return a.config
}

// Health returns the health status
func (a *Application) Health() map[string]interface{} {
	return map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"services":  a.container.Health(),
	}
}
