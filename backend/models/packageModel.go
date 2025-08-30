// backend/models/packageModel.go
package models

import (
	"time"
)

// Package struct sem o gorm.Model para desativar o soft delete.
type Package struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// O campo DeletedAt foi removido.

	TrackingID     string `gorm:"unique;not null"`
	Buffer         string `gorm:"not null"`
	Rua            string `gorm:"not null"`
	EntryTimestamp time.Time
}
