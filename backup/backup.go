package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qr-menu/logger"
)

// BackupManager gestisce i backup automatici del sistema
type BackupManager struct {
	mu                sync.Mutex
	basePath          string
	maxBackups        int
	compressBackups   bool
	lastBackupTime    time.Time
	isRunning         bool
	backupSchedule    time.Duration
	directoriesBackup []string // Directory da backuppare
}

// BackupMetadata contiene informazioni su un backup
type BackupMetadata struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Size         int64     `json:"size"`
	Status       string    `json:"status"` // success, failed, partial
	Duration     int64     `json:"duration"` // millisecondi
	FileCount    int       `json:"file_count"`
	CompressRate float64   `json:"compress_rate"`
	Hash         string    `json:"hash"` // SHA256 per integrità
}

// BackupSchedule definisce la pianificazione dei backup
type BackupSchedule struct {
	Type      string        // "hourly", "daily", "weekly", "monthly"
	Interval  time.Duration // Intervallo custom
	Hour      int           // Ora del giorno (0-23)
	Day       int           // Giorno della settimana (0-6) per weekly
	DayOfMonth int          // Giorno del mese (1-31) per monthly
}

var (
	defaultManager *BackupManager
	once           sync.Once
)

// GetBackupManager restituisce il singleton BackupManager
func GetBackupManager() *BackupManager {
	once.Do(func() {
		defaultManager = &BackupManager{
			basePath:       "backups",
			maxBackups:     30, // Mantiene ultimi 30 backup
			compressBackups: true,
			directoriesBackup: []string{
				"storage",
				"logs",
				"analytics",
			},
		}
	})
	return defaultManager
}

// Init inizializza il backup manager
func (bm *BackupManager) Init(basePath string, maxBackups int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.basePath = basePath
	bm.maxBackups = maxBackups

	// Crea la directory per i backup se non esiste
	if err := os.MkdirAll(bm.basePath, 0755); err != nil {
		return fmt.Errorf("errore creazione directory backup: %w", err)
	}

	logger.Info("Backup manager inizializzato", map[string]interface{}{
		"path":        bm.basePath,
		"max_backups": bm.maxBackups,
		"compress":    bm.compressBackups,
	})

	return nil
}

// StartScheduled avvia il backup automatico schedulato
func (bm *BackupManager) StartScheduled(schedule BackupSchedule) error {
	bm.mu.Lock()
	if bm.isRunning {
		bm.mu.Unlock()
		return fmt.Errorf("backup scheduler già in esecuzione")
	}
	bm.isRunning = true
	bm.mu.Unlock()

	// Calcola il prossimo tempo di backup
	nextBackup := bm.calculateNextBackupTime(schedule)
	
	logger.Info("Backup scheduler avviato", map[string]interface{}{
		"schedule": schedule.Type,
		"next_backup": nextBackup,
	})

	// Goroutine per scheduling
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Controlla ogni minuto
		defer ticker.Stop()

		for range ticker.C {
			if time.Now().After(nextBackup) {
				// È ora di fare il backup
				backupID, err := bm.CreateBackup()
				if err != nil {
					logger.Error("Errore nel backup automatico", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					logger.Info("Backup automatico completato", map[string]interface{}{
						"backup_id": backupID,
					})
				}
				
				// Calcola il prossimo backup
				nextBackup = bm.calculateNextBackupTime(schedule)
			}
		}
	}()

	return nil
}

