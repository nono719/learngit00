package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token 并获取用户信息
		// 这里简化处理，实际应该使用 JWT 或其他 token 验证机制
		// token 是用户 ID（数字字符串）
		var user models.User
		var err error
		
		// 先尝试将 token 转换为数字 ID
		if userID, parseErr := strconv.ParseUint(token, 10, 32); parseErr == nil {
			// token 是数字，作为用户 ID 查找
			err = db.Where("id = ? AND is_active = ?", uint(userID), true).First(&user).Error
		} else {
			// token 不是数字，作为用户名查找（向后兼容）
			err = db.Where("username = ? AND is_active = ?", token, true).First(&user).Error
		}
		
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
					"message": "用户不存在或已被禁用",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
					"message": "token 验证失败",
				})
			}
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user", &user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("user_domain", user.Domain)

		c.Next()
	}
}

// RequirePermission 权限检查中间件（支持多个权限，满足其一即可）
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		u := user.(*models.User)
		hasPermission := false
		for _, permission := range permissions {
			if u.HasPermission(permission) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required": permissions,
				"role": u.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		u := user.(*models.User)
		hasRole := false
		for _, role := range roles {
			if u.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient role",
				"required": roles,
				"current": u.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func OptionalAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				var user models.User
				var err error
				
				// 先尝试将 token 转换为数字 ID
				if userID, parseErr := strconv.ParseUint(token, 10, 32); parseErr == nil {
					// token 是数字，作为用户 ID 查找
					err = db.Where("id = ? AND is_active = ?", uint(userID), true).First(&user).Error
				} else {
					// token 不是数字，作为用户名查找（向后兼容）
					err = db.Where("username = ? AND is_active = ?", token, true).First(&user).Error
				}
				
				if err == nil {
					c.Set("user", &user)
					c.Set("user_id", user.ID)
					c.Set("user_role", user.Role)
					c.Set("user_domain", user.Domain)
				}
			}
		}
		c.Next()
	}
}

