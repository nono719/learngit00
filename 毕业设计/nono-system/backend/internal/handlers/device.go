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

// ListDevices 列出设备（带权限过滤）
func ListDevices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Query("domain")
		status := c.Query("status")

		var devices []models.Device
		query := db.Model(&models.Device{})

		// 应用数据权限过滤
		permissionType, exists := c.Get("data_permission")
		if exists {
			pt := permissionType.(string)
			if pt == "domain" {
				userDomain, exists := c.Get("user_domain")
				if exists && userDomain.(string) != "" {
					query = query.Where("domain = ?", userDomain.(string))
				}
			}
		}

		if domain != "" {
			// 检查域访问权限
			if !CheckDomainAccess(c, domain) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this domain"})
				return
			}
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

// CheckDomainAccess 检查域访问权限（辅助函数）
func CheckDomainAccess(c *gin.Context, domain string) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}
	u := user.(*models.User)
	return u.CanAccessDomain(domain)
}

// GetDeviceStatuses 获取设备状态列表（供预言机使用）
// 返回格式：[]DeviceStatus
func GetDeviceStatuses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var devices []models.Device
		query := db.Model(&models.Device{})

		// 应用数据权限过滤
		permissionType, exists := c.Get("data_permission")
		if exists {
			pt := permissionType.(string)
			if pt == "domain" {
				userDomain, exists := c.Get("user_domain")
				if exists && userDomain.(string) != "" {
					query = query.Where("domain = ?", userDomain.(string))
				}
			}
		}

		// 支持按域和状态过滤
		if domain := c.Query("domain"); domain != "" {
			if !CheckDomainAccess(c, domain) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this domain"})
				return
			}
			query = query.Where("domain = ?", domain)
		}
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		if err := query.Find(&devices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 转换为预言机需要的格式
		type DeviceStatus struct {
			DID       string    `json:"did"`
			Status    string    `json:"status"`    // active, suspicious, revoked
			Online    bool      `json:"online"`   // 根据状态判断：active=true, 其他=false
			LastSeen  time.Time `json:"last_seen"` // 使用 LastUpdated
			Source    string    `json:"source"`    // 数据源名称
			Timestamp time.Time `json:"timestamp"` // 当前时间
		}

		statuses := make([]DeviceStatus, len(devices))
		for i, device := range devices {
			statuses[i] = DeviceStatus{
				DID:       device.DID,
				Status:    device.Status,
				Online:    device.Status == "active",
				LastSeen:  device.LastUpdated,
				Source:    "backend_api",
				Timestamp: time.Now(),
			}
		}

		c.JSON(http.StatusOK, statuses)
	}
}
