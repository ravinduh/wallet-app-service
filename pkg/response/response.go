package response

import (
	"encoding/json"
	"net/http"

	"github.com/ravindu/wallet-app-service/pkg/errors"
)

// Response is the standard JSON response structure
type Response struct {
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// JSON sends a JSON response with the given data and status code
func JSON(w http.ResponseWriter, requestID string, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		RequestID: requestID,
		Data:      data,
	}

	// Handle encoding errors
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// If we can't encode the response, switch to an error response
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error encoding response: " + err.Error()))
	}
}

// Error sends a JSON error response
func Error(w http.ResponseWriter, errResponse *errors.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResponse.Code)

	resp := Response{
		RequestID: errResponse.RequestID,
		Error:     errResponse.Error,
	}

	// Handle encoding errors
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// If we can't even encode the error response, fallback to a simple error
		// This is an edge case that shouldn't happen often
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error encoding response: " + err.Error()))
	}
}