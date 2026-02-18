package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
	"gorm.io/gorm"
)

type mockConnPool struct{}

func (m *mockConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}
func (m *mockConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *mockConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

type mockTxCommitter struct {
	mockConnPool
	failOnCommit bool
}

func (m *mockTxCommitter) Commit() error {
	if m.failOnCommit {
		return errors.New("simulated commit failure")
	}
	return nil
}
func (m *mockTxCommitter) Rollback() error { return nil }

func newMockTx(failOnCommit bool) *gorm.DB {
	db := &gorm.DB{
		Config:    &gorm.Config{},
		Statement: &gorm.Statement{},
	}
	db.Statement.ConnPool = &mockTxCommitter{failOnCommit: failOnCommit}
	return db
}

type mockWalletRepository struct {
	mu            sync.Mutex
	wallets       map[string]*models.Wallet
	ledgerEntries map[string]*models.LedgerEntry
	allLedgers    []*models.LedgerEntry
	failOnCommit  bool
	failOnCreate  bool
	failOnUpdate  bool
	callCount     map[string]int
}

func newMockWalletRepository() *mockWalletRepository {
	return &mockWalletRepository{
		wallets:       make(map[string]*models.Wallet),
		ledgerEntries: make(map[string]*models.LedgerEntry),
		allLedgers:    make([]*models.LedgerEntry, 0),
		callCount:     make(map[string]int),
	}
}

func (m *mockWalletRepository) addWallet(w *models.Wallet) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := *w
	m.wallets[w.ID] = &copy
}

func (m *mockWalletRepository) getWalletState(id string) *models.Wallet {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wallets[id]
	if !ok {
		return nil
	}
	copy := *w
	return &copy
}

func (m *mockWalletRepository) getLedgerCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.allLedgers)
}

func (m *mockWalletRepository) getLedgerEntriesForWallet(walletID string) []*models.LedgerEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.LedgerEntry
	for _, e := range m.allLedgers {
		if e.WalletID == walletID {
			copy := *e
			result = append(result, &copy)
		}
	}
	return result
}

func (m *mockWalletRepository) GetWalletsByOwnerID(ownerID string) ([]models.Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.Wallet
	for _, w := range m.wallets {
		if w.OwnerID == ownerID {
			result = append(result, *w)
		}
	}
	return result, nil
}

func (m *mockWalletRepository) GetWalletByOwnerIDAndCurrency(ownerID string, currency string) (*models.Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.wallets {
		if w.OwnerID == ownerID && w.Currency == currency {
			copy := *w
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockWalletRepository) GetWalletByID(walletID string) (*models.Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wallets[walletID]
	if !ok {
		return nil, nil
	}
	copy := *w
	return &copy, nil
}

func (m *mockWalletRepository) CreateWallet(wallet *models.Wallet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("simulated create failure")
	}
	copy := *wallet
	m.wallets[wallet.ID] = &copy
	return nil
}

func (m *mockWalletRepository) BeginTx() *gorm.DB {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount["BeginTx"]++
	return newMockTx(m.failOnCommit)
}

func (m *mockWalletRepository) GetWalletByIDForUpdate(tx *gorm.DB, walletID string) (*models.Wallet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wallets[walletID]
	if !ok {
		return nil, nil
	}
	copy := *w
	return &copy, nil
}

func (m *mockWalletRepository) UpdateWalletBalance(tx *gorm.DB, walletID string, newBalance decimal.Decimal, currentVersion int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnUpdate {
		return gorm.ErrRecordNotFound
	}
	w, ok := m.wallets[walletID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if w.Version != currentVersion {
		return gorm.ErrRecordNotFound
	}
	w.Balance = newBalance
	w.Version = currentVersion + 1
	return nil
}

func (m *mockWalletRepository) UpdateWalletStatus(tx *gorm.DB, walletID string, status string, currentVersion int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnUpdate {
		return gorm.ErrRecordNotFound
	}
	w, ok := m.wallets[walletID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if w.Version != currentVersion {
		return gorm.ErrRecordNotFound
	}
	w.Status = status
	w.Version = currentVersion + 1
	return nil
}

func (m *mockWalletRepository) CreateLedgerEntry(tx *gorm.DB, entry *models.LedgerEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("simulated ledger create failure")
	}
	if _, exists := m.ledgerEntries[entry.ReferenceID]; exists {
		return errors.New("duplicate reference_id")
	}
	copy := *entry
	m.ledgerEntries[entry.ReferenceID] = &copy
	m.allLedgers = append(m.allLedgers, &copy)
	return nil
}

func (m *mockWalletRepository) GetLedgerEntryByReferenceID(tx *gorm.DB, referenceID string) (*models.LedgerEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.ledgerEntries[referenceID]; ok {
		copy := *e
		return &copy, nil
	}
	suffixes := []string{"-debit", "-credit"}
	for _, suffix := range suffixes {
		if e, ok := m.ledgerEntries[referenceID+suffix]; ok {
			copy := *e
			return &copy, nil
		}
	}
	return nil, nil
}
