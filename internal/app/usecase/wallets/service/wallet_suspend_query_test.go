package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
)

func TestSuspend_Success(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "500.00"))
	svc := NewWalletService(repo)

	resp, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)
	assert.Equal(t, "SUSPENDED", resp.Status)
	assert.Equal(t, "500.00", resp.Balance)
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, "SUSPENDED", repo.getWalletState("w-1").Status)
}

func TestSuspend_AlreadySuspended_Idempotent(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newSuspendedWallet("w-1", "USD", "500.00"))
	svc := NewWalletService(repo)

	resp, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)
	assert.Equal(t, "SUSPENDED", resp.Status)
	assert.Equal(t, "500.00", resp.Balance)
}

func TestSuspend_WalletNotFound(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("nonexistent")
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestSuspend_BalancePreserved(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "12345.67"))
	svc := NewWalletService(repo)

	resp, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)
	assert.Equal(t, "12345.67", resp.Balance)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestSuspend_ThenTopUpRejected(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)

	_, err = svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "ref-1"})
	assert.ErrorIs(t, err, ErrWalletSuspended)
}

func TestSuspend_ThenPaymentRejected(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)

	_, err = svc.PayWallet("w-1", domains.PayWalletRequest{Amount: "10.00", ReferenceID: "pay-1"})
	assert.ErrorIs(t, err, ErrWalletSuspended)
}

func TestSuspend_ThenTransferRejected_AsSender(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "100.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("w-a")
	require.NoError(t, err)

	_, err = svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletSuspended)
}

func TestSuspend_ThenTransferRejected_AsReceiver(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "100.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("w-b")
	require.NoError(t, err)

	_, err = svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletSuspended)
}

func TestGetWalletByID_Success(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "1234.56"))
	svc := NewWalletService(repo)

	resp, err := svc.GetWalletByID("w-1")
	require.NoError(t, err)
	assert.Equal(t, "w-1", resp.WalletID)
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, "1234.56", resp.Balance)
	assert.Equal(t, "ACTIVE", resp.Status)
}

func TestGetWalletByID_NotFound(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.GetWalletByID("nonexistent")
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestGetWalletByID_SuspendedWalletStillReadable(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newSuspendedWallet("w-1", "USD", "500.00"))
	svc := NewWalletService(repo)

	resp, err := svc.GetWalletByID("w-1")
	require.NoError(t, err)
	assert.Equal(t, "SUSPENDED", resp.Status)
	assert.Equal(t, "500.00", resp.Balance)
}

func TestGetWalletByID_ReadAfterWrite(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "ref-1"})
	require.NoError(t, err)

	resp, err := svc.GetWalletByID("w-1")
	require.NoError(t, err)
	assert.Equal(t, "150.00", resp.Balance)
}

func TestGetWalletByID_ReadAfterSuspend(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	_, err := svc.SuspendWallet("w-1")
	require.NoError(t, err)

	resp, err := svc.GetWalletByID("w-1")
	require.NoError(t, err)
	assert.Equal(t, "SUSPENDED", resp.Status)
	assert.Equal(t, "100.00", resp.Balance)
}

func TestCrashRecovery_FailedUpdateDoesNotChangeLedger(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-1", "USD", "100.00"))
	svc := NewWalletService(repo)

	repo.failOnUpdate = true
	_, err := svc.TopUpWallet("w-1", domains.TopUpWalletRequest{Amount: "50.00", ReferenceID: "crash-1"})
	assert.ErrorIs(t, err, ErrConcurrentUpdate)

	assert.Equal(t, "100.00", repo.getWalletState("w-1").Balance.StringFixed(2))
}
