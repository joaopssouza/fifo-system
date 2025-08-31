// backend/models/userModel.go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:'fifo'"`  // Roles: fifo, adminv
	Sector       string `gorm:"not null;default:'Geral'"` // Novo campo: Setor
}
