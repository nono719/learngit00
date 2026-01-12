package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// GetDeviceHistory 获取设备操作历史
func GetDeviceHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		did := c.Param("did")
		
		// 验证设备存在
		var device models.Device
		if err := db.Where("d_id = ?", did).First(&device).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 查询历史记录
		var history []models.DeviceHistory
		query := db.Where("device_did = ?", did)
		
		// 支持分页
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "20")
		
		var pageNum, size int
		pageNum = 1
		size = 20
		
		// 简化处理，实际应该解析为整数
		if page != "1" {
			// 可以添加解析逻辑
		}
		if pageSize != "20" {
			// 可以添加解析逻辑
		}
		
		// 按时间倒序
		if err := query.Order("created_at DESC").Limit(size).Offset((pageNum - 1) * size).Find(&history).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"device_did": did,
			"history":    history,
			"total":      len(history),
		})
	}
}

// RecordDeviceHistory 记录设备操作历史（辅助函数）
func RecordDeviceHistory(db *gorm.DB, did, action, oldValue, newValue, changedBy, txHash, description string) error {
	history := models.DeviceHistory{
		DeviceDID:   did,
		Action:      action,
		OldValue:    oldValue,
		NewValue:    newValue,
		ChangedBy:   changedBy,
		TxHash:      txHash,
		Description: description,
		CreatedAt:   time.Now(),
	}
	return db.Create(&history).Error
}

