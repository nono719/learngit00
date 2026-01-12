package oracle

import (
	"context"
	"fmt"
	"log"
	"strings"
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

	// 检查区块链配置是否有效
	if cfg.Blockchain.ContractAddr == "" || cfg.Blockchain.ContractAddr == "0x0000000000000000000000000000000000000000" {
		log.Printf("Warning: blockchain contract address is not configured (blockchain features will be disabled)")
		bcClient = nil
	} else if cfg.Blockchain.PrivateKey == "" || cfg.Blockchain.PrivateKey == "your_oracle_private_key_here" {
		log.Printf("Warning: blockchain private key is not configured (blockchain features will be disabled)")
		bcClient = nil
	} else {
		bcClient, err = blockchain.NewClient(cfg.Blockchain)
		if err != nil {
			log.Printf("ERROR: failed to create blockchain client: %v (blockchain features will be disabled)", err)
			log.Printf("  -> Please check:")
			log.Printf("     1. Ganache is running on %s", cfg.Blockchain.RPCURL)
			log.Printf("     2. Contract address is correct: %s", cfg.Blockchain.ContractAddr)
			log.Printf("     3. Private key format is correct (should start with 0x, no duplicate 0x)")
			bcClient = nil // 设置为 nil，表示区块链功能不可用
		} else {
			log.Printf("Successfully connected to blockchain at %s", cfg.Blockchain.RPCURL)
			log.Printf("Contract address: %s", cfg.Blockchain.ContractAddr)
		}
	}

	// 初始化数据源
	var sources []datasource.DataSource
	enabledCount := 0
	for _, dsCfg := range cfg.DataSources {
		if !dsCfg.Enabled {
			log.Printf("Data source %s is disabled, skipping", dsCfg.Name)
			continue
		}
		enabledCount++

		apiKeyStatus := "not configured"
		if dsCfg.APIKey != "" {
			apiKeyStatus = fmt.Sprintf("configured (value: %s)", dsCfg.APIKey)
		}
		log.Printf("Initializing data source: %s (type: %s, url: %s, api_key: %s)",
			dsCfg.Name, dsCfg.Type, dsCfg.URL, apiKeyStatus)

		ds, err := datasource.NewDataSource(dsCfg)
		if err != nil {
			log.Printf("ERROR: failed to create data source %s: %v", dsCfg.Name, err)
			continue
		}

		// 测试数据源连接和认证
		// 如果健康检查失败（特别是401认证失败），不添加该数据源
		if err := ds.HealthCheck(); err != nil {
			log.Printf("ERROR: data source %s health check failed: %v (will NOT be initialized)", dsCfg.Name, err)
			if err.Error() != "" && (err.Error() == "authentication failed (401)" ||
				strings.Contains(err.Error(), "401") ||
				strings.Contains(err.Error(), "authentication failed")) {
				log.Printf("  -> This is an authentication error. Please check:")
				log.Printf("     1. API Key is correct: %s", func() string {
					if dsCfg.APIKey != "" {
						return "configured"
					}
					return "NOT configured"
				}())
				log.Printf("     2. User exists and is active in database")
				log.Printf("     3. Backend service is running on %s", dsCfg.URL)
			}
			continue // 不添加失败的数据源
		}

		log.Printf("Data source %s health check passed", dsCfg.Name)
		sources = append(sources, ds)
		log.Printf("Successfully initialized data source: %s", dsCfg.Name)
	}

	if enabledCount == 0 {
		log.Printf("Warning: no enabled data sources found in configuration")
	} else if len(sources) == 0 {
		log.Printf("ERROR: %d data source(s) enabled but none were successfully initialized. Check logs above for errors.", enabledCount)
	} else {
		log.Printf("Successfully initialized %d/%d data source(s)", len(sources), enabledCount)
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

// GetStatus 获取预言机状态
func (o *Oracle) GetStatus() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	blockchainConnected := false
	if o.blockchain != nil {
		blockchainConnected = o.blockchain.IsConnected()
	}

	// 检查数据源健康状态
	dataSourceStatus := make([]map[string]interface{}, 0)
	if o.dataSources != nil {
		for _, ds := range o.dataSources {
			if ds == nil {
				continue
			}
			healthy := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						// HealthCheck 可能 panic，捕获它
						healthy = false
					}
				}()
				healthy = ds.HealthCheck() == nil
			}()

			dsStatus := map[string]interface{}{
				"name":    ds.Name(),
				"healthy": healthy,
			}
			dataSourceStatus = append(dataSourceStatus, dsStatus)
		}
	}

	interval := 0
	minConsensus := 0
	votingNodes := 0
	if o.config != nil && o.config.Oracle.Interval > 0 {
		interval = o.config.Oracle.Interval
		minConsensus = o.config.Oracle.MinConsensus
		votingNodes = o.config.Oracle.VotingNodes
	}

	return map[string]interface{}{
		"status":              "running",
		"data_sources":        len(o.dataSources),
		"data_sources_detail": dataSourceStatus,
		"blockchain":          blockchainConnected,
		"interval":            interval,
		"min_consensus":       minConsensus,
		"voting_nodes":        votingNodes,
	}
}

