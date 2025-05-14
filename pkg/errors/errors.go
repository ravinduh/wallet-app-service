package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error types we use throughout the app
var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrDatabaseError         = errors.New("database error")
	ErrInvalidAmount         = errors.New("amount must be positive")
	ErrSenderReceiverSame    = errors.New("sender and receiver cannot be the same")
	ErrTransactionFailed     = errors.New("transaction failed")
	ErrInvalidRequestID      = errors.New("invalid request ID")
	ErrCachingFailed         = errors.New("caching operation failed")
	ErrLockAcquisitionFailed = errors.New("could not acquire lock for operation")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrForbidden             = errors.New("forbidden action")
)

// WrapError adds more context to an error
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// ErrorResponse structure for our API errors
type ErrorResponse struct {
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
	Code      int    `json:"-"` // Just for internal use
}

// NewErrorResponse creates a basic error response
func NewErrorResponse(requestID, message string, code int) *ErrorResponse {
	return &ErrorResponse{
		RequestID: requestID,
		Error:     message,
		Code:      code,
	}
}

// BadRequestError for 400 errors
func BadRequestError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusBadRequest)
}

// PaymentRequiredError for 402 errors
func PaymentRequiredError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusPaymentRequired)
}

// NotFoundError for 404 errors
func NotFoundError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusNotFound)
}

// InternalServerError for 500 errors
func InternalServerError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusInternalServerError)
}

// UnauthorizedError for 401 errors
func UnauthorizedError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusUnauthorized)
}

// ForbiddenError for 403 errors
func ForbiddenError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusForbidden)
}

// TooManyRequestsError for 429 errors
func TooManyRequestsError(requestID, message string) *ErrorResponse {
	return NewErrorResponse(requestID, message, http.StatusTooManyRequests)
}

// MapErrorToResponse converts domain errors to HTTP responses
func MapErrorToResponse(requestID string, err error) *ErrorResponse {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrSenderReceiverSame):
		return BadRequestError(requestID, err.Error())
	case errors.Is(err, ErrInsufficientFunds):
		return PaymentRequiredError(requestID, "Insufficient funds for this operation")
	case errors.Is(err, ErrResourceNotFound), errors.Is(err, ErrUserNotFound), errors.Is(err, ErrWalletNotFound):
		return NotFoundError(requestID, err.Error())
	case errors.Is(err, ErrUnauthorized):
		return UnauthorizedError(requestID, err.Error())
	case errors.Is(err, ErrForbidden):
		return ForbiddenError(requestID, err.Error())
	case errors.Is(err, ErrLockAcquisitionFailed):
		return TooManyRequestsError(requestID, "Service is busy, please try again in a moment")
	default:
		// Don't leak internal errors to clients
		return InternalServerError(requestID, "An unexpected error occurred")
	}
}