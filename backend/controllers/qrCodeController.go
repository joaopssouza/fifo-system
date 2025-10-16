// backend/controllers/qrCodeController.go
package controllers

import (
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GenerateQRCodeData atualizado para verificar TODOS os registros, incluindo os inativos.
func GenerateQRCodeData(c *gin.Context) {
	var body struct {
		Quantity int `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A quantidade é obrigatória e deve ser maior que zero."})
		return
	}

	var nextIDs []string
	currentNumber := 1

	for len(nextIDs) < body.Quantity {
		nextIDStr := fmt.Sprintf("CG%06d", currentNumber)
		
		var count int64
		// --- ALTERAÇÃO CRÍTICA: Unscoped() ---
		// Procura em TODOS os registros, ignorando o status de soft delete (deleted_at).
		// Isso garante que um ID já usado NUNCA seja gerado novamente.
		initializers.DB.Model(&models.Package{}).Unscoped().Where("tracking_id = ?", nextIDStr).Count(&count)

		if count == 0 {
			nextIDs = append(nextIDs, nextIDStr)
		}
		
		currentNumber++
	}

	c.JSON(http.StatusOK, gin.H{"data": nextIDs})
}


// ConfirmQRCodeData permanece o mesmo.
func ConfirmQRCodeData(c *gin.Context) {
	var body struct {
		TrackingIDs []string `json:"trackingIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lista de IDs é obrigatória."})
		return
	}

	var newPackages []models.Package
	for _, id := range body.TrackingIDs {
		newPackages = append(newPackages, models.Package{
			TrackingID:     id,
			Buffer:         "PENDENTE",
			Rua:            "INDEFINIDA",
			EntryTimestamp: time.Time{},
		})
	}

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newPackages).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar os códigos no banco de dados."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Códigos confirmados e salvos com sucesso."})
}

// FindQRCodeData atualizado para encontrar TODOS os registros, incluindo inativos.
func FindQRCodeData(c *gin.Context) {
	trackingID := c.Param("trackingId")
	if trackingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tracking ID não fornecido."})
		return
	}

	var pkg models.Package
	// --- ALTERAÇÃO CRÍTICA: Unscoped() ---
	// Permite que o modal de "Buscar e Reimprimir" encontre um código mesmo que ele já
	// tenha saído da fila (e esteja marcado como inativo).
	if err := initializers.DB.Unscoped().Where("tracking_id = ?", trackingID).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Código não encontrado no banco de dados."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar o código."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": []string{pkg.TrackingID}})
}