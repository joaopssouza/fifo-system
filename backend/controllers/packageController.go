// backend/controllers/packageController.go
package controllers

import (
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
	"fifo-system/backend/websocket" // <-- Importar websocket
	"fmt"
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

		currentTime := services.GetBrasiliaTime()

		if err == gorm.ErrRecordNotFound {
			newPackage := models.Package{
				TrackingID:     body.TrackingID,
				Buffer:         body.Buffer,
				Rua:            body.Rua,
				EntryTimestamp: currentTime,
			}
			if err := tx.Create(&newPackage).Error; err != nil {
				return err
			}
		} else if err == nil {
			if pkg.Buffer != "PENDENTE" {
				return fmt.Errorf("o item %s já se encontra na fila (Buffer: %s, Rua: %s)", pkg.TrackingID, pkg.Buffer, pkg.Rua)
			}
			updates := models.Package{
				Buffer:         body.Buffer,
				Rua:            body.Rua,
				EntryTimestamp: currentTime,
			}
			if err := tx.Model(&pkg).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			return err
		}

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

	// --- NOTIFICA VIA WEBSOCKET APÓS SUCESSO ---
	go websocket.H.BroadcastQueueUpdate() // Executa em goroutine para não bloquear a resposta HTTP
	// --- FIM DA NOTIFICAÇÃO ---

	c.JSON(http.StatusOK, gin.H{"message": "Entrada do item registrada com sucesso."})
}

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
		if err := tx.First(&pkg, "tracking_id = ?", body.TrackingID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("item não encontrado")
			}
			return err
		}

		if pkg.Buffer == "PENDENTE" {
			return fmt.Errorf("este item existe, mas ainda não teve entrada na fila e não pode ser removido")
		}

		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s foi removida do buffer %s na rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User), "SAIDA", logDetails); err != nil {
			return err
		}

		// --- ALTERADO: Usar Soft Delete (gorm.Model faz isso por padrão) ---
		if err := tx.Delete(&pkg).Error; err != nil { // GORM fará UPDATE deleted_at = NOW()
			return err
		}
		// --- FIM DA ALTERAÇÃO ---

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

	// --- NOTIFICA VIA WEBSOCKET APÓS SUCESSO ---
	go websocket.H.BroadcastQueueUpdate()
	// --- FIM DA NOTIFICAÇÃO ---

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
		// --- CORREÇÃO: Busca apenas pacotes ativos (não deletados) ---
		if err := tx.First(&pkg, packageID).Error; err != nil {
			// Não precisa mais verificar DeletedAt explicitamente se gorm.Model estiver correto
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("item não encontrado ou já removido da fila")
			}
			return err
		}
        // --- FIM DA CORREÇÃO ---


		oldRua := pkg.Rua
		newRua := body.Rua

		// Verifica se a rua realmente mudou para evitar logs desnecessários
		if oldRua == newRua {
			return nil // Nenhuma alteração necessária
		}

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
		if err.Error() == "item não encontrado ou já removido da fila" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao mover o item."})
		return
	}

	// --- NOTIFICA VIA WEBSOCKET APÓS SUCESSO ---
	go websocket.H.BroadcastQueueUpdate()
	// --- FIM DA NOTIFICAÇÃO ---

	c.JSON(http.StatusOK, gin.H{"message": "Item movido com sucesso."})
}


// --- Funções GetFIFOQueue, GetBacklogCount e GetBufferCounts no controller podem ser mantidas ---
// Elas ainda são úteis para a carga inicial ou se o WebSocket falhar.
// No entanto, GetBufferCounts foi movida para o hub.go para evitar duplicação.

func GetFIFOQueue(c *gin.Context) {
	var packages []models.Package
	// Busca apenas pacotes ativos
	initializers.DB.Where("buffer <> ?", "PENDENTE").Order("entry_timestamp asc").Find(&packages)
	c.JSON(http.StatusOK, gin.H{"data": packages})
}

func GetBacklogCount(c *gin.Context) {
	var count int64
	// Conta apenas pacotes ativos
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
		// --- CORREÇÃO: Usar fuso horário correto se necessário ---
		// Se as datas no frontend não consideram fuso, pode ser preciso ajustar aqui
		// Ex: loc, _ := time.LoadLocation("America/Sao_Paulo")
		// startParsed, _ := time.ParseInLocation("2006-01-02", startDate, loc)
		// endParsed, _ := time.ParseInLocation("2006-01-02 15:04:05", endDateWithTime, loc)
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDateWithTime)
	}

	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// GetBufferCounts - Esta função pode ser removida do controller se a lógica
// estiver apenas no hub.go para evitar duplicação. Se for mantida para
// alguma API específica, garantir que conta apenas itens ativos.
func GetBufferCounts(c *gin.Context) {
    var rtsCount, ehaCount, salCount int64

    // Conta separadamente para cada buffer que está ativo na fila
    initializers.DB.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "RTS").Count(&rtsCount)
    initializers.DB.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "EHA").Count(&ehaCount)
    initializers.DB.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", "SAL").Count(&salCount)

    c.JSON(http.StatusOK, gin.H{
        "rts": rtsCount,
        "eha": ehaCount,
        "sal": salCount,
    })
}