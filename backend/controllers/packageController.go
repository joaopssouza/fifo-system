// backend/controllers/packageController.go
package controllers

import (
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
	"fmt"
	"log" // Adicionado para logar o aviso
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PackageEntry(c *gin.Context) {
	var body struct {
		TrackingID string `json:"trackingId" binding:"required"`
		Buffer     string `json:"buffer" binding:"required"`
		Rua        string `json:"rua" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input inválido"})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var existingPackage models.Package
		if err := tx.First(&existingPackage, "tracking_id = ?", body.TrackingID).Error; err == nil {
			return fmt.Errorf("Um item com este ID já existe na fila")
		}

		// --- INÍCIO DA ALTERAÇÃO ---
		// Obtém a hora autoritativa do nosso novo serviço
		currentTime, timeErr := services.GetWorldTime()
		if timeErr != nil {
			// Loga o erro, mas continua a usar o tempo do servidor como fallback
			log.Printf("AVISO: Falha ao buscar a hora mundial, a usar a hora do servidor: %v", timeErr)
		}

		pkg := models.Package{
			TrackingID:     body.TrackingID,
			Buffer:         body.Buffer,
			Rua:            body.Rua,
			EntryTimestamp: currentTime, // Usa a hora obtida
		}
		// --- FIM DA ALTERAÇÃO ---

		if err := tx.Create(&pkg).Error; err != nil {
			return err
		}

		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s entrou no buffer %s na rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User), "ENTRADA", logDetails); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "Um item com este ID já existe na fila" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Package registered successfully"})
}

// O resto das funções (PackageExit, MovePackage, etc.) permanecem as mesmas
// ... (COLE O RESTO DO SEU FICHEIRO packageController.go AQUI)
func PackageExit(c *gin.Context) {
	var body struct {
		TrackingID string `json:"trackingId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tracking ID is required"})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		if err := tx.First(&pkg, "tracking_id = ?", body.TrackingID).Error; err != nil {
			return fmt.Errorf("package not found")
		}

		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("Package %s removed from buffer %s at rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		// Passa o objeto user completo
		if err := services.CreateAuditLog(tx, user.(models.User), "SAIDA", logDetails); err != nil {
			return err
		}

		if err := tx.Delete(&pkg).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "package not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Package removed successfully"})
}

func MovePackage(c *gin.Context) {
	// Extrair o ID do pacote da URL
	packageID := c.Param("id")

	// Extrair a nova rua do corpo da requisição
	var body struct {
		Rua string `json:"rua" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'rua' é obrigatório."})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		// Encontra o pacote pelo seu ID primário
		if err := tx.First(&pkg, packageID).Error; err != nil {
			return fmt.Errorf("item não encontrado")
		}

		oldRua := pkg.Rua
		newRua := body.Rua

		// Atualiza o campo Rua
		if err := tx.Model(&pkg).Update("rua", newRua).Error; err != nil {
			return err
		}

		// Cria o registro de auditoria
		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s foi movida da rua %s para %s", pkg.TrackingID, oldRua, newRua)
		if err := services.CreateAuditLog(tx, user.(models.User), "MOVIMENTACAO", logDetails); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "item não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao mover o item."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item movido com sucesso."})
}

func GetFIFOQueue(c *gin.Context) {
	var packages []models.Package
	initializers.DB.Order("entry_timestamp asc").Find(&packages)
	c.JSON(http.StatusOK, gin.H{"data": packages})
}

func GetBacklogCount(c *gin.Context) {
	var count int64
	initializers.DB.Model(&models.Package{}).Count(&count)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetAuditLogs agora também pode filtrar pelo nome completo
func GetAuditLogs(c *gin.Context) {
	username := c.Query("username")
	fullname := c.Query("fullname") // <-- 1. Ler o novo parâmetro
	action := c.Query("action")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := initializers.DB.Order("created_at desc")

	// 2. Aplicar filtro de username se existir
	if username != "" {
		query = query.Where("username ILIKE ?", "%"+username+"%")
	}
	// 3. Aplicar filtro de fullname se existir
	if fullname != "" {
		// O nome da coluna no banco é 'user_fullname' conforme o model 'auditLogModel.go'
		query = query.Where("user_fullname ILIKE ?", "%"+fullname+"%")
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if startDate != "" && endDate != "" {
		endDateWithTime := endDate + " 23:59:59"
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDateWithTime)
	}

	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
