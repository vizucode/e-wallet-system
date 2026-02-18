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
	ErrInvalidCurrency     = errors.New("invalid currency code, must be a valid ISO 4217 currency (e.g. USD, EUR, IDR)")
	ErrWalletExists        = errors.New("wallet with this currency already exists for this user")
	ErrUserIDRequired      = errors.New("user_id is required")
	ErrCurrencyRequired    = errors.New("currency is required")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletSuspended     = errors.New("wallet is suspended, this operation is not allowed")
	ErrInvalidAmount       = errors.New("amount must be a valid decimal number")
	ErrAmountNotPositive   = errors.New("amount must be greater than zero")
	ErrAmountPrecision     = errors.New("amount must not have more than 2 decimal places (e.g. 12.50)")
	ErrReferenceRequired   = errors.New("reference_id is required")
	ErrDuplicateReference  = errors.New("a transaction with this reference_id has already been processed")
	ErrConcurrentUpdate    = errors.New("wallet was modified by another request, please retry")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrCurrencyMismatch    = errors.New("both wallets must use the same currency")
	ErrSameWallet          = errors.New("cannot transfer to the same wallet")
	ErrFromWalletRequired  = errors.New("from_wallet_id is required")
	ErrToWalletRequired    = errors.New("to_wallet_id is required")
)

var smallestUnit = decimal.NewFromFloat(0.01)

type WalletService interface {
	GetWalletsByUserID(userID string) (domains.GetUserWalletsResponse, error)
	CreateWallet(req domains.CreateWalletRequest) (domains.CreateWalletResponse, error)
	TopUpWallet(walletID string, req domains.TopUpWalletRequest) (domains.TopUpWalletResponse, error)
	PayWallet(walletID string, req domains.PayWalletRequest) (domains.PayWalletResponse, error)
	TransferWallet(req domains.TransferWalletRequest) (domains.TransferWalletResponse, error)
	SuspendWallet(walletID string) (domains.SuspendWalletResponse, error)
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

func (s *walletService) validateAmountInput(amountStr string) (decimal.Decimal, error) {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return decimal.Zero, ErrInvalidAmount
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w: '%s'", ErrInvalidAmount, amountStr)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrAmountNotPositive
	}

	if amount.LessThan(smallestUnit) {
		return decimal.Zero, ErrAmountPrecision
	}

	rounded := amount.Round(2)
	if !amount.Equal(rounded) {
		return decimal.Zero, fmt.Errorf("%w: received '%s', did you mean '%s'?", ErrAmountPrecision, amount.String(), rounded.String())
	}

	return rounded, nil
}

