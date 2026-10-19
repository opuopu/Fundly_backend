package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success    bool        `json:"success"`
	Data       any         `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Error      *Error      `json:"error,omitempty"`
}

// Pagination is the standard paginated-response metadata (spec §3.3).
type Pagination struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
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

// JSONPaginated sends a standard paginated JSON response (spec §3.3).
func JSONPaginated(w http.ResponseWriter, statusCode int, data any, p Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Response{
		Success:    statusCode >= 200 && statusCode < 300,
		Data:       data,
		Pagination: &p,
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
