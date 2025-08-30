// backend/controllers/packageController.go
package controllers

import (
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services" // <-- IMPORTA O NOVO PACOTE
	"fmt"
	"net/http"
	"time"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var existingPackage models.Package
		if err := tx.First(&existingPackage, "tracking_id = ?", body.TrackingID).Error; err == nil {
			return fmt.Errorf("um item com este ID já existe na fila")
		}

		pkg := models.Package{
			TrackingID:     body.TrackingID,
			Buffer:         body.Buffer,
			Rua:            body.Rua,
			EntryTimestamp: time.Now(),
		}
		if err := tx.Create(&pkg).Error; err != nil {
			return err
		}

		// --- LÓGICA DE LOG SUBSTITUÍDA POR CHAMADA AO SERVIÇO ---
		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("Package %s entered into buffer %s at rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User).Username, "ENTRADA", logDetails); err != nil {
			return err // Se o log falhar, toda a transação é desfeita (rollback)
		}
		// --- FIM DA ALTERAÇÃO ---

		return nil
	})

	if err != nil {
		if err.Error() == "um item com este ID já existe na fila" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Package registered successfully"})
}

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

		// --- LÓGICA DE LOG SUBSTITUÍDA POR CHAMADA AO SERVIÇO ---
		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("Package %s removed from buffer %s at rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		if err := services.CreateAuditLog(tx, user.(models.User).Username, "SAIDA", logDetails); err != nil {
			return err
		}
		// --- FIM DA ALTERAÇÃO ---

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

func GetAuditLogs(c *gin.Context) {
	// 1. Obter os parâmetros de filtro da URL (ex: /logs?username=admin)
	username := c.Query("username")
	action := c.Query("action")
	startDate := c.Query("startDate") // Espera o formato 'YYYY-MM-DD'
	endDate := c.Query("endDate")     // Espera o formato 'YYYY-MM-DD'

	// 2. Iniciar a construção da consulta ao banco de dados
	query := initializers.DB.Order("created_at desc")

	// 3. Adicionar filtros à consulta dinamicamente
	if username != "" {
		// Usa ILIKE para uma busca de texto que não diferencia maiúsculas de minúsculas
		query = query.Where("username ILIKE ?", "%"+username+"%")
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if startDate != "" && endDate != "" {
		// Adiciona um '23:59:59' ao final da data de término para incluir todo o dia
		endDateWithTime := endDate + " 23:59:59"
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDateWithTime)
	}

	// 4. Executar a consulta final
	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
