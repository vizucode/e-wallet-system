package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/vizucode/e-wallet-system/internal/app/dto/domains"
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
	"github.com/vizucode/e-wallet-system/internal/app/usecase/wallets/repository"
)

var supportedCurrencies = map[string]bool{
	"USD": true, "EUR": true, "GBP": true, "JPY": true,
	"IDR": true, "SGD": true, "MYR": true, "AUD": true,
	"CAD": true, "CHF": true, "CNY": true, "HKD": true,
	"KRW": true, "THB": true, "PHP": true, "INR": true,
}

var (
	ErrInvalidCurrency   = errors.New("invalid currency code, must be a valid ISO 4217 currency (e.g. USD, EUR, IDR)")
	ErrWalletExists      = errors.New("wallet with this currency already exists for this user")
	ErrUserIDRequired    = errors.New("user_id is required")
	ErrCurrencyRequired  = errors.New("currency is required")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrWalletSuspended   = errors.New("wallet is suspended, top-up is not allowed")
	ErrInvalidAmount     = errors.New("amount must be a valid decimal number")
	ErrAmountNotPositive = errors.New("amount must be greater than zero")
	ErrAmountPrecision   = errors.New("amount must not have more than 2 decimal places (e.g. 12.50)")
	ErrReferenceRequired = errors.New("reference_id is required")
	ErrDuplicateTopUp    = errors.New("a top-up with this reference_id has already been processed")
	ErrConcurrentUpdate  = errors.New("wallet was modified by another request, please retry")
)

var smallestUnit = decimal.NewFromFloat(0.01)

type WalletService interface {
	GetWalletsByUserID(userID string) (domains.GetUserWalletsResponse, error)
	CreateWallet(req domains.CreateWalletRequest) (domains.CreateWalletResponse, error)
	TopUpWallet(walletID string, req domains.TopUpWalletRequest) (domains.TopUpWalletResponse, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
}

func NewWalletService(walletRepo repository.WalletRepository) WalletService {
	return &walletService{walletRepo: walletRepo}
}

func (s *walletService) GetWalletsByUserID(userID string) (domains.GetUserWalletsResponse, error) {
	wallets, err := s.walletRepo.GetWalletsByOwnerID(userID)
	if err != nil {
		return domains.GetUserWalletsResponse{}, err
	}

	walletResponses := make([]domains.WalletResponse, 0, len(wallets))
	for _, w := range wallets {
		walletResponses = append(walletResponses, domains.WalletResponse{
			WalletID: w.ID,
			Currency: w.Currency,
			Balance:  w.Balance.StringFixed(2),
			Status:   w.Status,
		})
	}

	return domains.GetUserWalletsResponse{
		UserID:  userID,
		Wallets: walletResponses,
	}, nil
}

func (s *walletService) CreateWallet(req domains.CreateWalletRequest) (domains.CreateWalletResponse, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return domains.CreateWalletResponse{}, ErrUserIDRequired
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		return domains.CreateWalletResponse{}, ErrCurrencyRequired
	}
	if !supportedCurrencies[currency] {
		return domains.CreateWalletResponse{}, fmt.Errorf("%w: '%s' is not supported", ErrInvalidCurrency, currency)
	}

	existing, err := s.walletRepo.GetWalletByOwnerIDAndCurrency(req.UserID, currency)
	if err != nil {
		return domains.CreateWalletResponse{}, fmt.Errorf("failed to check existing wallet: %w", err)
	}
	if existing != nil {
		return domains.CreateWalletResponse{}, fmt.Errorf("%w: user already has a %s wallet", ErrWalletExists, currency)
	}

	wallet := &models.Wallet{
		ID:       uuid.New().String(),
		OwnerID:  req.UserID,
		Currency: currency,
		Balance:  decimal.NewFromFloat(0),
		Status:   "ACTIVE",
		Version:  0,
	}

	if err := s.walletRepo.CreateWallet(wallet); err != nil {
		return domains.CreateWalletResponse{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	return domains.CreateWalletResponse{
		WalletID: wallet.ID,
		UserID:   wallet.OwnerID,
		Currency: wallet.Currency,
		Balance:  wallet.Balance.StringFixed(2),
		Status:   wallet.Status,
	}, nil
}

func (s *walletService) TopUpWallet(walletID string, req domains.TopUpWalletRequest) (domains.TopUpWalletResponse, error) {
	referenceID := strings.TrimSpace(req.ReferenceID)
	if referenceID == "" {
		return domains.TopUpWalletResponse{}, ErrReferenceRequired
	}

	amountStr := strings.TrimSpace(req.Amount)
	if amountStr == "" {
		return domains.TopUpWalletResponse{}, ErrInvalidAmount
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return domains.TopUpWalletResponse{}, fmt.Errorf("%w: '%s'", ErrInvalidAmount, amountStr)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return domains.TopUpWalletResponse{}, ErrAmountNotPositive
	}

	if amount.LessThan(smallestUnit) {
		return domains.TopUpWalletResponse{}, ErrAmountPrecision
	}

	rounded := amount.Round(2)
	if !amount.Equal(rounded) {
		return domains.TopUpWalletResponse{}, fmt.Errorf("%w: received '%s', did you mean '%s'?", ErrAmountPrecision, amount.String(), rounded.String())
	}
	amount = rounded

	tx := s.walletRepo.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	existingEntry, err := s.walletRepo.GetLedgerEntryByReferenceID(tx, referenceID)
	if err != nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, fmt.Errorf("failed to check reference_id: %w", err)
	}
	if existingEntry != nil {
		tx.Rollback()
		wallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, walletID)
		if err != nil || wallet == nil {
			return domains.TopUpWalletResponse{}, ErrDuplicateTopUp
		}
		return domains.TopUpWalletResponse{
			WalletID: wallet.ID,
			Currency: strings.TrimSpace(wallet.Currency),
			Balance:  wallet.Balance.StringFixed(2),
			Status:   wallet.Status,
		}, nil
	}

	wallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, walletID)
	if err != nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if wallet == nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, ErrWalletNotFound
	}

	if wallet.Status == "SUSPENDED" {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, ErrWalletSuspended
	}

	newBalance := wallet.Balance.Add(amount)

	ledgerEntry := &models.LedgerEntry{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		Currency:    strings.TrimSpace(wallet.Currency),
		Amount:      amount,
		EntryType:   "TOP_UP",
		ReferenceID: referenceID,
	}

	if err := s.walletRepo.CreateLedgerEntry(tx, ledgerEntry); err != nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	if err := s.walletRepo.UpdateWalletBalance(tx, wallet.ID, newBalance, wallet.Version); err != nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, ErrConcurrentUpdate
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return domains.TopUpWalletResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return domains.TopUpWalletResponse{
		WalletID: wallet.ID,
		Currency: strings.TrimSpace(wallet.Currency),
		Balance:  newBalance.StringFixed(2),
		Status:   wallet.Status,
	}, nil
}