// CreateBackup crea un backup manuale
func (bm *BackupManager) CreateBackup() (string, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backupID := fmt.Sprintf("backup-%d", time.Now().Unix())
	startTime := time.Now()

	logger.Info("Inizio backup", map[string]interface{}{
		"backup_id": backupID,
	})

	// Crea un file zip contenente tutti i dati
	var zipPath string
	if bm.compressBackups {
		zipPath = filepath.Join(bm.basePath, backupID+".zip")
	} else {
		zipPath = filepath.Join(bm.basePath, backupID)
	}

	if bm.compressBackups {
		err := bm.createCompressedBackup(zipPath, backupID)
		if err != nil {
			logger.Error("Errore nel backup compresso", map[string]interface{}{
				"backup_id": backupID,
				"error": err.Error(),
			})
			return "", err
		}
	} else {
		err := bm.createUncompressedBackup(zipPath, backupID)
		if err != nil {
			logger.Error("Errore nel backup non compresso", map[string]interface{}{
				"backup_id": backupID,
				"error": err.Error(),
			})
			return "", err
		}
	}

	// Registra i metadati del backup
	duration := time.Since(startTime).Milliseconds()
	metadata := BackupMetadata{
		ID:        backupID,
		Timestamp: startTime,
		Status:    "success",
		Duration:  duration,
	}

	bm.lastBackupTime = startTime

	// Pulisci i backup vecchi
	err := bm.cleanupOldBackups()
	if err != nil {
		logger.Warn("Errore nella pulizia backup vecchi", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Backup completato", map[string]interface{}{
		"backup_id":     backupID,
		"duration_ms":   duration,
		"compressed":    bm.compressBackups,
	})

	// Salva i metadati
	bm.saveBackupMetadata(metadata)

	return backupID, nil
}

// createCompressedBackup crea un backup compresso
func (bm *BackupManager) createCompressedBackup(zipPath string, backupID string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("errore creazione zip: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileCount := 0

	// Aggiungi ogni directory al backup
	for _, dir := range bm.directoriesBackup {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Salta le directory
			if info.IsDir() {
				return nil
			}

			// Apri il file
			fileData, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fileData.Close()

			// Aggiungi allo zip
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = filepath.Join(backupID, path)
			header.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, fileData)
			fileCount++
			return err
		})

		if err != nil {
			return fmt.Errorf("errore durante backup di %s: %w", dir, err)
		}
	}

	return nil
}

// createUncompressedBackup crea un backup non compresso (copia)
func (bm *BackupManager) createUncompressedBackup(backupDir string, backupID string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("errore creazione directory backup: %w", err)
	}

	for _, dir := range bm.directoriesBackup {
		// Copia la directory
		err := bm.copyDirectory(dir, filepath.Join(backupDir, dir))
		if err != nil {
			return fmt.Errorf("errore copia directory %s: %w", dir, err)
		}
	}

	return nil
}

// copyDirectory copia una intera directory
func (bm *BackupManager) copyDirectory(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := bm.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := bm.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copia un file
func (bm *BackupManager) copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// RestoreBackup ripristina un backup
func (bm *BackupManager) RestoreBackup(backupID string, restorePath string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	logger.Info("Inizio restore backup", map[string]interface{}{
		"backup_id":     backupID,
		"restore_path": restorePath,
	})

	// Trova il file backup
	var backupFile string
	backups, err := os.ReadDir(bm.basePath)
	if err != nil {
		return fmt.Errorf("errore lettura directory backup: %w", err)
	}

	for _, backup := range backups {
		if strings.Contains(backup.Name(), backupID) {
			backupFile = filepath.Join(bm.basePath, backup.Name())
			break
		}
	}

	if backupFile == "" {
		return fmt.Errorf("backup non trovato: %s", backupID)
	}

	// Se è un file zip, estrai
	if strings.HasSuffix(backupFile, ".zip") {
		err := bm.extractZipBackup(backupFile, restorePath)
		if err != nil {
			logger.Error("Errore estrazione backup", map[string]interface{}{
				"backup_id": backupID,
				"error": err.Error(),
			})
			return err
		}
	} else {
		// Altrimenti è una directory, copia da lì
		err := bm.copyDirectory(backupFile, restorePath)
		if err != nil {
			logger.Error("Errore copia backup", map[string]interface{}{
				"backup_id": backupID,
				"error": err.Error(),
			})
			return err
		}
	}

	logger.Info("Restore backup completato", map[string]interface{}{
		"backup_id": backupID,
	})

	return nil
}

// extractZipBackup estrae un backup compresso
func (bm *BackupManager) extractZipBackup(zipPath string, destPath string) error {
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("errore apertura zip: %w", err)
	}
	defer zipFile.Close()

	stat, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("errore stat zip: %w", err)
	}

	zipReader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return fmt.Errorf("errore lettura zip: %w", err)
	}

	for _, file := range zipReader.File {
		path := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		inFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, inFile)
		outFile.Close()
		inFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// ListBackups elenca tutti i backup disponibili