func (s *walletService) TopUpWallet(walletID string, req domains.TopUpWalletRequest) (domains.TopUpWalletResponse, error) {
	referenceID := strings.TrimSpace(req.ReferenceID)
	if referenceID == "" {
		return domains.TopUpWalletResponse{}, ErrReferenceRequired
	}

	amount, err := s.validateAmountInput(req.Amount)
	if err != nil {
		return domains.TopUpWalletResponse{}, err
	}

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
			return domains.TopUpWalletResponse{}, ErrDuplicateReference
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

func (s *walletService) PayWallet(walletID string, req domains.PayWalletRequest) (domains.PayWalletResponse, error) {
	referenceID := strings.TrimSpace(req.ReferenceID)
	if referenceID == "" {
		return domains.PayWalletResponse{}, ErrReferenceRequired
	}

	amount, err := s.validateAmountInput(req.Amount)
	if err != nil {
		return domains.PayWalletResponse{}, err
	}

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
		return domains.PayWalletResponse{}, fmt.Errorf("failed to check reference_id: %w", err)
	}
	if existingEntry != nil {
		tx.Rollback()
		wallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, walletID)
		if err != nil || wallet == nil {
			return domains.PayWalletResponse{}, ErrDuplicateReference
		}
		return domains.PayWalletResponse{
			WalletID: wallet.ID,
			Currency: strings.TrimSpace(wallet.Currency),
			Balance:  wallet.Balance.StringFixed(2),
			Status:   wallet.Status,
		}, nil
	}

	wallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, walletID)
	if err != nil {
		tx.Rollback()
		return domains.PayWalletResponse{}, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if wallet == nil {
		tx.Rollback()
		return domains.PayWalletResponse{}, ErrWalletNotFound
	}

	if wallet.Status == "SUSPENDED" {
		tx.Rollback()
		return domains.PayWalletResponse{}, ErrWalletSuspended
	}

	if wallet.Balance.LessThan(amount) {
		tx.Rollback()
		return domains.PayWalletResponse{}, fmt.Errorf(
			"%w: available balance is %s %s, but payment requires %s",
			ErrInsufficientBalance,
			strings.TrimSpace(wallet.Currency),
			wallet.Balance.StringFixed(2),
			amount.StringFixed(2),
		)
	}

	newBalance := wallet.Balance.Sub(amount)

	ledgerEntry := &models.LedgerEntry{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		Currency:    strings.TrimSpace(wallet.Currency),
		Amount:      amount.Neg(),
		EntryType:   "PAYMENT",
		ReferenceID: referenceID,
	}

	if err := s.walletRepo.CreateLedgerEntry(tx, ledgerEntry); err != nil {
		tx.Rollback()
		return domains.PayWalletResponse{}, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	if err := s.walletRepo.UpdateWalletBalance(tx, wallet.ID, newBalance, wallet.Version); err != nil {
		tx.Rollback()
		return domains.PayWalletResponse{}, ErrConcurrentUpdate
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return domains.PayWalletResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return domains.PayWalletResponse{
		WalletID: wallet.ID,
		Currency: strings.TrimSpace(wallet.Currency),
		Balance:  newBalance.StringFixed(2),
		Status:   wallet.Status,
	}, nil
}

func (s *walletService) TransferWallet(req domains.TransferWalletRequest) (domains.TransferWalletResponse, error) {
	fromWalletID := strings.TrimSpace(req.FromWalletID)
	if fromWalletID == "" {
		return domains.TransferWalletResponse{}, ErrFromWalletRequired
	}

	toWalletID := strings.TrimSpace(req.ToWalletID)
	if toWalletID == "" {
		return domains.TransferWalletResponse{}, ErrToWalletRequired
	}

	if fromWalletID == toWalletID {
		return domains.TransferWalletResponse{}, ErrSameWallet
	}

	referenceID := strings.TrimSpace(req.ReferenceID)
	if referenceID == "" {
		return domains.TransferWalletResponse{}, ErrReferenceRequired
	}

	amount, err := s.validateAmountInput(req.Amount)
	if err != nil {
		return domains.TransferWalletResponse{}, err
	}

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
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to check reference_id: %w", err)
	}
	if existingEntry != nil {
		tx.Rollback()
		fromWallet, err1 := s.walletRepo.GetWalletByIDForUpdate(tx, fromWalletID)
		toWallet, err2 := s.walletRepo.GetWalletByIDForUpdate(tx, toWalletID)
		if err1 != nil || err2 != nil || fromWallet == nil || toWallet == nil {
			return domains.TransferWalletResponse{}, ErrDuplicateReference
		}
		return domains.TransferWalletResponse{
			FromWallet: domains.TransferWalletDetail{
				WalletID: fromWallet.ID,
				Currency: strings.TrimSpace(fromWallet.Currency),
				Balance:  fromWallet.Balance.StringFixed(2),
				Status:   fromWallet.Status,
			},
			ToWallet: domains.TransferWalletDetail{
				WalletID: toWallet.ID,
				Currency: strings.TrimSpace(toWallet.Currency),
				Balance:  toWallet.Balance.StringFixed(2),
				Status:   toWallet.Status,
			},
		}, nil
	}

	firstID, secondID := fromWalletID, toWalletID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	firstWallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, firstID)
	if err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if firstWallet == nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("%w: wallet %s does not exist", ErrWalletNotFound, firstID)
	}

	secondWallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, secondID)
	if err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if secondWallet == nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("%w: wallet %s does not exist", ErrWalletNotFound, secondID)
	}

	var fromWallet, toWallet *models.Wallet
	if firstID == fromWalletID {
		fromWallet, toWallet = firstWallet, secondWallet
	} else {
		fromWallet, toWallet = secondWallet, firstWallet
	}

	fromCurrency := strings.TrimSpace(fromWallet.Currency)
	toCurrency := strings.TrimSpace(toWallet.Currency)

	if fromCurrency != toCurrency {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf(
			"%w: sender wallet uses %s but receiver wallet uses %s",
			ErrCurrencyMismatch, fromCurrency, toCurrency,
		)
	}

	if fromWallet.Status == "SUSPENDED" {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf(
			"%w: sender wallet %s is suspended",
			ErrWalletSuspended, fromWallet.ID,
		)
	}

	if toWallet.Status == "SUSPENDED" {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf(
			"%w: receiver wallet %s is suspended",
			ErrWalletSuspended, toWallet.ID,
		)
	}

	if fromWallet.Balance.LessThan(amount) {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf(
			"%w: available balance is %s %s, but transfer requires %s",
			ErrInsufficientBalance,
			fromCurrency,
			fromWallet.Balance.StringFixed(2),
			amount.StringFixed(2),
		)
	}

	newFromBalance := fromWallet.Balance.Sub(amount)
	newToBalance := toWallet.Balance.Add(amount)

	debitEntry := &models.LedgerEntry{
		ID:              uuid.New().String(),
		WalletID:        fromWallet.ID,
		RelatedWalletID: &toWallet.ID,
		Currency:        fromCurrency,
		Amount:          amount.Neg(),
		EntryType:       "TRANSFER_OUT",
		ReferenceID:     referenceID + "-debit",
	}

	if err := s.walletRepo.CreateLedgerEntry(tx, debitEntry); err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to create debit ledger entry: %w", err)
	}

	creditEntry := &models.LedgerEntry{
		ID:              uuid.New().String(),
		WalletID:        toWallet.ID,
		RelatedWalletID: &fromWallet.ID,
		Currency:        fromCurrency,
		Amount:          amount,
		EntryType:       "TRANSFER_IN",
		ReferenceID:     referenceID + "-credit",
	}

	if err := s.walletRepo.CreateLedgerEntry(tx, creditEntry); err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to create credit ledger entry: %w", err)
	}

	if err := s.walletRepo.UpdateWalletBalance(tx, fromWallet.ID, newFromBalance, fromWallet.Version); err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, ErrConcurrentUpdate
	}

	if err := s.walletRepo.UpdateWalletBalance(tx, toWallet.ID, newToBalance, toWallet.Version); err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, ErrConcurrentUpdate
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return domains.TransferWalletResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return domains.TransferWalletResponse{
		FromWallet: domains.TransferWalletDetail{
			WalletID: fromWallet.ID,
			Currency: fromCurrency,
			Balance:  newFromBalance.StringFixed(2),
			Status:   fromWallet.Status,
		},
		ToWallet: domains.TransferWalletDetail{
			WalletID: toWallet.ID,
			Currency: fromCurrency,
			Balance:  newToBalance.StringFixed(2),
			Status:   toWallet.Status,
		},
	}, nil
}

