package handlers

import (
    "context"
    "html/template"
    "log"
    "net/http"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"

    "github.com/google/uuid"
    "qr-menu/db"
    "qr-menu/models"
)

var csrfMutex sync.Mutex

// renderTemplate renders a named template from the cached templates or fails gracefully
func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    t := GetTemplates()
    if t == nil {
        // fallback: try parsing single file
        tmpl, err := template.ParseFiles("templates/" + name + ".html")
        if err != nil {
            log.Printf("renderTemplate fallback parse error: %v", err)
            http.Error(w, "Template error", http.StatusInternalServerError)
            return
        }
        if err := tmpl.Execute(w, data); err != nil {
            log.Printf("renderTemplate execute error: %v", err)
        }
        return
    }
    if err := t.ExecuteTemplate(w, name+".html", data); err != nil {
        // Log template name + error to help debugging which template failed
        log.Printf("renderTemplate execute cached error for template %s: %v", name+".html", err)
        // Return a minimal error with template name to aid diagnosis (no sensitive data)
        http.Error(w, "Template render error: "+name+"", http.StatusInternalServerError)
    }
}

// setSecurityHeaders adds basic security headers to responses
func setSecurityHeaders(w http.ResponseWriter) {
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
    w.Header().Set("Content-Security-Policy", "default-src 'self'")
}

// generateUniqueRestaurantUsername creates a slug-like username with a short suffix
func generateUniqueRestaurantUsername(ctx context.Context, name string) (string, error) {
    // Basic slug: lowercase, replace non-alnum with hyphens
    slug := strings.ToLower(name)
    re := regexp.MustCompile(`[^a-z0-9]+`)
    slug = re.ReplaceAllString(slug, "-")
    slug = strings.Trim(slug, "-")
    if len(slug) > 24 {
        slug = slug[:24]
    }
    // append short suffix to reduce collisions
    suffix := uuid.New().String()[:6]
    username := slug + "-" + suffix
    return username, nil
}

// cleanupCSRFTokens periodically removes expired CSRF tokens from the in-memory map
func cleanupCSRFTokens() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        now := time.Now()
        csrfMutex.Lock()
        for token, ts := range csrfTokens {
            if now.Sub(ts) > time.Hour {
                delete(csrfTokens, token)
            }
        }
        csrfMutex.Unlock()
    }
}

// generateCSRFToken crea e memorizza un token CSRF temporaneo
func generateCSRFToken() string {
    token := uuid.New().String()
    csrfMutex.Lock()
    csrfTokens[token] = time.Now()
    csrfMutex.Unlock()
    return token
}

// requireValidCSRF verifica la presenza di un token CSRF valido
func requireValidCSRF(w http.ResponseWriter, r *http.Request) bool {
    token := r.FormValue("csrf_token")
    if token == "" {
        token = r.Header.Get("X-CSRF-Token")
    }
    if token == "" {
        http.Error(w, "CSRF token missing", http.StatusBadRequest)
        return false
    }
    csrfMutex.Lock()
    ts, ok := csrfTokens[token]
    csrfMutex.Unlock()
    if !ok || time.Since(ts) > time.Hour {
        http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
        return false
    }
    return true
}

// getBaseURL costruisce la base URL dalla request
func getBaseURL(r *http.Request) string {
    scheme := "http"
    if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
        scheme = "https"
    }
    host := r.Host
    return scheme + "://" + host
}

// ensureRestaurantUsername garantisce che un ristorante abbia un username, lo crea se mancante
func ensureRestaurantUsername(ctx context.Context, restaurant *models.Restaurant) (string, error) {
    if restaurant.Username != "" {
        return restaurant.Username, nil
    }
    username, err := generateUniqueRestaurantUsername(ctx, restaurant.Name)
    if err != nil {
        return "", err
    }
    restaurant.Username = username
    if err := db.MongoInstance.UpdateRestaurant(ctx, restaurant); err != nil {
        return "", err
    }
    return username, nil
}

// createDirectories crea le directory necessarie per il package handlers
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
