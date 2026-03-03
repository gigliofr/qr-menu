package http

import (
	"encoding/json"
	"net/http"

	"qr-menu/pkg/errors"
)

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int64       `json:"total"`
	Pages   int64       `json:"pages"`
}

// MetaResponse adds metadata to response
type MetaResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// JSON sends a JSON response with status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// Success sends a successful response with data
func Success(w http.ResponseWriter, message string, data interface{}) error {
	return JSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, message string, data interface{}) error {
	return JSON(w, http.StatusCreated, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Accepted sends a 202 Accepted response
func Accepted(w http.ResponseWriter, message string, data interface{}) error {
	return JSON(w, http.StatusAccepted, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Error sends an error response
func Error(w http.ResponseWriter, appErr *errors.AppError) error {
	statusCode := appErr.HTTPCode
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return JSON(w, statusCode, Response{
		Status:  "error",
		Message: appErr.Message,
		Error: map[string]interface{}{
			"code":     appErr.Code,
			"severity": appErr.Severity,
			"details":  appErr.Details,
		},
	})
}

// ErrorMessage sends a simple error message
func ErrorMessage(w http.ResponseWriter, statusCode int, message string) error {
	return JSON(w, statusCode, Response{
		Status:  "error",
		Message: message,
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) error {
	appErr := errors.BadRequest(message)
	return Error(w, appErr)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) error {
	appErr := errors.Unauthorized(message)
	return Error(w, appErr)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) error {
	appErr := errors.Forbidden(message)
	return Error(w, appErr)
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, resource string) error {
	appErr := errors.NotFound(resource)
	return Error(w, appErr)
}

// Conflict sends a 409 Conflict response
func Conflict(w http.ResponseWriter, message string) error {
	appErr := errors.Conflict(message)
	return Error(w, appErr)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) error {
	appErr := errors.InternalError(message)
	return Error(w, appErr)
}

// ValidationError sends validation error response
func ValidationError(w http.ResponseWriter, errors map[string]string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(Response{
		Status:  "error",
		Message: "Validation failed",
		Error:   errors,
	})
}

// Paginated sends a paginated response
func Paginated(w http.ResponseWriter, data interface{}, page, limit int, total int64) error {
	pages := (total + int64(limit) - 1) / int64(limit)

	return JSON(w, http.StatusOK, PaginatedResponse{
		Status: "success",
		Data:   data,
		Page:   page,
		Limit:  limit,
		Total:  total,
		Pages:  pages,
	})
}

// WithMeta sends a response with metadata
func WithMeta(w http.ResponseWriter, data interface{}, meta map[string]interface{}) error {
	return JSON(w, http.StatusOK, MetaResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	})
}

// File serves a file with custom headers
func File(w http.ResponseWriter, filePath string, filename string) {
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, nil, filePath)
}

// Redirect sends a redirect response
func Redirect(w http.ResponseWriter, r *http.Request, location string, statusCode int) {
	http.Redirect(w, r, location, statusCode)
}

// Stream sets headers for streaming response
func Stream(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}
