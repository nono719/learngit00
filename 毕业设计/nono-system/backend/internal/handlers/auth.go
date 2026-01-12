package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/blockchain"
	"nono-system/backend/internal/models"
)

// RequestCrossDomainAuth 请求跨域认证
func RequestCrossDomainAuth(db *gorm.DB, bcClient *blockchain.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DeviceDID    string `json:"device_did" binding:"required"`
			SourceDomain string `json:"source_domain" binding:"required"`
			TargetDomain string `json:"target_domain" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查询设备
		var device models.Device
		if err := db.Where("d_id = ?", req.DeviceDID).First(&device).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 检查设备访问权限
		user, exists := c.Get("user")
		if exists {
			u := user.(*models.User)
			// 操作人员只能发起自己域设备的跨域认证
			if u.Role == models.RoleOperator && u.Domain != device.Domain {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot request cross-domain auth for devices in other domains"})
				return
			}
		}

		// 检查设备状态
		if device.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Device is not active"})
			return
		}

		// 验证设备的域是否与源域匹配
		if device.Domain != req.SourceDomain {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "源域错误",
				"message": "设备所属域与源域不匹配",
				"device_domain": device.Domain,
				"source_domain": req.SourceDomain,
			})
			return
		}

		// 验证目标域是否存在
		var targetDomain models.Domain
		if err := db.Where("name = ?", req.TargetDomain).First(&targetDomain).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "目标域不存在",
					"message": "指定的目标域在系统中不存在",
					"target_domain": req.TargetDomain,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 调用区块链合约进行跨域认证
		var txHash string
		var authorized bool
		var err error

		if bcClient != nil && bcClient.IsConnected() {
			// 调用区块链合约
			txHash, authorized, err = bcClient.RequestCrossDomainAuth(
				req.DeviceDID,
				req.SourceDomain,
				req.TargetDomain,
			)
			if err != nil {
				log.Printf("Failed to call blockchain contract: %v", err)
				// 如果区块链调用失败，使用本地状态作为后备
				authorized = device.Status == "active"
				txHash = ""
			} else {
				log.Printf("Cross-domain auth transaction submitted: txHash=%s, authorized=%v", txHash, authorized)
			}
		} else {
			// 区块链未连接，使用本地状态
			log.Printf("Blockchain not connected, using local device status")
			authorized = device.Status == "active"
			txHash = ""
		}

		// 记录认证记录
		authRecord := models.AuthRecord{
			DeviceDID:    req.DeviceDID,
			SourceDomain: req.SourceDomain,
			TargetDomain: req.TargetDomain,
			Authorized:   authorized,
			TxHash:       txHash,
			Timestamp:    time.Now(),
		}

		if err := db.Create(&authRecord).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 记录日志
		action := "success"
		if !authorized {
			action = "failed"
		}

		message := "Cross-domain authentication"
		if txHash != "" {
			message += " (txHash: " + txHash + ")"
		}

		authLog := models.AuthLog{
			DeviceDID:    req.DeviceDID,
			SourceDomain: req.SourceDomain,
			TargetDomain: req.TargetDomain,
			Action:       action,
			Message:      message,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		}
		db.Create(&authLog)

		response := gin.H{
			"authorized": authorized,
			"record_id":  authRecord.ID,
		}
		if txHash != "" {
			response["tx_hash"] = txHash
		}

		c.JSON(http.StatusOK, response)
	}
}

// SyncAuthRecord 同步前端上链的认证记录到数据库
func SyncAuthRecord(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DeviceDID    string `json:"device_did" binding:"required"`
			SourceDomain string `json:"source_domain" binding:"required"`
			TargetDomain string `json:"target_domain" binding:"required"`
			TxHash       string `json:"tx_hash" binding:"required"`
			Authorized   bool   `json:"authorized"`
			BlockNumber  *uint64 `json:"block_number,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查设备是否存在
		var device models.Device
		if err := db.Where("d_id = ?", req.DeviceDID).First(&device).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 验证设备的域是否与源域匹配
		if device.Domain != req.SourceDomain {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "源域错误",
				"message": "设备所属域与源域不匹配",
				"device_domain": device.Domain,
				"source_domain": req.SourceDomain,
			})
			return
		}

		// 验证目标域是否存在
		var targetDomain models.Domain
		if err := db.Where("name = ?", req.TargetDomain).First(&targetDomain).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "目标域不存在",
					"message": "指定的目标域在系统中不存在",
					"target_domain": req.TargetDomain,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 检查是否已存在相同的交易哈希记录
		var existingRecord models.AuthRecord
		if err := db.Where("tx_hash = ?", req.TxHash).First(&existingRecord).Error; err == nil {
			// 记录已存在，返回现有记录
			c.JSON(http.StatusOK, gin.H{
				"message":   "Record already exists",
				"record_id": existingRecord.ID,
				"record":    existingRecord,
			})
			return
		}

		// 检查是否存在相同设备、源域、目标域但没有交易哈希的记录
		var existingRecordWithoutTx models.AuthRecord
		if err := db.Where("device_did = ? AND source_domain = ? AND target_domain = ? AND (tx_hash = '' OR tx_hash IS NULL)", 
			req.DeviceDID, req.SourceDomain, req.TargetDomain).First(&existingRecordWithoutTx).Error; err == nil {
			// 更新现有记录的交易哈希
			existingRecordWithoutTx.TxHash = req.TxHash
			existingRecordWithoutTx.Authorized = req.Authorized
			if err := db.Save(&existingRecordWithoutTx).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			// 记录日志
			message := "Cross-domain authentication (frontend onchain sync, txHash: " + req.TxHash + ")"
			authLog := models.AuthLog{
				DeviceDID:    req.DeviceDID,
				SourceDomain: req.SourceDomain,
				TargetDomain: req.TargetDomain,
				Action:       "success",
				Message:      message,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
			}
			db.Create(&authLog)

			c.JSON(http.StatusOK, gin.H{
				"message":   "Record updated successfully",
				"record_id": existingRecordWithoutTx.ID,
				"record":    existingRecordWithoutTx,
			})
			return
		}

		// 创建新的认证记录
		authRecord := models.AuthRecord{
			DeviceDID:    req.DeviceDID,
			SourceDomain: req.SourceDomain,
			TargetDomain: req.TargetDomain,
			Authorized:   req.Authorized,
			TxHash:       req.TxHash,
			Timestamp:    time.Now(),
		}

		if err := db.Create(&authRecord).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 记录日志
		message := "Cross-domain authentication (frontend onchain, txHash: " + req.TxHash + ")"
		authLog := models.AuthLog{
			DeviceDID:    req.DeviceDID,
			SourceDomain: req.SourceDomain,
			TargetDomain: req.TargetDomain,
			Action:       "success",
			Message:      message,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		}
		db.Create(&authLog)

		c.JSON(http.StatusCreated, gin.H{
			"message":   "Record synced successfully",
			"record_id": authRecord.ID,
			"record":    authRecord,
		})
	}
}

