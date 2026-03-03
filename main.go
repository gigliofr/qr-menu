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
	// MongoDB è OBBLIGATORIO - l'app non parte senza
	log.Println("🔌 Connessione a MongoDB Atlas...")
	
	// Verifica configurazione
	if os.Getenv("MONGODB_URI") == "" {
		log.Fatal("❌ MONGODB_URI non configurato - impossibile avviare l'applicazione")
	}
	
	if os.Getenv("MONGODB_CERT_CONTENT") == "" && os.Getenv("MONGODB_CERT_PATH") == "" {
		log.Fatal("❌ Certificato MongoDB non configurato - imposta MONGODB_CERT_CONTENT o MONGODB_CERT_PATH")
	}
	
	// Connetti a MongoDB Atlas (obbligatorio)
	if err := db.Connect(); err != nil {
		log.Fatalf("❌ Errore connessione MongoDB: %v\nL'applicazione richiede MongoDB per funzionare.", err)
	}
	
	log.Println("✅ MongoDB connesso con successo")
	
	defer func() {
		if db.MongoInstance != nil {
			db.MongoInstance.Disconnect()
		}
	}()

	// Prova migrazione da file storage a MongoDB (idempotente)
	if shouldMigrate := os.Getenv("MIGRATE_FROM_FILES"); shouldMigrate == "true" || shouldMigrate == "1" {
		log.Println("🔄 Avvio migrazione da file storage a MongoDB...")
		if err := db.MongoInstance.MigrateFromFileStorage(); err != nil {
			log.Printf("⚠️  Errore durante la migrazione: %v (continuando comunque)", err)
		} else {
			log.Println("✅ Migrazione completata")
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
