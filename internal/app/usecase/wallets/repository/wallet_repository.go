package repository

import (
	"github.com/shopspring/decimal"
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
	"gorm.io/gorm"
)

type WalletRepository interface {
	GetWalletsByOwnerID(ownerID string) ([]models.Wallet, error)
	GetWalletByOwnerIDAndCurrency(ownerID string, currency string) (*models.Wallet, error)
	CreateWallet(wallet *models.Wallet) error
	BeginTx() *gorm.DB
	GetWalletByIDForUpdate(tx *gorm.DB, walletID string) (*models.Wallet, error)
	UpdateWalletBalance(tx *gorm.DB, walletID string, newBalance decimal.Decimal, currentVersion int) error
	CreateLedgerEntry(tx *gorm.DB, entry *models.LedgerEntry) error
	GetLedgerEntryByReferenceID(tx *gorm.DB, referenceID string) (*models.LedgerEntry, error)
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetWalletsByOwnerID(ownerID string) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Select("id", "owner_id", "currency", "balance", "status").
		Where("owner_id = ?", ownerID).
		Find(&wallets).Error
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepository) GetWalletByOwnerIDAndCurrency(ownerID string, currency string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("owner_id = ? AND currency = ?", ownerID, currency).
		First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) CreateWallet(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

func (r *walletRepository) GetWalletByIDForUpdate(tx *gorm.DB, walletID string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := tx.Where("id = ?", walletID).
		First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateWalletBalance(tx *gorm.DB, walletID string, newBalance decimal.Decimal, currentVersion int) error {
	result := tx.Model(&models.Wallet{}).
		Where("id = ? AND version = ?", walletID, currentVersion).
		Updates(map[string]interface{}{
			"balance": newBalance,
			"version": currentVersion + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *walletRepository) CreateLedgerEntry(tx *gorm.DB, entry *models.LedgerEntry) error {
	return tx.Create(entry).Error
}

func (r *walletRepository) GetLedgerEntryByReferenceID(tx *gorm.DB, referenceID string) (*models.LedgerEntry, error) {
	var entry models.LedgerEntry
	err := tx.Where("reference_id = ?", referenceID).First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}