func (s *walletService) SuspendWallet(walletID string) (domains.SuspendWalletResponse, error) {
	tx := s.walletRepo.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	wallet, err := s.walletRepo.GetWalletByIDForUpdate(tx, walletID)
	if err != nil {
		tx.Rollback()
		return domains.SuspendWalletResponse{}, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if wallet == nil {
		tx.Rollback()
		return domains.SuspendWalletResponse{}, ErrWalletNotFound
	}

	if wallet.Status == "SUSPENDED" {
		tx.Rollback()
		return domains.SuspendWalletResponse{
			WalletID: wallet.ID,
			Currency: strings.TrimSpace(wallet.Currency),
			Balance:  wallet.Balance.StringFixed(2),
			Status:   wallet.Status,
		}, nil
	}

	if err := s.walletRepo.UpdateWalletStatus(tx, wallet.ID, "SUSPENDED", wallet.Version); err != nil {
		tx.Rollback()
		return domains.SuspendWalletResponse{}, ErrConcurrentUpdate
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return domains.SuspendWalletResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return domains.SuspendWalletResponse{
		WalletID: wallet.ID,
		Currency: strings.TrimSpace(wallet.Currency),
		Balance:  wallet.Balance.StringFixed(2),
		Status:   "SUSPENDED",
	}, nil
}
