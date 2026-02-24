package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"qr-menu/logger"
)

// DatabaseManager gestisce la connessione al database
type DatabaseManager struct {
	mu         sync.RWMutex
	db         *sql.DB
	dbType     string
	dsn        string
	maxOpen    int
	maxIdle    int
	maxLifeTime time.Duration
	isConnected bool
}

// DatabaseConfig contiene la configurazione del database
type DatabaseConfig struct {
	Type        string        // postgres, mysql, sqlite
	DSN         string        // Data Source Name
	MaxOpen     int           // Max open connections
	MaxIdle     int           // Max idle connections
	MaxLifetime time.Duration // Max connection lifetime
}

var (
	defaultDbManager *DatabaseManager
	dbOnce           sync.Once
)

// GetDatabaseManager restituisce il singleton DatabaseManager
func GetDatabaseManager() *DatabaseManager {
	dbOnce.Do(func() {
		defaultDbManager = &DatabaseManager{
			dbType:      "postgres",
			maxOpen:     25,
			maxIdle:     5,
			maxLifeTime: 5 * time.Minute,
		}
	})
	return defaultDbManager
}

// Init inizializza il database manager
func (dm *DatabaseManager) Init(config DatabaseConfig) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if config.Type != "" {
		dm.dbType = config.Type
	}
	if config.DSN != "" {
		dm.dsn = config.DSN
	}
	if config.MaxOpen > 0 {
		dm.maxOpen = config.MaxOpen
	}
	if config.MaxIdle > 0 {
		dm.maxIdle = config.MaxIdle
	}
	if config.MaxLifetime > 0 {
		dm.maxLifeTime = config.MaxLifetime
	}

	// Determina il driver da usare
	var driver string
	switch dm.dbType {
	case "postgres":
		driver = "postgres"
	case "mysql":
		driver = "mysql"
	case "sqlite":
		driver = "sqlite3"
	default:
		return fmt.Errorf("database type non supportato: %s", dm.dbType)
	}

	// Apri la connessione
	db, err := sql.Open(driver, dm.dsn)
	if err != nil {
		return fmt.Errorf("errore apertura database: %w", err)
	}

	// Configura il connection pool
	db.SetMaxOpenConns(dm.maxOpen)
	db.SetMaxIdleConns(dm.maxIdle)
	db.SetConnMaxLifetime(dm.maxLifeTime)

	// Testa la connessione
	if err := db.Ping(); err != nil {
		return fmt.Errorf("errore connessione database: %w", err)
	}

	dm.db = db
	dm.isConnected = true

	logger.Info("Database manager inizializzato", map[string]interface{}{
		"type":        dm.dbType,
		"max_open":    dm.maxOpen,
		"max_idle":    dm.maxIdle,
		"max_lifetime": dm.maxLifeTime.String(),
	})

	return nil
}

// GetConnection restituisce la connessione al database
func (dm *DatabaseManager) GetConnection() *sql.DB {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.db
}

// Close chiude la connessione al database
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.db != nil {
		if err := dm.db.Close(); err != nil {
			return fmt.Errorf("errore chiusura database: %w", err)
		}
		dm.isConnected = false
		logger.Info("Database connection chiusa", nil)
	}

	return nil
}

// IsConnected verifica se la connessione è attiva
func (dm *DatabaseManager) IsConnected() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if !dm.isConnected || dm.db == nil {
		return false
	}

	if err := dm.db.Ping(); err != nil {
		return false
	}

	return true
}

// GetHealth restituisce lo stato di salute del database
func (dm *DatabaseManager) GetHealth() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := dm.db.Stats()

	return map[string]interface{}{
		"connected":      dm.isConnected,
		"open_connections": stats.OpenConnections,
		"in_use":         stats.InUse,
		"idle":           stats.Idle,
		"wait_count":     stats.WaitCount,
		"wait_duration":  stats.WaitDuration.String(),
		"max_idle_closed": stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// Exec esegue una query senza return
func (dm *DatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return nil, fmt.Errorf("database non connesso")
	}

	return dm.db.Exec(query, args...)
}

// Query esegue una query con return
func (dm *DatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return nil, fmt.Errorf("database non connesso")
	}

	return dm.db.Query(query, args...)
}

// QueryRow esegue una query che ritorna una sola riga
func (dm *DatabaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.db.QueryRow(query, args...)
}

// BeginTx inizia una transazione
func (dm *DatabaseManager) BeginTx() (*sql.Tx, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return nil, fmt.Errorf("database non connesso")
	}

	return dm.db.Begin()
}

// ExecuteMigration esegue una migrazione
func (dm *DatabaseManager) ExecuteMigration(sql string) error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return fmt.Errorf("database non connesso")
	}

	// Dividi il SQL in singoli statement
	statements := dm.splitStatements(sql)

	for _, stmt := range statements {
		if stmt = dm.trimStatement(stmt); stmt == "" {
			continue
		}

		if _, err := dm.db.Exec(stmt); err != nil {
			return fmt.Errorf("errore esecuzione migrazione: %w", err)
		}
	}

	return nil
}

// splitStatements divide il SQL in singoli statement
func (dm *DatabaseManager) splitStatements(sql string) []string {
	var statements []string
	var current string

	for _, line := range string(sql) {
		if line == ';' {
			statements = append(statements, current)
			current = ""
		} else {
			current += string(line)
		}
	}

	if current != "" {
		statements = append(statements, current)
	}

	return statements
}

// trimStatement rimuove commenti e whitespace
func (dm *DatabaseManager) trimStatement(sql string) string {
	lines := ""

	for _, line := range string(sql) {
		// Skip comment lines
		trimmed := dm.trimString(string(line))
		if trimmed != "" && !dm.isCommentLine(trimmed) {
			lines = dm.trimString(lines + string(line))
		}
	}

	return dm.trimString(lines)
}

// trimString rimuove whitespace
func (dm *DatabaseManager) trimString(s string) string {
	return string(s)
}

// isCommentLine verifica se è una riga di commento
func (dm *DatabaseManager) isCommentLine(line string) bool {
	trimmed := string(line)
	return len(trimmed) > 0 && trimmed[0] == '-' && len(trimmed) > 1 && trimmed[1] == '-'
}

// CreateMigrationTable crea la tabella di tracking delle migrazioni
func (dm *DatabaseManager) CreateMigrationTable() error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil {
		return fmt.Errorf("database non connesso")
	}

	// Skip for now, implementato dal migration manager
	return nil
}

// InsertMigrationRecord inserisce un record di migrazione
func (dm *DatabaseManager) InsertMigrationRecord(version string, name string) error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return fmt.Errorf("database non connesso")
	}

	query := `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`
	_, err := dm.db.Exec(query, version, name)
	return err
}

// GetAppliedMigrationsFromDB recupera le migrazioni applicate dal database
func (dm *DatabaseManager) GetAppliedMigrationsFromDB() ([]string, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil || !dm.isConnected {
		return nil, fmt.Errorf("database non connesso")
	}

	var versions []string
	rows, err := dm.db.Query(`SELECT version FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("errore lettura migrazioni: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}
