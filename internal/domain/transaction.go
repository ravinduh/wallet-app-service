package domain

import (
	"time"
)

// TransactionType represents the type of wallet transaction
type TransactionType string

const (
	// Deposit represents money added to wallet
	Deposit TransactionType = "DEPOSIT"
	// Withdrawal represents money removed from wallet
	Withdrawal TransactionType = "WITHDRAWAL"
	// Transfer represents money sent to another user
	Transfer TransactionType = "TRANSFER"
)

// Transaction represents a wallet transaction
type Transaction struct {
	ID              int64           `json:"id"`
	WalletID        int64           `json:"wallet_id"`
	DestWalletID    *int64          `json:"dest_wallet_id,omitempty"`
	Type            TransactionType `json:"type"`
	Amount          float64         `json:"amount"`
	BalanceBefore   float64         `json:"balance_before"`
	BalanceAfter    float64         `json:"balance_after"`
	Description     string          `json:"description"`
	TransactionTime time.Time       `json:"transaction_time"`
	CreatedAt       time.Time       `json:"created_at"`
}