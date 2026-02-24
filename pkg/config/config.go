package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Backup         BackupConfig
	Notifications  NotificationConfig
	Localization   LocalizationConfig
	Logger         LoggerConfig
	Analytics      AnalyticsConfig
	Security       SecurityConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxBodySize  int64
	Environment  string // dev, staging, prod
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN              string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
	Engine           string // postgres, mysql, sqlite
	MigrationPath    string
	AutoMigrate      bool
}

// BackupConfig holds backup service configuration
type BackupConfig struct {
	QueueSize          int
	MaxBackups         int
	ScheduleTime       string        // HH:MM format, default "02:00"
	Enabled            bool
	CompressionLevel   int           // 1-9
	RetentionDays      int
	RotationInterval   time.Duration
	StoragePath        string
}

// NotificationConfig holds notification service configuration
type NotificationConfig struct {
	Workers           int
	QueueSize         int
	BatchSize         int
	BatchTimeout      time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
	FCMCredentialsURL string
	Enabled           bool
}

// LocalizationConfig holds localization configuration
type LocalizationConfig struct {
	DefaultLanguage string
	SupportedLanguages []string
	DateFormat      string
	TimeFormat      string
	TimezoneOffset  int // hours
	CurrencySymbols map[string]string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string        // debug, info, warn, error, fatal
	Format      string        // json, text
	OutputFile  string        // path to log file
	MaxSize     int           // MB
	MaxBackups  int
	MaxAge      int           // days
	Compress    bool
	Development bool          // true for dev, false for prod
}

// AnalyticsConfig holds analytics configuration
type AnalyticsConfig struct {
	Enabled         bool
	TrackingEnabled bool
	StoragePath     string
	CleanupInterval time.Duration
	RetentionDays   int
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	SessionTimeout   time.Duration
	PasswordMinLen   int
	PasswordRequireSpecial bool
	PasswordRequireNumbers bool
	RateLimitPerSecond int
	RateLimitBurst     int
	CORSEnabled      bool
	CORSAllowedOrigins []string
	EnableHTTPS      bool
	CertFile         string
	KeyFile          string
}

