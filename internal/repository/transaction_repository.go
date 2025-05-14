package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ravindu/wallet-app-service/internal/domain"
)

type transactionRepository struct {
	db *pgxpool.Pool
}

// NewTransactionRepository creates a new PostgreSQL transaction repository
func NewTransactionRepository(db *pgxpool.Pool) domain.TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	now := time.Now()
	transaction.CreatedAt = now
	transaction.TransactionTime = now

	query := `
		INSERT INTO transactions (
			wallet_id, dest_wallet_id, type, amount, 
			balance_before, balance_after, description, 
			transaction_time, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		transaction.WalletID,
		transaction.DestWalletID,
		transaction.Type,
		transaction.Amount,
		transaction.BalanceBefore,
		transaction.BalanceAfter,
		transaction.Description,
		transaction.TransactionTime,
		transaction.CreatedAt,
	).Scan(&transaction.ID)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) GetByWalletID(ctx context.Context, walletID int64, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT 
			id, wallet_id, dest_wallet_id, type, 
			amount, balance_before, balance_after, 
			description, transaction_time, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY transaction_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		tr := &domain.Transaction{}
		err := rows.Scan(
			&tr.ID,
			&tr.WalletID,
			&tr.DestWalletID,
			&tr.Type,
			&tr.Amount,
			&tr.BalanceBefore,
			&tr.BalanceAfter,
			&tr.Description,
			&tr.TransactionTime,
			&tr.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, tr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) CountByWalletID(ctx context.Context, walletID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM transactions
		WHERE wallet_id = $1
	`

	var count int
	err := r.db.QueryRow(ctx, query, walletID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}