// GetAuthRecords 获取认证记录
func GetAuthRecords(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		did := c.Param("did")

		var records []models.AuthRecord
		if err := db.Where("device_did = ?", did).Order("timestamp DESC").Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, records)
	}
}

// GetAuthLogs 获取认证日志
func GetAuthLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		did := c.Query("device_did")
		domain := c.Query("domain")

		var logs []models.AuthLog
		query := db.Model(&models.AuthLog{})

		if did != "" {
			query = query.Where("device_did = ?", did)
		}
		if domain != "" {
			query = query.Where("source_domain = ? OR target_domain = ?", domain, domain)
		}

		if err := query.Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}

// VerifyTransaction 验证区块链交易
func VerifyTransaction(bcClient *blockchain.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Param("txHash")

		if bcClient == nil || !bcClient.IsConnected() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Blockchain client not available",
			})
			return
		}

		receipt, err := bcClient.GetTransactionReceipt(txHash)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Transaction not found",
				"details": err.Error(),
			})
			return
		}

		// 解析事件
		var events []gin.H
		if len(receipt.Logs) > 0 {
			for _, log := range receipt.Logs {
				events = append(events, gin.H{
					"address":     log.Address.Hex(),
					"topics":      log.Topics,
					"data":        log.Data,
					"block_number": receipt.BlockNumber.Uint64(),
					"tx_index":     receipt.TransactionIndex,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"tx_hash":       txHash,
			"status":        receipt.Status == 1,
			"block_number":  receipt.BlockNumber.Uint64(),
			"gas_used":      receipt.GasUsed,
			"events":        events,
			"confirmations": 1, // 简化版，实际应该计算当前区块与交易区块的差值
		})
	}
}

