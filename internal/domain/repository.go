package domain

import "context"

// UserRepository defines operations for user management
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
}

// WalletRepository defines operations for wallet management
type WalletRepository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetByUserID(ctx context.Context, userID int64) (*Wallet, error)
	Update(ctx context.Context, wallet *Wallet) error
}

// TransactionRepository defines operations for transaction management
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetByWalletID(ctx context.Context, walletID int64, limit, offset int) ([]*Transaction, error)
	CountByWalletID(ctx context.Context, walletID int64) (int, error)
}