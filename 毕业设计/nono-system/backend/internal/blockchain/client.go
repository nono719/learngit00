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

	"nono-system/backend/internal/config"
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
	// 如果未配置 RPC URL，返回 nil（允许服务在没有区块链的情况下运行）
	if cfg.RPCURL == "" || cfg.RPCURL == "http://localhost:8545" {
		return nil, nil
	}

	// 连接区块链节点
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	// 解析合约地址
	contractAddr := common.HexToAddress(cfg.ContractAddr)
	if contractAddr == (common.Address{}) {
		return nil, fmt.Errorf("invalid contract address")
	}

	// 加载合约ABI
	contractABI, err := abi.JSON(strings.NewReader(DeviceIdentityABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	// 解析私钥（如果配置了）
	var privateKey *ecdsa.PrivateKey
	var auth *bind.TransactOpts
	if cfg.PrivateKey != "" {
		privateKey, err = crypto.HexToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		chainID := big.NewInt(cfg.ChainID)
		auth, err = bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return nil, fmt.Errorf("failed to create transactor: %w", err)
		}
	}

	return &Client{
		client:       client,
		contractAddr: contractAddr,
		contractABI:  contractABI,
		auth:         auth,
		privateKey:   privateKey,
		chainID:      big.NewInt(cfg.ChainID),
	}, nil
}

// RequestCrossDomainAuth 请求跨域认证
func (c *Client) RequestCrossDomainAuth(did, sourceDomain, targetDomain string) (string, bool, error) {
	if c == nil || c.client == nil {
		return "", false, fmt.Errorf("blockchain client not initialized")
	}

	// 获取nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), c.auth.From)
	if err != nil {
		return "", false, fmt.Errorf("failed to get nonce: %w", err)
	}

	// 构造调用数据
	data, err := c.contractABI.Pack("requestCrossDomainAuth", did, sourceDomain, targetDomain)
	if err != nil {
		return "", false, fmt.Errorf("failed to pack function call: %w", err)
	}

	// 获取gas价格
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", false, fmt.Errorf("failed to get gas price: %w", err)
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		c.contractAddr,
		nil,
		200000, // gas limit
		gasPrice,
		data,
	)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.privateKey)
	if err != nil {
		return "", false, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	if err := c.client.SendTransaction(context.Background(), signedTx); err != nil {
		return "", false, fmt.Errorf("failed to send transaction: %w", err)
	}

	// 等待交易确认（可选，也可以异步处理）
	receipt, err := bind.WaitMined(context.Background(), c.client, signedTx)
	if err != nil {
		return "", false, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// 检查交易状态
	if receipt.Status == 0 {
		return signedTx.Hash().Hex(), false, fmt.Errorf("transaction failed")
	}

	// 解析事件获取授权结果
	authorized := true // 默认授权，实际应该从事件中解析
	if len(receipt.Logs) > 0 {
		// 解析 CrossDomainAuthCompleted 事件
		for _, log := range receipt.Logs {
			if log.Address == c.contractAddr {
				// 尝试解析事件
				event, err := c.contractABI.EventByID(log.Topics[0])
				if err == nil && event != nil && event.Name == "CrossDomainAuthCompleted" {
					// 解析事件数据
					var authEvent struct {
						Did          string
						SourceDomain string
						TargetDomain string
						Authorized   bool
					}
					if err := c.contractABI.UnpackIntoInterface(&authEvent, "CrossDomainAuthCompleted", log.Data); err == nil {
						authorized = authEvent.Authorized
					}
				}
			}
		}
	}

	return signedTx.Hash().Hex(), authorized, nil
}

// GetTransactionReceipt 获取交易收据
func (c *Client) GetTransactionReceipt(txHash string) (*types.Receipt, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("blockchain client not initialized")
	}

	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return receipt, nil
}

// IsConnected 检查是否连接到区块链
func (c *Client) IsConnected() bool {
	if c == nil || c.client == nil {
		return false
	}
	_, err := c.client.ChainID(context.Background())
	return err == nil
}

// GetDeviceStatus 查询设备状态
func (c *Client) GetDeviceStatus(did string) (int, error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("blockchain client not initialized")
	}

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

	// 解析结果（简化版，只解析状态）
	var device struct {
		Did          string
		Metadata     string
		Status       uint8
		Owner        common.Address
		RegisteredAt *big.Int
		LastUpdated  *big.Int
		Exists       bool
	}

	if err := c.contractABI.UnpackIntoInterface(&device, "devices", result); err != nil {
		return 0, fmt.Errorf("failed to unpack result: %w", err)
	}

	return int(device.Status), nil
}

