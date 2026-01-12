package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// ExportDevices 导出设备数据为CSV
func ExportDevices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var devices []models.Device
		
		// 支持过滤条件
		query := db.Model(&models.Device{})
		
		if domain := c.Query("domain"); domain != "" {
			query = query.Where("domain = ?", domain)
		}
		
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		if err := query.Find(&devices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 设置响应头
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=devices_"+time.Now().Format("20060102_150405")+".csv")

		// 创建CSV写入器
		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// 写入表头
		headers := []string{
			"ID", "DID", "Device ID", "Device Type", "Manufacturer", 
			"Model", "Firmware", "Domain", "Status", "Owner", 
			"Registered At", "Last Updated",
		}
		writer.Write(headers)

		// 写入数据
		for _, device := range devices {
			record := []string{
				strconv.Itoa(int(device.ID)),
				device.DID,
				device.DeviceID,
				device.DeviceType,
				device.Manufacturer,
				device.Model,
				device.Firmware,
				device.Domain,
				device.Status,
				device.Owner,
				device.RegisteredAt.Format(time.RFC3339),
				device.LastUpdated.Format(time.RFC3339),
			}
			writer.Write(record)
		}
	}
}

// ExportAuthRecords 导出认证记录为CSV
func ExportAuthRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var records []models.AuthRecord
		
		query := db.Model(&models.AuthRecord{})
		
		if did := c.Query("device_did"); did != "" {
			query = query.Where("device_did = ?", did)
		}
		
		if err := query.Order("timestamp DESC").Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 设置响应头
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=auth_records_"+time.Now().Format("20060102_150405")+".csv")

		// 创建CSV写入器
		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// 写入表头
		headers := []string{
			"ID", "Device DID", "Source Domain", "Target Domain", 
			"Authorized", "Tx Hash", "Timestamp",
		}
		writer.Write(headers)

		// 写入数据
		for _, record := range records {
			authorized := "false"
			if record.Authorized {
				authorized = "true"
			}
			
			csvRecord := []string{
				strconv.Itoa(int(record.ID)),
				record.DeviceDID,
				record.SourceDomain,
				record.TargetDomain,
				authorized,
				record.TxHash,
				record.Timestamp.Format(time.RFC3339),
			}
			writer.Write(csvRecord)
		}
	}
}

