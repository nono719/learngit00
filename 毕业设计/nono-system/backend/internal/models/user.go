package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"` // 不返回密码
	Email       string    `gorm:"index" json:"email"`
	Role        string    `gorm:"index;not null" json:"role"` // 角色：admin, operator, oracle, auditor, user
	Domain      string    `gorm:"index" json:"domain"` // 所属域（操作人员、预言机节点需要）
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	LastLogin   time.Time `json:"last_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Role 角色常量
const (
	RoleAdmin   = "admin"   // 系统管理员
	RoleOperator = "operator" // 系统操作人员
	RoleOracle  = "oracle"  // 预言机节点
	RoleAuditor = "auditor" // 管理/审计人员
	RoleUser    = "user"    // 普通用户
)

// Permission 权限常量
const (
	// 系统配置管理权限
	PermConfigCreate   = "config:create"
	PermConfigUpdate   = "config:update"
	PermConfigDelete   = "config:delete"
	PermConfigQuery    = "config:query"

	// 域信息管理权限
	PermDomainCreate   = "domain:create"
	PermDomainUpdate   = "domain:update"
	PermDomainDelete   = "domain:delete"
	PermDomainQuery    = "domain:query"

	// 设备身份管理权限
	PermDeviceRegister = "device:register"
	PermDeviceUpdate   = "device:update"
	PermDeviceRevoke   = "device:revoke"
	PermDeviceQuery    = "device:query"

	// 设备身份注册权限（操作人员）
	PermDeviceRegisterDomain = "device:register:domain"

	// 跨域认证权限
	PermAuthRequest    = "auth:request"
	PermAuthQuery      = "auth:query"

	// 设备状态权限（预言机）
	PermDeviceStatusReport = "device:status:report"
	PermDeviceStatusUpdate = "device:status:update"

	// 审计权限
	PermAuditQuery    = "audit:query"
	PermAuditStats    = "audit:stats"

	// 系统信息查看权限
	PermSystemView    = "system:view"
)

// GetRolePermissions 获取角色的权限列表
func GetRolePermissions(role string) []string {
	permissions := make(map[string][]string)

	// 系统管理员 - 全权限
	permissions[RoleAdmin] = []string{
		PermConfigCreate, PermConfigUpdate, PermConfigDelete, PermConfigQuery,
		PermDomainCreate, PermDomainUpdate, PermDomainDelete, PermDomainQuery,
		PermDeviceRegister, PermDeviceUpdate, PermDeviceRevoke, PermDeviceQuery,
		PermAuthRequest, PermAuthQuery,
		PermAuditQuery, PermAuditStats,
		PermSystemView,
	}

	// 系统操作人员 - 域级权限
	permissions[RoleOperator] = []string{
		PermDeviceRegisterDomain, PermDeviceQuery,
		PermAuthRequest, PermAuthQuery,
		PermSystemView,
	}

	// 预言机节点 - 受限权限
	permissions[RoleOracle] = []string{
		PermDeviceStatusReport, PermDeviceStatusUpdate,
		PermDeviceQuery, // 只能查询，不能修改基础信息
	}

	// 管理/审计人员 - 只读权限
	permissions[RoleAuditor] = []string{
		PermAuditQuery, PermAuditStats,
		PermAuthQuery,
		PermSystemView,
	}

	// 普通用户 - 受限只读权限
	permissions[RoleUser] = []string{
		PermSystemView,
		PermDeviceQuery, // 仅可查看公开或授权的设备信息
	}

	if perms, ok := permissions[role]; ok {
		return perms
	}
	return []string{}
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(permission string) bool {
	permissions := GetRolePermissions(u.Role)
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// CanAccessDomain 检查用户是否可以访问指定域
func (u *User) CanAccessDomain(domain string) bool {
	// 系统管理员可以访问所有域
	if u.Role == RoleAdmin {
		return true
	}
	// 操作人员和预言机节点只能访问自己的域
	if u.Role == RoleOperator || u.Role == RoleOracle {
		return u.Domain == domain
	}
	// 审计人员和普通用户可以查看（但受其他权限限制）
	return true
}

// CanAccessDevice 检查用户是否可以访问指定设备
func (u *User) CanAccessDevice(device *Device) bool {
	// 系统管理员可以访问所有设备
	if u.Role == RoleAdmin {
		return true
	}
	// 操作人员和预言机节点只能访问自己域的设备
	if u.Role == RoleOperator || u.Role == RoleOracle {
		return u.Domain == device.Domain
	}
	// 审计人员和普通用户可以查看（但受其他权限限制）
	return true
}

