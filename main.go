package main

import (
	"log"
	"net/http"
	"os"

	"qr-menu/analytics"
	"qr-menu/api"
	"qr-menu/backup"
	"qr-menu/handlers"
	"qr-menu/logger"
	"qr-menu/middleware"

	"github.com/gorilla/mux"
)

func main() {
	// Inizializza il sistema di logging
	if err := logger.Init(logger.INFO, "logs"); err != nil {
		log.Fatalf("Errore nell'inizializzazione del logger: %v", err)
	}
	defer logger.Close()

	// Inizializzazione del sistema analytics (usa il singleton globale)
	_ = analytics.GetAnalytics()

	// Inizializzazione del sistema di backup
	bm := backup.GetBackupManager()
	if err := bm.Init("backups", 30); err != nil {
		logger.Warn("Errore nell'inizializzazione del backup manager", map[string]interface{}{"error": err.Error()})
	} else {
		// Avvia il backup automatico giornaliero alle 02:00
		schedule := backup.BackupSchedule{
			Type: "daily",
			Hour: 2,
		}
		if err := bm.StartScheduled(schedule); err != nil {
			logger.Warn("Errore nell'avvio del backup scheduler", map[string]interface{}{"error": err.Error()})
		}
		defer bm.Stop()
	}
	if err := logger.CleanOldLogs(30); err != nil {
		logger.Warn("Errore nella pulizia dei log", map[string]interface{}{"error": err.Error()})
	}

	logger.Info("Avvio QR Menu System", map[string]interface{}{
		"version": "2.0.0",
		"mode":    "production",
	})
	// Crea le directory necessarie se non esistono
	createDirectories()

	// Inizializza il router
	r := mux.NewRouter()

	// Route per servire file statici
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Configura i middleware di logging e sicurezza
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.SecurityMiddleware)
	r.Use(middleware.AuthMiddleware)

	// Route pubbliche (non richiedono autenticazione)
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")

	// Route per visualizzazione menu pubblico (non richiedono auth)
	r.HandleFunc("/menu/{id}", handlers.PublicMenuHandler).Methods("GET")
	r.HandleFunc("/r/{username}", handlers.GetActiveMenuHandler).Methods("GET")

	// Route per servire i QR codes (pubblico)
	r.PathPrefix("/qr/").Handler(http.StripPrefix("/qr/", http.FileServer(http.Dir("./static/qrcodes/"))))

	// Route per tracking analytics (pubblico)
	r.HandleFunc("/api/track/share", handlers.TrackShareHandler).Methods("POST")

	// Route protette (richiedono autenticazione)
	r.HandleFunc("/admin", handlers.RequireAuth(handlers.AdminHandler)).Methods("GET")
	r.HandleFunc("/admin/analytics", handlers.RequireAuth(handlers.AnalyticsDashboardHandler)).Methods("GET")
	r.HandleFunc("/api/analytics", handlers.RequireAuth(handlers.AnalyticsAPIHandler)).Methods("GET")
	r.HandleFunc("/admin/menu/create", handlers.RequireAuth(handlers.CreateMenuHandler)).Methods("GET")
	r.HandleFunc("/admin/menu/create", handlers.RequireAuth(handlers.CreateMenuPostHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}", handlers.RequireAuth(handlers.EditMenuHandler)).Methods("GET")
	r.HandleFunc("/admin/menu/{id}/update", handlers.RequireAuth(handlers.UpdateMenuHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}/complete", handlers.RequireAuth(handlers.CompleteMenuHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}/activate", handlers.RequireAuth(handlers.SetActiveMenuHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}/delete", handlers.RequireAuth(handlers.DeleteMenuHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}/duplicate", handlers.RequireAuth(handlers.DuplicateMenuHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/duplicate", handlers.RequireAuth(handlers.DuplicateItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/edit", handlers.RequireAuth(handlers.EditItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/delete", handlers.RequireAuth(handlers.DeleteItemHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{menuId}/category/{categoryId}/item/{itemId}/upload-image", handlers.RequireAuth(handlers.UploadItemImageHandler)).Methods("POST")
	r.HandleFunc("/admin/menu/{id}/add-item", handlers.RequireAuth(handlers.AddItemHandler)).Methods("POST")
	r.HandleFunc("/menu/{id}/share", handlers.ShareMenuHandler).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET", "POST")

	// Route per l'API JSON (richiedono autenticazione)
	r.HandleFunc("/api/menus", handlers.RequireAuth(handlers.GetMenusHandler)).Methods("GET")
	r.HandleFunc("/api/menu/{id}", handlers.GetMenuHandler).Methods("GET") // Pubblico per compatibilità
	r.HandleFunc("/api/menu", handlers.RequireAuth(handlers.CreateMenuAPIHandler)).Methods("POST")
	r.HandleFunc("/api/menu/{id}/generate-qr", handlers.RequireAuth(handlers.GenerateQRHandler)).Methods("POST")

	// Setup delle nuove API REST v2
	api.SetupAPIRoutes(r)

	// Route per il sistema di backup (richiedono autenticazione)
	r.HandleFunc("/api/backup/create", handlers.RequireAuth(handlers.CreateBackupHandler)).Methods("POST")
	r.HandleFunc("/api/backup/list", handlers.RequireAuth(handlers.ListBackupsHandler)).Methods("GET")
	r.HandleFunc("/api/backup/delete", handlers.RequireAuth(handlers.DeleteBackupHandler)).Methods("DELETE")
	r.HandleFunc("/api/backup/restore", handlers.RequireAuth(handlers.RestoreBackupHandler)).Methods("POST")
	r.HandleFunc("/api/backup/status", handlers.RequireAuth(handlers.GetBackupStatusHandler)).Methods("GET")
	r.HandleFunc("/api/backup/schedule", handlers.RequireAuth(handlers.ScheduleBackupHandler)).Methods("POST")
	r.HandleFunc("/api/backup/stats", handlers.RequireAuth(handlers.GetBackupStatsHandler)).Methods("GET")
	r.HandleFunc("/api/backup/download", handlers.RequireAuth(handlers.DownloadBackupHandler)).Methods("GET")

	// Avvia il server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server QR Menu System avviato", map[string]interface{}{
		"port":       ":" + port,
		"admin_url":  "http://localhost:" + port + "/admin",
		"login_url":  "http://localhost:" + port + "/login",
		"api_docs":   "http://localhost:" + port + "/api/v1/docs",
		"api_health": "http://localhost:" + port + "/api/v1/health",
	})

	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatal("Errore nell'avvio del server", map[string]interface{}{"error": err.Error()})
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
