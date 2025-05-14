package domain_test

import (
	"testing"
	"time"

	"github.com/ravindu/wallet-app-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestWallet_Deposit(t *testing.T) {
	tests := []struct {
		name          string
		wallet        domain.Wallet
		amount        float64
		expectedError bool
	}{
		{
			name: "valid deposit",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        50.0,
			expectedError: false,
		},
		{
			name: "zero amount",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        0.0,
			expectedError: true,
		},
		{
			name: "negative amount",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        -50.0,
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			tc.wallet.CreatedAt = now
			tc.wallet.UpdatedAt = now

			initialBalance := tc.wallet.Balance
			err := tc.wallet.Deposit(tc.amount)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, initialBalance, tc.wallet.Balance)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialBalance+tc.amount, tc.wallet.Balance)
				assert.True(t, tc.wallet.UpdatedAt.After(now) || tc.wallet.UpdatedAt.Equal(now))
			}
		})
	}
}

func TestWallet_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		wallet        domain.Wallet
		amount        float64
		expectedError bool
	}{
		{
			name: "valid withdrawal",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        50.0,
			expectedError: false,
		},
		{
			name: "insufficient funds",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        150.0,
			expectedError: true,
		},
		{
			name: "zero amount",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        0.0,
			expectedError: true,
		},
		{
			name: "negative amount",
			wallet: domain.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  100.0,
				Currency: domain.USD,
			},
			amount:        -50.0,
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			tc.wallet.CreatedAt = now
			tc.wallet.UpdatedAt = now

			initialBalance := tc.wallet.Balance
			err := tc.wallet.Withdraw(tc.amount)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Equal(t, initialBalance, tc.wallet.Balance)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialBalance-tc.amount, tc.wallet.Balance)
				assert.True(t, tc.wallet.UpdatedAt.After(now) || tc.wallet.UpdatedAt.Equal(now))
			}
		})
	}
}