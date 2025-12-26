package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// JSON writes any payload as JSON with a status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Error writes a consistent JSON error response: { "message": "..." }
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Message: message})
}
