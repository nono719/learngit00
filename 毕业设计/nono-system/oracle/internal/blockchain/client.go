package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"nono-system/oracle/internal/config"
)

// Client 区块链客户端
type Client struct {
	client       *ethclient.Client
	contractAddr common.Address
	contractABI  abi.ABI
	auth         *bind.TransactOpts
	privateKey   *ecdsa.PrivateKey
	chainID      *big.Int
}

// NewClient 创建新的区块链客户端
func NewClient(cfg config.BlockchainConfig) (*Client, error) {
	// 检查配置是否有效
	if cfg.RPCURL == "" {
		return nil, fmt.Errorf("blockchain RPC URL is not configured")
	}

	// 检查合约地址
	if cfg.ContractAddr == "" || cfg.ContractAddr == "0x0000000000000000000000000000000000000000" {
		return nil, fmt.Errorf("blockchain contract address is not configured or is default value")
	}

	// 检查私钥
	if cfg.PrivateKey == "" || cfg.PrivateKey == "your_oracle_private_key_here" {
		return nil, fmt.Errorf("blockchain private key is not configured or is default value")
	}

	// 连接区块链节点
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain RPC (%s): %w", cfg.RPCURL, err)
	}

	// 解析合约地址
	contractAddr := common.HexToAddress(cfg.ContractAddr)
	if contractAddr == (common.Address{}) {
		return nil, fmt.Errorf("invalid contract address: %s", cfg.ContractAddr)
	}

	// 加载合约ABI（简化版，实际应从文件加载）
	contractABI, err := abi.JSON(strings.NewReader(DeviceIdentityABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	// 清理私钥格式（移除可能的重复0x前缀）
	privateKeyStr := cfg.PrivateKey
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = strings.TrimPrefix(privateKeyStr, "0x")
	}
	// 如果还有0x前缀，再次移除（处理0x0x...的情况）
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = strings.TrimPrefix(privateKeyStr, "0x")
	}

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key (format error): %w", err)
	}

	chainID := big.NewInt(cfg.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	return &Client{
		client:       client,
		contractAddr: contractAddr,
		contractABI:  contractABI,
		auth:         auth,
		privateKey:   privateKey,
		chainID:      chainID,
	}, nil
}

// UpdateDeviceStatus 更新设备状态
func (c *Client) UpdateDeviceStatus(did string, status int) error {
	// 获取nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), c.auth.From)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// 构造调用数据
	data, err := c.contractABI.Pack("updateDeviceStatus", did, status)
	if err != nil {
		return fmt.Errorf("failed to pack function call: %w", err)
	}

	// 获取gas价格
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		c.contractAddr,
		nil,
		100000, // gas limit
		gasPrice,
		data,
	)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	if err := c.client.SendTransaction(context.Background(), signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

// IsConnected 检查是否连接到区块链
func (c *Client) IsConnected() bool {
	_, err := c.client.ChainID(context.Background())
	return err == nil
}

// GetDeviceStatus 查询设备状态
func (c *Client) GetDeviceStatus(did string) (int, error) {
	// 构造查询调用
	data, err := c.contractABI.Pack("devices", did)
	if err != nil {
		return 0, fmt.Errorf("failed to pack function call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &c.contractAddr,
		Data: data,
	}

	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	// 解析结果（简化版）
	// 实际应该解析完整的设备结构体
	var status int
	if err := c.contractABI.UnpackIntoInterface(&status, "devices", result); err != nil {
		return 0, fmt.Errorf("failed to unpack result: %w", err)
	}

	return status, nil
}
