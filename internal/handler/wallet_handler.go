package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ravindu/wallet-app-service/internal/domain"
	apperrors "github.com/ravindu/wallet-app-service/pkg/errors"
	"github.com/ravindu/wallet-app-service/pkg/logging"
	"github.com/ravindu/wallet-app-service/pkg/request"
	"github.com/ravindu/wallet-app-service/pkg/response"
)

type WalletHandler struct {
	walletUsecase domain.WalletUsecase
	logger        *logging.Logger
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletUsecase domain.WalletUsecase) *WalletHandler {
	return &WalletHandler{
		walletUsecase: walletUsecase,
		logger:        logging.NewLogger(),
	}
}

// getRequestID extracts the request ID from context
func getRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value(request.RequestIDKey).(string); ok {
		return requestID
	}
	return "no-request-id"
}

// DepositHandler handles deposit requests
func (h *WalletHandler) DepositHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	ctx := r.Context()
	
	h.logger.Info(ctx, "Processing deposit request")
	
	var req domain.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode deposit request: "+err.Error())
		errResp := apperrors.BadRequestError(requestID, "Invalid request format, please check your JSON payload")
		response.Error(w, errResp)
		return
	}

	// Validate request
	if req.Amount <= 0 {
		h.logger.Error(ctx, "Invalid deposit amount: "+fmt.Sprintf("%f", req.Amount))
		errResp := apperrors.BadRequestError(requestID, "Amount must be positive")
		response.Error(w, errResp)
		return
	}

	transaction, err := h.walletUsecase.Deposit(ctx, req)
	if err != nil {
		h.logger.Error(ctx, "Deposit failed: "+err.Error())
		// Map the domain error to the appropriate HTTP response
		errResp := apperrors.MapErrorToResponse(requestID, err)
		response.Error(w, errResp)
		return
	}

	h.logger.Info(ctx, "Deposit successful")
	response.JSON(w, requestID, transaction, http.StatusOK)
}

// WithdrawHandler handles withdrawal requests
func (h *WalletHandler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	ctx := r.Context()
	
	h.logger.Info(ctx, "Processing withdrawal request")
	
	var req domain.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode withdrawal request: "+err.Error())
		errResp := apperrors.BadRequestError(requestID, "Invalid request format, please check your JSON payload")
		response.Error(w, errResp)
		return
	}

	// Validate request
	if req.Amount <= 0 {
		h.logger.Error(ctx, "Invalid withdrawal amount: "+fmt.Sprintf("%f", req.Amount))
		errResp := apperrors.BadRequestError(requestID, "Amount must be positive")
		response.Error(w, errResp)
		return
	}

	transaction, err := h.walletUsecase.Withdraw(ctx, req)
	if err != nil {
		h.logger.Error(ctx, "Withdrawal failed: "+err.Error())
		
		// Handle specific errors with appropriate responses
		if errors.Is(err, apperrors.ErrInsufficientFunds) {
			errResp := apperrors.PaymentRequiredError(requestID, "Insufficient funds for this withdrawal")
			response.Error(w, errResp)
			return
		}
		
		// Map other domain errors to HTTP responses
		errResp := apperrors.MapErrorToResponse(requestID, err)
		response.Error(w, errResp)
		return
	}

	h.logger.Info(ctx, "Withdrawal successful")
	response.JSON(w, requestID, transaction, http.StatusOK)
}