// GetConfig 获取配置信息（不包含敏感信息如私钥）
func (o *Oracle) GetConfig() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// 获取配置的数据源信息（包括未启用的）
	configuredSources := make([]map[string]interface{}, 0)
	if o.config != nil && o.config.DataSources != nil {
		for _, dsCfg := range o.config.DataSources {
			configuredSources = append(configuredSources, map[string]interface{}{
				"name":               dsCfg.Name,
				"type":               dsCfg.Type,
				"url":                dsCfg.URL,
				"enabled":            dsCfg.Enabled,
				"api_key_configured": dsCfg.APIKey != "" && dsCfg.APIKey != "your_api_key_here",
			})
		}
	}

	blockchainInfo := map[string]interface{}{
		"rpc_url":                "",
		"chain_id":               0,
		"contract_addr":          "",
		"configured":             false,
		"private_key_configured": false,
	}

	if o.config != nil {
		contractAddr := o.config.Blockchain.ContractAddr
		privateKey := o.config.Blockchain.PrivateKey

		blockchainInfo["rpc_url"] = o.config.Blockchain.RPCURL
		blockchainInfo["chain_id"] = o.config.Blockchain.ChainID
		blockchainInfo["contract_addr"] = contractAddr
		blockchainInfo["configured"] = contractAddr != "" && contractAddr != "0x0000000000000000000000000000000000000000"
		blockchainInfo["private_key_configured"] = privateKey != "" && privateKey != "your_oracle_private_key_here"

		// 调试日志
		log.Printf("[GetConfig] Blockchain - ContractAddr: '%s', configured: %v", contractAddr, blockchainInfo["configured"])
		log.Printf("[GetConfig] Blockchain - PrivateKey length: %d, configured: %v", len(privateKey), blockchainInfo["private_key_configured"])
	}

	oracleConfig := map[string]interface{}{
		"interval":      0,
		"voting_nodes":  0,
		"min_consensus": 0,
	}

	if o.config != nil {
		oracleConfig["interval"] = o.config.Oracle.Interval
		oracleConfig["voting_nodes"] = o.config.Oracle.VotingNodes
		oracleConfig["min_consensus"] = o.config.Oracle.MinConsensus
	}

	return map[string]interface{}{
		"oracle":     oracleConfig,
		"blockchain": blockchainInfo,
		"data_sources": map[string]interface{}{
			"configured": configuredSources,
			"loaded": func() int {
				if o.dataSources != nil {
					return len(o.dataSources)
				}
				return 0
			}(),
		},
	}
}

// GetDataSources 获取数据源列表
func (o *Oracle) GetDataSources() []map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sources := make([]map[string]interface{}, 0)
	if o.dataSources != nil {
		for _, ds := range o.dataSources {
			if ds == nil {
				continue
			}
			healthy := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						// HealthCheck 可能 panic，捕获它
						healthy = false
					}
				}()
				healthy = ds.HealthCheck() == nil
			}()

			source := map[string]interface{}{
				"name":    ds.Name(),
				"healthy": healthy,
			}
			sources = append(sources, source)
		}
	}

	return sources
}

// GetAllDevicesStatus 获取所有设备状态
func (o *Oracle) GetAllDevicesStatus() map[string]*models.DeviceStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()

	deviceStatuses := make(map[string][]models.DeviceStatus)

	// 从所有数据源获取状态
	for _, ds := range o.dataSources {
		statuses, err := ds.FetchDeviceStatuses()
		if err != nil {
			continue
		}

		for _, status := range statuses {
			deviceStatuses[status.DID] = append(deviceStatuses[status.DID], status)
		}
	}

	// 对每个设备进行多数投票
	result := make(map[string]*models.DeviceStatus)
	for did, statuses := range deviceStatuses {
		consensusStatus := o.majorityVote(statuses)
		if consensusStatus != nil {
			result[did] = consensusStatus
		}
	}

	return result
}

// GetConsensusStatus 获取设备共识状态详情
func (o *Oracle) GetConsensusStatus(did string) (map[string]interface{}, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

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

	// 统计投票结果
	statusCount := make(map[string]int)
	for _, status := range allStatuses {
		statusCount[status.Status]++
	}

	// 使用多数投票
	consensusStatus := o.majorityVote(allStatuses)

	result := map[string]interface{}{
		"device_did":       did,
		"total_votes":      len(allStatuses),
		"vote_details":     statusCount,
		"consensus":        consensusStatus != nil,
		"consensus_status": nil,
		"min_consensus":    o.config.Oracle.MinConsensus,
		"all_statuses":     allStatuses,
	}

	if consensusStatus != nil {
		result["consensus_status"] = consensusStatus
	}

	return result, nil
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
		"status":       "healthy",
		"data_sources": len(o.dataSources),
		"blockchain":   blockchainConnected,
	}

	return health
}
