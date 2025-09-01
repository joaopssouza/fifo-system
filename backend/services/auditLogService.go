// backend/services/auditLogService.go
package services

import (
	"fifo-system/backend/models"
	"fmt"

	"gorm.io/gorm"
)

// CreateAuditLog agora também regista o nome completo do utilizador.
func CreateAuditLog(tx *gorm.DB, user models.User, action string, details string) error {
	auditLog := models.AuditLog{
		Username:     user.Username,
		UserFullname: user.FullName, // <-- NOVO CAMPO
		Action:       action,
		Details:      details,
	}

	if err := tx.Create(&auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
