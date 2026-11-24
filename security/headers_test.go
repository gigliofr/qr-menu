package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultContentSecurityPolicyAllowsTemplateStyles(t *testing.T) {
	policy := DefaultContentSecurityPolicy()

	for _, directive := range []string{
		"style-src 'self' 'unsafe-inline'",
		"style-src-elem 'self' 'unsafe-inline'",
		"style-src-attr 'unsafe-inline'",
		"https://fonts.googleapis.com",
		"https://fonts.gstatic.com",
	} {
		if !strings.Contains(policy, directive) {
			t.Errorf("CSP missing %q: %s", directive, policy)
		}
	}
}

func TestSecurityHeadersMiddlewareSetsFullCSP(t *testing.T) {
	middleware := NewSecurityHeadersMiddleware(DefaultSecurityHeadersConfig())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(recorder, request)

	policy := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(policy, "style-src 'self' 'unsafe-inline'") {
		t.Fatalf("middleware returned incomplete CSP: %q", policy)
	}
}
