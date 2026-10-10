package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldIssue `json:"details,omitempty"`
}

type FieldIssue struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// JSON sends a standard JSON response with appropriate status code and envelope.
func JSON(w http.ResponseWriter, statusCode int, data any, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
		Message: message,
	})
}

// JSONError sends a standard error JSON response with error code and details.
func JSONError(w http.ResponseWriter, statusCode int, code, message string, details ...FieldIssue) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
