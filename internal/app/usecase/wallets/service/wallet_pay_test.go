package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
)

func TestPay_Success(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "1000.00"))
	svc := NewWalletService(repo)

	resp, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "25.75", ReferenceID: "pay-1"})
	require.NoError(t, err)
	assert.Equal(t, "974.25", resp.Balance)
	assert.Equal(t, 1, repo.getLedgerCount())

	entries := repo.getLedgerEntriesForWallet("w-1")
	assert.True(t, entries[0].Amount.Equal(decimal.NewFromFloat(-25.75)), "ledger amount should be negative")
	assert.Equal(t, "PAYMENT", entries[0].EntryType)
}

func TestPay_InsufficientBalance(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "10.00"))
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "10.01", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrInsufficientBalance)
	assert.Equal(t, 0, repo.getLedgerCount())
	assert.Equal(t, "10.00", repo.getWalletState("w-1").Balance.StringFixed(2))
}

func TestPay_ExactBalance(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "50.00"))
	svc := NewWalletService(repo)

	resp, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "50.00", ReferenceID: "pay-1"})
	require.NoError(t, err)
	assert.Equal(t, "0.00", resp.Balance)
}

func TestPay_ZeroAmount(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "0.00", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrAmountNotPositive)
}

func TestPay_NegativeAmount(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "-5.00", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrAmountNotPositive)
}

func TestPay_PrecisionRejection(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "0.001", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrAmountPrecision)
}

func TestPay_WalletNotFound(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("nonexistent", domains.PayWalletRequest{Amount: "10.00", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestPay_SuspendedWallet(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newSuspendedWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "10.00", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrWalletSuspended)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestPay_DuplicateReferenceID_Idempotent(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	resp1, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "30.00", ReferenceID: "pay-dup"})
	require.NoError(t, err)
	assert.Equal(t, "70.00", resp1.Balance)

	resp2, err := svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "30.00", ReferenceID: "pay-dup"})
	require.NoError(t, err)
	assert.Equal(t, "70.00", resp2.Balance)
	assert.Equal(t, 1, repo.getLedgerCount())
}

func TestPay_LedgerConsistency(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "0.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "200.00", ReferenceID: "top-1"})
	require.NoError(t, err)
	_, err = svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "75.50", ReferenceID: "pay-1"})
	require.NoError(t, err)

	entries := repo.getLedgerEntriesForWallet("w-1")
	assert.Len(t, entries, 2)

	ledgerSum := decimal.Zero
	for _, e := range entries {
		ledgerSum = ledgerSum.Add(e.Amount)
	}
	wallet := repo.getWalletState("w-1")
	assert.True(t, wallet.Balance.Equal(ledgerSum), "balance %s != ledger sum %s", wallet.Balance, ledgerSum)
	assert.Equal(t, "124.50", wallet.Balance.StringFixed(2))
}

func TestPay_OutOfOrderOperations(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "0.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "100.00", ReferenceID: "t-1"})
	require.NoError(t, err)
	_, err = svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "30.00", ReferenceID: "p-1"})
	require.NoError(t, err)
	_, err = svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "20.00", ReferenceID: "t-2"})
	require.NoError(t, err)
	_, err = svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "50.00", ReferenceID: "p-2"})
	require.NoError(t, err)

	wallet := repo.getWalletState("w-1")
	assert.Equal(t, "40.00", wallet.Balance.StringFixed(2))

	entries := repo.getLedgerEntriesForWallet("w-1")
	ledgerSum := decimal.Zero
	for _, e := range entries {
		ledgerSum = ledgerSum.Add(e.Amount)
	}
	assert.True(t, wallet.Balance.Equal(ledgerSum))
}
