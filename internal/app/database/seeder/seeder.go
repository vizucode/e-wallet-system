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
	// Create 5 users
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

	// Create wallets for each user (IDR + USD)
	now := time.Now()
	wallets := []models.Wallet{
		// Alice wallets
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
		// Bob wallets
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
		// Charlie wallet
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
		// Diana wallet
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
		// Eve wallet
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

	// Create ledger entries (sample transactions)
	aliceIDR := wallets[0]   // Alice IDR wallet
	bobIDR := wallets[2]     // Bob IDR wallet
	charlieIDR := wallets[4] // Charlie IDR wallet
	dianaIDR := wallets[5]   // Diana IDR wallet

	bobWalletID := bobIDR.ID
	aliceWalletID := aliceIDR.ID
	charlieWalletID := charlieIDR.ID

	ledgerEntries := []models.LedgerEntry{
		// Alice tops up IDR 5,000,000
		{
			ID:          uuid.New().String(),
			WalletID:    aliceIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(5000000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-48 * time.Hour),
		},
		// Alice transfers 1,000,000 to Bob (debit from Alice)
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
		// Bob receives 1,000,000 from Alice (credit to Bob)
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
		// Charlie tops up IDR 10,000,000
		{
			ID:          uuid.New().String(),
			WalletID:    charlieIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(10000000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-12 * time.Hour),
		},
		// Diana tops up IDR 7,500,000
		{
			ID:          uuid.New().String(),
			WalletID:    dianaIDR.ID,
			Currency:    "IDR",
			Amount:      decimal.NewFromFloat(7500000.00),
			EntryType:   "TOP_UP",
			ReferenceID: fmt.Sprintf("TU-%s", uuid.New().String()[:8]),
			CreatedAt:   now.Add(-6 * time.Hour),
		},
		// Charlie transfers 2,000,000 to Diana
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
		// Diana receives 2,000,000 from Charlie
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
