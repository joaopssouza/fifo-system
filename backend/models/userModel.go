// backend/models/userModel.go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model   // gorm.Model já contém o campo ID como chave primária (que é indexada)
	FullName     string `gorm:"not null"`
	// Adiciona um índice único na coluna 'username' para buscas ultra-rápidas.
	Username     string `gorm:"unique;not null;index"` 
	PasswordHash string `gorm:"not null"`
	Sector       string `gorm:"not null;default:'Geral'"`
	RoleID       uint   // Chave estrangeira
	Role         Role   // Relacionamento
}