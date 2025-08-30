// backend/services/auditLogService.go
package services

import (
	"fifo-system/backend/models"
	"fmt"

	"gorm.io/gorm"
)

// CreateAuditLog cria um registro de auditoria dentro de uma transação de banco de dados.
// É crucial que ele receba o 'tx *gorm.DB' para garantir que a criação do log
// faça parte da mesma operação atômica que a ação principal (entrada/saída de pacote).
func CreateAuditLog(tx *gorm.DB, username string, action string, details string) error {
	auditLog := models.AuditLog{
		Username: username,
		Action:   action,
		Details:  details,
	}

	if err := tx.Create(&auditLog).Error; err != nil {
		// Retorna um erro formatado para facilitar a depuração
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
