package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// RegisterDevice 注册设备
func RegisterDevice(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DID         string `json:"did" binding:"required"`
			DeviceID    string `json:"device_id" binding:"required"`
			DeviceType  string `json:"device_type"`
			Manufacturer string `json:"manufacturer"`
			Model       string `json:"model"`
			Firmware    string `json:"firmware"`
			Domain      string `json:"domain" binding:"required"`
			Metadata    string `json:"metadata"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		device := models.Device{
			DID:         req.DID,
			DeviceID:    req.DeviceID,
			DeviceType:  req.DeviceType,
			Manufacturer: req.Manufacturer,
			Model:       req.Model,
			Firmware:    req.Firmware,
			Domain:      req.Domain,
			Status:      "active",
			Metadata:    req.Metadata,
			RegisteredAt: time.Now(),
			LastUpdated:  time.Now(),
		}

		if err := db.Create(&device).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// TODO: 调用区块链合约注册设备

		c.JSON(http.StatusCreated, device)
	}
}

// GetDevice 获取设备信息
func GetDevice(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 DID 参数，Gin 会自动解码 URL 编码
		did := c.Param("did")

		var device models.Device
		if err := db.Where("d_id = ?", did).First(&device).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, device)
	}
}

// UpdateDeviceStatus 更新设备状态
func UpdateDeviceStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 DID 参数，Gin 会自动解码 URL 编码
		did := c.Param("did")

		var req struct {
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var device models.Device
		if err := db.Where("d_id = ?", did).First(&device).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		device.Status = req.Status
		device.LastUpdated = time.Now()

		if err := db.Save(&device).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// TODO: 调用区块链合约更新状态

		c.JSON(http.StatusOK, device)
	}
}

// RevokeDevice 吊销设备
func RevokeDevice(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 DID 参数，Gin 会自动解码 URL 编码
		did := c.Param("did")
		
		// 添加日志用于调试
		log.Printf("RevokeDevice: received DID = %s", did)

		var device models.Device
		if err := db.Where("d_id = ?", did).First(&device).Error; err != nil {
			log.Printf("RevokeDevice: database query error for DID %s: %v", did, err)
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Device not found",
					"did":   did,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"did":   did,
			})
			return
		}

		log.Printf("RevokeDevice: found device ID=%d, DID=%s, Status=%s", device.ID, device.DID, device.Status)

		device.Status = "revoked"
		device.LastUpdated = time.Now()

		if err := db.Save(&device).Error; err != nil {
			log.Printf("RevokeDevice: save error for DID %s: %v", did, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"did":   did,
			})
			return
		}

		log.Printf("RevokeDevice: successfully revoked device DID=%s", did)

		// TODO: 调用区块链合约吊销设备

		c.JSON(http.StatusOK, gin.H{"message": "Device revoked successfully"})
	}
}

// ListDevices 列出设备
func ListDevices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Query("domain")
		status := c.Query("status")

		var devices []models.Device
		query := db.Model(&models.Device{})

		if domain != "" {
			query = query.Where("domain = ?", domain)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}

		if err := query.Find(&devices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, devices)
	}
}

