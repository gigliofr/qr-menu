package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"qr-menu/logger"
)

// Migration rappresenta una singola migrazione
type Migration struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	AppliedAt time.Time `json:"applied_at,omitempty"`
	Status    string    `json:"status"` // pending, applied, failed
	SQL       string    `json:"-"`
}

// MigrationManager gestisce le migrazioni del database
type MigrationManager struct {
	mu                 sync.RWMutex
	migrationsPath     string
	appliedMigrations  []Migration
	pendingMigrations  []Migration
	databaseType       string // postgres, mysql, sqlite
	schemaVersion      int
}

// MigrationConfig contiene la configurazione per le migrazioni
type MigrationConfig struct {
	MigrationsPath string // Path ai file di migrazione
	DatabaseType   string // postgres, mysql, sqlite
}

var (
	defaultManager *MigrationManager
	once           sync.Once

	// Default migrations
	defaultMigrations = map[string]string{
		"001_initial_schema": `-- Create restaurants table
CREATE TABLE IF NOT EXISTS restaurants (
	id VARCHAR(36) PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	restaurant_name VARCHAR(255) NOT NULL,
	restaurant_type VARCHAR(100),
	phone VARCHAR(20),
	address TEXT,
	city VARCHAR(100),
	postal_code VARCHAR(20),
	country VARCHAR(100),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_username (username),
	INDEX idx_email (email)
);

-- Create menus table
CREATE TABLE IF NOT EXISTS menus (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	menu_name VARCHAR(255) NOT NULL,
	menu_description TEXT,
	is_active BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_is_active (is_active)
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
	id VARCHAR(36) PRIMARY KEY,
	menu_id VARCHAR(36) NOT NULL,
	category_name VARCHAR(255) NOT NULL,
	category_description TEXT,
	display_order INT DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
	INDEX idx_menu_id (menu_id)
);

-- Create menu_items table
CREATE TABLE IF NOT EXISTS menu_items (
	id VARCHAR(36) PRIMARY KEY,
	category_id VARCHAR(36) NOT NULL,
	item_name VARCHAR(255) NOT NULL,
	item_description TEXT,
	price DECIMAL(10, 2) NOT NULL,
	image_url VARCHAR(255),
	is_available BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
	INDEX idx_category_id (category_id),
	INDEX idx_is_available (is_available)
);`,

		"002_add_orders_table": `-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	customer_name VARCHAR(255),
	customer_phone VARCHAR(20),
	customer_email VARCHAR(255),
	total_amount DECIMAL(10, 2) NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	notes TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_status (status),
	INDEX idx_created_at (created_at)
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
	id VARCHAR(36) PRIMARY KEY,
	order_id VARCHAR(36) NOT NULL,
	menu_item_id VARCHAR(36),
	item_name VARCHAR(255) NOT NULL,
	quantity INT NOT NULL,
	unit_price DECIMAL(10, 2) NOT NULL,
	total_price DECIMAL(10, 2) NOT NULL,
	FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
	FOREIGN KEY (menu_item_id) REFERENCES menu_items(id),
	INDEX idx_order_id (order_id)
);`,

		"003_add_analytics_table": `-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	event_type VARCHAR(100) NOT NULL,
	event_data JSON,
	user_id VARCHAR(36),
	session_id VARCHAR(36),
	ip_address VARCHAR(45),
	user_agent TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_event_type (event_type),
	INDEX idx_created_at (created_at)
);

-- Create analytics_sessions table
CREATE TABLE IF NOT EXISTS analytics_sessions (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	session_start TIMESTAMP NOT NULL,
	session_end TIMESTAMP,
	duration_seconds INT,
	total_events INT DEFAULT 0,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_session_start (session_start)
);`,

		"004_add_backups_table": `-- Create backups table
CREATE TABLE IF NOT EXISTS backups (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36),
	backup_path VARCHAR(255) NOT NULL,
	backup_size BIGINT NOT NULL,
	file_count INT,
	compress_rate DECIMAL(5, 2),
	hash VARCHAR(64),
	status VARCHAR(50) DEFAULT 'success',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_created_at (created_at),
	INDEX idx_status (status)
);

-- Create backup_schedules table
CREATE TABLE IF NOT EXISTS backup_schedules (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36),
	schedule_type VARCHAR(50) NOT NULL,
	schedule_hour INT,
	schedule_day INT,
	is_active BOOLEAN DEFAULT TRUE,
	next_run TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_restaurant_id (restaurant_id)
);`,

		"005_add_notifications_table": `-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	enable_push BOOLEAN DEFAULT TRUE,
	enable_email BOOLEAN DEFAULT TRUE,
	enable_sms BOOLEAN DEFAULT FALSE,
	order_notifications BOOLEAN DEFAULT TRUE,
	reservation_notifications BOOLEAN DEFAULT TRUE,
	promo_notifications BOOLEAN DEFAULT TRUE,
	quiet_hours_start VARCHAR(5),
	quiet_hours_end VARCHAR(5),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	UNIQUE KEY uniq_restaurant_id (restaurant_id)
);

-- Create notification_history table
CREATE TABLE IF NOT EXISTS notification_history (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	type VARCHAR(50) NOT NULL,
	title VARCHAR(255) NOT NULL,
	body TEXT,
	status VARCHAR(50) DEFAULT 'sent',
	read_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_created_at (created_at)
);`,
	}
)

