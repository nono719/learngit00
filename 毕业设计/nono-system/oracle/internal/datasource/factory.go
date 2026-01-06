package datasource

import (
	"fmt"

	"nono-system/oracle/internal/config"
)

// NewDataSource 根据配置创建数据源
func NewDataSource(cfg config.DataSourceConfig) (DataSource, error) {
	switch cfg.Type {
	case "api":
		return NewAPIDataSource(cfg)
	case "monitoring":
		return NewMonitoringDataSource(cfg)
	case "certificate":
		return NewCertificateDataSource(cfg)
	default:
		return nil, fmt.Errorf("unknown data source type: %s", cfg.Type)
	}
}

