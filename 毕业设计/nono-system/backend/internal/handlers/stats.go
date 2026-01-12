package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// GetStatistics 获取系统统计信息
func GetStatistics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := make(map[string]interface{})

		// 设备统计
		var totalDevices int64
		var activeDevices int64
		var revokedDevices int64
		var suspiciousDevices int64

		db.Model(&models.Device{}).Count(&totalDevices)
		db.Model(&models.Device{}).Where("status = ?", "active").Count(&activeDevices)
		db.Model(&models.Device{}).Where("status = ?", "revoked").Count(&revokedDevices)
		db.Model(&models.Device{}).Where("status = ?", "suspicious").Count(&suspiciousDevices)

		stats["devices"] = gin.H{
			"total":      totalDevices,
			"active":     activeDevices,
			"revoked":    revokedDevices,
			"suspicious": suspiciousDevices,
		}

		// 域统计
		var totalDomains int64
		db.Model(&models.Domain{}).Count(&totalDomains)

		// 按域分组统计设备数量
		var domainStats []gin.H
		rows, err := db.Model(&models.Device{}).
			Select("domain, COUNT(*) as count, status").
			Group("domain, status").
			Rows()
		if err == nil {
			defer rows.Close()
			domainMap := make(map[string]gin.H)
			for rows.Next() {
				var domain string
				var count int64
				var status string
				rows.Scan(&domain, &count, &status)
				
				if _, exists := domainMap[domain]; !exists {
					domainMap[domain] = gin.H{
						"domain": domain,
						"total":  0,
						"active": 0,
						"revoked": 0,
						"suspicious": 0,
					}
				}
				domainInfo := domainMap[domain]
				domainInfo["total"] = domainInfo["total"].(int) + int(count)
				if status == "active" {
					domainInfo["active"] = domainInfo["active"].(int) + int(count)
				} else if status == "revoked" {
					domainInfo["revoked"] = domainInfo["revoked"].(int) + int(count)
				} else if status == "suspicious" {
					domainInfo["suspicious"] = domainInfo["suspicious"].(int) + int(count)
				}
				domainMap[domain] = domainInfo
			}
			for _, v := range domainMap {
				domainStats = append(domainStats, v)
			}
		}

		stats["domains"] = gin.H{
			"total":  totalDomains,
			"details": domainStats,
		}

		// 认证统计
		var totalAuthRecords int64
		var successfulAuths int64
		var failedAuths int64
		var onChainAuths int64

		db.Model(&models.AuthRecord{}).Count(&totalAuthRecords)
		db.Model(&models.AuthRecord{}).Where("authorized = ?", true).Count(&successfulAuths)
		db.Model(&models.AuthRecord{}).Where("authorized = ?", false).Count(&failedAuths)
		db.Model(&models.AuthRecord{}).Where("tx_hash IS NOT NULL AND tx_hash != ''").Count(&onChainAuths)

		stats["authentication"] = gin.H{
			"total":          totalAuthRecords,
			"successful":     successfulAuths,
			"failed":         failedAuths,
			"on_chain":       onChainAuths,
			"success_rate":   calculateRate(successfulAuths, totalAuthRecords),
			"on_chain_rate":  calculateRate(onChainAuths, totalAuthRecords),
		}

		// 最近活动统计
		var recentDevices int64
		var recentAuths int64
		oneWeekAgo := db.NowFunc().AddDate(0, 0, -7)
		
		db.Model(&models.Device{}).Where("created_at > ?", oneWeekAgo).Count(&recentDevices)
		db.Model(&models.AuthRecord{}).Where("timestamp > ?", oneWeekAgo).Count(&recentAuths)

		stats["recent_activity"] = gin.H{
			"devices_registered_7d": recentDevices,
			"authentications_7d":     recentAuths,
		}

		c.JSON(http.StatusOK, stats)
	}
}

// calculateRate 计算百分比
func calculateRate(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}

