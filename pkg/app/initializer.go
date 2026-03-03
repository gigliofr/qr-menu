package app

import (
	"fmt"
	"qr-menu/analytics"
	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/security"
)

// Services contiene i servizi core inizializzati
type Services struct {
	Analytics  *analytics.Analytics
	Database   *db.DatabaseManager

	// Security services
	RateLimiter     *security.RateLimiter
	AuditLogger     *security.AuditLogger
	GDPRManager     *security.GDPRManager
	SecurityHeaders *security.SecurityHeadersMiddleware
	CORSMiddleware  *security.CORSMiddleware
}

// Config contiene la configurazione per l'inizializzazione
type Config struct {
	LogLevel    logger.LogLevel
	LogDir      string
	DatabaseURL string
}

// DefaultConfig ritorna la configurazione di default
func DefaultConfig() Config {
	return Config{
		LogLevel: logger.INFO,
		LogDir:   "logs",
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
		"version": "2.0.0-simplified",
	})

	// 2. Analytics
	services.Analytics = analytics.GetAnalytics()

	// 3. Security Services
	services.RateLimiter = security.NewRateLimiter()
	services.AuditLogger = security.NewAuditLogger(10000)
	services.GDPRManager = security.NewGDPRManager(services.AuditLogger)
	services.SecurityHeaders = security.NewSecurityHeadersMiddleware(security.DefaultSecurityHeadersConfig())
	services.CORSMiddleware = security.NewCORSMiddleware(security.DefaultCORSConfig())

	// 4. Pulizia log vecchi
	logger.CleanOldLogs(30)

	logger.Info("All core services initialized successfully", map[string]interface{}{
		"analytics": true,
		"security":  true,
	})

	return services, nil
}

// Shutdown ferma gracefully tutti i servizi
func (s *Services) Shutdown() {
	logger.Info("Shutting down services...", nil)

	if s.RateLimiter != nil {
		s.RateLimiter.Stop()
	}

	logger.Close()
}
