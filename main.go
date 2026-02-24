package main

import (
	"log"
	"net/http"
	"os"

	"qr-menu/logger"
	"qr-menu/pkg/app"
)

func main() {
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
