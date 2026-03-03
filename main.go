package main

import (
	"log"
	"net/http"
	"os"

	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/pkg/app"
)

func main() {
	// Connetti a MongoDB Atlas (OBBLIGATORIO)
	log.Println("🔄 Connessione a MongoDB Atlas...")
	
	if err := db.Connect(); err != nil {
		log.Fatalf("❌ Errore connessione MongoDB: %v\n\n"+
			"Configura le variabili d'ambiente:\n"+
			"  - MONGODB_URI: connection string MongoDB Atlas\n"+
			"  - MONGODB_CERT_CONTENT: contenuto del certificato PEM (per Railway/Cloud)\n"+
			"  - MONGODB_CERT_PATH: path al file certificato (per sviluppo locale)\n"+
			"  - MONGODB_DB_NAME: nome del database (default: qr-menu)", err)
	}
	log.Println("✓ MongoDB connesso con successo")
	
	defer func() {
		if db.MongoInstance != nil {
			db.MongoInstance.Disconnect()
		}
	}()

	// Prova migrazione da file storage a MongoDB (idempotente)
	if shouldMigrate := os.Getenv("MIGRATE_FROM_FILES"); shouldMigrate == "true" || shouldMigrate == "1" {
		if err := db.MongoInstance.MigrateFromFileStorage(); err != nil {
			log.Printf("⚠️  Errore durante la migrazione: %v (continuando comunque)", err)
		}
	}

	// Configurazione
	cfg := app.DefaultConfig()
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	// Inizializza tutti i servizi
	services, err := app.InitializeServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.Shutdown()

	// Crea le directory necessarie
	createDirectories()

	// Setup router con tutte le route
	router := app.SetupRouter(services)

	// Determina porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log startup
	logger.Info("QR Menu System ready", map[string]interface{}{
		"port":       port,
		"admin_url":  "http://localhost:" + port + "/admin",
		"login_url":  "http://localhost:" + port + "/login",
		"api_docs":   "http://localhost:" + port + "/api/v1/docs",
		"api_health": "http://localhost:" + port + "/api/v1/health",
	})

	// Avvia server
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal("Server failed", map[string]interface{}{"error": err.Error()})
	}
}

func createDirectories() {
	dirs := []string{
		"storage",
		"static/qrcodes",
		"static/css",
		"static/js",
		"templates",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			log.Printf("Errore nella creazione della directory %s: %v", dir, err)
		}
	}
}
