package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/ravindu/wallet-app-service/internal/domain"
	"github.com/ravindu/wallet-app-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}


type mockWalletRepository struct {
	mock.Mock
}

func (m *mockWalletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *mockWalletRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wallet), args.Error(1)
}

func (m *mockWalletRepository) Update(ctx context.Context, wallet *domain.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

type mockTransactionRepository struct {
	mock.Mock
}

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *mockTransactionRepository) GetByWalletID(ctx context.Context, walletID int64, limit, offset int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, walletID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *mockTransactionRepository) CountByWalletID(ctx context.Context, walletID int64) (int, error) {
	args := m.Called(ctx, walletID)
	return args.Int(0), args.Error(1)
}

func TestDeposit(t *testing.T) {
	// Setup
	ctx := context.Background()
	now := time.Now()
	
	mockUser := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	mockWallet := &domain.Wallet{
		ID:        1,
		UserID:    1,
		Balance:   100.0,
		Currency:  domain.USD,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Create mocks
	userRepo := new(mockUserRepository)
	walletRepo := new(mockWalletRepository)
	transactionRepo := new(mockTransactionRepository)
	
	// Setup expectations
	userRepo.On("GetByID", ctx, int64(1)).Return(mockUser, nil)
	walletRepo.On("GetByUserID", ctx, int64(1)).Return(mockWallet, nil)
	walletRepo.On("Update", ctx, mock.AnythingOfType("*domain.Wallet")).Return(nil)
	transactionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	
	// Create usecase with mocks
	uc := usecase.NewWalletUsecase(userRepo, walletRepo, transactionRepo)
	
	// Test success case
	req := domain.DepositRequest{
		UserID:  1,
		Amount:  50.0,
		Comment: "Test deposit",
	}
	
	transaction, err := uc.Deposit(ctx, req)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, domain.Deposit, transaction.Type)
	assert.Equal(t, 50.0, transaction.Amount)
	assert.Equal(t, 100.0, transaction.BalanceBefore)
	assert.Equal(t, 150.0, transaction.BalanceAfter)
	
	// Verify expectations
	userRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestWithdraw(t *testing.T) {
	// Setup
	ctx := context.Background()
	now := time.Now()
	
	mockUser := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	mockWallet := &domain.Wallet{
		ID:        1,
		UserID:    1,
		Balance:   100.0,
		Currency:  domain.USD,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Create mocks
	userRepo := new(mockUserRepository)
	walletRepo := new(mockWalletRepository)
	transactionRepo := new(mockTransactionRepository)
	
	// Setup expectations
	userRepo.On("GetByID", ctx, int64(1)).Return(mockUser, nil)
	walletRepo.On("GetByUserID", ctx, int64(1)).Return(mockWallet, nil)
	walletRepo.On("Update", ctx, mock.AnythingOfType("*domain.Wallet")).Return(nil)
	transactionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	
	// Create usecase with mocks
	uc := usecase.NewWalletUsecase(userRepo, walletRepo, transactionRepo)
	
	// Test success case
	req := domain.WithdrawRequest{
		UserID:  1,
		Amount:  50.0,
		Comment: "Test withdrawal",
	}
	
	transaction, err := uc.Withdraw(ctx, req)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, domain.Withdrawal, transaction.Type)
	assert.Equal(t, 50.0, transaction.Amount)
	assert.Equal(t, 100.0, transaction.BalanceBefore)
	assert.Equal(t, 50.0, transaction.BalanceAfter)
	
	// Test insufficient funds
	insufficientReq := domain.WithdrawRequest{
		UserID:  1,
		Amount:  200.0,
		Comment: "Insufficient withdrawal",
	}
	
	// Reset wallet for this test
	mockWallet.Balance = 100.0
	
	transaction, err = uc.Withdraw(ctx, insufficientReq)
	
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, domain.ErrInsufficientBalance, err)
	
	// Verify expectations
	userRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestTransfer(t *testing.T) {
	// Setup
	ctx := context.Background()
	now := time.Now()
	
	sender := &domain.User{
		ID:        1,
		Username:  "sender",
		Email:     "sender@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	receiver := &domain.User{
		ID:        2,
		Username:  "receiver",
		Email:     "receiver@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	senderWallet := &domain.Wallet{
		ID:        1,
		UserID:    1,
		Balance:   100.0,
		Currency:  domain.USD,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	receiverWallet := &domain.Wallet{
		ID:        2,
		UserID:    2,
		Balance:   50.0,
		Currency:  domain.USD,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Create mocks
	userRepo := new(mockUserRepository)
	walletRepo := new(mockWalletRepository)
	transactionRepo := new(mockTransactionRepository)
	
	// Setup expectations
	userRepo.On("GetByID", ctx, int64(1)).Return(sender, nil)
	userRepo.On("GetByID", ctx, int64(2)).Return(receiver, nil)
	walletRepo.On("GetByUserID", ctx, int64(1)).Return(senderWallet, nil)
	walletRepo.On("GetByUserID", ctx, int64(2)).Return(receiverWallet, nil)
	walletRepo.On("Update", ctx, mock.AnythingOfType("*domain.Wallet")).Return(nil)
	transactionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)
	
	// Create usecase with mocks
	uc := usecase.NewWalletUsecase(userRepo, walletRepo, transactionRepo)
	
	// Test success case
	req := domain.TransferRequest{
		SenderID:   1,
		ReceiverID: 2,
		Amount:     30.0,
		Comment:    "Test transfer",
	}
	
	transaction, err := uc.Transfer(ctx, req)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, domain.Transfer, transaction.Type)
	assert.Equal(t, 30.0, transaction.Amount)
	assert.Equal(t, 100.0, transaction.BalanceBefore)
	assert.Equal(t, 70.0, transaction.BalanceAfter)
	
	// Verify expectations
	userRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}