package middleware

import (
	"net/http"
	"time"

	"qr-menu/pkg/metrics"
)

// BusinessMetricsMiddleware tracks business-level events
func BusinessMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		operation := r.Method + " " + r.URL.Path

		next.ServeHTTP(w, r)

		// Record metrics for operation duration
		duration := time.Since(start).Seconds()
		metrics.OperationDuration.WithLabelValues(operation).Observe(duration)

		// Track specific business operations
		if r.URL.Path == "/api/v1/menus/create" {
			metrics.MenusCreated.WithLabelValues("").Inc()
		} else if r.URL.Path == "/api/v1/qrcode" {
			metrics.QRCodesGenerated.WithLabelValues("").Inc()
		}
	})
}
