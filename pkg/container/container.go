package container

import (
	"context"
	"fmt"
	"sync"

	"qr-menu/analytics"
	"qr-menu/logger"
	"qr-menu/pkg/cache"
	"qr-menu/pkg/config"
	"qr-menu/pkg/errors"
)

// ServiceContainer holds all service instances and manages their lifecycle
type ServiceContainer struct {
	config           *config.Config
	analytics        *analytics.Analytics
	responseCache    *cache.ResponseCache
	queryCache       *cache.QueryResultCache
	isInitialized    bool
	mu               sync.RWMutex
	shutdownHandlers []func(ctx context.Context) error
}

// NewServiceContainer creates and initializes a new service container
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	if cfg == nil {
		return nil, errors.New(
			errors.CodeValidation,
			"Configuration cannot be nil",
			errors.SeverityFatal,
		)
	}

	c := &ServiceContainer{
		config:           cfg,
		shutdownHandlers: make([]func(ctx context.Context) error, 0),
	}

	// Initialize in dependency order (simplified)
	if err := c.initLogger(); err != nil {
		return nil, err
	}

	if err := c.initAnalytics(); err != nil {
		logger.Warn("Analytics initialization failed", map[string]interface{}{"error": err.Error()})
		// Don't fail container creation for analytics
	}

	if err := c.initCache(); err != nil {
		logger.Warn("Cache initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	c.isInitialized = true
	logger.Info("Service container initialized successfully", map[string]interface{}{
		"services": "logger, analytics, cache",
	})

	return c, nil
}

// Initialization methods

func (c *ServiceContainer) initLogger() error {
	if err := logger.Init(logLevelToInt(c.config.Logger.Level), c.config.Logger.OutputFile); err != nil {
		return errors.InitializationError("logger", err).WithDetails(err.Error())
	}
	c.registerShutdownHandler(func(ctx context.Context) error {
		logger.Close()
		return nil
	})
	return nil
}

func (c *ServiceContainer) initAnalytics() error {
	a := analytics.GetAnalytics()
	if a == nil {
		return errors.InitializationError("analytics", fmt.Errorf("failed to get analytics instance"))
	}
	c.analytics = a
	return nil
}

func (c *ServiceContainer) initCache() error {
	if !c.config.Cache.Enabled {
		logger.Info("Cache is disabled", nil)
		return nil
	}

	// Initialize in-memory cache instances
	responseCoreCache := cache.NewInMemoryCache()
	queryCoreCache := cache.NewInMemoryCache()

	// Initialize response cache wrapper
	respCache := cache.NewResponseCache(responseCoreCache)

	// Initialize query result cache wrapper
	queryCache := cache.NewQueryResultCache(queryCoreCache)

	c.responseCache = respCache
	c.queryCache = queryCache

	logger.Info("Cache initialized successfully", map[string]interface{}{
		"response_cache_max_size": c.config.Cache.MaxResponseCacheSize,
		"query_cache_max_size":    c.config.Cache.MaxQueryCacheSize,
		"response_cache_ttl":      c.config.Cache.ResponseCacheTTL.String(),
		"query_cache_ttl":         c.config.Cache.QueryCacheTTL.String(),
	})

	return nil
}

// Getter methods

// Config returns the configuration
func (c *ServiceContainer) Config() *config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// Analytics returns the analytics service
func (c *ServiceContainer) Analytics() *analytics.Analytics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.analytics
}

// ResponseCache returns the response cache
func (c *ServiceContainer) ResponseCache() *cache.ResponseCache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.responseCache
}

// QueryCache returns the query result cache
func (c *ServiceContainer) QueryCache() *cache.QueryResultCache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.queryCache
}

// IsInitialized returns true if container is fully initialized
func (c *ServiceContainer) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isInitialized
}

// Shutdown gracefully shuts down all services
func (c *ServiceContainer) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	handlers := c.shutdownHandlers
	c.mu.Unlock()

	// Call handlers in reverse order (LIFO)
	for i := len(handlers) - 1; i >= 0; i-- {
		if err := handlers[i](ctx); err != nil {
			logger.Warn("Error during shutdown", map[string]interface{}{"error": err.Error()})
			// Continue with other handlers even if one fails
		}
	}

	c.mu.Lock()
	c.isInitialized = false
	c.mu.Unlock()

	logger.Info("Service container shutdown complete", nil)
	return nil
}

// registerShutdownHandler registers a function to be called during shutdown
func (c *ServiceContainer) registerShutdownHandler(handler func(ctx context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdownHandlers = append(c.shutdownHandlers, handler)
}

// Health returns health status of all services
func (c *ServiceContainer) Health() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := map[string]interface{}{
		"initialized": c.isInitialized,
		"services": map[string]bool{
			"analytics":      c.analytics != nil,
			"response_cache": c.responseCache != nil,
			"query_cache":    c.queryCache != nil,
		},
	}

	return health
}

// Helper functions

func logLevelToInt(level string) logger.LogLevel {
	switch level {
	case "debug":
		return logger.DEBUG
	case "info":
		return logger.INFO
	case "warn":
		return logger.WARN
	case "error":
		return logger.ERROR
	case "fatal":
		return logger.FATAL
	default:
		return logger.INFO
	}
}
