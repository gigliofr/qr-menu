package main

import (
	"log"
	"net/http"
	"os"

	"qr-menu/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Crea le directory necessarie se non esistono
	createDirectories()

	// Inizializza il router
	r := mux.NewRouter()

	// Route per servire file statici
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Route pubbliche (non richiedono autenticazione)
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	
	// Route per visualizzazione menu pubblico (non richiedono auth)
	r.HandleFunc("/menu/{id}", handlers.PublicMenuHandler).Methods("GET")
	r.HandleFunc("/r/{username}", handlers.GetActiveMenuHandler).Methods("GET")

	// Route per servire i QR codes (pubblico)
	r.PathPrefix("/qr/").Handler(http.StripPrefix("/qr/", http.FileServer(http.Dir("./static/qrcodes/"))))

	// Route protette (richiedono autenticazione)
	r.HandleFunc("/admin", handlers.RequireAuth(handlers.AdminHandler)).Methods("GET")
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

	// Avvia il server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server QR Menu System avviato su http://localhost:%s", port)
	log.Printf("🔐 Login: http://localhost:%s/login", port)
	log.Printf("📝 Registrazione: http://localhost:%s/register", port)
	log.Printf("⚙️  Interfaccia admin: http://localhost:%s/admin", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
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