package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"nono-system/backend/internal/models"
)

// RegisterUser 注册用户
func RegisterUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
			Email    string `json:"email" binding:"required,email"`
			Role     string `json:"role" binding:"required"`
			Domain   string `json:"domain"` // 操作人员和预言机节点需要
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 验证角色
		validRoles := []string{models.RoleAdmin, models.RoleOperator, models.RoleOracle, models.RoleAuditor, models.RoleUser}
		validRole := false
		for _, role := range validRoles {
			if req.Role == role {
				validRole = true
				break
			}
		}
		if !validRole {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}

		// 操作人员和预言机节点必须指定域
		if (req.Role == models.RoleOperator || req.Role == models.RoleOracle) && req.Domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required for operator and oracle roles"})
			return
		}

		// 检查用户名是否已存在
		var existingUser models.User
		if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}

		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user := models.User{
			Username:  req.Username,
			Password:  string(hashedPassword),
			Email:     req.Email,
			Role:      req.Role,
			Domain:    req.Domain,
			IsActive:  true,
			CreatedAt: time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 不返回密码
		user.Password = ""
		c.JSON(http.StatusCreated, user)
	}
}

// Login 用户登录
func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Where("username = ? AND is_active = ?", req.Username, true).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// 验证密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// 更新最后登录时间
		user.LastLogin = time.Now()
		db.Save(&user)

		// 返回用户信息和 token（简化处理，实际应该使用 JWT）
		// 这里使用用户 ID 作为 token（实际应该生成 JWT）
		c.JSON(http.StatusOK, gin.H{
			"token":    user.ID,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"domain":   user.Domain,
			},
			"permissions": models.GetRolePermissions(user.Role),
		})
	}
}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		u := user.(*models.User)
		u.Password = "" // 不返回密码

		c.JSON(http.StatusOK, gin.H{
			"user": u,
			"permissions": models.GetRolePermissions(u.Role),
		})
	}
}

// ListUsers 列出用户（仅管理员）
func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		query := db.Model(&models.User{})

		// 支持过滤
		if role := c.Query("role"); role != "" {
			query = query.Where("role = ?", role)
		}

		if domain := c.Query("domain"); domain != "" {
			query = query.Where("domain = ?", domain)
		}

		if err := query.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 不返回密码
		for i := range users {
			users[i].Password = ""
		}

		c.JSON(http.StatusOK, users)
	}
}

