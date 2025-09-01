// backend/models/userModel.go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FullName     string `gorm:"not null"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Sector       string `gorm:"not null;default:'Geral'"`
	RoleID       uint   // Chave estrangeira
	Role         Role   // Relacionamento
}
