// backend/controllers/packageController.go
package controllers

import (
	"fifo-system/backend/initializers"
	"fifo-system/backend/models"
	"fifo-system/backend/services"
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

		user, _ := c.Get("user")
		logDetails := fmt.Sprintf("A Gaiola %s entrou no buffer %s na rua %s", pkg.TrackingID, pkg.Buffer, pkg.Rua)
		// Passa o objeto user completo
		if err := services.CreateAuditLog(tx, user.(models.User), "ENTRADA", logDetails); err != nil {
			return err
		}

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
	action := c.Query("action")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := initializers.DB.Order("created_at desc")

	if username != "" {
		// Procura tanto no nome de utilizador como no nome completo
		searchPattern := "%" + username + "%"
		query = query.Where("username ILIKE ? OR user_full_name ILIKE ?", searchPattern, searchPattern)
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
