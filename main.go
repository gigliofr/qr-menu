package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"qr-menu/db"
	"qr-menu/logger"
	"qr-menu/pkg/app"
)

func main() {
	// Inizializza il logger PRIMA di tutto
	logLevel := logger.INFO
	if lvl := os.Getenv("LOG_LEVEL"); lvl == "DEBUG" {
		logLevel = logger.DEBUG
	}
	
	// Su Railway/Cloud usa directory temporanea, in locale usa ./logs
	logDir := "./logs"
	if os.Getenv("PORT") != "" {
		// In produzione (Railway) usa /tmp per i log
		logDir = "/tmp/logs"
	}
	
	if err := logger.Init(logLevel, logDir); err != nil {
		log.Printf("⚠️ Errore nell'inizializzazione del logger: %v (continuo con log.Println)", err)
	}
	defer logger.Close()
	
	logger.Info("🚀 QR Menu System starting...", map[string]interface{}{
		"version": "1.0.0",
		"env":     os.Getenv("PORT") != "",
	})
	
	// Connetti a MongoDB Atlas (OBBLIGATORIO)
	log.Println("🔄 Connessione a MongoDB Atlas...")
	logger.Info("Connessione a MongoDB Atlas", nil)
	
	if err := db.Connect(); err != nil {
		errMsg := fmt.Sprintf("❌ Errore connessione MongoDB: %v\n\n"+
			"Configura le variabili d'ambiente:\n"+
			"  - MONGODB_URI: connection string MongoDB Atlas\n"+
			"  - MONGODB_CERT_CONTENT: contenuto del certificato PEM (per Railway/Cloud)\n"+
			"  - MONGODB_CERT_PATH: path al file certificato (per sviluppo locale)\n"+
			"  - MONGODB_DB_NAME: nome del database (default: qr-menu)", err)
		log.Fatalf(errMsg)
		logger.Fatal("Errore connessione MongoDB", map[string]interface{}{"error": err.Error()})
	}
	log.Println("✓ MongoDB connesso con successo")
	logger.Info("✅ MongoDB connesso con successo", nil)
	
	// Carica i template HTML (con embed per Railway)
	log.Println("🔄 Caricamento template HTML...")
	InitTemplates()
	
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

	// HTTPS Redirect Middleware (solo in staging/production)
	env := os.Getenv("ENVIRONMENT")
	if env == "production" || env == "staging" {
		router.Use(httpsRedirectMiddleware)
		logger.Info("HTTPS redirect enabled", map[string]interface{}{"env": env})
	}

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

// httpsRedirectMiddleware forza HTTPS in produzione/staging
func httpsRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Railway usa X-Forwarded-Proto header
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			target := "https://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
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
