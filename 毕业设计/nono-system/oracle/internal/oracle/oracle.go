package oracle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"nono-system/oracle/internal/blockchain"
	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/datasource"
	"nono-system/oracle/internal/models"
)

// Oracle 预言机服务
type Oracle struct {
	config      *config.Config
	blockchain  *blockchain.Client
	dataSources []datasource.DataSource
	mu          sync.RWMutex
}

// New 创建新的预言机实例
func New(cfg *config.Config) (*Oracle, error) {
	// 初始化区块链客户端（如果连接失败，只记录警告，不阻止服务启动）
	var bcClient *blockchain.Client
	var err error
	bcClient, err = blockchain.NewClient(cfg.Blockchain)
	if err != nil {
		log.Printf("Warning: failed to create blockchain client: %v (blockchain features will be disabled)", err)
		bcClient = nil // 设置为 nil，表示区块链功能不可用
	}

	// 初始化数据源
	var sources []datasource.DataSource
	for _, dsCfg := range cfg.DataSources {
		if !dsCfg.Enabled {
			continue
		}

		ds, err := datasource.NewDataSource(dsCfg)
		if err != nil {
			log.Printf("Warning: failed to create data source %s: %v", dsCfg.Name, err)
			continue
		}
		sources = append(sources, ds)
	}

	if len(sources) == 0 {
		log.Printf("Warning: no enabled data sources found, service will start but data collection will be limited")
	}

	return &Oracle{
		config:      cfg,
		blockchain:  bcClient,
		dataSources: sources,
	}, nil
}

// StartDataCollection 启动数据采集任务
func (o *Oracle) StartDataCollection(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(o.config.Oracle.Interval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	if err := o.collectAndUpdate(); err != nil {
		log.Printf("Error in initial data collection: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := o.collectAndUpdate(); err != nil {
				log.Printf("Error in data collection: %v", err)
			}
		}
	}
}

// collectAndUpdate 采集数据并更新区块链
func (o *Oracle) collectAndUpdate() error {
	// 从所有数据源采集数据
	deviceStatuses := make(map[string][]models.DeviceStatus)

	for _, ds := range o.dataSources {
		statuses, err := ds.FetchDeviceStatuses()
		if err != nil {
			log.Printf("Error fetching from data source %s: %v", ds.Name(), err)
			continue
		}

		for _, status := range statuses {
			deviceStatuses[status.DID] = append(deviceStatuses[status.DID], status)
		}
	}

	// 对每个设备的状态进行多数投票聚合
	for did, statuses := range deviceStatuses {
		if len(statuses) == 0 {
			continue
		}

		// 多数投票机制
		consensusStatus := o.majorityVote(statuses)
		if consensusStatus == nil {
			log.Printf("No consensus reached for device %s", did)
			continue
		}

		// 更新区块链
		if err := o.updateBlockchain(did, *consensusStatus); err != nil {
			log.Printf("Error updating blockchain for device %s: %v", did, err)
		}
	}

	return nil
}

// majorityVote 多数投票机制
func (o *Oracle) majorityVote(statuses []models.DeviceStatus) *models.DeviceStatus {
	if len(statuses) == 0 {
		return nil
	}

	// 统计每种状态的出现次数
	statusCount := make(map[string]int)
	for _, status := range statuses {
		statusCount[status.Status]++
	}

	// 找到出现次数最多的状态
	maxCount := 0
	var consensusStatus string
	for status, count := range statusCount {
		if count > maxCount {
			maxCount = count
			consensusStatus = status
		}
	}

	// 检查是否达到最小共识数量
	if maxCount < o.config.Oracle.MinConsensus {
		return nil
	}

	// 返回共识状态
	for _, status := range statuses {
		if status.Status == consensusStatus {
			return &status
		}
	}

	return nil
}

// updateBlockchain 更新区块链上的设备状态
func (o *Oracle) updateBlockchain(did string, status models.DeviceStatus) error {
	// 如果区块链客户端不可用，跳过更新
	if o.blockchain == nil {
		return nil
	}

	// 将状态转换为智能合约中的枚举值
	var contractStatus int
	switch status.Status {
	case "active":
		contractStatus = 0 // DeviceStatus.Active
	case "suspicious":
		contractStatus = 1 // DeviceStatus.Suspicious
	case "revoked":
		contractStatus = 2 // DeviceStatus.Revoked
	default:
		return fmt.Errorf("unknown status: %s", status.Status)
	}

	// 调用智能合约更新状态
	return o.blockchain.UpdateDeviceStatus(did, contractStatus)
}

// GetDeviceStatus 获取设备状态
func (o *Oracle) GetDeviceStatus(did string) (*models.DeviceStatus, error) {
	// 从所有数据源获取状态
	var allStatuses []models.DeviceStatus
	for _, ds := range o.dataSources {
		statuses, err := ds.FetchDeviceStatuses()
		if err != nil {
			continue
		}

		for _, status := range statuses {
			if status.DID == did {
				allStatuses = append(allStatuses, status)
			}
		}
	}

	if len(allStatuses) == 0 {
		return nil, fmt.Errorf("device not found: %s", did)
	}

	// 使用多数投票
	consensusStatus := o.majorityVote(allStatuses)
	if consensusStatus == nil {
		return nil, fmt.Errorf("no consensus reached for device: %s", did)
	}

	return consensusStatus, nil
}

// HealthCheck 健康检查
func (o *Oracle) HealthCheck() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	blockchainConnected := false
	if o.blockchain != nil {
		blockchainConnected = o.blockchain.IsConnected()
	}

	health := map[string]interface{}{
		"status":      "healthy",
		"data_sources": len(o.dataSources),
		"blockchain":  blockchainConnected,
	}

	return health
}

