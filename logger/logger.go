package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel definisce i livelli di logging
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// LogEntry rappresenta una singola voce di log strutturata
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// Logger rappresenta il logger personalizzato
type Logger struct {
	level      LogLevel
	output     io.Writer
	fileWriter *os.File
	logDir     string
}

var (
	defaultLogger *Logger
)

// Init inizializza il sistema di logging
func Init(level LogLevel, logDir string) error {
	// Crea la directory dei log se non esiste
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("errore nella creazione della directory log: %v", err)
	}

	// Determina il file di log per oggi
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("qr-menu-%s.log", today))

	// Apri il file di log
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("errore nell'apertura del file di log: %v", err)
	}

	// Crea il logger con output multiplo (console + file)
	multiWriter := io.MultiWriter(os.Stdout, file)

	defaultLogger = &Logger{
		level:      level,
		output:     multiWriter,
		fileWriter: file,
		logDir:     logDir,
	}

	// Log di inizializzazione
	Info("Logger inizializzato", map[string]interface{}{
		"level":    levelNames[level],
		"log_dir":  logDir,
		"log_file": logFile,
	})

	return nil
}

// Close chiude il logger e i file aperti
func Close() {
	if defaultLogger != nil && defaultLogger.fileWriter != nil {
		defaultLogger.fileWriter.Close()
	}
}

// getSource ottiene informazioni sulla funzione chiamante
func getSource() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}
	
	// Estrae solo il nome del file senza il percorso completo
	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}

// writeLog scrive una voce di log
func (l *Logger) writeLog(level LogLevel, message string, data map[string]interface{}, userID, ip, userAgent string) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     levelNames[level],
		Message:   message,
		Source:    getSource(),
		Data:      data,
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
	}

	// Serializza in JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Errore nella serializzazione del log: %v", err)
		return
	}

	// Scrive nel output
	fmt.Fprintf(l.output, "%s\n", string(jsonData))
}

// Funzioni di logging pubbliche

// Debug scrive un log di debug
func Debug(message string, data map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.writeLog(DEBUG, message, data, "", "", "")
	}
}

// Info scrive un log informativo
func Info(message string, data map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.writeLog(INFO, message, data, "", "", "")
	}
}

// Warn scrive un log di warning
func Warn(message string, data map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.writeLog(WARN, message, data, "", "", "")
	}
}

// Error scrive un log di errore
func Error(message string, data map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.writeLog(ERROR, message, data, "", "", "")
	}
}

// Fatal scrive un log critico e termina l'applicazione
func Fatal(message string, data map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.writeLog(FATAL, message, data, "", "", "")
	}
	os.Exit(1)
}

// Funzioni con contesto utente per audit e sicurezza

// InfoWithContext scrive un log informativo con contesto utente
func InfoWithContext(message string, data map[string]interface{}, userID, ip, userAgent string) {
	if defaultLogger != nil {
		defaultLogger.writeLog(INFO, message, data, userID, ip, userAgent)
	}
}

// WarnWithContext scrive un log di warning con contesto utente
func WarnWithContext(message string, data map[string]interface{}, userID, ip, userAgent string) {
	if defaultLogger != nil {
		defaultLogger.writeLog(WARN, message, data, userID, ip, userAgent)
	}
}

// ErrorWithContext scrive un log di errore con contesto utente
func ErrorWithContext(message string, data map[string]interface{}, userID, ip, userAgent string) {
	if defaultLogger != nil {
		defaultLogger.writeLog(ERROR, message, data, userID, ip, userAgent)
	}
}

// Funzioni specializzate per sicurezza

// SecurityEvent registra un evento di sicurezza
func SecurityEvent(eventType, message string, userID, ip, userAgent string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["security_event"] = true
	data["event_type"] = eventType
	
	WarnWithContext(fmt.Sprintf("SECURITY: %s - %s", eventType, message), data, userID, ip, userAgent)
}

// AuditLog registra un evento di audit
func AuditLog(action, resource, message string, userID, ip, userAgent string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["audit"] = true
	data["action"] = action
	data["resource"] = resource
	
	InfoWithContext(fmt.Sprintf("AUDIT: %s on %s - %s", action, resource, message), data, userID, ip, userAgent)
}

// PerformanceLog registra metriche di performance
func PerformanceLog(operation string, duration time.Duration, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["performance"] = true
	data["operation"] = operation
	data["duration_ms"] = duration.Milliseconds()
	
	if duration > time.Second {
		Warn(fmt.Sprintf("PERFORMANCE: Operazione lenta - %s (durata: %v)", operation, duration), data)
	} else {
		Debug(fmt.Sprintf("PERFORMANCE: %s (durata: %v)", operation, duration), data)
	}
}

// CleanOldLogs rimuove i file di log più vecchi di N giorni
func CleanOldLogs(daysToKeep int) error {
	if defaultLogger == nil {
		return fmt.Errorf("logger non inizializzato")
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)
	
	entries, err := os.ReadDir(defaultLogger.logDir)
	if err != nil {
		return fmt.Errorf("errore nella lettura della directory log: %v", err)
	}

	deletedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "qr-menu-") && strings.HasSuffix(entry.Name(), ".log") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			if info.ModTime().Before(cutoff) {
				filePath := filepath.Join(defaultLogger.logDir, entry.Name())
				if err := os.Remove(filePath); err == nil {
					deletedCount++
				}
			}
		}
	}

	Info("Pulizia log completata", map[string]interface{}{
		"files_deleted": deletedCount,
		"days_kept":     daysToKeep,
	})

	return nil
}