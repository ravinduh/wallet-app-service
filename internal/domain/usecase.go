package domain

import "context"

// DepositRequest represents deposit parameters
type DepositRequest struct {
	UserID  int64   `json:"user_id"`
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment,omitempty"`
}

// WithdrawRequest represents withdrawal parameters
type WithdrawRequest struct {
	UserID  int64   `json:"user_id"`
	Amount  float64 `json:"amount"`
	Comment string  `json:"comment,omitempty"`
}

// TransferRequest represents transfer parameters
type TransferRequest struct {
	SenderID   int64   `json:"sender_id"`
	ReceiverID int64   `json:"receiver_id"`
	Amount     float64 `json:"amount"`
	Comment    string  `json:"comment,omitempty"`
}

// PaginationRequest for limiting result sets
type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// TransactionHistoryResponse for transaction listings
type TransactionHistoryResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int            `json:"total"`
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
}

// WalletUsecase defines business logic for wallet operations
type WalletUsecase interface {
	Deposit(ctx context.Context, req DepositRequest) (*Transaction, error)
	Withdraw(ctx context.Context, req WithdrawRequest) (*Transaction, error)
	Transfer(ctx context.Context, req TransferRequest) (*Transaction, error)
	GetBalance(ctx context.Context, userID int64) (*Wallet, error)
	GetTransactionHistory(ctx context.Context, userID int64, pagination PaginationRequest) (*TransactionHistoryResponse, error)
}