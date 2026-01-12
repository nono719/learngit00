package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Blockchain  BlockchainConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Oracle      OracleConfig
	DataSources []DataSourceConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type BlockchainConfig struct {
	RPCURL      string `mapstructure:"rpc_url" yaml:"rpc_url"`
	ChainID     int64  `mapstructure:"chain_id" yaml:"chain_id"`
	ContractAddr string `mapstructure:"contract_addr" yaml:"contract_addr"`
	PrivateKey  string `mapstructure:"private_key" yaml:"private_key"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type OracleConfig struct {
	Interval      int // 数据采集间隔（秒）
	VotingNodes   int // 投票节点数量（用于多数投票）
	MinConsensus  int // 最小共识数量
}

type DataSourceConfig struct {
	Name     string `mapstructure:"name" yaml:"name"`
	Type     string `mapstructure:"type" yaml:"type"` // "api", "certificate", "monitoring"
	URL      string `mapstructure:"url" yaml:"url"`
	APIKey   string `mapstructure:"api_key" yaml:"api_key"`
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// 设置默认值
	setDefaults()

	// 从环境变量读取
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用环境变量和默认值
		fmt.Printf("Config file not found, using defaults and environment variables: %v\n", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 备用：如果自动映射失败，手动读取关键字段
	if cfg.Blockchain.ContractAddr == "" {
		cfg.Blockchain.ContractAddr = viper.GetString("blockchain.contract_addr")
	}
	if cfg.Blockchain.PrivateKey == "" {
		cfg.Blockchain.PrivateKey = viper.GetString("blockchain.private_key")
	}
	if cfg.Blockchain.RPCURL == "" {
		cfg.Blockchain.RPCURL = viper.GetString("blockchain.rpc_url")
	}
	
	// 手动读取数据源（如果自动映射失败）
	if len(cfg.DataSources) == 0 && viper.IsSet("data_sources") {
		var dataSources []DataSourceConfig
		if err := viper.UnmarshalKey("data_sources", &dataSources); err == nil && len(dataSources) > 0 {
			cfg.DataSources = dataSources
		}
	}

	// 从环境变量覆盖配置
	overrideFromEnv(&cfg)

	// 调试：打印配置信息
	fmt.Printf("Loaded config - Blockchain ContractAddr: '%s', PrivateKey length: %d\n", 
		cfg.Blockchain.ContractAddr, len(cfg.Blockchain.PrivateKey))
	fmt.Printf("Loaded config - DataSources count: %d\n", len(cfg.DataSources))
	for i, ds := range cfg.DataSources {
		fmt.Printf("  DataSource[%d]: name=%s, type=%s, url=%s, api_key=%s, enabled=%v\n",
			i, ds.Name, ds.Type, ds.URL, func() string {
				if ds.APIKey != "" {
					return "configured"
				}
				return "not configured"
			}(), ds.Enabled)
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 9000)
	viper.SetDefault("blockchain.rpc_url", "http://localhost:8545")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "nono")
	viper.SetDefault("database.password", "nono123")
	viper.SetDefault("database.db_name", "nono_system")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("oracle.interval", 30)
	viper.SetDefault("oracle.voting_nodes", 3)
	viper.SetDefault("oracle.min_consensus", 2)
}

func overrideFromEnv(cfg *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}
	if rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL"); rpcURL != "" {
		cfg.Blockchain.RPCURL = rpcURL
	}
}

