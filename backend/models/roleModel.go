// backend/models/roleModel.go
package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
	Users       []User       `gorm:"foreignKey:RoleID"`           // Relacionamento com User
	Permissions []Permission `gorm:"many2many:role_permissions;"` // Relacionamento N-para-N
}
