package datasource

import (
	"time"

	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/models"
)

// CertificateDataSource 证书数据源
type CertificateDataSource struct {
	name string
	url  string
}

// NewCertificateDataSource 创建证书数据源
func NewCertificateDataSource(cfg config.DataSourceConfig) (*CertificateDataSource, error) {
	return &CertificateDataSource{
		name: cfg.Name,
		url:  cfg.URL,
	}, nil
}

// Name 返回数据源名称
func (ds *CertificateDataSource) Name() string {
	return ds.name
}

// FetchDeviceStatuses 获取所有设备状态
func (ds *CertificateDataSource) FetchDeviceStatuses() ([]models.DeviceStatus, error) {
	// 从证书服务查询设备状态
	// 这里应该实际调用证书查询API
	return []models.DeviceStatus{}, nil
}

// FetchDeviceStatus 获取指定设备状态
func (ds *CertificateDataSource) FetchDeviceStatus(did string) (*models.DeviceStatus, error) {
	// 查询证书状态
	// 如果证书有效，返回active；如果吊销，返回revoked
	return &models.DeviceStatus{
		DID:      did,
		Status:   "active",
		Online:   true,
		LastSeen: time.Now(),
		Source:   ds.name,
		Timestamp: time.Now(),
	}, nil
}

// HealthCheck 健康检查
func (ds *CertificateDataSource) HealthCheck() error {
	return nil
}

