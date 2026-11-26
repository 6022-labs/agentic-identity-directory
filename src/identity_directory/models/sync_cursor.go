package models

import "time"

// SyncCursor is the last block a chain is indexed up to, included.
type SyncCursor struct {
	ChainId   uint64 `gorm:"primaryKey;autoIncrement:false"`
	Block     uint64 `gorm:"not null"`
	UpdatedAt time.Time
}
