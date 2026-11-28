package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gigliofr/gokit/authx/session"
)

// Percorso reale: template veri + middleware CSRF di gokit, come in pkg/app.
func TestLoginPageCSRFFlow(t *testing.T) {
	tmpl, err := template.ParseGlob("../templates/*.html")
	if err != nil {
		t.Skipf("template non disponibili: %v", err)
	}
	old := GetTemplates()
	SetTemplates(tmpl)
	defer SetTemplates(old)

	reached := false
	h := session.CSRF(session.CSRFOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			renderTemplate(w, r, "login", nil)
			return
		}
		reached = true
	}))

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, httptest.NewRequest("GET", "https://qr.example/login", nil))
	if rw.Code != 200 {
		t.Fatalf("GET /login: %d %s", rw.Code, rw.Body.String())
	}
	var cookie *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == "csrf_token" {
			cookie = c
		}
	}
	m := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`).FindStringSubmatch(rw.Body.String())
	if cookie == nil || m == nil || m[1] != cookie.Value {
		t.Fatalf("il form di login deve contenere il token del cookie (cookie=%v, campo=%v)", cookie, m)
	}

	post := func(tok string) int {
		form := url.Values{"username": {"a"}, "password": {"b"}}
		if tok != "" {
			form.Set("csrf_token", tok)
		}
		req := httptest.NewRequest("POST", "https://qr.example/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "https://qr.example")
		req.AddCookie(cookie)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		return rw.Code
	}
	if post("") != http.StatusForbidden || reached {
		t.Fatal("POST senza token deve essere bloccato")
	}
	if post(m[1]); !reached {
		t.Fatal("POST con il token del form deve passare")
	}
}
