package seeder

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/vizucode/e-wallet-system/internal/app/dto/models"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) (err error) {
	users := []models.User{
		{
			ID:        uuid.New().String(),
			Email:     "alice@example.com",
			Name:      "Alice Johnson",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Email:     "bob@example.com",
			Name:      "Bob Smith",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Email:     "charlie@example.com",
			Name:      "Charlie Brown",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Email:     "diana@example.com",
			Name:      "Diana Prince",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Email:     "eve@example.com",
			Name:      "Eve Williams",
			CreatedAt: time.Now(),
		},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", users[i].Name, err)
		}
	}

	now := time.Now()
	wallets := []models.Wallet{
		{
			ID:        uuid.New().String(),
			OwnerID:   users[0].ID,
			Currency:  "IDR",
			Balance:   decimal.NewFromFloat(5000000.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[0].ID,
			Currency:  "USD",
			Balance:   decimal.NewFromFloat(500.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[1].ID,
			Currency:  "IDR",
			Balance:   decimal.NewFromFloat(3000000.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[1].ID,
			Currency:  "USD",
			Balance:   decimal.NewFromFloat(250.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[2].ID,
			Currency:  "IDR",
			Balance:   decimal.NewFromFloat(10000000.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[3].ID,
			Currency:  "IDR",
			Balance:   decimal.NewFromFloat(7500000.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New().String(),
			OwnerID:   users[4].ID,
			Currency:  "IDR",
			Balance:   decimal.NewFromFloat(1500000.00),
			Status:    "ACTIVE",
			Version:   0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for i := range wallets {
		if err := db.Create(&wallets[i]).Error; err != nil {
			return fmt.Errorf("failed to seed wallet for user %s: %w", wallets[i].OwnerID, err)
		}
	}

	aliceIDR := wallets[0]
	bobIDR := wallets[2]
	charlieIDR := wallets[4]
	dianaIDR := wallets[5]

	bobWalletID := bobIDR.ID
	aliceWalletID := aliceIDR.ID
	charlieWalletID := charlieIDR.ID

	ledgerEntries := []models.LedgerEntry{
		{
			ID:          uuid.New().String(),
			WalletID:    aliceIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(5000000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-48 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			WalletID:        aliceIDR.ID,
			RelatedWalletID: &bobWalletID,
			Currency:        "IDR",
			Amount:          decimal.NewFromFloat(-1000000.00),
			EntryType:       "TRANSFER_OUT",
			ReferenceID:     fmt.Sprintf("TF-%s", uuid.New().String()[:8]),
			CreatedAt:       now.Add(-24 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			WalletID:        bobIDR.ID,
			RelatedWalletID: &aliceWalletID,
			Currency:        "IDR",
			Amount:          decimal.NewFromFloat(1000000.00),
			EntryType:       "TRANSFER_IN",
			ReferenceID:     fmt.Sprintf("TF-%s", uuid.New().String()[:8]),
			CreatedAt:       now.Add(-24 * time.Hour),
		},
		{
			ID:          uuid.New().String(),
			WalletID:    charlieIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(10000000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-12 * time.Hour),
		},
		{
			ID:          uuid.New().String(),
			WalletID:    dianaIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(7500000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-6 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			WalletID:        charlieIDR.ID,
			RelatedWalletID: &dianaIDR.ID,
			Currency:        "IDR",
			Amount:          decimal.NewFromFloat(-2000000.00),
			EntryType:       "TRANSFER_OUT",
			ReferenceID:     fmt.Sprintf("TF-%s", uuid.New().String()[:8]),
			CreatedAt:       now.Add(-3 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			WalletID:        dianaIDR.ID,
			RelatedWalletID: &charlieWalletID,
			Currency:        "IDR",
			Amount:          decimal.NewFromFloat(2000000.00),
			EntryType:       "TRANSFER_IN",
			ReferenceID:     fmt.Sprintf("TF-%s", uuid.New().String()[:8]),
			CreatedAt:       now.Add(-3 * time.Hour),
		},
	}

	for i := range ledgerEntries {
		if err := db.Create(&ledgerEntries[i]).Error; err != nil {
			return fmt.Errorf("failed to seed ledger entry: %w", err)
		}
	}

	return nil
}
