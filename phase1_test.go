package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"qr-menu/pkg/config"
	"qr-menu/pkg/errors"
	httputil "qr-menu/pkg/http"
)

// TestErrorPackage tests the error wrapper functionality
func TestErrorPackage(t *testing.T) {
	tests := []struct {
		name        string
		createErr   func() *errors.AppError
		expectedCode string
		expectedSeverity string
	}{
		{
			name:         "Validation Error",
			createErr:    func() *errors.AppError { return errors.ValidationError("Invalid email") },
			expectedCode: errors.CodeValidation,
			expectedSeverity: errors.SeverityWarning,
		},
		{
			name:         "Not Found Error",
			createErr:    func() *errors.AppError { return errors.NotFound("Restaurant") },
			expectedCode: errors.CodeNotFound,
			expectedSeverity: errors.SeverityWarning,
		},
		{
			name:         "Database Error",
			createErr:    func() *errors.AppError { return errors.DatabaseError("Connection failed") },
			expectedCode: errors.CodeDatabaseConnection,
			expectedSeverity: errors.SeverityError,
		},
		{
			name:         "Custom Error with Details",
			createErr: func() *errors.AppError {
				return errors.New(errors.CodeIOError, "File read failed", errors.SeverityError).
					WithDetails("File not found: /path/to/file.txt")
			},
			expectedCode: errors.CodeIOError,
			expectedSeverity: errors.SeverityError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.createErr()

			if appErr.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, appErr.Code)
			}

			if appErr.Severity != tt.expectedSeverity {
				t.Errorf("Expected severity %s, got %s", tt.expectedSeverity, appErr.Severity)
			}

			// Test error string representation
			errStr := appErr.Error()
			if errStr == "" {
				t.Error("Error string should not be empty")
			}
		})
	}
}

// TestConfigPackage tests the config loading functionality
func TestConfigPackage(t *testing.T) {
	cfg := config.Load()

	if cfg.Server.Port == 0 {
		t.Error("Server port should not be 0")
	}

	if cfg.Server.Host == "" {
		t.Error("Server host should not be empty")
	}

	if cfg.Database.Engine == "" {
		t.Error("Database engine should not be empty")
	}

	if cfg.Backup.MaxBackups == 0 {
		t.Error("Backup max backups should not be 0")
	}

	if !cfg.IsDevelopment() && !cfg.IsProduction() && !cfg.IsStaging() {
		t.Error("Environment should be dev, staging, or prod")
	}

	t.Logf("Config loaded: Server %s:%d, DB Engine: %s", 
		cfg.Server.Host, cfg.Server.Port, cfg.Database.Engine)
}

// TestHTTPResponsePackage tests the HTTP response helpers
func TestHTTPResponsePackage(t *testing.T) {
	tests := []struct {
		name           string
		sendResponse   func(w http.ResponseWriter)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success Response",
			sendResponse: func(w http.ResponseWriter) {
				_ = httputil.Success(w, "Operation successful", map[string]string{"id": "123"})
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "Created Response",
			sendResponse: func(w http.ResponseWriter) {
				_ = httputil.Created(w, "Resource created", map[string]string{"id": "456"})
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "success",
		},
		{
			name: "Not Found Error",
			sendResponse: func(w http.ResponseWriter) {
				_ = httputil.NotFound(w, "Menu")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "error",
		},
		{
			name: "Bad Request Error",
			sendResponse: func(w http.ResponseWriter) {
				_ = httputil.BadRequest(w, "Invalid input")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.sendResponse(w)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if !bytes.Contains([]byte(body), []byte(tt.expectedBody)) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, body)
			}

			// Verify it's valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Response body is not valid JSON: %v", err)
			}

			if _, ok := response["status"]; !ok {
				t.Error("Response should contain 'status' field")
			}
		})
	}
}

// TestErrorResponseIntegration tests error wrapping with HTTP responses
func TestErrorResponseIntegration(t *testing.T) {
	w := httptest.NewRecorder()

	// Create an error and send it as HTTP response
	appErr := errors.ValidationError("Email is required").WithDetails("User must provide a valid email address")
	_ = httputil.Error(w, appErr)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if response["status"] != "error" {
		t.Error("Response status should be 'error'")
	}

	if response["message"] == nil {
		t.Error("Response should contain error message")
	}

	if response["error"] == nil {
		t.Error("Response should contain error details")
	}
}

// TestConfigWithEnvironmentVariables tests loading config with env vars
func TestConfigWithEnvironmentVariables(t *testing.T) {
	// Note: this test just verifies the config loads without panicking
	cfg := config.Load()

	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify all major sections are initialized
	if cfg.Server.Port == 0 {
		t.Error("Server config not initialized")
	}
	if cfg.Database.Engine == "" {
		t.Error("Database config not initialized")
	}
	if cfg.Backup.MaxBackups == 0 {
		t.Error("Backup config not initialized")
	}
	if cfg.Notifications.Workers == 0 {
		t.Error("Notifications config not initialized")
	}

	t.Log("✅ All config sections initialized successfully")
}
