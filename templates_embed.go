package main

import (
	"embed"
	"html/template"
	"log"
	"qr-menu/handlers"
)

//go:embed templates/*.html
var templateFS embed.FS

var Templates *template.Template

func InitTemplates() {
	var err error
	Templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Printf("❌ Errore caricamento embedded templates: %v", err)
		// Fallback a filesystem locale
		Templates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			log.Printf("❌ Errore caricamento templates da filesystem: %v", err)
			Templates = nil
		} else {
			log.Printf("✅ Template caricati da filesystem locale")
		}
	} else {
		log.Printf("✅ Template caricati da embedded files (Railway)")
	}
	
	// Passa i template al package handlers
	if Templates != nil {
		handlers.SetTemplates(Templates)
	}
}
