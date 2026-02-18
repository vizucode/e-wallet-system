package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID        string          `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	OwnerID   string          `gorm:"column:owner_id;type:uuid;not null" json:"owner_id"`
	Currency  string          `gorm:"column:currency;type:char(3);not null" json:"currency"`
	Balance   decimal.Decimal `gorm:"column:balance;type:decimal(20,2);not null;default:0.00" json:"balance"`
	Status    string          `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'" json:"status"`
	Version   int             `gorm:"column:version;not null;default:0" json:"version"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Owner         User          `gorm:"foreignKey:OwnerID;references:ID" json:"owner,omitempty"`
	LedgerEntries []LedgerEntry `gorm:"foreignKey:WalletID;references:ID" json:"ledger_entries,omitempty"`
}
