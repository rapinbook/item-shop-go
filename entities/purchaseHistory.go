package entities

import (
	"time"
)

type PurchaseHistory struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	ItemID          string    `gorm:"type:varchar(64);"`
	PlayerID        string    `gorm:"type:varchar(64);unique;not null;"`
	ItemName        string    `gorm:"type:varchar(64);"`
	ItemPicture     string    `gorm:"type:varchar(128);not null;"`
	ItemDescription string    `gorm:"type:varchar(128);not null;"`
	ItemPrice       uint      `gorm:"not null;"`
	Quantity        uint      `gorm:"not null;"`
	IsBuying        bool      `gorm:"type:boolean;not null;"`
	CreatedAt       time.Time `gorm:"not null;autoCreateTime;"`
	UpdatedAt       time.Time `gorm:"not null;autoUpdateTime;"`
}
