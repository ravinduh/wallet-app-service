package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ravindu/wallet-app-service/internal/domain"
	apperrors "github.com/ravindu/wallet-app-service/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	balanceCacheTTL = 5 * time.Second
)

type walletUsecase struct {
	userRepo        domain.UserRepository
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
	redisClient     *redis.Client
}

// NewWalletUsecase creates a wallet use case with all the necessary repos
func NewWalletUsecase(
	userRepo domain.UserRepository,
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	redisClient *redis.Client,
) domain.WalletUsecase {
	return &walletUsecase{
		userRepo:        userRepo,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		redisClient:     redisClient,
	}
}

// Deposit adds money to a user's wallet
func (u *walletUsecase) Deposit(ctx context.Context, req domain.DepositRequest) (*domain.Transaction, error) {
	if req.Amount <= 0 {
		return nil, apperrors.ErrInvalidAmount
	}

	// Find the user
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get user")
	}

	// Get their wallet
	wallet, err := u.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get wallet")
	}

	balanceBefore := wallet.Balance

	// Add the money
	if err := wallet.Deposit(req.Amount); err != nil {
		return nil, err // No need to wrap - just pass through domain errors
	}

	// Save the updated wallet
	if err := u.walletRepo.Update(ctx, wallet); err != nil {
		return nil, apperrors.WrapError(err, "failed to update wallet")
	}
	
	// Clear cache since balance changed
	if u.redisClient != nil {
		cacheKey := fmt.Sprintf("wallet:balance:%d", req.UserID)
		u.redisClient.Del(ctx, cacheKey)
	}

	// Record the transaction
	transaction := &domain.Transaction{
		WalletID:      wallet.ID,
		Type:          domain.Deposit,
		Amount:        req.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Description:   req.Comment,
	}

	if err := u.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, apperrors.WrapError(err, "failed to create transaction record")
	}

	return transaction, nil
}

// Withdraw takes money from a user's wallet
func (u *walletUsecase) Withdraw(ctx context.Context, req domain.WithdrawRequest) (*domain.Transaction, error) {
	if req.Amount <= 0 {
		return nil, apperrors.ErrInvalidAmount
	}

	// Find the user
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get user")
	}

	// Get their wallet
	wallet, err := u.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get wallet")
	}

	balanceBefore := wallet.Balance

	// Take out the money
	if err := wallet.Withdraw(req.Amount); err != nil {
		if errors.Is(err, apperrors.ErrInsufficientFunds) {
			return nil, apperrors.ErrInsufficientFunds
		}
		return nil, err // Just pass through domain errors
	}

	// Save the updated wallet
	if err := u.walletRepo.Update(ctx, wallet); err != nil {
		return nil, apperrors.WrapError(err, "failed to update wallet")
	}
	
	// Clear cache since balance changed
	if u.redisClient != nil {
		cacheKey := fmt.Sprintf("wallet:balance:%d", req.UserID)
		u.redisClient.Del(ctx, cacheKey)
	}

	// Record the transaction
	transaction := &domain.Transaction{
		WalletID:      wallet.ID,
		Type:          domain.Withdrawal,
		Amount:        req.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Description:   req.Comment,
	}

	if err := u.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, apperrors.WrapError(err, "failed to create transaction record")
	}

	return transaction, nil
}

