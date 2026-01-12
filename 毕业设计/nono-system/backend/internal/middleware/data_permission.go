package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// FilterByDataPermission 数据权限过滤中间件
// 根据用户角色和域限制，过滤查询结果
func FilterByDataPermission(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			// 如果没有用户信息，使用默认的只读权限
			c.Set("data_permission", "readonly")
			c.Next()
			return
		}

		u := user.(*models.User)
		
		// 设置数据权限类型
		var permissionType string
		switch u.Role {
		case models.RoleAdmin:
			permissionType = "all" // 全域数据权限
		case models.RoleOperator, models.RoleOracle:
			permissionType = "domain" // 域级数据权限
		case models.RoleAuditor:
			permissionType = "readonly" // 只读数据权限
		case models.RoleUser:
			permissionType = "restricted_readonly" // 受限只读权限
		default:
			permissionType = "readonly"
		}

		c.Set("data_permission", permissionType)
		c.Set("user_domain", u.Domain)
		c.Next()
	}
}

// ApplyDomainFilter 应用域过滤到查询
func ApplyDomainFilter(c *gin.Context, query *gorm.DB, tableName string) *gorm.DB {
	permissionType, exists := c.Get("data_permission")
	if !exists {
		return query
	}

	pt := permissionType.(string)
	
	// 系统管理员可以访问所有数据
	if pt == "all" {
		return query
	}

	// 域级权限：只能访问自己域的数据
	if pt == "domain" {
		userDomain, exists := c.Get("user_domain")
		if exists && userDomain.(string) != "" {
			return query.Where(tableName+".domain = ?", userDomain.(string))
		}
	}

	// 只读权限和受限只读权限：可以查询，但受其他权限限制
	return query
}

// CheckDomainAccess 检查用户是否可以访问指定域
func CheckDomainAccess(c *gin.Context, domain string) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}

	u := user.(*models.User)
	return u.CanAccessDomain(domain)
}

// CheckDeviceAccess 检查用户是否可以访问指定设备
func CheckDeviceAccess(c *gin.Context, device *models.Device) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}

	u := user.(*models.User)
	return u.CanAccessDevice(device)
}

// RequireDomainAccess 要求域访问权限的中间件
func RequireDomainAccess(domainParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Param(domainParam)
		if domain == "" {
			domain = c.Query(domainParam)
		}

		if !CheckDomainAccess(c, domain) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to this domain",
				"domain": domain,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

