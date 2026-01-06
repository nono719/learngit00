package models

import "time"

// DeviceStatus 设备状态
type DeviceStatus struct {
	DID         string    `json:"did"`
	Status      string    `json:"status"`      // "active", "suspicious", "revoked"
	Online      bool      `json:"online"`      // 是否在线
	LastSeen    time.Time `json:"last_seen"`   // 最后在线时间
	Metadata    string    `json:"metadata"`   // 元数据（JSON字符串）
	Source      string    `json:"source"`      // 数据源名称
	Timestamp   time.Time `json:"timestamp"`  // 数据采集时间
}

// DeviceMetadata 设备元数据
type DeviceMetadata struct {
	DeviceID      string `json:"device_id"`
	DeviceType    string `json:"device_type"`
	Manufacturer  string `json:"manufacturer"`
	Model         string `json:"model"`
	Firmware      string `json:"firmware"`
	Domain        string `json:"domain"`
	Location      string `json:"location,omitempty"`
	Capabilities  []string `json:"capabilities,omitempty"`
}

