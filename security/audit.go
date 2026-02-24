package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AuditEvent represents a security or compliance event
type AuditEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Success    bool                   `json:"success"`
}

// AuditLogger manages audit log collection
type AuditLogger struct {
	mu     sync.RWMutex
	events []AuditEvent
	maxSize int
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(maxSize int) *AuditLogger {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &AuditLogger{
		events:  make([]AuditEvent, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log records an audit event
func (al *AuditLogger) Log(event AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Implement circular buffer
	if len(al.events) >= al.maxSize {
		al.events = al.events[1:]
	}
	al.events = append(al.events, event)
}

// LogAuth logs authentication events
func (al *AuditLogger) LogAuth(userID, action string, success bool, r *http.Request, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  "authentication",
		Method:    r.Method,
		Path:      r.URL.Path,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Success:   success,
		Details:   details,
	}
	al.Log(event)
}

// LogDataAccess logs data access events (GDPR compliance)
func (al *AuditLogger) LogDataAccess(userID, resource, action string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Success:   true,
		Details:   details,
	}
	al.Log(event)
}

// LogDataModification logs data changes (GDPR compliance)
func (al *AuditLogger) LogDataModification(userID, resource, action string, before, after interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Success:   true,
		Details: map[string]interface{}{
			"before": before,
			"after":  after,
		},
	}
	al.Log(event)
}

// LogDeletion logs data deletion (GDPR compliance)
func (al *AuditLogger) LogDeletion(userID, resource string, data interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    "delete",
		Resource:  resource,
		Success:   true,
		Details: map[string]interface{}{
			"deleted_data": data,
		},
	}
	al.Log(event)
}

// GetEvents retrieves audit events with filtering
func (al *AuditLogger) GetEvents(filter func(AuditEvent) bool, limit int) []AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	results := make([]AuditEvent, 0)
	for i := len(al.events) - 1; i >= 0; i-- {
		if filter == nil || filter(al.events[i]) {
			results = append(results, al.events[i])
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results
}

// GetEventsByUser retrieves events for a specific user
func (al *AuditLogger) GetEventsByUser(userID string, limit int) []AuditEvent {
	return al.GetEvents(func(e AuditEvent) bool {
		return e.UserID == userID
	}, limit)
}

// GetEventsByAction retrieves events by action type
func (al *AuditLogger) GetEventsByAction(action string, limit int) []AuditEvent {
	return al.GetEvents(func(e AuditEvent) bool {
		return e.Action == action
	}, limit)
}

// GetEventsInTimeRange retrieves events within a time range
func (al *AuditLogger) GetEventsInTimeRange(start, end time.Time, limit int) []AuditEvent {
	return al.GetEvents(func(e AuditEvent) bool {
		return e.Timestamp.After(start) && e.Timestamp.Before(end)
	}, limit)
}

// ExportJSON exports audit logs as JSON
func (al *AuditLogger) ExportJSON() ([]byte, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	return json.MarshalIndent(al.events, "", "  ")
}

// AuditMiddleware wraps HTTP handlers to log requests
type AuditMiddleware struct {
	logger *AuditLogger
}

// NewAuditMiddleware creates audit logging middleware
func NewAuditMiddleware(logger *AuditLogger) *AuditMiddleware {
	return &AuditMiddleware{logger: logger}
}

// Middleware returns the HTTP middleware
func (am *AuditMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Get user ID from context if available
		userID := r.Header.Get("X-User-ID")

		start := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		// Log the request
		event := AuditEvent{
			Timestamp:  start,
			UserID:     userID,
			Action:     "http_request",
			Resource:   r.URL.Path,
			Method:     r.Method,
			Path:       r.URL.Path,
			IPAddress:  r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			StatusCode: rw.statusCode,
			Success:    rw.statusCode < 400,
			Details: map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"query":       r.URL.RawQuery,
			},
		}
		am.logger.Log(event)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
