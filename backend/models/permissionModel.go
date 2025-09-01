// backend/models/permissionModel.go
package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	Name        string `gorm:"unique;not null"` // Ex: "EDIT_USER", "VIEW_LOGS"
	Description string
}
