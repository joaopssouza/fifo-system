// backend/models/auditLogModel.go
package models

import (
	"gorm.io/gorm"
)

type AuditLog struct {
	gorm.Model
	Username     string `gorm:"not null"`
	UserFullname string `gorm:"not null"`
	Action       string `gorm:"not null"` // Ex: "ENTRADA", "SAIDA"
	Details      string `gorm:"not null"` // Ex: "Package CG01 at RTS-001"
}
