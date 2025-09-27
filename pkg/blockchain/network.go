package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// NetworkConfig represents a blockchain network configuration
type NetworkConfig struct {
	Name     string `yaml:"name" json:"name"`
	ChainID  int64  `yaml:"chain_id" json:"chain_id"`
	RPCURL   string `yaml:"rpc_url" json:"rpc_url"`
	Symbol   string `yaml:"symbol" json:"symbol"`
	Decimals int    `yaml:"decimals" json:"decimals"`
}

// PredefinedNetworks contains predefined network configurations
var PredefinedNetworks = map[string]NetworkConfig{
	"ethereum": {
		Name:     "Ethereum Mainnet",
		ChainID:  1,
		RPCURL:   "https://mainnet.infura.io/v3/YOUR_KEY",
		Symbol:   "ETH",
		Decimals: 18,
	},
	"goerli": {
		Name:     "Ethereum Goerli Testnet",
		ChainID:  5,
		RPCURL:   "https://goerli.infura.io/v3/YOUR_KEY",
		Symbol:   "ETH",
		Decimals: 18,
	},
	"sepolia": {
		Name:     "Ethereum Sepolia Testnet",
		ChainID:  11155111,
		RPCURL:   "https://sepolia.infura.io/v3/YOUR_KEY",
		Symbol:   "ETH",
		Decimals: 18,
	},
	"bsc": {
		Name:     "BSC Mainnet",
		ChainID:  56,
		RPCURL:   "https://bsc-dataseed.binance.org/",
		Symbol:   "BNB",
		Decimals: 18,
	},
	"bsc-testnet": {
		Name:     "BSC Chapel Testnet",
		ChainID:  97,
		RPCURL:   "https://data-seed-prebsc-1-s1.binance.org:8545/",
		Symbol:   "BNB",
		Decimals: 18,
	},
	"polygon": {
		Name:     "Polygon Mainnet",
		ChainID:  137,
		RPCURL:   "https://polygon-rpc.com/",
		Symbol:   "MATIC",
		Decimals: 18,
	},
	"mumbai": {
		Name:     "Polygon Mumbai Testnet",
		ChainID:  80001,
		RPCURL:   "https://rpc-mumbai.maticvigil.com/",
		Symbol:   "MATIC",
		Decimals: 18,
	},
}

// NetworkManager manages blockchain network connections
type NetworkManager struct {
	config NetworkConfig
	client *ethclient.Client
}

// NewNetworkManager creates a new NetworkManager instance
func NewNetworkManager(networkName string, customRPCURL string) (*NetworkManager, error) {
	var config NetworkConfig
	var exists bool

	if networkName == "custom" {
		if customRPCURL == "" {
			return nil, fmt.Errorf("custom RPC URL is required for custom network")
		}
		config = NetworkConfig{
			Name:     "Custom Network",
			ChainID:  0, // Will be detected
			RPCURL:   customRPCURL,
			Symbol:   "ETH",
			Decimals: 18,
		}
	} else {
		config, exists = PredefinedNetworks[networkName]
		if !exists {
			return nil, fmt.Errorf("unknown network: %s", networkName)
		}

		// Override RPC URL if provided
		if customRPCURL != "" {
			config.RPCURL = customRPCURL
		}
	}

	nm := &NetworkManager{
		config: config,
	}

	if err := nm.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to network: %w", err)
	}

	return nm, nil
}

// Connect establishes connection to the blockchain network
func (nm *NetworkManager) Connect() error {
	client, err := ethclient.Dial(nm.config.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", nm.config.RPCURL, err)
	}

	nm.client = client

	// Verify connection and get chain ID if not set
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Update chain ID if it was 0 (custom network)
	if nm.config.ChainID == 0 {
		nm.config.ChainID = chainID.Int64()
	} else if nm.config.ChainID != chainID.Int64() {
		return fmt.Errorf("chain ID mismatch: expected %d, got %d", nm.config.ChainID, chainID.Int64())
	}

	return nil
}

// GetClient returns the Ethereum client
func (nm *NetworkManager) GetClient() *ethclient.Client {
	return nm.client
}

// GetConfig returns the network configuration
func (nm *NetworkManager) GetConfig() NetworkConfig {
	return nm.config
}

// GetBalance gets the balance of an address
func (nm *NetworkManager) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	balance, err := nm.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for %s: %w", address.Hex(), err)
	}
	return balance, nil
}

// GetBalanceByAddress 通过地址字符串获取余额
func (nm *NetworkManager) GetBalanceByAddress(ctx context.Context, address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("无效的地址格式: %s", address)
	}
	
	balance, err := nm.client.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}
	return balance, nil
}

// GetNonce 获取账户nonce
func (nm *NetworkManager) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	nonce, err := nm.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("获取nonce失败: %w", err)
	}
	
	return nonce, nil
}

// EstimateGas estimates gas for a transaction
func (nm *NetworkManager) EstimateGas(from, to common.Address, value *big.Int, data []byte) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gasLimit, err := nm.client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return gasLimit, nil
}

// SuggestGasPrice suggests gas price
func (nm *NetworkManager) SuggestGasPrice() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gasPrice, err := nm.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	return gasPrice, nil
}

// SendTransaction 发送交易
func (nm *NetworkManager) SendTransaction(ctx context.Context, tx *types.Transaction) (common.Hash, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err := nm.client.SendTransaction(ctx, tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("发送交易失败: %w", err)
	}
	
	return tx.Hash(), nil
}

// WaitForTransaction 等待交易确认
func (nm *NetworkManager) WaitForTransaction(ctx context.Context, txHash common.Hash, confirmations int) (*types.Receipt, error) {
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("等待交易确认超时: %s", txHash.Hex())
		case <-ticker.C:
			receipt, err := nm.client.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue // 交易可能还未被打包
			}
			
			// 获取当前区块号
			currentBlock, err := nm.client.BlockNumber(ctx)
			if err != nil {
				continue
			}
			
			// 检查确认数
			confirmationCount := currentBlock - receipt.BlockNumber.Uint64()
			if confirmationCount >= uint64(confirmations) {
				return receipt, nil
			}
		}
	}
}

// GetChainID 获取链ID
func (nm *NetworkManager) GetChainID() *big.Int {
	return big.NewInt(nm.config.ChainID)
}

// Close 关闭网络连接
func (nm *NetworkManager) Close() {
	if nm.client != nil {
		nm.client.Close()
	}
}

// HealthCheck 执行网络连接健康检查
func (nm *NetworkManager) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := nm.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("network health check failed: %w", err)
	}

	return nil
}
