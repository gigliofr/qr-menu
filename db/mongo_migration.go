package db

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"qr-menu/models"
)

// MigrateFromFileStorage migra i dati da storage JSON a MongoDB
func (m *MongoClient) MigrateFromFileStorage() error {
	log.Println("🔄 Inizio migrazione da file storage a MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Migra restaurants
	if err := m.migrateRestaurants(ctx); err != nil {
		return err
	}

	// Migra menus
	if err := m.migrateMenus(ctx); err != nil {
		return err
	}

	// Migra sessions
	if err := m.migrateSessions(ctx); err != nil {
		return err
	}

	log.Println("✓ Migrazione completata con successo!")
	return nil
}

// migrateRestaurants migra i ristoranti da JSON a MongoDB
func (m *MongoClient) migrateRestaurants(ctx context.Context) error {
	files, err := filepath.Glob("storage/restaurant_*.json")
	if err != nil {
		return err
	}

	successCount := 0
	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("⚠️  Errore apertura %s: %v", filename, err)
			continue
		}

		var restaurant models.Restaurant
		if err := json.NewDecoder(file).Decode(&restaurant); err != nil {
			log.Printf("⚠️  Errore decode %s: %v", filename, err)
			file.Close()
			continue
		}
		file.Close()

		// Verifica se esiste già
		existing, err := m.GetRestaurantByID(ctx, restaurant.ID)
		if err == nil && existing != nil {
			log.Printf("⏭️  Restaurant %s già exists, skip", restaurant.ID)
			continue
		}

		// Salva in MongoDB
		if err := m.CreateRestaurant(ctx, &restaurant); err != nil {
			log.Printf("⚠️  Errore save restaurant %s: %v", restaurant.ID, err)
			continue
		}

		successCount++
		log.Printf("✓ Migrato restaurant: %s (%s)", restaurant.Name, restaurant.ID)
	}

	log.Printf("📊 Restaurant migrati: %d/%d", successCount, len(files))
	return nil
}

// migrateMenus migra i menu da JSON a MongoDB
func (m *MongoClient) migrateMenus(ctx context.Context) error {
	files, err := filepath.Glob("storage/menu_*.json")
	if err != nil {
		return err
	}

	successCount := 0
	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("⚠️  Errore apertura %s: %v", filename, err)
			continue
		}

		var menu models.Menu
		if err := json.NewDecoder(file).Decode(&menu); err != nil {
			log.Printf("⚠️  Errore decode %s: %v", filename, err)
			file.Close()
			continue
		}
		file.Close()

		// Verifica se esiste già
		existing, err := m.GetMenuByID(ctx, menu.ID)
		if err == nil && existing != nil {
			log.Printf("⏭️  Menu %s già exists, skip", menu.ID)
			continue
		}

		// Salva in MongoDB
		if err := m.CreateMenu(ctx, &menu); err != nil {
			log.Printf("⚠️  Errore save menu %s: %v", menu.ID, err)
			continue
		}

		successCount++
		log.Printf("✓ Migrato menu: %s (%s)", menu.Name, menu.ID)
	}

	log.Printf("📊 Menu migrati: %d/%d", successCount, len(files))
	return nil
}

// migrateSessions migra le sessioni da JSON a MongoDB
func (m *MongoClient) migrateSessions(ctx context.Context) error {
	files, err := filepath.Glob("storage/session_*.json")
	if err != nil {
		return err
	}

	successCount := 0
	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("⚠️  Errore apertura %s: %v", filename, err)
			continue
		}

		var session models.Session
		if err := json.NewDecoder(file).Decode(&session); err != nil {
			log.Printf("⚠️  Errore decode %s: %v", filename, err)
			file.Close()
			continue
		}
		file.Close()

		// Salta sessioni scadute
		if time.Since(session.LastAccessed) > 24*time.Hour {
			log.Printf("⏭️  Session %s scaduta, skip", session.ID)
			continue
		}

		// Verifica se esiste già
		existing, err := m.GetSessionByID(ctx, session.ID)
		if err == nil && existing != nil {
			log.Printf("⏭️  Session %s già exists, skip", session.ID)
			continue
		}

		// Salva in MongoDB
		if err := m.CreateSession(ctx, &session); err != nil {
			log.Printf("⚠️  Errore save session %s: %v", session.ID, err)
			continue
		}

		successCount++
		log.Printf("✓ Migrato session: %s", session.ID)
	}

	log.Printf("📊 Session migrationcompleted: %d/%d", successCount, len(files))
	return nil
}

// BackupToJSON esporta i dati da MongoDB a JSON (backup)
func (m *MongoClient) BackupToJSON(backupDir string) error {
	log.Println("💾 Backup in corso...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Crea directory backup
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Backup restaurants
	restaurants, err := m.GetAllRestaurants(ctx)
	if err != nil {
		return err
	}

	for _, rest := range restaurants {
		data, _ := json.MarshalIndent(rest, "", "  ")
		filename := filepath.Join(backupDir, "restaurant_"+rest.ID+".json")
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			log.Printf("⚠️  Errore backup restaurant: %v", err)
		}
	}

	// Backup menus
	menus, err := m.GetAllMenus(ctx)
	if err != nil {
		return err
	}

	for _, menu := range menus {
		data, _ := json.MarshalIndent(menu, "", "  ")
		filename := filepath.Join(backupDir, "menu_"+menu.ID+".json")
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			log.Printf("⚠️  Errore backup menu: %v", err)
		}
	}

	log.Printf("✓ Backup completato: %d restaurants, %d menus", len(restaurants), len(menus))
	return nil
}
