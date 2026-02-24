package handlers

import (
	"net/http"

	"qr-menu/pwa"
)

// ManifestHandler serve il manifest.json con i corretti header
func ManifestHandler(w http.ResponseWriter, r *http.Request) {
	pm := pwa.GetPWAManager()
	manifest, err := pm.GetManifest()
	if err != nil {
		http.Error(w, "Manifest not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(manifest))
}

// ServiceWorkerHandler serve il service worker con i corretti header
func ServiceWorkerHandler(w http.ResponseWriter, r *http.Request) {
	pm := pwa.GetPWAManager()
	sw, err := pm.GetServiceWorker()
	if err != nil {
		http.Error(w, "Service Worker not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Service-Worker-Allowed", "/")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sw))
}

// OfflineHandler serve la pagina offline
func OfflineHandler(w http.ResponseWriter, r *http.Request) {
	pm := pwa.GetPWAManager()
	offlineHTML := pm.GetOfflinePageHTML()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(offlineHTML))
}

// PWAHeadersMiddleware aggiunge gli header necessari per il PWA
func PWAHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Aggiungi header per HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Aggiungi header per PWA
		w.Header().Set("X-UA-Compatible", "IE=edge")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Aggiungi link per manifest nel header di risposta
		if r.Method == "GET" && (r.RequestURI == "/" || r.RequestURI == "/admin") {
			w.Header().Set("Link", "</manifest.json>; rel=\"manifest\"")
		}
		
		next.ServeHTTP(w, r)
	})
}

// HealthCheckHandler per verificare la connessione (usado dal service worker)
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
