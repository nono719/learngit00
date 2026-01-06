package datasource

import "nono-system/oracle/internal/models"

// DataSource 数据源接口
type DataSource interface {
	// Name 返回数据源名称
	Name() string

	// FetchDeviceStatuses 从数据源获取设备状态列表
	FetchDeviceStatuses() ([]models.DeviceStatus, error)

	// FetchDeviceStatus 获取指定设备的状态
	FetchDeviceStatus(did string) (*models.DeviceStatus, error)

	// HealthCheck 健康检查
	HealthCheck() error
}

