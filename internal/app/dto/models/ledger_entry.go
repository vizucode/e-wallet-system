package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type LedgerEntry struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	WalletID        string          `gorm:"column:wallet_id;type:uuid;not null" json:"wallet_id"`
	RelatedWalletID *string         `gorm:"column:related_wallet_id;type:uuid" json:"related_wallet_id,omitempty"`
	Currency        string          `gorm:"column:currency;type:char(3);not null" json:"currency"`
	Amount          decimal.Decimal `gorm:"column:amount;type:decimal(20,2);not null" json:"amount"`
	EntryType       string          `gorm:"column:entry_type;type:varchar(30);not null" json:"entry_type"`
	ReferenceID     string          `gorm:"column:reference_id;type:varchar(100);not null;uniqueIndex" json:"reference_id"`
	CreatedAt       time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Wallet        Wallet  `gorm:"foreignKey:WalletID;references:ID" json:"wallet,omitempty"`
	RelatedWallet *Wallet `gorm:"foreignKey:RelatedWalletID;references:ID" json:"related_wallet,omitempty"`
}
