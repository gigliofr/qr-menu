package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Severity levels for errors
const (
	SeverityFatal   = "FATAL"
	SeverityError   = "ERROR"
	SeverityWarning = "WARN"
	SeverityInfo    = "INFO"
)

// Common error codes
const (
	CodeValidation          = "VALIDATION_ERROR"
	CodeNotFound            = "NOT_FOUND"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeConflict            = "CONFLICT"
	CodeInternalServer      = "INTERNAL_SERVER_ERROR"
	CodeBadRequest          = "BAD_REQUEST"
	CodeDatabaseConnection  = "DB_CONNECTION_FAILED"
	CodeInitFailed          = "INIT_FAILED"
	CodeParseFailed         = "PARSE_FAILED"
	CodeIOError             = "IO_ERROR"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
	CodeTimeout             = "TIMEOUT"
	CodeRateLimited         = "RATE_LIMITED"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
)

// AppError represents a standardized application error
type AppError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Details   string `json:"details,omitempty"`
	Err       error  `json:"-"` // Underlying error, not serialized
	HTTPCode  int    `json:"-"` // HTTP status code
}

// New creates a new AppError with given code, message, and severity
func New(code, message, severity string) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Severity: severity,
		HTTPCode: http.StatusInternalServerError,
	}
}

// WithError adds the underlying error
func (ae *AppError) WithError(err error) *AppError {
	ae.Err = err
	return ae
}

// WithDetails adds error details
func (ae *AppError) WithDetails(details string) *AppError {
	ae.Details = details
	return ae
}

// WithHTTPCode sets the HTTP status code
func (ae *AppError) WithHTTPCode(code int) *AppError {
	ae.HTTPCode = code
	return ae
}

// Error implements the error interface
func (ae *AppError) Error() string {
	if ae.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", ae.Code, ae.Message, ae.Err)
	}
	return fmt.Sprintf("[%s] %s", ae.Code, ae.Message)
}

// Unwrap returns the underlying error
func (ae *AppError) Unwrap() error {
	return ae.Err
}

// MarshalJSON provides JSON marshaling for the error
func (ae *AppError) MarshalJSON() ([]byte, error) {
	type appErrorJSON struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Details  string `json:"details,omitempty"`
	}

	return json.Marshal(appErrorJSON{
		Code:     ae.Code,
		Message:  ae.Message,
		Severity: ae.Severity,
		Details:  ae.Details,
	})
}

// Factory functions for common errors

// ValidationError creates a validation error
func ValidationError(message string) *AppError {
	return New(CodeValidation, message, SeverityWarning).WithHTTPCode(http.StatusBadRequest)
}

// NotFound creates a not found error
func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), SeverityWarning).WithHTTPCode(http.StatusNotFound)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, SeverityWarning).WithHTTPCode(http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, SeverityWarning).WithHTTPCode(http.StatusForbidden)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message, SeverityWarning).WithHTTPCode(http.StatusConflict)
}

// InternalError creates an internal server error
func InternalError(message string) *AppError {
	return New(CodeInternalServer, message, SeverityError).WithHTTPCode(http.StatusInternalServerError)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, SeverityWarning).WithHTTPCode(http.StatusBadRequest)
}

// DatabaseError creates a database connection error
func DatabaseError(message string) *AppError {
	return New(CodeDatabaseConnection, message, SeverityError).WithHTTPCode(http.StatusInternalServerError)
}

// InitializationError creates an initialization error
func InitializationError(service string, err error) *AppError {
	return New(
		CodeInitFailed,
		fmt.Sprintf("Failed to initialize %s", service),
		SeverityFatal,
	).WithError(err).WithHTTPCode(http.StatusServiceUnavailable)
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(service string) *AppError {
	return New(
		CodeServiceUnavailable,
		fmt.Sprintf("%s is currently unavailable", service),
		SeverityError,
	).WithHTTPCode(http.StatusServiceUnavailable)
}

// RateLimited creates a rate limited error
func RateLimited(message string) *AppError {
	return New(CodeRateLimited, message, SeverityWarning).WithHTTPCode(http.StatusTooManyRequests)
}

// Is checks if an error is of a specific type/code
func Is(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// As extracts an AppError if present
func As(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
