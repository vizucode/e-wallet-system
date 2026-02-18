package models

import "time"

type User struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"column:email;type:varchar(100)" json:"email"`
	Name      string    `gorm:"column:name;type:varchar(100)" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Wallets []Wallet `gorm:"foreignKey:OwnerID;references:ID" json:"wallets,omitempty"`
}
