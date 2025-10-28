// backend/controllers/packageController.go
package controllers

import (
	"errors" // Certifique-se de que errors está importado
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
	"fifo-system/backend/websocket"
	"fmt"
	"log" // Adicionado para logar erros no cálculo de tempo
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PackageEntry - Lógica ajustada para perfil opcional e log condicional
func PackageEntry(c *gin.Context) {
	var body struct {
		TrackingID string `json:"trackingId" binding:"required"`
		Buffer     string `json:"buffer" binding:"required"`
		Rua        string `json:"rua" binding:"required"`
		Profile    string `json:"profile"` // Perfil é opcional no bind
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input inválido. TrackingID, Buffer e Rua são necessários."})
		return
	}

	var profileValue int
	var profileCode string = "N/A" // Padrão

	if body.Buffer != "SAL" {
		if body.Profile == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Perfil é obrigatório para buffers RTS e EHA."})
			return
		}
		switch body.Profile {
		case "P":
			profileValue = 250
			profileCode = "P"
		case "M":
			profileValue = 80
			profileCode = "M"
		case "G":
			profileValue = 10
			profileCode = "G"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Perfil inválido. Use 'P', 'M', ou 'G'."})
			return
		}
	} else {
		profileValue = 0
		profileCode = "N/A"
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		// Usar Unscoped para encontrar mesmo se existir mas estiver "deletado" (soft delete)
		// Isso evita criar um ID duplicado se ele já existiu antes.
		err := tx.Unscoped().Where("tracking_id = ?", body.TrackingID).First(&pkg).Error

		currentTime := services.GetBrasiliaTime()

		// Cenário 1: Pacote NUNCA existiu
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newPackage := models.Package{
				TrackingID:     body.TrackingID,
				Buffer:         body.Buffer,
				Rua:            body.Rua,
				EntryTimestamp: currentTime,
				Profile:        profileCode,
				ProfileValue:   profileValue,
				// gorm.Model já inclui CreatedAt, UpdatedAt. DeletedAt será NULL.
			}
			if err := tx.Create(&newPackage).Error; err != nil {
				return fmt.Errorf("falha ao criar novo pacote: %w", err)
			}
			// Cenário 2: Pacote JÁ EXISTE
		} else if err == nil {
			// Sub-cenário 2.1: Pacote existe e está ATIVO na fila (Buffer não PENDENTE e DeletedAt é NULL)
			if pkg.Buffer != "PENDENTE" && pkg.DeletedAt.Valid == false {
				return fmt.Errorf("o item %s já se encontra na fila (Buffer: %s, Rua: %s)", pkg.TrackingID, pkg.Buffer, pkg.Rua)
			}
			// Sub-cenário 2.2: Pacote existe mas está como PENDENTE ou foi DELETADO (soft delete)
			// Podemos "reativá-lo" ou atualizar seu estado de PENDENTE para ativo.
			updates := map[string]interface{}{
				"Buffer":         body.Buffer,
				"Rua":            body.Rua,
				"EntryTimestamp": currentTime,
				"Profile":        profileCode,
				"ProfileValue":   profileValue,
				"DeletedAt":      nil, // Garante que o soft delete seja removido se existir
			}
			// Usar Unscoped aqui também para garantir que atualizamos mesmo se estiver deletado
			if err := tx.Unscoped().Model(&pkg).Where("tracking_id = ?", body.TrackingID).Updates(updates).Error; err != nil {
				return fmt.Errorf("falha ao atualizar pacote existente: %w", err)
			}
			// Cenário 3: Outro erro de banco de dados
		} else {
			return fmt.Errorf("erro ao buscar pacote: %w", err)
		}

		user, _ := c.Get("user")
		logDetails := ""
		if profileCode != "N/A" {
			logDetails = fmt.Sprintf("A Gaiola %s perfil de pacote %s entrou no buffer %s na rua %s", body.TrackingID, profileCode, body.Buffer, body.Rua)
		} else {
			logDetails = fmt.Sprintf("A Gaiola %s entrou no buffer %s na rua %s", body.TrackingID, body.Buffer, body.Rua)
		}
		if err := services.CreateAuditLog(tx, user.(models.User), "ENTRADA", logDetails); err != nil {
			return err // Erro já vem formatado do service
		}

		return nil
	})

	if err != nil {
		// Ajustar o status code baseado no tipo de erro pode ser útil
		if strings.Contains(err.Error(), "já se encontra na fila") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			log.Printf("Erro na transação de PackageEntry: %v", err) // Log detalhado no servidor
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao processar a entrada."})
		}
		return
	}

	go websocket.H.BroadcastQueueUpdate()
	c.JSON(http.StatusOK, gin.H{"message": "Entrada do item registrada com sucesso."})
}

