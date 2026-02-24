package security

import (
	"encoding/json"
	"fmt"
	"time"
)

// GDPRManager handles GDPR compliance operations
type GDPRManager struct {
	auditLogger *AuditLogger
}

// NewGDPRManager creates a new GDPR manager
func NewGDPRManager(auditLogger *AuditLogger) *GDPRManager {
	return &GDPRManager{
		auditLogger: auditLogger,
	}
}

// ConsentRecord tracks user consent
type ConsentRecord struct {
	UserID      string    `json:"user_id"`
	ConsentType string    `json:"consent_type"`
	Granted     bool      `json:"granted"`
	Timestamp   time.Time `json:"timestamp"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
}

// DataExportRequest represents a GDPR data export request
type DataExportRequest struct {
	UserID      string    `json:"user_id"`
	RequestedAt time.Time `json:"requested_at"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	Format      string    `json:"format"` // json, csv
}

// DataDeletionRequest represents a GDPR right to be forgotten request
type DataDeletionRequest struct {
	UserID      string    `json:"user_id"`
	RequestedAt time.Time `json:"requested_at"`
	ScheduledAt time.Time `json:"scheduled_at"` // 30-day grace period
	Status      string    `json:"status"`       // pending, scheduled, completed, cancelled
	Reason      string    `json:"reason,omitempty"`
}

// UserDataExport contains all user data for GDPR export
type UserDataExport struct {
	User         interface{}     `json:"user"`
	Restaurants  []interface{}   `json:"restaurants"`
	Menus        []interface{}   `json:"menus"`
	Orders       []interface{}   `json:"orders,omitempty"`
	Analytics    interface{}     `json:"analytics,omitempty"`
	AuditLogs    []AuditEvent    `json:"audit_logs"`
	Consents     []ConsentRecord `json:"consents"`
	ExportedAt   time.Time       `json:"exported_at"`
	ExportFormat string          `json:"export_format"`
}

// Consent types
const (
	ConsentMarketing   = "marketing"
	ConsentAnalytics   = "analytics"
	ConsentDataSharing = "data_sharing"
	ConsentCookies     = "cookies"
)

var consentStore = make(map[string][]ConsentRecord)

// RecordConsent records a user's consent
func (gm *GDPRManager) RecordConsent(record ConsentRecord) error {
	if record.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if record.ConsentType == "" {
		return fmt.Errorf("consent_type is required")
	}
	
	record.Timestamp = time.Now()
	
	// Store consent
	if _, exists := consentStore[record.UserID]; !exists {
		consentStore[record.UserID] = make([]ConsentRecord, 0)
	}
	consentStore[record.UserID] = append(consentStore[record.UserID], record)
	
	// Audit log
	gm.auditLogger.LogDataAccess(record.UserID, "consent", "record", map[string]interface{}{
		"consent_type": record.ConsentType,
		"granted":      record.Granted,
	})
	
	return nil
}

// GetConsents retrieves all consents for a user
func (gm *GDPRManager) GetConsents(userID string) []ConsentRecord {
	if records, exists := consentStore[userID]; exists {
		return records
	}
	return []ConsentRecord{}
}

// HasConsent checks if user has granted specific consent
func (gm *GDPRManager) HasConsent(userID, consentType string) bool {
	records, exists := consentStore[userID]
	if !exists {
		return false
	}
	
	// Get latest consent for this type
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].ConsentType == consentType {
			return records[i].Granted
		}
	}
	return false
}

// ExportUserData generates a complete data export for a user
func (gm *GDPRManager) ExportUserData(userID string, user interface{}, restaurants, menus, orders []interface{}, analytics interface{}) (*UserDataExport, error) {
	// Get audit logs for this user
	auditLogs := gm.auditLogger.GetEventsByUser(userID, 1000)
	
	// Get consents
	consents := gm.GetConsents(userID)
	
	export := &UserDataExport{
		User:         user,
		Restaurants:  restaurants,
		Menus:        menus,
		Orders:       orders,
		Analytics:    analytics,
		AuditLogs:    auditLogs,
		Consents:     consents,
		ExportedAt:   time.Now(),
		ExportFormat: "json",
	}
	
	// Audit log the export
	gm.auditLogger.LogDataAccess(userID, "user_data", "export", map[string]interface{}{
		"export_timestamp": export.ExportedAt,
		"records_count":    len(auditLogs),
	})
	
	return export, nil
}

