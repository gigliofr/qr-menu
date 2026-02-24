package app

import (
	"fmt"
	"qr-menu/analytics"
	"qr-menu/backup"
	"qr-menu/db"
	"qr-menu/localization"
	"qr-menu/logger"
	"qr-menu/notifications"
	"qr-menu/pwa"
	"qr-menu/security"
)

// Services contiene tutti i servizi inizializzati
type Services struct {
	Analytics     *analytics.Analytics
	Backup        *backup.BackupManager
	Notifications *notifications.NotificationManager
	Localization  *localization.LocalizationManager
	PWA           *pwa.PWAManager
	Migration     *db.MigrationManager
	Database      *db.DatabaseManager
	
	// Security services
	RateLimiter     *security.RateLimiter
	AuditLogger     *security.AuditLogger
	GDPRManager     *security.GDPRManager
	SecurityHeaders *security.SecurityHeadersMiddleware
	CORSMiddleware  *security.CORSMiddleware
}

// Config contiene la configurazione per l'inizializzazione
type Config struct {
	LogLevel        logger.LogLevel
	LogDir          string
	BackupDir       string
	BackupRetention int
	TranslationDir  string
	MigrationDir    string
	DatabaseURL     string
	PWAConfig       pwa.PWAConfig
}

// DefaultConfig ritorna la configurazione di default
func DefaultConfig() Config {
	return Config{
		LogLevel:        logger.INFO,
		LogDir:          "logs",
		BackupDir:       "backups",
		BackupRetention: 30,
		TranslationDir:  "translations",
		MigrationDir:    "db/migrations",
		PWAConfig: pwa.PWAConfig{
			AppName:            "QR Menu System",
			AppShortName:       "QR Menu",
			AppDescription:     "Digital QR Code Menu System for Restaurants",
			AppStartURL:        "/",
			AppScope:           "/",
			AppThemeColor:      "#2E7D32",
			AppBackgroundColor: "#FFFFFF",
			StaticPath:         "static",
		},
	}
}

// InitializeServices inizializza tutti i servizi dell'applicazione
func InitializeServices(cfg Config) (*Services, error) {
	services := &Services{}
	
	// 1. Logger (critico - se fallisce, fermiamo tutto)
	if err := logger.Init(cfg.LogLevel, cfg.LogDir); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	logger.Info("Starting QR Menu System initialization", map[string]interface{}{
		"version": "2.0.0",
	})
	
	// 2. Analytics (non critico)
	services.Analytics = analytics.GetAnalytics()
	
	// 3. Backup Manager (non critico)
	services.Backup = backup.GetBackupManager()
	if err := services.Backup.Init(cfg.BackupDir, cfg.BackupRetention); err != nil {
		logger.Warn("Backup system not initialized", map[string]interface{}{"error": err.Error()})
	} else {
		// Avvia backup automatico giornaliero
		schedule := backup.BackupSchedule{Type: "daily", Hour: 2}
		if err := services.Backup.StartScheduled(schedule); err != nil {
			logger.Warn("Backup scheduler not started", map[string]interface{}{"error": err.Error()})
		}
	}
	
	// 4. Notification Manager (non critico)
	services.Notifications = notifications.GetNotificationManager()
	if err := services.Notifications.Init(100); err != nil {
		logger.Warn("Notification system not initialized", map[string]interface{}{"error": err.Error()})
	} else {
		if err := services.Notifications.Start(3); err != nil {
			logger.Warn("Notification workers not started", map[string]interface{}{"error": err.Error()})
		}
	}
	
	// 5. Localization Manager (non critico)
	services.Localization = localization.GetLocalizationManager()
	services.Localization.CreateDefaultTranslationFiles(cfg.TranslationDir)
	if err := services.Localization.Init(cfg.TranslationDir); err != nil {
		logger.Warn("Localization not initialized", map[string]interface{}{"error": err.Error()})
	}
	
	// 6. PWA Manager (non critico)
	services.PWA = pwa.GetPWAManager()
	if err := services.PWA.Init(cfg.PWAConfig); err != nil {
		logger.Warn("PWA not initialized", map[string]interface{}{"error": err.Error()})
	}
	
	// 7. Migration Manager (non critico)
	services.Migration = db.GetMigrationManager()
	if err := services.Migration.Init(db.MigrationConfig{
		MigrationsPath: cfg.MigrationDir,
		DatabaseType:   "postgres",
	}); err != nil {
		logger.Warn("Migration manager not initialized", map[string]interface{}{"error": err.Error()})
	} else {
		services.Migration.CreateDefaultMigrations()
	}
	
	// 8. Database Manager (opzionale - solo se DATABASE_URL è configurato)
	if cfg.DatabaseURL != "" {
		services.Database = db.GetDatabaseManager()
		if err := services.Database.Init(db.DatabaseConfig{
			Type: "postgres",
			DSN:  cfg.DatabaseURL,
		}); err != nil {
			logger.Warn("Database not initialized", map[string]interface{}{"error": err.Error()})
		}
	}
	
	// 9. Security Services
	services.RateLimiter = security.NewRateLimiter()
	services.AuditLogger = security.NewAuditLogger(10000)
	services.GDPRManager = security.NewGDPRManager(services.AuditLogger)
	services.SecurityHeaders = security.NewSecurityHeadersMiddleware(security.DefaultSecurityHeadersConfig())
	services.CORSMiddleware = security.NewCORSMiddleware(security.DefaultCORSConfig())
	
	// 10. Pulizia log vecchi
	logger.CleanOldLogs(30)
	
	logger.Info("All services initialized successfully", map[string]interface{}{
		"analytics":     true,
		"backup":        services.Backup != nil,
		"notifications": services.Notifications != nil,
		"localization":  services.Localization != nil,
		"pwa":           services.PWA != nil,
		"database":      services.Database != nil,
		"security":      true,
	})
	
	return services, nil
}

// Shutdown ferma gracefully tutti i servizi
func (s *Services) Shutdown() {
	logger.Info("Shutting down services...", nil)
	
	if s.RateLimiter != nil {
		s.RateLimiter.Stop()
	}
	
	if s.Backup != nil {
		s.Backup.Stop()
	}
	
	if s.Notifications != nil {
		s.Notifications.Stop()
	}
	
	if s.Database != nil {
		s.Database.Close()
	}
	
	logger.Close()
}
