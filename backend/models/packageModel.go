// backend/models/packageModel.go
package models

import (
	"time"
	"gorm.io/gorm"
)

// --- MODELO ATUALIZADO PARA USAR GORM.MODEL (HABILITANDO SOFT DELETE) ---
type Package struct {
	gorm.Model // Contém ID, CreatedAt, UpdatedAt e DeletedAt (para o soft delete)

	TrackingID     string `gorm:"unique;not null"`
	Buffer         string `gorm:"not null"`
	Rua            string `gorm:"not null"`
	EntryTimestamp time.Time
}