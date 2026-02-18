package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
)

func TestTopUp_Success(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	resp, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "ref-1"})
	require.NoError(t, err)
	assert.Equal(t, "150.00", resp.Balance)
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, 1, repo.getLedgerCount())
}

func TestTopUp_LargeBalance(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "999999999.00"))
	svc := NewWalletService(repo)

	resp, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "1.00", ReferenceID: "ref-1"})
	require.NoError(t, err)
	assert.Equal(t, "1000000000.00", resp.Balance)
}

func TestTopUp_DecimalPrecisionThreeDecimalPlaces(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "12.345", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrAmountPrecision)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestTopUp_AmountBelowSmallestUnit(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "0.001", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrAmountPrecision)
}

func TestTopUp_ZeroAmount(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "0.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrAmountNotPositive)
}

func TestTopUp_NegativeAmount(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "-10.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrAmountNotPositive)
}

func TestTopUp_InvalidAmountString(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "abc", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrInvalidAmount)
}

func TestTopUp_EmptyReferenceID(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "10.00", ReferenceID: ""})
	assert.ErrorIs(t, err, ErrReferenceRequired)
}

func TestTopUp_WalletNotFound(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("nonexistent", domains.TopUpWalletRequest{Amount: "10.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestTopUp_SuspendedWallet(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newSuspendedWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "10.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrWalletSuspended)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestTopUp_DuplicateReferenceID_Idempotent(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	resp1, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "ref-dup"})
	require.NoError(t, err)
	assert.Equal(t, "150.00", resp1.Balance)

	resp2, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "ref-dup"})
	require.NoError(t, err)
	assert.Equal(t, "150.00", resp2.Balance)
	assert.Equal(t, 1, repo.getLedgerCount())
}

func TestTopUp_ConcurrentUpdate_VersionMismatch(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	repo.failOnUpdate = true
	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "10.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrConcurrentUpdate)
}

func TestTopUp_LedgerConsistency(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "0.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "100.00", ReferenceID: "ref-1"})
	require.NoError(t, err)
	_, err = svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.50", ReferenceID: "ref-2"})
	require.NoError(t, err)

	entries := repo.getLedgerEntriesForWallet("w-1")
	assert.Len(t, entries, 2)

	ledgerSum := entries[0].Amount.Add(entries[1].Amount)
	wallet := repo.getWalletState("w-1")
	assert.True(t, wallet.Balance.Equal(ledgerSum), "balance %s != ledger sum %s", wallet.Balance, ledgerSum)
}
