package handlers

import (
	"html/template"
	"log"
	"sync"
)

var (
	// Template caching (moved here to centralize template utilities)
	templateCache    *template.Template
	templateInitOnce sync.Once
)

// SetTemplates initializes the template cache once
func SetTemplates(t *template.Template) {
	templateInitOnce.Do(func() {
		templateCache = t
		log.Printf("Templates cached via handlers/utils.go")
	})
}

// GetTemplates returns cached templates
func GetTemplates() *template.Template {
	return templateCache
}
