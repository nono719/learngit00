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
	RPCURL      string
	ChainID     int64
	ContractAddr string
	PrivateKey  string
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
	Name     string
	Type     string // "api", "certificate", "monitoring"
	URL      string
	APIKey   string
	Enabled  bool
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

	// 从环境变量覆盖配置
	overrideFromEnv(&cfg)

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