func (bm *BackupManager) ListBackups() ([]BackupMetadata, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	var backups []BackupMetadata

	entries, err := os.ReadDir(bm.basePath)
	if err != nil {
		return nil, fmt.Errorf("errore lettura directory backup: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || (bm.compressBackups && !strings.HasSuffix(entry.Name(), ".zip")) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		metadata := BackupMetadata{
			ID:        strings.TrimSuffix(entry.Name(), ".zip"),
			Timestamp: info.ModTime(),
			Size:      info.Size(),
			Status:    "success",
		}

		backups = append(backups, metadata)
	}

	return backups, nil
}

// DeleteBackup elimina uno specifico backup
func (bm *BackupManager) DeleteBackup(backupID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backups, err := os.ReadDir(bm.basePath)
	if err != nil {
		return fmt.Errorf("errore lettura directory backup: %w", err)
	}

	for _, backup := range backups {
		if strings.Contains(backup.Name(), backupID) {
			backupPath := filepath.Join(bm.basePath, backup.Name())
			err := os.RemoveAll(backupPath)
			if err != nil {
				return fmt.Errorf("errore eliminazione backup: %w", err)
			}
			logger.Info("Backup eliminato", map[string]interface{}{
				"backup_id": backupID,
			})
			return nil
		}
	}

	return fmt.Errorf("backup non trovato: %s", backupID)
}

// cleanupOldBackups elimina i backup più vecchi oltre il limite
func (bm *BackupManager) cleanupOldBackups() error {
	backups, err := bm.ListBackups()
	if err != nil {
		return err
	}

	// Ordina per timestamp (più recenti primo)
	if len(backups) > bm.maxBackups {
		// Elimina i più vecchi
		for i := bm.maxBackups; i < len(backups); i++ {
			err := bm.DeleteBackup(backups[i].ID)
			if err != nil {
				logger.Warn("Errore eliminazione backup vecchio", map[string]interface{}{
					"backup_id": backups[i].ID,
					"error": err.Error(),
				})
			}
		}
	}

	return nil
}

// calculateNextBackupTime calcola il prossimo tempo di backup
func (bm *BackupManager) calculateNextBackupTime(schedule BackupSchedule) time.Time {
	now := time.Now()

	switch schedule.Type {
	case "hourly":
		return now.Add(1 * time.Hour)
	case "daily":
		next := now.AddDate(0, 0, 1)
		next = time.Date(next.Year(), next.Month(), next.Day(), schedule.Hour, 0, 0, 0, next.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return next
	case "weekly":
		next := now.AddDate(0, 0, 1)
		for next.Weekday() != time.Weekday(schedule.Day) {
			next = next.AddDate(0, 0, 1)
		}
		next = time.Date(next.Year(), next.Month(), next.Day(), schedule.Hour, 0, 0, 0, next.Location())
		return next
	case "monthly":
		next := time.Date(now.Year(), now.Month(), schedule.DayOfMonth, schedule.Hour, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}
		return next
	default:
		return now.Add(schedule.Interval)
	}
}

// saveBackupMetadata salva i metadati del backup
func (bm *BackupManager) saveBackupMetadata(metadata BackupMetadata) {
	// In futuro, salvare in database o file JSON
	// Per ora, semplicemente logghiamo
	logger.Info("Metadati backup salvati", map[string]interface{}{
		"backup_id": metadata.ID,
		"timestamp": metadata.Timestamp,
		"size": metadata.Size,
	})
}

// GetLastBackupTime restituisce il timestamp dell'ultimo backup
func (bm *BackupManager) GetLastBackupTime() time.Time {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.lastBackupTime
}

// GetTotalBackupSize restituisce la dimensione totale di tutti i backup
func (bm *BackupManager) GetTotalBackupSize() int64 {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	var totalSize int64
	entries, err := os.ReadDir(bm.basePath)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	return totalSize
}

// Stop ferma il backup scheduler
func (bm *BackupManager) Stop() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.isRunning = false
	logger.Info("Backup scheduler fermato", nil)
}
