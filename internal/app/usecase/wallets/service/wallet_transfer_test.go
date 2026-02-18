package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
)

func TestTransfer_Success(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "1000.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "200.00"))
	svc := NewWalletService(repo)

	resp, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "150.75", ReferenceID: "xfer-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "849.25", resp.FromWallet.Balance)
	assert.Equal(t, "350.75", resp.ToWallet.Balance)
	assert.Equal(t, 2, repo.getLedgerCount())
}

func TestTransfer_CurrencyMismatch(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-usd", "USD", "500.00"))
	repo.addWallet(newActiveWallet("w-eur", "EUR", "500.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-usd", ToWalletID: "w-eur",
		Amount: "100.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrCurrencyMismatch)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestTransfer_SameWalletRejected(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-1", ToWalletID: "w-1",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrSameWallet)
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "50.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "200.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "50.01", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrInsufficientBalance)
	assert.Equal(t, "50.00", repo.getWalletState("w-a").Balance.StringFixed(2))
	assert.Equal(t, "200.00", repo.getWalletState("w-b").Balance.StringFixed(2))
}

func TestTransfer_SenderSuspended(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newSuspendedWallet("w-a", "USD", "100.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "200.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletSuspended)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestTransfer_ReceiverSuspended(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "100.00"))
	repo.addWallet(newSuspendedWallet("w-b", "USD", "200.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletSuspended)
	assert.Equal(t, 0, repo.getLedgerCount())
}

func TestTransfer_SenderNotFound(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-b", "USD", "200.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "nonexistent", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestTransfer_ReceiverNotFound(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "200.00"))
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "nonexistent",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrWalletNotFound)
}

func TestTransfer_ZeroAmount(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "0.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrAmountNotPositive)
}

func TestTransfer_EmptyFromWalletID(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "", ToWalletID: "w-b",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrFromWalletRequired)
}

func TestTransfer_EmptyToWalletID(t *testing.T) {
	repo := newMockWalletRepository()
	svc := NewWalletService(repo)

	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "",
		Amount: "10.00", ReferenceID: "xfer-1",
	})
	assert.ErrorIs(t, err, ErrToWalletRequired)
}

func TestTransfer_DuplicateReferenceID_Idempotent(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "500.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "100.00"))
	svc := NewWalletService(repo)

	resp1, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "100.00", ReferenceID: "xfer-dup",
	})
	require.NoError(t, err)
	assert.Equal(t, "400.00", resp1.FromWallet.Balance)
	assert.Equal(t, "200.00", resp1.ToWallet.Balance)

	resp2, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "100.00", ReferenceID: "xfer-dup",
	})
	require.NoError(t, err)
	assert.Equal(t, "400.00", resp2.FromWallet.Balance)
	assert.Equal(t, "200.00", resp2.ToWallet.Balance)
	assert.Equal(t, 2, repo.getLedgerCount())
}

func TestTransfer_LedgerConsistency(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "0.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "0.00"))
	svc := NewWalletService(repo)

	_, _ = svc.TopUpWallet("w-a", domains.TopUpWalletRequest{Amount: "500.00", ReferenceID: "t-1"})
	_, _ = svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "200.00", ReferenceID: "xfer-1",
	})

	for _, wid := range []string{"w-a", "w-b"} {
		entries := repo.getLedgerEntriesForWallet(wid)
		ledgerSum := decimal.Zero
		for _, e := range entries {
			ledgerSum = ledgerSum.Add(e.Amount)
		}
		wallet := repo.getWalletState(wid)
		assert.True(t, wallet.Balance.Equal(ledgerSum),
			"wallet %s: balance %s != ledger sum %s", wid, wallet.Balance, ledgerSum)
	}

	assert.Equal(t, "300.00", repo.getWalletState("w-a").Balance.StringFixed(2))
	assert.Equal(t, "200.00", repo.getWalletState("w-b").Balance.StringFixed(2))
}

func TestTransfer_PartialFailure_NoStateChange(t *testing.T) {
	repo := newMockWalletRepository()
	repo.addWallet(newActiveWallet("w-a", "USD", "500.00"))
	repo.addWallet(newActiveWallet("w-b", "USD", "100.00"))
	svc := NewWalletService(repo)

	repo.failOnUpdate = true
	_, err := svc.TransferWallet(domains.TransferWalletRequest{
		FromWalletID: "w-a", ToWalletID: "w-b",
		Amount: "100.00", ReferenceID: "xfer-fail",
	})
	assert.ErrorIs(t, err, ErrConcurrentUpdate)
	assert.Equal(t, "500.00", repo.getWalletState("w-a").Balance.StringFixed(2))
	assert.Equal(t, "100.00", repo.getWalletState("w-b").Balance.StringFixed(2))
}
