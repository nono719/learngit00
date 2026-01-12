package models

import (
	"time"
)

// DeviceHistory 设备操作历史记录
type DeviceHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceDID   string    `gorm:"column:device_did;index;not null" json:"device_did"`
	Action      string    `gorm:"column:action;index" json:"action"`               // register, update, revoke, status_change
	OldValue    string    `gorm:"column:old_value;type:text" json:"old_value"`     // JSON格式的旧值
	NewValue    string    `gorm:"column:new_value;type:text" json:"new_value"`     // JSON格式的新值
	ChangedBy   string    `gorm:"column:changed_by" json:"changed_by"`             // 操作者
	TxHash      string    `gorm:"column:tx_hash" json:"tx_hash"`                   // 区块链交易哈希（如果有）
	Description string    `gorm:"column:description;type:text" json:"description"` // 操作描述
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (DeviceHistory) TableName() string {
	return "device_history"
}