// Transfer moves money between wallets
func (u *walletUsecase) Transfer(ctx context.Context, req domain.TransferRequest) (*domain.Transaction, error) {
	if req.Amount <= 0 {
		return nil, apperrors.ErrInvalidAmount
	}

	if req.SenderID == req.ReceiverID {
		return nil, apperrors.ErrSenderReceiverSame
	}

	// Lock wallets if Redis is available to prevent concurrent transfers
	if u.redisClient != nil {
		// Lock in ID order to prevent deadlocks
		var firstID, secondID int64
		if req.SenderID < req.ReceiverID {
			firstID, secondID = req.SenderID, req.ReceiverID
		} else {
			firstID, secondID = req.ReceiverID, req.SenderID
		}
		
		// Set up lock keys
		firstLockKey := fmt.Sprintf("lock:wallet:%d", firstID)
		secondLockKey := fmt.Sprintf("lock:wallet:%d", secondID)
		
		// Get first lock
		firstLock, err := u.redisClient.SetNX(ctx, firstLockKey, "1", 10*time.Second).Result()
		if err != nil || !firstLock {
			return nil, apperrors.ErrLockAcquisitionFailed
		}
		
		defer u.redisClient.Del(ctx, firstLockKey)
		
		// Get second lock
		secondLock, err := u.redisClient.SetNX(ctx, secondLockKey, "1", 10*time.Second).Result()
		if err != nil || !secondLock {
			return nil, apperrors.ErrLockAcquisitionFailed
		}
		
		defer u.redisClient.Del(ctx, secondLockKey)
	}

	// Get sender
	sender, err := u.userRepo.GetByID(ctx, req.SenderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get sender")
	}

	// Get receiver
	receiver, err := u.userRepo.GetByID(ctx, req.ReceiverID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get receiver")
	}

	// Get both wallets
	senderWallet, err := u.walletRepo.GetByUserID(ctx, sender.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get sender wallet")
	}

	receiverWallet, err := u.walletRepo.GetByUserID(ctx, receiver.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get receiver wallet")
	}

	senderBalanceBefore := senderWallet.Balance

	// Take from sender
	if err := senderWallet.Withdraw(req.Amount); err != nil {
		if errors.Is(err, apperrors.ErrInsufficientFunds) {
			return nil, apperrors.ErrInsufficientFunds
		}
		return nil, err
	}

	// Give to receiver
	if err := receiverWallet.Deposit(req.Amount); err != nil {
		// Should never happen since we've already validated the amount
		return nil, err
	}

	// Save sender's wallet
	if err := u.walletRepo.Update(ctx, senderWallet); err != nil {
		return nil, apperrors.WrapError(err, "failed to update sender wallet")
	}

	// Save receiver's wallet
	if err := u.walletRepo.Update(ctx, receiverWallet); err != nil {
		// This is bad - we already took money from sender but couldn't give to receiver
		// In a real app, we'd need transactions or a way to roll back
		return nil, apperrors.WrapError(err, "failed to update receiver wallet")
	}

	// Clear both caches
	if u.redisClient != nil {
		senderCacheKey := fmt.Sprintf("wallet:balance:%d", req.SenderID)
		receiverCacheKey := fmt.Sprintf("wallet:balance:%d", req.ReceiverID)
		u.redisClient.Del(ctx, senderCacheKey, receiverCacheKey)
	}

	// Record the transaction
	transaction := &domain.Transaction{
		WalletID:      senderWallet.ID,
		DestWalletID:  &receiverWallet.ID,
		Type:          domain.Transfer,
		Amount:        req.Amount,
		BalanceBefore: senderBalanceBefore,
		BalanceAfter:  senderWallet.Balance,
		Description:   req.Comment,
	}

	if err := u.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, apperrors.WrapError(err, "failed to create transaction record")
	}

	return transaction, nil
}

// GetBalance returns a user's current wallet balance
func (u *walletUsecase) GetBalance(ctx context.Context, userID int64) (*domain.Wallet, error) {
	// Try cache first
	if u.redisClient != nil {
		cacheKey := fmt.Sprintf("wallet:balance:%d", userID)
		cachedData, err := u.redisClient.Get(ctx, cacheKey).Bytes()
		
		if err == nil {
			var wallet domain.Wallet
			if err := json.Unmarshal(cachedData, &wallet); err == nil {
				return &wallet, nil
			}
			// If unmarshal fails, just continue to DB lookup
		}
	}
	
	// Get from database
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get user")
	}

	wallet, err := u.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get wallet")
	}
	
	// Cache the result
	if u.redisClient != nil {
		cacheKey := fmt.Sprintf("wallet:balance:%d", userID)
		if walletData, err := json.Marshal(wallet); err == nil {
			// Cache for a short time since balance changes frequently
			u.redisClient.Set(ctx, cacheKey, walletData, balanceCacheTTL)
		}
	}

	return wallet, nil
}

// GetTransactionHistory returns a user's past transactions
func (u *walletUsecase) GetTransactionHistory(
	ctx context.Context,
	userID int64,
	pagination domain.PaginationRequest,
) (*domain.TransactionHistoryResponse, error) {
	// Set defaults for pagination
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}
	if pagination.Offset < 0 {
		pagination.Offset = 0
	}

	// Get the user
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get user")
	}

	// Get their wallet
	wallet, err := u.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.WrapError(err, "failed to get wallet")
	}

	// Get their transactions
	transactions, err := u.transactionRepo.GetByWalletID(
		ctx, wallet.ID, pagination.Limit, pagination.Offset,
	)
	if err != nil {
		return nil, apperrors.WrapError(err, "failed to get transactions")
	}

	// Get total count for pagination info
	total, err := u.transactionRepo.CountByWalletID(ctx, wallet.ID)
	if err != nil {
		return nil, apperrors.WrapError(err, "failed to count transactions")
	}

	response := &domain.TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
		Limit:        pagination.Limit,
		Offset:       pagination.Offset,
	}

	return response, nil
}