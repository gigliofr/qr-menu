package middleware

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "qr_menu_http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "path", "status"})

    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "qr_menu_http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "path"})
)

// MetricsMiddleware registra metriche Prometheus per ogni richiesta
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // usiamo responseWriter definito in logging.go
        rw := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(rw, r)

        duration := time.Since(start).Seconds()

        path := r.URL.Path
        method := r.Method
        status := rw.statusCode

        httpRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
        httpRequestDuration.WithLabelValues(method, path).Observe(duration)
    })
}