// ExportUserDataJSON exports user data as JSON
func (gm *GDPRManager) ExportUserDataJSON(userID string, user interface{}, restaurants, menus, orders []interface{}, analytics interface{}) ([]byte, error) {
	export, err := gm.ExportUserData(userID, user, restaurants, menus, orders, analytics)
	if err != nil {
		return nil, err
	}
	
	return json.MarshalIndent(export, "", "  ")
}

var deletionRequests = make(map[string]*DataDeletionRequest)

// RequestDataDeletion submits a data deletion request (right to be forgotten)
func (gm *GDPRManager) RequestDataDeletion(userID, reason string) (*DataDeletionRequest, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	
	// Check if already requested
	if req, exists := deletionRequests[userID]; exists && req.Status != "cancelled" {
		return nil, fmt.Errorf("deletion already requested")
	}
	
	now := time.Now()
	request := &DataDeletionRequest{
		UserID:      userID,
		RequestedAt: now,
		ScheduledAt: now.Add(30 * 24 * time.Hour), // 30-day grace period
		Status:      "scheduled",
		Reason:      reason,
	}
	
	deletionRequests[userID] = request
	
	// Audit log
	gm.auditLogger.Log(AuditEvent{
		Timestamp: now,
		UserID:    userID,
		Action:    "data_deletion_requested",
		Resource:  "user_data",
		Success:   true,
		Details: map[string]interface{}{
			"scheduled_at": request.ScheduledAt,
			"reason":       reason,
		},
	})
	
	return request, nil
}

// CancelDataDeletion cancels a pending deletion request
func (gm *GDPRManager) CancelDataDeletion(userID string) error {
	req, exists := deletionRequests[userID]
	if !exists {
		return fmt.Errorf("no deletion request found")
	}
	
	if req.Status == "completed" {
		return fmt.Errorf("deletion already completed")
	}
	
	req.Status = "cancelled"
	
	gm.auditLogger.Log(AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    "data_deletion_cancelled",
		Resource:  "user_data",
		Success:   true,
	})
	
	return nil
}

// GetDeletionRequest retrieves a deletion request
func (gm *GDPRManager) GetDeletionRequest(userID string) (*DataDeletionRequest, error) {
	req, exists := deletionRequests[userID]
	if !exists {
		return nil, fmt.Errorf("no deletion request found")
	}
	return req, nil
}

// ProcessScheduledDeletions processes all scheduled deletions that are due
func (gm *GDPRManager) ProcessScheduledDeletions() []string {
	now := time.Now()
	deleted := make([]string, 0)
	
	for userID, req := range deletionRequests {
		if req.Status == "scheduled" && now.After(req.ScheduledAt) {
			// Mark as completed (actual deletion would happen here)
			req.Status = "completed"
			deleted = append(deleted, userID)
			
			gm.auditLogger.Log(AuditEvent{
				Timestamp: now,
				UserID:    userID,
				Action:    "data_deletion_completed",
				Resource:  "user_data",
				Success:   true,
			})
		}
	}
	
	return deleted
}

// AnonymizeData pseudonymizes sensitive data for analytics
func (gm *GDPRManager) AnonymizeData(data map[string]interface{}) map[string]interface{} {
	anonymized := make(map[string]interface{})
	
	sensitiveFields := []string{"email", "phone", "address", "name", "ip_address"}
	
	for key, value := range data {
		isSensitive := false
		for _, field := range sensitiveFields {
			if key == field {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			anonymized[key] = "[REDACTED]"
		} else {
			anonymized[key] = value
		}
	}
	
	return anonymized
}
