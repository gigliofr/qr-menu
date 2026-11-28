package handlers

import (
	"bytes"

	"github.com/gigliofr/gokit/authx/session"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"qr-menu/db"
	"qr-menu/models"
	"qr-menu/security"
)

// renderTemplate esegue il template e inserisce il token CSRF della richiesta in
// ogni form POST (e in una <meta name="csrf-token"> per il JavaScript).
func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	var buf bytes.Buffer
	t := GetTemplates()
	if t == nil {
		// fallback: try parsing single file
		tmpl, err := template.ParseFiles("templates/" + name + ".html")
		if err != nil {
			log.Printf("renderTemplate fallback parse error: %v", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			log.Printf("renderTemplate execute error: %v", err)
			http.Error(w, "Template render error: "+name, http.StatusInternalServerError)
			return
		}
	} else if err := t.ExecuteTemplate(&buf, name+".html", data); err != nil {
		// Log template name + error to help debugging which template failed
		log.Printf("renderTemplate execute cached error for template %s: %v", name+".html", err)
		// Return a minimal error with template name to aid diagnosis (no sensitive data)
		http.Error(w, "Template render error: "+name+"", http.StatusInternalServerError)
		return
	}
	out := buf.Bytes()
	if tok := csrfToken(r); tok != "" {
		out = injectCSRF(out, tok)
	}
	_, _ = w.Write(out)
}

// setSecurityHeaders adds basic security headers to responses
func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
	w.Header().Set("Content-Security-Policy", security.DefaultContentSecurityPolicy())
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

// csrfToken restituisce il token CSRF della richiesta, gestito dal middleware
// gokit/authx/session.CSRF (cookie + controllo dell'origine su ogni POST).
func csrfToken(r *http.Request) string {
	if r == nil {
		return ""
	}
	return session.CSRFToken(r)
}

var (
	formOpenRe   = regexp.MustCompile(`(?is)<form\b[^>]*>`)
	formMethodRe = regexp.MustCompile(`(?is)\bmethod\s*=\s*["']?post\b`)
	headOpenRe   = regexp.MustCompile(`(?is)<head\b[^>]*>`)
)

// injectCSRF aggiunge il campo csrf_token ai form POST che non lo hanno già e
// la <meta name="csrf-token"> nell'<head>. Il token è base64url: non va escapato.
func injectCSRF(html []byte, token string) []byte {
	field := []byte(`<input type="hidden" name="csrf_token" value="` + token + `">`)
	var out bytes.Buffer
	last := 0
	for _, loc := range formOpenRe.FindAllIndex(html, -1) {
		out.Write(html[last:loc[1]])
		last = loc[1]
		if !formMethodRe.Match(html[loc[0]:loc[1]]) {
			continue
		}
		end := bytes.Index(bytes.ToLower(html[loc[1]:]), []byte("</form"))
		body := html[loc[1]:]
		if end >= 0 {
			body = body[:end]
		}
		if !bytes.Contains(body, []byte(`name="csrf_token"`)) {
			out.Write(field)
		}
	}
	out.Write(html[last:])
	res := out.Bytes()
	if loc := headOpenRe.FindIndex(res); loc != nil {
		meta := []byte(`<meta name="csrf-token" content="` + token + `">`)
		res = append(res[:loc[1]:loc[1]], append(meta, res[loc[1]:]...)...)
	}
	return res
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
