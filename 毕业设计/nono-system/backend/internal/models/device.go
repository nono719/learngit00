package models

import (
	"time"

	"gorm.io/gorm"
)

// Device 设备模型
type Device struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DID         string    `gorm:"column:d_id;uniqueIndex;not null" json:"did"`
	DeviceID    string    `gorm:"column:device_id;index" json:"device_id"`
	DeviceType  string    `gorm:"column:device_type" json:"device_type"`
	Manufacturer string   `gorm:"column:manufacturer" json:"manufacturer"`
	Model       string    `gorm:"column:model" json:"model"`
	Firmware    string    `gorm:"column:firmware" json:"firmware"`
	Domain      string    `gorm:"column:domain;index" json:"domain"`
	Status      string    `gorm:"column:status;default:active" json:"status"` // active, suspicious, revoked
	Metadata    string    `gorm:"column:metadata;type:text" json:"metadata"`    // JSON格式
	Owner       string    `gorm:"column:owner" json:"owner"`
	RegisteredAt time.Time `gorm:"column:registered_at" json:"registered_at"`
	LastUpdated  time.Time `gorm:"column:last_updated" json:"last_updated"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// Domain 域模型
type Domain struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuthRecord 认证记录
type AuthRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeviceDID    string    `gorm:"column:device_did;index;not null" json:"device_did"`
	SourceDomain string    `gorm:"column:source_domain;index" json:"source_domain"`
	TargetDomain string    `gorm:"column:target_domain;index" json:"target_domain"`
	Authorized   bool      `gorm:"column:authorized" json:"authorized"`
	TxHash       string    `gorm:"column:tx_hash" json:"tx_hash"` // 区块链交易哈希
	Timestamp    time.Time `gorm:"column:timestamp" json:"timestamp"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// AuthLog 认证日志
type AuthLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeviceDID    string    `gorm:"column:device_did;index" json:"device_did"`
	SourceDomain string    `gorm:"column:source_domain" json:"source_domain"`
	TargetDomain string    `gorm:"column:target_domain" json:"target_domain"`
	Action       string    `gorm:"column:action" json:"action"` // request, success, failed
	Message      string    `gorm:"column:message;type:text" json:"message"`
	IPAddress    string    `gorm:"column:ip_address" json:"ip_address"`
	UserAgent    string    `gorm:"column:user_agent" json:"user_agent"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

