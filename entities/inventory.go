package entities

import (
	"time"
)

type Inventory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	ItemID    string    `gorm:"type:varchar(64);"`
	PlayerID  string    `gorm:"type:varchar(64);unique;not null;"`
	IsDeleted bool      `gorm:"not null;default:false;"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime;"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime;"`
}
