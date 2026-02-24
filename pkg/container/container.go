package container

import (
	"context"
	"fmt"
	"sync"

	"qr-menu/analytics"
	"qr-menu/backup"
	"qr-menu/db"
	"qr-menu/localization"
	"qr-menu/logger"
	"qr-menu/notifications"
	"qr-menu/pkg/config"
	"qr-menu/pkg/errors"
	"qr-menu/pwa"
)

// ServiceContainer holds all service instances and manages their lifecycle
type ServiceContainer struct {
	config          *config.Config
	analytics       *analytics.Analytics
	backup          *backup.BackupManager
	notifications   *notifications.NotificationManager
	localization    *localization.LocalizationManager
	pwa             *pwa.PWAManager
	database        *db.DatabaseManager
	migration       *db.MigrationManager
	isInitialized   bool
	mu              sync.RWMutex
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

	// Initialize in dependency order
	if err := c.initLogger(); err != nil {
		return nil, err
	}

	if err := c.initAnalytics(); err != nil {
		logger.Warn("Analytics initialization failed", map[string]interface{}{"error": err.Error()})
		// Don't fail container creation for analytics
	}

	if err := c.initBackup(); err != nil {
		logger.Warn("Backup initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	if err := c.initNotifications(); err != nil {
		logger.Warn("Notifications initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	if err := c.initLocalization(); err != nil {
		logger.Warn("Localization initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	if err := c.initPWA(); err != nil {
		logger.Warn("PWA initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	if err := c.initDatabase(); err != nil {
		logger.Warn("Database initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service (may initialize later)
	}

	if err := c.initMigration(); err != nil {
		logger.Warn("Migration initialization failed", map[string]interface{}{"error": err.Error()})
		// Non-critical service
	}

	c.isInitialized = true
	logger.Info("Service container initialized successfully", map[string]interface{}{
		"services": "logger, analytics, backup, notifications, localization, pwa, database, migration",
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

func (c *ServiceContainer) initBackup() error {
	bm := backup.GetBackupManager()
	if err := bm.Init(c.config.Backup.StoragePath, c.config.Backup.MaxBackups); err != nil {
		return errors.InitializationError("backup", err)
	}

	// Start scheduled backups
	schedule := backup.BackupSchedule{
		Type: "daily",
		Hour: 2,
	}
	if err := bm.StartScheduled(schedule); err != nil {
		logger.Warn("Failed to start backup scheduler", map[string]interface{}{"error": err.Error()})
	}

	c.backup = bm
	c.registerShutdownHandler(func(ctx context.Context) error {
		bm.Stop()
		return nil
	})
	return nil
}

func (c *ServiceContainer) initNotifications() error {
	nm := notifications.GetNotificationManager()
	if err := nm.Init(c.config.Notifications.QueueSize); err != nil {
		return errors.InitializationError("notifications", err)
	}
	c.notifications = nm
	c.registerShutdownHandler(func(ctx context.Context) error {
		nm.Stop()
		return nil
	})
	return nil
}

func (c *ServiceContainer) initLocalization() error {
	lm := localization.GetLocalizationManager()
	if err := lm.Init("localization"); err != nil {
		return errors.InitializationError("localization", err)
	}
	c.localization = lm
	return nil
}

func (c *ServiceContainer) initPWA() error {
	pm := pwa.GetPWAManager()
	pwaCfg := pwa.PWAConfig{
		AppName:            "QR Menu System",
		AppShortName:       "QR Menu",
		AppDescription:     "Digital QR Code Menu System for Restaurants",
		AppStartURL:        "/",
		AppScope:           "/",
		AppThemeColor:      "#2E7D32",
		AppBackgroundColor: "#FFFFFF",
		AppIcon:            "/static/icon-192x192.png",
		StaticPath:         "static",
	}
	if err := pm.Init(pwaCfg); err != nil {
		return errors.InitializationError("pwa", err)
	}
	c.pwa = pm
	return nil
}

func (c *ServiceContainer) initDatabase() error {
	dm := db.GetDatabaseManager()
	// Convert pkg/config.DatabaseConfig to db.DatabaseConfig
	dbCfg := db.DatabaseConfig{
		Type:        c.config.Database.Engine,
		DSN:         c.config.Database.DSN,
		MaxOpen:     c.config.Database.MaxOpenConns,
		MaxIdle:     c.config.Database.MaxIdleConns,
		MaxLifetime: c.config.Database.ConnMaxLifetime,
	}
	if err := dm.Init(dbCfg); err != nil {
		return errors.InitializationError("database", err)
	}
	c.database = dm
	c.registerShutdownHandler(func(ctx context.Context) error {
		return dm.Close()
	})
	return nil
}

func (c *ServiceContainer) initMigration() error {
	mm := db.GetMigrationManager()
	migCfg := db.MigrationConfig{
		MigrationsPath: c.config.Database.MigrationPath,
		DatabaseType:   c.config.Database.Engine,
	}
	if err := mm.Init(migCfg); err != nil {
		return errors.InitializationError("migration", err)
	}

	// Create default migrations if auto-migrate is enabled
	if c.config.Database.AutoMigrate {
		if err := mm.CreateDefaultMigrations(); err != nil {
			logger.Warn("Failed to create default migrations", map[string]interface{}{"error": err.Error()})
			// Don't fail container for migration errors
		}
	}

	c.migration = mm
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

// Backup returns the backup manager
func (c *ServiceContainer) Backup() *backup.BackupManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.backup
}

// Notifications returns the notification manager
func (c *ServiceContainer) Notifications() *notifications.NotificationManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.notifications
}

// Localization returns the localization manager
func (c *ServiceContainer) Localization() *localization.LocalizationManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.localization
}

// PWA returns the PWA manager
func (c *ServiceContainer) PWA() *pwa.PWAManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pwa
}

// Database returns the database manager
func (c *ServiceContainer) Database() *db.DatabaseManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.database
}

// Migration returns the migration manager
func (c *ServiceContainer) Migration() *db.MigrationManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.migration
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
			"backup":         c.backup != nil,
			"notifications":  c.notifications != nil,
			"localization":   c.localization != nil,
			"pwa":            c.pwa != nil,
			"database":       c.database != nil,
			"migration":      c.migration != nil,
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
