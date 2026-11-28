package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestInjectCSRF(t *testing.T) {
	in := `<html><head><title>x</title></head><body>
<form method="POST" action="/a"><input name="x"></form>
<form method='post'><input type="hidden" name="csrf_token" value="TOK"></form>
<FORM action="/s" method=get><input name="q"></FORM>
<form action="/b" METHOD="post" class="c">
</form></body></html>`
	out := string(injectCSRF([]byte(in), "TOK"))
	if got := strings.Count(out, `name="csrf_token"`); got != 3 {
		t.Fatalf("attesi 3 campi csrf_token (2 iniettati + 1 esistente), trovati %d:\n%s", got, out)
	}
	if !strings.Contains(out, `<head><meta name="csrf-token" content="TOK">`) {
		t.Fatalf("meta csrf-token mancante:\n%s", out)
	}
	if strings.Contains(out, `method=get><input type="hidden" name="csrf_token"`) {
		t.Fatal("i form GET non devono ricevere il token")
	}
	if !strings.Contains(out, `<form method="POST" action="/a"><input type="hidden" name="csrf_token" value="TOK"><input name="x">`) {
		t.Fatalf("token non inserito subito dopo l'apertura del form:\n%s", out)
	}
}

func TestInjectCSRFNoHeadNoForm(t *testing.T) {
	in := "<p>ciao</p>"
	if out := string(injectCSRF([]byte(in), "T")); out != in {
		t.Fatalf("pagina senza form né head modificata: %q", out)
	}
}

func TestLoginLockoutBlocksBeforeDatabase(t *testing.T) {
	key := "bloccato@example.com"
	defer loginLocks.Reset(key)
	for i := 0; i < 5; i++ {
		loginLocks.Fail(key)
	}
	form := url.Values{"username": {"Bloccato@Example.com"}, "password": {"x"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()
	LoginHandler(rw, req) // con il blocco attivo non tocca MongoDB
	if rw.Code != http.StatusTooManyRequests {
		t.Fatalf("atteso 429, ottenuto %d", rw.Code)
	}
}
