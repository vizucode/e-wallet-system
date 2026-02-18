package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
)

func TestCreateWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	resp, err := svc.CreateWallet(domains.CreateWalletRequest{
		UserID:   "user-1",
		Currency: "USD",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, "0.00", resp.Balance)
	assert.Equal(t, "ACTIVE", resp.Status)
	assert.NotEmpty(t, resp.WalletID)
}

func TestCreateWallet_MultipleWalletsDifferentCurrencies(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	resp1, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, "USD", resp1.Currency)

	resp2, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: "EUR"})
	require.NoError(t, err)
	assert.Equal(t, "EUR", resp2.Currency)
	assert.NotEqual(t, resp1.WalletID, resp2.WalletID)
}

func TestCreateWallet_DuplicateCurrencyRejected(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: "USD"})
	require.NoError(t, err)

	_, err = svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: "USD"})
	assert.ErrorIs(t, err, ErrWalletExists)
}

func TestCreateWallet_EmptyUserID(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "", Currency: "USD"})
	assert.ErrorIs(t, err, ErrUserIDRequired)
}

func TestCreateWallet_EmptyCurrency(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: ""})
	assert.ErrorIs(t, err, ErrCurrencyRequired)
}

func TestCreateWallet_InvalidCurrency(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: "XYZ"})
	assert.ErrorIs(t, err, ErrInvalidCurrency)
}

func TestCreateWallet_CurrencyNormalization(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	resp, err := svc.CreateWallet(domains.CreateWalletRequest{UserID: "user-1", Currency: " usd "})
	require.NoError(t, err)
	assert.Equal(t, "USD", resp.Currency)
}

func newActiveWallet(id, currency string, balance string) *models.Wallet {
	b, _ := decimal.NewFromString(balance)
	return &models.Wallet{
		ID: id, OwnerID: "user-1", Currency: currency,
		Balance: b, Status: "ACTIVE", Version: 0,
	}
}

func newSuspendedWallet(id, currency string, balance string) *models.Wallet {
	b, _ := decimal.NewFromString(balance)
	return &models.Wallet{
		ID: id, OwnerID: "user-1", Currency: currency,
		Balance: b, Status: "SUSPENDED", Version: 0,
	}
}
