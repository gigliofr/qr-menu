package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireMetricsToken(t *testing.T) {
	h := requireMetricsToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("metriche")) }))
	call := func(auth string) int {
		req := httptest.NewRequest("GET", "/metrics", nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		return rw.Code
	}
	t.Setenv("METRICS_TOKEN", "")
	if call("Bearer qualsiasi") != 404 {
		t.Fatal("senza METRICS_TOKEN le metriche devono restare nascoste")
	}
	t.Setenv("METRICS_TOKEN", "segreto-lungo")
	if call("") != 404 || call("Bearer sbagliato") != 404 {
		t.Fatal("token mancante o errato accettato")
	}
	if call("Bearer segreto-lungo") != 200 {
		t.Fatal("token corretto rifiutato")
	}
}