// GetMigrationManager restituisce il singleton MigrationManager
func GetMigrationManager() *MigrationManager {
	once.Do(func() {
		defaultManager = &MigrationManager{
			migrationsPath: "db/migrations",
			databaseType:   "postgres",
			schemaVersion:  0,
		}
	})
	return defaultManager
}

// Init inizializza il migration manager
func (mm *MigrationManager) Init(config MigrationConfig) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if config.MigrationsPath != "" {
		mm.migrationsPath = config.MigrationsPath
	}
	if config.DatabaseType != "" {
		mm.databaseType = config.DatabaseType
	}

	// Crea la directory per le migrazioni
	if err := os.MkdirAll(mm.migrationsPath, 0755); err != nil {
		return fmt.Errorf("errore creazione directory migrazioni: %w", err)
	}

	// Carica le migrazioni da file
	if err := mm.loadMigrations(); err != nil {
		logger.Warn("Errore caricamento migrazioni", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Migration manager inizializzato", map[string]interface{}{
		"migrations_path": mm.migrationsPath,
		"database_type":   mm.databaseType,
		"total":           len(mm.appliedMigrations) + len(mm.pendingMigrations),
	})

	return nil
}

// CreateDefaultMigrations crea i file di migrazione di default
func (mm *MigrationManager) CreateDefaultMigrations() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for name, sql := range defaultMigrations {
		filePath := filepath.Join(mm.migrationsPath, name+".sql")

		// Verifica se il file esiste già
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		data := []byte(sql)
		err := os.WriteFile(filePath, data, 0644)
		if err != nil {
			logger.Error("Errore scrittura migration file", map[string]interface{}{
				"file":  name,
				"error": err.Error(),
			})
			continue
		}

		logger.Info("File migrazione creato", map[string]interface{}{
			"file": name,
			"path": filePath,
		})
	}

	return nil
}

// loadMigrations carica tutte le migrazioni dai file
func (mm *MigrationManager) loadMigrations() error {
	entries, err := os.ReadDir(mm.migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory non esiste, è OK
			return nil
		}
		return fmt.Errorf("errore lettura directory migrazioni: %w", err)
	}

	migrations := []Migration{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Estrai versione dal nome file (es. 001_initial_schema.sql -> 001)
		version := strings.TrimSuffix(entry.Name(), ".sql")
		parts := strings.Split(version, "_")
		if len(parts) < 1 {
			continue
		}

		filePath := filepath.Join(mm.migrationsPath, entry.Name())
		sqlData, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Errore lettura migration file", map[string]interface{}{
				"file":  entry.Name(),
				"error": err.Error(),
			})
			continue
		}

		migration := Migration{
			Version: parts[0],
			Name:    version,
			SQL:     string(sqlData),
			Status:  "pending",
		}

		migrations = append(migrations, migration)
	}

	// Ordina le migrazioni per versione
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	mm.pendingMigrations = migrations

	return nil
}

