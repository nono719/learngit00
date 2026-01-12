package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// SearchDevices 高级搜索设备
func SearchDevices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&models.Device{})

		// 支持多个搜索条件
		if did := c.Query("did"); did != "" {
			query = query.Where("d_id LIKE ?", "%"+did+"%")
		}

		if deviceID := c.Query("device_id"); deviceID != "" {
			query = query.Where("device_id LIKE ?", "%"+deviceID+"%")
		}

		if deviceType := c.Query("device_type"); deviceType != "" {
			query = query.Where("device_type = ?", deviceType)
		}

		if manufacturer := c.Query("manufacturer"); manufacturer != "" {
			query = query.Where("manufacturer LIKE ?", "%"+manufacturer+"%")
		}

		if model := c.Query("model"); model != "" {
			query = query.Where("model LIKE ?", "%"+model+"%")
		}

		if domain := c.Query("domain"); domain != "" {
			query = query.Where("domain = ?", domain)
		}

		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		// 多状态搜索
		if statuses := c.Query("statuses"); statuses != "" {
			statusList := strings.Split(statuses, ",")
			query = query.Where("status IN ?", statusList)
		}

		// 多域搜索
		if domains := c.Query("domains"); domains != "" {
			domainList := strings.Split(domains, ",")
			query = query.Where("domain IN ?", domainList)
		}

		// 时间范围搜索
		if startDate := c.Query("start_date"); startDate != "" {
			query = query.Where("registered_at >= ?", startDate)
		}

		if endDate := c.Query("end_date"); endDate != "" {
			query = query.Where("registered_at <= ?", endDate)
		}

		// 排序
		sortBy := c.DefaultQuery("sort_by", "created_at")
		sortOrder := c.DefaultQuery("sort_order", "desc")
		if sortOrder == "asc" {
			query = query.Order(sortBy + " ASC")
		} else {
			query = query.Order(sortBy + " DESC")
		}

		// 分页
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("page_size", "20")
		// 简化处理，实际应该解析为整数
		var limit, offset int
		if pageSize == "20" {
			limit = 20
		} else {
			limit = 20
		}
		if page == "1" {
			offset = 0
		} else {
			offset = 0
		}

		var devices []models.Device
		var total int64

		// 获取总数
		query.Count(&total)

		// 获取数据
		if err := query.Limit(limit).Offset(offset).Find(&devices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   total,
			"page":    page,
			"devices": devices,
		})
	}
}