// PackageExit - Lógica ajustada para log condicional
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
		// Busca apenas pacotes ATIVOS (não deletados)
		if err := tx.Where("tracking_id = ?", body.TrackingID).First(&pkg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("item não encontrado na fila ativa") // Mensagem mais clara
			}
			return fmt.Errorf("erro ao buscar pacote para saída: %w", err)
		}

		// Itens PENDENTES não deveriam estar ativos, mas checamos por segurança.
		if pkg.Buffer == "PENDENTE" {
			return errors.New("este item está como pendente e não pode ser removido pela saída normal")
		}

		user, _ := c.Get("user")
		logDetails := ""
		if pkg.Profile != "N/A" {
			logDetails = fmt.Sprintf("A Gaiola %s perfil de pacote %s foi removida do buffer %s na rua %s", pkg.TrackingID, pkg.Profile, pkg.Buffer, pkg.Rua)
		} else {
			logDetails = fmt.Sprintf("A Gaiola %s foi removida do buffer %s na rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		}
		if err := services.CreateAuditLog(tx, user.(models.User), "SAIDA", logDetails); err != nil {
			return err
		}

		// Soft Delete padrão do GORM
		if err := tx.Delete(&pkg).Error; err != nil {
			return fmt.Errorf("falha ao realizar soft delete: %w", err)
		}

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "item não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item não encontrado na fila ativa."})
		} else if strings.Contains(err.Error(), "pendente") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			log.Printf("Erro na transação de PackageExit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao processar a saída."})
		}
		return
	}

	go websocket.H.BroadcastQueueUpdate()
	c.JSON(http.StatusOK, gin.H{"message": "Item removido da fila com sucesso."})
}

