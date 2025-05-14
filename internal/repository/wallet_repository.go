package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ravindu/wallet-app-service/internal/domain"
)

type walletRepository struct {
	db *pgxpool.Pool
}

// NewWalletRepository creates a new PostgreSQL wallet repository
func NewWalletRepository(db *pgxpool.Pool) domain.WalletRepository {
	return &walletRepository{
		db: db,
	}
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	now := time.Now()
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	query := `
		INSERT INTO wallets (user_id, balance, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		wallet.UserID,
		wallet.Balance,
		wallet.Currency,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	).Scan(&wallet.ID)

	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`

	wallet := &domain.Wallet{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get wallet by user ID: %w", err)
	}

	return wallet, nil
}

func (r *walletRepository) Update(ctx context.Context, wallet *domain.Wallet) error {
	wallet.UpdatedAt = time.Now()

	query := `
		UPDATE wallets
		SET balance = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query,
		wallet.Balance,
		wallet.UpdatedAt,
		wallet.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}