// TransferHandler handles transfer requests
func (h *WalletHandler) TransferHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	ctx := r.Context()
	
	h.logger.Info(ctx, "Processing transfer request")
	
	var req domain.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode transfer request: "+err.Error())
		errResp := apperrors.BadRequestError(requestID, "Invalid request format, please check your JSON payload")
		response.Error(w, errResp)
		return
	}

	// Validate request
	if req.Amount <= 0 {
		h.logger.Error(ctx, "Invalid transfer amount: "+fmt.Sprintf("%f", req.Amount))
		errResp := apperrors.BadRequestError(requestID, "Amount must be positive")
		response.Error(w, errResp)
		return
	}

	if req.SenderID == req.ReceiverID {
		h.logger.Error(ctx, "Transfer rejected: sender and receiver are the same (ID: "+fmt.Sprintf("%d", req.SenderID)+")")
		errResp := apperrors.BadRequestError(requestID, "Sender and receiver cannot be the same")
		response.Error(w, errResp)
		return
	}

	transaction, err := h.walletUsecase.Transfer(ctx, req)
	if err != nil {
		h.logger.Error(ctx, "Transfer failed: "+err.Error())
		
		// Handle specific errors with appropriate responses
		switch {
		case errors.Is(err, apperrors.ErrInsufficientFunds):
			errResp := apperrors.PaymentRequiredError(requestID, "Insufficient funds for this transfer")
			response.Error(w, errResp)
			return
		case errors.Is(err, apperrors.ErrUserNotFound):
			errResp := apperrors.NotFoundError(requestID, "One of the users in this transfer does not exist")
			response.Error(w, errResp)
			return
		case errors.Is(err, apperrors.ErrLockAcquisitionFailed):
			errResp := apperrors.TooManyRequestsError(requestID, "Transfer currently unavailable, please try again in a moment")
			response.Error(w, errResp)
			return
		default:
			// Map other domain errors to HTTP responses
			errResp := apperrors.MapErrorToResponse(requestID, err)
			response.Error(w, errResp)
			return
		}
	}

	h.logger.Info(ctx, "Transfer successful")
	response.JSON(w, requestID, transaction, http.StatusOK)
}

// GetBalanceHandler handles balance requests
func (h *WalletHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	ctx := r.Context()
	
	h.logger.Info(ctx, "Processing balance request")
	
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Error(ctx, "Invalid user ID format: "+userIDStr)
		errResp := apperrors.BadRequestError(requestID, "User ID must be a valid number")
		response.Error(w, errResp)
		return
	}

	wallet, err := h.walletUsecase.GetBalance(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get balance: "+err.Error())
		
		// Handle specific error cases
		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrWalletNotFound) {
			errResp := apperrors.NotFoundError(requestID, "No wallet found for this user")
			response.Error(w, errResp)
			return
		}
		
		// For other errors, use the generic mapper
		errResp := apperrors.MapErrorToResponse(requestID, err)
		response.Error(w, errResp)
		return
	}

	h.logger.Info(ctx, "Balance request successful")
	response.JSON(w, requestID, wallet, http.StatusOK)
}

// GetTransactionHistoryHandler handles transaction history requests
func (h *WalletHandler) GetTransactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	ctx := r.Context()
	
	h.logger.Info(ctx, "Processing transaction history request")
	
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Error(ctx, "Invalid user ID format: "+userIDStr)
		errResp := apperrors.BadRequestError(requestID, "User ID must be a valid number")
		response.Error(w, errResp)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			h.logger.Error(ctx, "Invalid limit parameter format: "+limitStr)
			errResp := apperrors.BadRequestError(requestID, "Limit must be a positive number")
			response.Error(w, errResp)
			return
		}
		
		if limitInt <= 0 {
			h.logger.Error(ctx, "Invalid limit value: "+limitStr)
			errResp := apperrors.BadRequestError(requestID, "Limit must be a positive number")
			response.Error(w, errResp)
			return
		}
		
		// Enforce a reasonable maximum limit to prevent overloading
		if limitInt > 100 {
			h.logger.Warn(ctx, "Limit too large, capping at 100: "+limitStr)
			limitInt = 100
		}
		
		limit = limitInt
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil {
			h.logger.Error(ctx, "Invalid offset parameter format: "+offsetStr)
			errResp := apperrors.BadRequestError(requestID, "Offset must be a non-negative number")
			response.Error(w, errResp)
			return
		}
		
		if offsetInt < 0 {
			h.logger.Error(ctx, "Invalid offset value: "+offsetStr)
			errResp := apperrors.BadRequestError(requestID, "Offset must be a non-negative number")
			response.Error(w, errResp)
			return
		}
		
		offset = offsetInt
	}

	pagination := domain.PaginationRequest{
		Limit:  limit,
		Offset: offset,
	}

	h.logger.Debug(ctx, "Getting transaction history")
	history, err := h.walletUsecase.GetTransactionHistory(ctx, userID, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get transaction history: "+err.Error())
		
		// Handle specific error cases
		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrWalletNotFound) {
			errResp := apperrors.NotFoundError(requestID, "No wallet found for this user")
			response.Error(w, errResp)
			return
		}
		
		// For other errors, use the generic mapper
		errResp := apperrors.MapErrorToResponse(requestID, err)
		response.Error(w, errResp)
		return
	}

	h.logger.Info(ctx, "Transaction history request successful")
	response.JSON(w, requestID, history, http.StatusOK)
}