// GetMigrations restituisce tutte le migrazioni
func (mm *MigrationManager) GetMigrations() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return map[string]interface{}{
		"applied":  mm.appliedMigrations,
		"pending":  mm.pendingMigrations,
		"total":    len(mm.appliedMigrations) + len(mm.pendingMigrations),
		"version":  mm.schemaVersion,
	}
}

// GetMigrationStatus restituisce lo status delle migrazioni
func (mm *MigrationManager) GetMigrationStatus() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return map[string]interface{}{
		"applied_count":  len(mm.appliedMigrations),
		"pending_count":  len(mm.pendingMigrations),
		"current_version": mm.schemaVersion,
		"last_applied":   func() *time.Time {
			if len(mm.appliedMigrations) > 0 {
				return &mm.appliedMigrations[len(mm.appliedMigrations)-1].AppliedAt
			}
			return nil
		}(),
	}
}

// MarkMigrationApplied marca una migrazione come applicata
func (mm *MigrationManager) MarkMigrationApplied(version string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Trova e rimuovi dalla pending
	for i, m := range mm.pendingMigrations {
		if m.Version == version {
			m.AppliedAt = time.Now()
			m.Status = "applied"
			mm.appliedMigrations = append(mm.appliedMigrations, m)
			mm.pendingMigrations = append(mm.pendingMigrations[:i], mm.pendingMigrations[i+1:]...)

			logger.Info("Migrazione marcata come applicata", map[string]interface{}{
				"version": version,
				"name":    m.Name,
			})

			return nil
		}
	}

	return fmt.Errorf("migrazione non trovata: %s", version)
}

// MarkMigrationFailed marca una migrazione come fallita
func (mm *MigrationManager) MarkMigrationFailed(version string, errMsg string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for i, m := range mm.pendingMigrations {
		if m.Version == version {
			m.Status = "failed"
			mm.pendingMigrations[i] = m

			logger.Error("Migrazione fallita", map[string]interface{}{
				"version": version,
				"name":    m.Name,
				"error":   errMsg,
			})

			return nil
		}
	}

	return fmt.Errorf("migrazione non trovata: %s", version)
}

// RollbackMigration esegue il rollback di una migrazione
func (mm *MigrationManager) RollbackMigration(version string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for i, m := range mm.appliedMigrations {
		if m.Version == version {
			m.Status = "pending"
			mm.pendingMigrations = append(mm.pendingMigrations, m)
			mm.appliedMigrations = append(mm.appliedMigrations[:i], mm.appliedMigrations[i+1:]...)

			logger.Info("Migrazione rollback", map[string]interface{}{
				"version": version,
				"name":    m.Name,
			})

			return nil
		}
	}

	return fmt.Errorf("migrazione applicata non trovata: %s", version)
}

// GetNextMigration restituisce la prossima migrazione da applicare
func (mm *MigrationManager) GetNextMigration() *Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if len(mm.pendingMigrations) > 0 {
		return &mm.pendingMigrations[0]
	}

	return nil
}

// GetAppliedMigrations restituisce tutte le migrazioni applicate
func (mm *MigrationManager) GetAppliedMigrations() []Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.appliedMigrations
}

// GetPendingMigrations restituisce tutte le migrazioni in attesa
func (mm *MigrationManager) GetPendingMigrations() []Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.pendingMigrations
}

// UpdateSchemaVersion aggiorna il numero di versione dello schema
func (mm *MigrationManager) UpdateSchemaVersion(version int) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.schemaVersion = version
	logger.Info("Schema version aggiornato", map[string]interface{}{
		"version": version,
	})
}
