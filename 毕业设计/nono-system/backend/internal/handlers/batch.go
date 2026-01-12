package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// BatchRegisterDevices 批量注册设备
func BatchRegisterDevices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Devices []struct {
				DID         string `json:"did" binding:"required"`
				DeviceID    string `json:"device_id" binding:"required"`
				DeviceType  string `json:"device_type"`
				Manufacturer string `json:"manufacturer"`
				Model       string `json:"model"`
				Firmware    string `json:"firmware"`
				Domain      string `json:"domain" binding:"required"`
				Metadata    string `json:"metadata"`
			} `json:"devices" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Devices) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "devices array is empty"})
			return
		}

		if len(req.Devices) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 100 devices per batch"})
			return
		}

		var results []gin.H
		var successCount int
		var failCount int

		for _, deviceReq := range req.Devices {
			device := models.Device{
				DID:          deviceReq.DID,
				DeviceID:    deviceReq.DeviceID,
				DeviceType:  deviceReq.DeviceType,
				Manufacturer: deviceReq.Manufacturer,
				Model:       deviceReq.Model,
				Firmware:    deviceReq.Firmware,
				Domain:      deviceReq.Domain,
				Status:      "active",
				Metadata:    deviceReq.Metadata,
				RegisteredAt: time.Now(),
				LastUpdated:  time.Now(),
			}

			if err := db.Create(&device).Error; err != nil {
				results = append(results, gin.H{
					"did":     deviceReq.DID,
					"success": false,
					"error":   err.Error(),
				})
				failCount++
			} else {
				// 记录历史
				oldValue, _ := json.Marshal(nil)
				newValue, _ := json.Marshal(device)
				_ = RecordDeviceHistory(db, device.DID, "register", string(oldValue), string(newValue), "system", "", "批量注册设备")
				
				results = append(results, gin.H{
					"did":     deviceReq.DID,
					"success": true,
					"device":  device,
				})
				successCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(req.Devices),
			"success": successCount,
			"failed":  failCount,
			"results": results,
		})
	}
}

// BatchUpdateDeviceStatus 批量更新设备状态
func BatchUpdateDeviceStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Devices []struct {
				DID    string `json:"did" binding:"required"`
				Status string `json:"status" binding:"required"`
			} `json:"devices" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(req.Devices) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 100 devices per batch"})
			return
		}

		var results []gin.H
		var successCount int
		var failCount int

		for _, deviceReq := range req.Devices {
			var device models.Device
			if err := db.Where("d_id = ?", deviceReq.DID).First(&device).Error; err != nil {
				results = append(results, gin.H{
					"did":     deviceReq.DID,
					"success": false,
					"error":   "Device not found",
				})
				failCount++
				continue
			}

			oldStatus := device.Status
			device.Status = deviceReq.Status
			device.LastUpdated = time.Now()

			if err := db.Save(&device).Error; err != nil {
				results = append(results, gin.H{
					"did":     deviceReq.DID,
					"success": false,
					"error":   err.Error(),
				})
				failCount++
			} else {
				// 记录历史
				oldValue, _ := json.Marshal(gin.H{"status": oldStatus})
				newValue, _ := json.Marshal(gin.H{"status": deviceReq.Status})
				_ = RecordDeviceHistory(db, device.DID, "status_change", string(oldValue), string(newValue), "system", "", "批量更新状态")
				
				results = append(results, gin.H{
					"did":     deviceReq.DID,
					"success": true,
					"status":  deviceReq.Status,
				})
				successCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   len(req.Devices),
			"success": successCount,
			"failed":  failCount,
			"results": results,
		})
	}
}

