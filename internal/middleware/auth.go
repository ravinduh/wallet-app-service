package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ravindu/wallet-app-service/pkg/errors"
	"github.com/ravindu/wallet-app-service/pkg/logging"
	"github.com/ravindu/wallet-app-service/pkg/response"
)

// AuthContextKey is the key used to store user ID in the context
type AuthContextKey string

// UserIDKey is the context key for the authenticated user ID
const UserIDKey AuthContextKey = "user_id"

// AuthMiddleware provides authentication for API endpoints
func AuthMiddleware(next http.Handler) http.Handler {
	logger := logging.NewLogger()
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := GetRequestID(ctx)
		
		// Get authorization header
		authHeader := r.Header.Get("Authorization")
		
		// Check if authorization header exists
		if authHeader == "" {
			// TODO: Implement proper unauthorized response
			logger.Error(ctx, "Missing Authorization header")
			errResp := errors.UnauthorizedError(requestID, "Missing Authorization header")
			response.Error(w, errResp)
			return
		}
		
		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// TODO: Implement proper error handling for invalid auth format
			logger.Error(ctx, "Invalid Authorization format")
			errResp := errors.UnauthorizedError(requestID, "Invalid Authorization format")
			response.Error(w, errResp)
			return
		}
		
		// Extract token for validation
		_ = parts[1] // Ignoring until token validation is implemented
		
		// TODO: Implement actual token validation logic
		// - Parse and validate JWT token
		// - Check token expiration
		// - Verify signature with secret key
		// - Check if token is blacklisted
		
		// TODO: Get user ID from token claims
		var userID int64 = 0 // Placeholder value, replace with actual user ID from token
		
		// TODO: Check if user exists in database

		// Add user ID to context for downstream handlers
		ctx = context.WithValue(ctx, UserIDKey, userID)
		
		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the authenticated user ID from the context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// RequireAuth checks if a user is authenticated
func RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := GetRequestID(ctx)
		
		userID, ok := GetUserID(ctx)
		if !ok || userID == 0 {
			// TODO: Implement proper unauthorized response
			errResp := errors.UnauthorizedError(requestID, "Authentication required")
			response.Error(w, errResp)
			return
		}
		
		handler(w, r)
	}
}