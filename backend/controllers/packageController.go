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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input inválido. Todos os campos são necessários."})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		err := tx.First(&pkg, "tracking_id = ?", body.TrackingID).Error

		// Obtém a hora atual para a transação.
		currentTime, timeErr := services.GetWorldTime()
		if timeErr != nil {
			log.Printf("AVISO: Falha ao buscar a hora mundial, usando a hora do servidor: %v", timeErr)
		}

		// CASO 1: O QR Code NÃO EXISTE no banco de dados (é um código antigo/legado).
		if err == gorm.ErrRecordNotFound {
			// Cria um novo registro de pacote já ativo.
			newPackage := models.Package{
				TrackingID:     body.TrackingID,
				Buffer:         body.Buffer,
				Rua:            body.Rua,
				EntryTimestamp: currentTime,
			}
			if err := tx.Create(&newPackage).Error; err != nil {
				return err
			}
			// CASO 2: O QR Code EXISTE no banco de dados.
		} else if err == nil {
			// Verifica se já está ativo.
			if pkg.Buffer != "PENDENTE" {
				return fmt.Errorf("O item %s já se encontra na fila (Buffer: %s, Rua: %s)", pkg.TrackingID, pkg.Buffer, pkg.Rua)
			}
			// Se estiver "PENDENTE", ativa-o atualizando os campos.
			updates := models.Package{
				Buffer:         body.Buffer,
				Rua:            body.Rua,
				EntryTimestamp: currentTime,
			}
			if err := tx.Model(&pkg).Updates(updates).Error; err != nil {
				return err
			}
			// CASO 3: Ocorreu outro erro de banco de dados.
		} else {
			return err
		}

		// Cria o log de auditoria para ambos os casos (criação ou atualização).
		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s entrou no buffer %s na rua %s", body.TrackingID, body.Buffer, body.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User), "ENTRADA", logDetails); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entrada do item registrada com sucesso."})
}

// --- FUNÇÃO PackageExit CORRIGIDA E SEGURA ---
func PackageExit(c *gin.Context) {
	var body struct {
		TrackingID string `json:"trackingId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O Tracking ID é obrigatório."})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		// 1. Encontra o pacote.
		if err := tx.First(&pkg, "tracking_id = ?", body.TrackingID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("item não encontrado")
			}
			return err
		}

		// 2. VALIDAÇÃO CRÍTICA: Se o buffer for "PENDENTE", a operação é abortada.
		if pkg.Buffer == "PENDENTE" {
			return fmt.Errorf("este item existe, mas ainda não teve entrada na fila e não pode ser removido")
		}

		// 3. Apenas se a validação passar, o log e a remoção são executados.
		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s foi removida do buffer %s na rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User), "SAIDA", logDetails); err != nil {
			return err
		}

		if err := tx.Delete(&pkg).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "item não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado na fila."})
			return
		}
		if err.Error() == "este item existe, mas ainda não teve entrada na fila e não pode ser removido" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao remover o item."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removido da fila com sucesso."})
}

func MovePackage(c *gin.Context) {
	packageID := c.Param("id")

	var body struct {
		Rua string `json:"rua" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'rua' é obrigatório."})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		if err := tx.First(&pkg, packageID).Error; err != nil {
			return fmt.Errorf("item não encontrado")
		}

		oldRua := pkg.Rua
		newRua := body.Rua

		if err := tx.Model(&pkg).Update("rua", newRua).Error; err != nil {
			return err
		}

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
	initializers.DB.Where("buffer <> ?", "PENDENTE").Order("entry_timestamp asc").Find(&packages)
	c.JSON(http.StatusOK, gin.H{"data": packages})
}

func GetBacklogCount(c *gin.Context) {
	var count int64
	initializers.DB.Model(&models.Package{}).Where("buffer <> ?", "PENDENTE").Count(&count)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func GetAuditLogs(c *gin.Context) {
	username := c.Query("username")
	fullname := c.Query("fullname")
	action := c.Query("action")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := initializers.DB.Order("created_at desc")

	if username != "" {
		query = query.Where("username ILIKE ?", "%"+username+"%")
	}
	if fullname != "" {
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

// GetBufferCounts retorna a contagem de pacotes para cada buffer principal.
func GetBufferCounts(c *gin.Context) {
	var rtsCount, ehaCount, salCount int64

	// Conta separadamente para cada buffer que está ativo na fila
	initializers.DB.Model(&models.Package{}).Where("buffer = ?", "RTS").Count(&rtsCount)
	initializers.DB.Model(&models.Package{}).Where("buffer = ?", "EHA").Count(&ehaCount)
	initializers.DB.Model(&models.Package{}).Where("buffer = ?", "SAL").Count(&salCount)

	c.JSON(http.StatusOK, gin.H{
		"rts": rtsCount,
		"eha": ehaCount,
		"sal": salCount,
	})
}
