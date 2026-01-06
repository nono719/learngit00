package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// CreateDomain 创建域
func CreateDomain(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Owner       string `json:"owner"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		domain := models.Domain{
			Name:        req.Name,
			Description: req.Description,
			Owner:       req.Owner,
		}

		if err := db.Create(&domain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, domain)
	}
}

// GetDomain 获取域信息
func GetDomain(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		var domain models.Domain
		if err := db.Where("name = ?", name).First(&domain).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, domain)
	}
}

// ListDomains 列出所有域
func ListDomains(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var domains []models.Domain
		if err := db.Find(&domains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, domains)
	}
}

