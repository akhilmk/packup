// Package httputil provides HTTP response utilities.
package httputil

import (
	"encoding/json"
	"net/http"
)

// APIError represents a standard API error response.
type APIError struct {
	Error string `json:"error"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a JSON error response with the given message and status code.
func WriteError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{Error: message})
}

// WriteSuccess writes a JSON success response.
func WriteSuccess(w http.ResponseWriter) {
	WriteJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// BadRequest writes a 400 Bad Request error response.
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusBadRequest)
}

// Unauthorized writes a 401 Unauthorized error response.
func Unauthorized(w http.ResponseWriter) {
	WriteError(w, "unauthorized", http.StatusUnauthorized)
}

// Forbidden writes a 403 Forbidden error response.
func Forbidden(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusForbidden)
}

// NotFound writes a 404 Not Found error response.
func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusNotFound)
}

// InternalError writes a 500 Internal Server Error response.
// Note: Avoid exposing internal error details to clients in production.
func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusInternalServerError)
}
