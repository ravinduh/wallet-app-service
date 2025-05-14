package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/ravindu/wallet-app-service/pkg/request"
)

// Header name for passing request IDs
const RequestIDHeader = "Request-Id"

// RequestID adds a tracking ID to every request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use existing ID if client provided one
		requestID := r.Header.Get(RequestIDHeader)
		
		// Or make a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Send it back in response headers
		w.Header().Set(RequestIDHeader, requestID)
		
		// Add to context for logging/errors
		ctx := context.WithValue(r.Context(), request.RequestIDKey, requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID grabs the ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return "no-request-id"
	}

	requestID, ok := ctx.Value(request.RequestIDKey).(string)
	if !ok || requestID == "" {
		return "no-request-id"
	}

	return requestID
}