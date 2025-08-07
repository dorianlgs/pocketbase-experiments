package main

import (
	"encoding/json"
	"net/http"
)

// JSONResponse is a helper function to send JSON responses
func JSONResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSONErrorResponse sends a standardized JSON error response
func JSONErrorResponse(w http.ResponseWriter, message string, status int) {
	JSONResponse(w, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}, status)
}