// MovePackage - Lógica ajustada para log condicional
func MovePackage(c *gin.Context) {
	packageIDStr := c.Param("id") // Renomeado para evitar conflito com variável 'packageID' int
	packageID, err := strconv.ParseUint(packageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pacote inválido."})
		return
	}

	var body struct {
		Rua string `json:"rua" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'rua' é obrigatório."})
		return
	}

	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		var pkg models.Package
		// Busca pacote ativo pelo ID numérico
		if err := tx.First(&pkg, uint(packageID)).Error; err != nil { // Convertendo para uint
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("item não encontrado ou já removido da fila")
			}
			return fmt.Errorf("erro ao buscar pacote para mover: %w", err)
		}

		oldRua := pkg.Rua
		newRua := body.Rua

		if oldRua == newRua {
			return nil // Nenhuma alteração
		}

		if err := tx.Model(&pkg).Update("rua", newRua).Error; err != nil {
			return fmt.Errorf("falha ao atualizar rua: %w", err)
		}

		user, _ := c.Get("user")
		logDetails := ""
		if pkg.Profile != "N/A" {
			logDetails = fmt.Sprintf("A Gaiola %s perfil de pacote %s foi movida da rua %s para %s", pkg.TrackingID, pkg.Profile, oldRua, newRua)
		} else {
			logDetails = fmt.Sprintf("A Gaiola %s foi movida da rua %s para %s", pkg.TrackingID, oldRua, newRua)
		}
		if err := services.CreateAuditLog(tx, user.(models.User), "MOVIMENTACAO", logDetails); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "item não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Erro na transação de MovePackage: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao mover o item."})
		}
		return
	}

	go websocket.H.BroadcastQueueUpdate()
	c.JSON(http.StatusOK, gin.H{"message": "Item movido com sucesso."})
}

// GetFIFOQueue - Inalterado
func GetFIFOQueue(c *gin.Context) {
	var packages []models.Package
	// Busca apenas pacotes ativos (DeletedAt IS NULL é implícito no GORM por padrão)
	initializers.DB.Where("buffer <> ?", "PENDENTE").Order("entry_timestamp asc").Find(&packages)
	c.JSON(http.StatusOK, gin.H{"data": packages})
}

// GetBacklogCount - Inalterado (já excluía SAL na versão anterior)
func GetBacklogCount(c *gin.Context) {
	var count int64
	var value int64
	db := initializers.DB.Model(&models.Package{}).Where("buffer <> ? AND buffer <> ? AND deleted_at IS NULL", "PENDENTE", "SAL")

	db.Count(&count)
	db.Select("COALESCE(SUM(profile_value), 0)").Row().Scan(&value)

	c.JSON(http.StatusOK, gin.H{"count": count, "value": value})
}

// GetAuditLogs - Inalterado
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

// GetBufferCounts - Refatorado para maior clareza e robustez no cálculo do tempo médio
func GetBufferCounts(c *gin.Context) {
	type BufferStats struct {
		Count   int64
		Value   int64
		AvgTime float64 // Tempo médio em segundos
	}
	stats := make(map[string]BufferStats)
	now := services.GetBrasiliaTime()

	buffersToProcess := []string{"RTS", "EHA", "SAL"}

	for _, bufferName := range buffersToProcess {
		var count int64
		var value int64 = 0
		var totalSeconds float64 = 0.0

		// *** INÍCIO DA CORREÇÃO ***
		// Query base para o buffer atual
		baseQuery := initializers.DB.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", bufferName)

		// 1. Obter a contagem (usando a query base)
		if err := baseQuery.Count(&count).Error; err != nil {
			log.Printf("Erro ao contar pacotes para buffer %s: %v", bufferName, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter contagem do buffer " + bufferName})
			return
		}

		// 2. Calcular valor e tempo médio apenas se houver pacotes e não for SAL
		if count > 0 && bufferName != "SAL" {
			// Calcular valor (usando a query base)
			if err := baseQuery.Select("COALESCE(SUM(profile_value), 0)").Row().Scan(&value); err != nil {
				log.Printf("Erro ao calcular valor para buffer %s: %v", bufferName, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter valor do buffer " + bufferName})
				return
			}

			// Calcular soma dos tempos de permanência (usando a query base separadamente para Pluck)
			var entryTimestamps []time.Time
			// Recria a query APENAS para o Pluck, garantindo que só seleciona o timestamp
			pluckQuery := initializers.DB.Model(&models.Package{}).Where("buffer = ? AND deleted_at IS NULL", bufferName)
			if err := pluckQuery.Pluck("entry_timestamp", &entryTimestamps).Error; err != nil {
				log.Printf("Erro ao buscar timestamps para buffer %s: %v", bufferName, err)
				// Não retorna erro aqui, apenas loga. Tempo médio será 0.
			} else {
				for _, entryTime := range entryTimestamps {
					if !entryTime.IsZero() {
						duration := now.Sub(entryTime).Seconds()
						if duration > 0 {
							totalSeconds += duration
						}
					}
				}
			}
		}

		// Calcular tempo médio (evita divisão por zero)
		var avgTime float64 = 0.0
		if count > 0 && bufferName != "SAL" {
			avgTime = totalSeconds / float64(count)
		}

		// Armazenar estatísticas
		stats[bufferName] = BufferStats{
			Count:   count,
			Value:   value,   // Será 0 para SAL ou se count for 0
			AvgTime: avgTime, // Será 0 para SAL ou se count for 0
		}
	}

	// Montar a resposta
	c.JSON(http.StatusOK, gin.H{
		"counts": map[string]int64{
			"RTS": stats["RTS"].Count,
			"EHA": stats["EHA"].Count,
			"SAL": stats["SAL"].Count,
		},
		"values": map[string]int64{
			"RTS": stats["RTS"].Value,
			"EHA": stats["EHA"].Value,
			"SAL": stats["SAL"].Value, // Sempre 0
		},
		"avgTimes": map[string]float64{
			"RTS": stats["RTS"].AvgTime,
			"EHA": stats["EHA"].AvgTime,
			// "SAL" não é incluído aqui, pois não calculamos
		},
	})
}
