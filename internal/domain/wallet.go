package domain

import (
	"time"
	
	apperrors "github.com/ravindu/wallet-app-service/pkg/errors"
)

// Currency type used for the wallet
type Currency string

const (
	// USD - US Dollars
	USD Currency = "USD"
)

// Wallet holds user's money and related info
type Wallet struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  Currency  `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Deposit money into the wallet
func (w *Wallet) Deposit(amount float64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}
	
	w.Balance += amount
	w.UpdatedAt = time.Now()
	return nil
}

// Withdraw money from the wallet
func (w *Wallet) Withdraw(amount float64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}
	
	if w.Balance < amount {
		return apperrors.ErrInsufficientFunds
	}
	
	w.Balance -= amount
	w.UpdatedAt = time.Now()
	return nil
}