// Load loads configuration from environment variables and returns a Config struct
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "localhost"),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			MaxBodySize:  getEnvInt64("SERVER_MAX_BODY_SIZE", 10*1024*1024), // 10MB
			Environment:  getEnv("ENVIRONMENT", "dev"),
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DATABASE_DSN", "host=localhost port=5432 user=postgres password=password dbname=qrmenu sslmode=disable"),
			MaxOpenConns:    getEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DATABASE_CONN_MAX_IDLE_TIME", 10*time.Minute),
			Engine:          getEnv("DATABASE_ENGINE", "postgres"),
			MigrationPath:   getEnv("DATABASE_MIGRATION_PATH", "./migrations"),
			AutoMigrate:     getEnvBool("DATABASE_AUTO_MIGRATE", true),
		},
		Backup: BackupConfig{
			QueueSize:        getEnvInt("BACKUP_QUEUE_SIZE", 100),
			MaxBackups:       getEnvInt("BACKUP_MAX_BACKUPS", 30),
			ScheduleTime:     getEnv("BACKUP_SCHEDULE_TIME", "02:00"),
			Enabled:          getEnvBool("BACKUP_ENABLED", true),
			CompressionLevel: getEnvInt("BACKUP_COMPRESSION_LEVEL", 6),
			RetentionDays:    getEnvInt("BACKUP_RETENTION_DAYS", 90),
			RotationInterval: getEnvDuration("BACKUP_ROTATION_INTERVAL", 24*time.Hour),
			StoragePath:      getEnv("BACKUP_STORAGE_PATH", "./backups"),
		},
		Notifications: NotificationConfig{
			Workers:           getEnvInt("NOTIFICATIONS_WORKERS", 3),
			QueueSize:         getEnvInt("NOTIFICATIONS_QUEUE_SIZE", 100),
			BatchSize:         getEnvInt("NOTIFICATIONS_BATCH_SIZE", 10),
			BatchTimeout:      getEnvDuration("NOTIFICATIONS_BATCH_TIMEOUT", 5*time.Second),
			MaxRetries:        getEnvInt("NOTIFICATIONS_MAX_RETRIES", 3),
			RetryDelay:        getEnvDuration("NOTIFICATIONS_RETRY_DELAY", 10*time.Second),
			FCMCredentialsURL: getEnv("NOTIFICATIONS_FCM_CREDENTIALS_URL", ""),
			Enabled:           getEnvBool("NOTIFICATIONS_ENABLED", true),
		},
		Localization: LocalizationConfig{
			DefaultLanguage: getEnv("LOCALIZATION_DEFAULT_LANG", "it"),
			SupportedLanguages: []string{"it", "en", "es", "fr", "de", "pt", "ja", "zh", "ar"},
			DateFormat:      getEnv("LOCALIZATION_DATE_FORMAT", "2006-01-02"),
			TimeFormat:      getEnv("LOCALIZATION_TIME_FORMAT", "15:04:05"),
			TimezoneOffset:  getEnvInt("LOCALIZATION_TIMEZONE_OFFSET", 1),
			CurrencySymbols: map[string]string{
				"EUR": "€",
				"USD": "$",
				"GBP": "£",
				"JPY": "¥",
			},
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOGGER_LEVEL", "info"),
			Format:      getEnv("LOGGER_FORMAT", "json"),
			OutputFile:  getEnv("LOGGER_OUTPUT_FILE", "./logs/qr-menu.log"),
			MaxSize:     getEnvInt("LOGGER_MAX_SIZE", 100),
			MaxBackups:  getEnvInt("LOGGER_MAX_BACKUPS", 10),
			MaxAge:      getEnvInt("LOGGER_MAX_AGE", 30),
			Compress:    getEnvBool("LOGGER_COMPRESS", true),
			Development: getEnv("ENVIRONMENT", "dev") == "dev",
		},
		Analytics: AnalyticsConfig{
			Enabled:         getEnvBool("ANALYTICS_ENABLED", true),
			TrackingEnabled: getEnvBool("ANALYTICS_TRACKING_ENABLED", true),
			StoragePath:     getEnv("ANALYTICS_STORAGE_PATH", "./analytics"),
			CleanupInterval: getEnvDuration("ANALYTICS_CLEANUP_INTERVAL", 24*time.Hour),
			RetentionDays:   getEnvInt("ANALYTICS_RETENTION_DAYS", 90),
		},
		Security: SecurityConfig{
			SessionTimeout:     getEnvDuration("SECURITY_SESSION_TIMEOUT", 24*time.Hour),
			PasswordMinLen:     getEnvInt("SECURITY_PASSWORD_MIN_LEN", 8),
			PasswordRequireSpecial: getEnvBool("SECURITY_PASSWORD_REQUIRE_SPECIAL", true),
			PasswordRequireNumbers: getEnvBool("SECURITY_PASSWORD_REQUIRE_NUMBERS", true),
			RateLimitPerSecond: getEnvInt("SECURITY_RATE_LIMIT_PER_SEC", 10),
			RateLimitBurst:     getEnvInt("SECURITY_RATE_LIMIT_BURST", 100),
			CORSEnabled:        getEnvBool("SECURITY_CORS_ENABLED", true),
			CORSAllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
			EnableHTTPS:        getEnvBool("SECURITY_ENABLE_HTTPS", false),
			CertFile:           getEnv("SECURITY_CERT_FILE", ""),
			KeyFile:            getEnv("SECURITY_KEY_FILE", ""),
		},
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getEnvBool(key string, defaultValue bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolVal
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

// IsDevelopment returns true if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "dev"
}

// IsProduction returns true if environment is production
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "prod"
}

// IsStaging returns true if environment is staging
func (c *Config) IsStaging() bool {
	return c.Server.Environment == "staging"
}
