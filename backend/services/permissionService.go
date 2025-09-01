// backend/services/permissionService.go
package services

import (
    "fifo-system/backend/initializers"
    "fifo-system/backend/models"
    "fmt"
)

// CanPerformAction agora verifica as permissões no banco de dados.
func CanPerformAction(actingUser models.User, targetUser *models.User, actionName string) (bool, error) {
    // 1. Carrega o papel do usuário com suas permissões
    var userRole models.Role
    if err := initializers.DB.Preload("Permissions").First(&userRole, actingUser.RoleID).Error; err != nil {
        return false, fmt.Errorf("could not find role for user %s", actingUser.Username)
    }

    // 2. Verifica se a permissão existe para o papel
    hasPermission := false
    for _, p := range userRole.Permissions {
        if p.Name == actionName {
            hasPermission = true
            break
        }
    }

    if !hasPermission {
        return false, nil // Não tem a permissão básica
    }

    // 3. Aplica regras de hierarquia (esta lógica pode ser mantida)
    if targetUser != nil {
        if actingUser.ID == targetUser.ID {
            return false, nil
        }

        var targetRole models.Role
        if err := initializers.DB.First(&targetRole, targetUser.RoleID).Error; err != nil {
             return false, fmt.Errorf("could not find role for target user %s", targetUser.Username)
        }

        // Exemplo: Um 'leader' não pode agir sobre um 'admin'
        if userRole.Name == "leader" && targetRole.Name == "admin" {
            return false, nil
        }
    }

    return true, nil
}