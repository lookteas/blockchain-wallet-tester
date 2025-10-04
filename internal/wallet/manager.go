package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v3"
)

// Manager 钱包管理器
type Manager struct {
	privateKeys []*ecdsa.PrivateKey
	addresses   []common.Address
	client      *ethclient.Client
	network     string
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	RPCURL   string
	ChainID  *big.Int
	Name     string
	Explorer string
}

// 默认网络配置（仅包含链ID和名称，RPC URL从配置文件读取）
var defaultNetworkConfigs = map[string]NetworkConfig{
	"sepolia": {
		ChainID:  big.NewInt(11155111),
		Name:     "Sepolia",
		Explorer: "https://sepolia.etherscan.io/tx/",
	},
	"goerli": {
		ChainID:  big.NewInt(5),
		Name:     "Goerli",
		Explorer: "https://goerli.etherscan.io/tx/",
	},
	"mainnet": {
		ChainID:  big.NewInt(1),
		Name:     "Mainnet",
		Explorer: "https://etherscan.io/tx/",
	},
	"bnb": {
		ChainID:  big.NewInt(56),
		Name:     "BNB Smart Chain",
		Explorer: "https://bscscan.com/tx/",
	},
	"polygon": {
		ChainID:  big.NewInt(137),
		Name:     "Polygon",
		Explorer: "https://polygonscan.com/tx/",
	},
}

// NewManager 创建钱包管理器
func NewManager(envFile, network string) (*Manager, error) {
	// 自动加载环境变量
	loadEnvFile(envFile)
	return NewManagerWithRPC(envFile, network, nil)
}

// NewManagerWithRPC 创建钱包管理器（支持自定义RPC）
func NewManagerWithRPC(envFile, network string, customRPCs map[string]string) (*Manager, error) {
	// 自动加载环境变量
	loadEnvFile(envFile)

	// 加载私钥
	privateKeys, err := loadPrivateKeys(envFile)
	if err != nil {
		return nil, fmt.Errorf("加载私钥失败: %v", err)
	}

	// 推导地址
	addresses := make([]common.Address, len(privateKeys))
	for i, pk := range privateKeys {
		addresses[i] = crypto.PubkeyToAddress(pk.PublicKey)
	}

	// 检查网络是否支持
	if _, exists := defaultNetworkConfigs[network]; !exists {
		return nil, fmt.Errorf("不支持的网络: %s", network)
	}

	// 获取RPC URL（优先级：自定义RPC > 环境变量 > 全局配置）
	rpcURL := getRPCURL(envFile, network, customRPCs, "")
	if rpcURL == "" {
		return nil, fmt.Errorf("未找到网络 %s 的RPC配置，请检查配置文件", network)
	}

	// 连接以太坊客户端
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接网络失败: %v", err)
	}

	return &Manager{
		privateKeys: privateKeys,
		addresses:   addresses,
		client:      client,
		network:     network,
	}, nil
}

// loadPrivateKeys 从.env文件加载私钥
func loadPrivateKeys(envFile string) ([]*ecdsa.PrivateKey, error) {
	// 检查文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return nil, fmt.Errorf(".env文件不存在: %s", envFile)
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(envFile)
	if err != nil {
		return nil, fmt.Errorf("读取.env文件失败: %v", err)
	}

	// 解析PRIVATE_KEYS
	lines := strings.Split(string(content), "\n")
	var privateKeysStr string
	for _, line := range lines {
		if strings.HasPrefix(line, "PRIVATE_KEYS=") {
			privateKeysStr = strings.TrimPrefix(line, "PRIVATE_KEYS=")
			break
		}
	}

	if privateKeysStr == "" {
		return nil, fmt.Errorf("未找到PRIVATE_KEYS配置")
	}

	// 分割私钥
	keys := strings.Split(privateKeysStr, ",")
	if len(keys) == 0 {
		return nil, fmt.Errorf("私钥列表为空")
	}

	// 解析每个私钥
	privateKeys := make([]*ecdsa.PrivateKey, 0, len(keys))
	for i, keyStr := range keys {
		keyStr = strings.TrimSpace(keyStr)
		if keyStr == "" {
			continue
		}

		// 确保私钥以0x开头
		if !strings.HasPrefix(keyStr, "0x") {
			keyStr = "0x" + keyStr
		}

		privateKey, err := crypto.HexToECDSA(keyStr[2:]) // 去掉0x前缀
		if err != nil {
			return nil, fmt.Errorf("解析第%d个私钥失败: %v", i+1, err)
		}

		privateKeys = append(privateKeys, privateKey)
	}

	if len(privateKeys) == 0 {
		return nil, fmt.Errorf("没有有效的私钥")
	}

	return privateKeys, nil
}

// GetAddresses 获取所有地址
func (m *Manager) GetAddresses() []common.Address {
	return m.addresses
}

// GetFirstAddress 获取第一个地址（用于send命令）
func (m *Manager) GetFirstAddress() common.Address {
	if len(m.addresses) == 0 {
		return common.Address{}
	}
	return m.addresses[0]
}

// GetAddressByIndex 根据索引获取地址（用于batch命令轮询）
func (m *Manager) GetAddressByIndex(index int) common.Address {
	if len(m.addresses) == 0 {
		return common.Address{}
	}
	return m.addresses[index%len(m.addresses)]
}

// GetPrivateKeyByIndex 根据索引获取私钥
func (m *Manager) GetPrivateKeyByIndex(index int) *ecdsa.PrivateKey {
	if len(m.privateKeys) == 0 {
		return nil
	}
	return m.privateKeys[index%len(m.privateKeys)]
}

// GetClient 获取以太坊客户端
func (m *Manager) GetClient() *ethclient.Client {
	return m.client
}

// GetNetworkConfig 获取网络配置
func (m *Manager) GetNetworkConfig() NetworkConfig {
	return defaultNetworkConfigs[m.network]
}

// GetChainID 获取链ID
func (m *Manager) GetChainID() *big.Int {
	return defaultNetworkConfigs[m.network].ChainID
}

// GetExplorerURL 获取区块浏览器URL
func (m *Manager) GetExplorerURL(txHash string) string {
	return defaultNetworkConfigs[m.network].Explorer + txHash
}

// ValidateAddress 验证以太坊地址格式
func ValidateAddress(address string) error {
	if !common.IsHexAddress(address) {
		return fmt.Errorf("无效的以太坊地址格式: %s", address)
	}
	return nil
}

// ParseAmount 解析金额（ETH转Wei）
func ParseAmount(amountStr string) (*big.Int, error) {
	amount, ok := new(big.Float).SetString(amountStr)
	if !ok {
		return nil, fmt.Errorf("无效的金额格式: %s", amountStr)
	}

	if amount.Cmp(big.NewFloat(0)) <= 0 {
		return nil, fmt.Errorf("金额必须大于0")
	}

	// 转换为Wei (1 ETH = 10^18 Wei)
	weiFloat := new(big.Float).Mul(amount, big.NewFloat(1e18))
	wei, _ := weiFloat.Int(nil)

	return wei, nil
}

// FormatAmount 格式化金额（Wei转ETH）
func FormatAmount(wei *big.Int) string {
	eth := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return fmt.Sprintf("%.6f", eth)
}

// GetBalance 获取地址余额
func (m *Manager) GetBalance(address common.Address) (*big.Int, error) {
	ctx := context.Background()
	balance, err := m.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %v", err)
	}
	return balance, nil
}

// GetGasPrice 获取当前Gas价格
func (m *Manager) GetGasPrice() (*big.Int, error) {
	ctx := context.Background()
	gasPrice, err := m.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取Gas价格失败: %v", err)
	}
	return gasPrice, nil
}

// EstimateGas 估算Gas消耗
func (m *Manager) EstimateGas(from, to common.Address, value *big.Int, data []byte) (uint64, error) {
	ctx := context.Background()
	msg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	}

	gasLimit, err := m.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("估算Gas失败: %v", err)
	}

	// 增加20%的缓冲
	gasLimit = gasLimit * 120 / 100
	return gasLimit, nil
}

// CreateTransactor 创建交易发送者
func (m *Manager) CreateTransactor(privateKey *ecdsa.PrivateKey) *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(privateKey)
	auth.GasLimit = 21000 // 标准转账的Gas限制
	return auth
}

// getRPCURL 获取RPC URL，按优先级选择
func getRPCURL(envFile, network string, customRPCs map[string]string, defaultURL string) string {
	// 1. 优先使用自定义RPC配置
	if customRPCs != nil {
		if customRPC, exists := customRPCs[network]; exists && customRPC != "" {
			return customRPC
		}
	}

	// 2. 尝试从环境变量读取
	envRPC := getRPCFromEnv(envFile, network)
	if envRPC != "" {
		return envRPC
	}

	// 3. 尝试从全局RPC配置文件读取
	globalRPC, err := loadGlobalRPCConfig()
	if err == nil && globalRPC != nil {
		if rpcURL, exists := globalRPC[network]; exists && rpcURL != "" {
			return rpcURL
		}
	}

	// 4. 如果都没有配置，返回错误
	return ""
}

// getRPCFromEnv 从环境变量文件读取RPC配置
func getRPCFromEnv(envFile, network string) string {
	// 检查文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return ""
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(envFile)
	if err != nil {
		return ""
	}

	// 解析环境变量
	lines := strings.Split(string(content), "\n")
	envVars := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// 根据网络返回对应的RPC URL
	switch network {
	case "sepolia":
		if url, exists := envVars["SEPOLIA_RPC_URL"]; exists {
			return url
		}
	case "goerli":
		if url, exists := envVars["GOERLI_RPC_URL"]; exists {
			return url
		}
	case "mainnet":
		if url, exists := envVars["MAINNET_RPC_URL"]; exists {
			return url
		}
	case "bnb":
		if url, exists := envVars["BNB_RPC_URL"]; exists {
			return url
		}
	case "polygon":
		if url, exists := envVars["POLYGON_RPC_URL"]; exists {
			return url
		}
	}

	return ""
}

// loadEnvFile 加载环境变量文件
func loadEnvFile(envFile string) {
	// 检查文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(envFile)
	if err != nil {
		return
	}

	// 解析环境变量并设置到系统环境变量中
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 只有当环境变量不存在时才设置
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// loadGlobalRPCConfig 加载全局RPC配置
func loadGlobalRPCConfig() (map[string]string, error) {
	// 尝试加载合并后的配置文件
	configFile := "configs/config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	var config struct {
		RPCConfig map[string]string `yaml:"rpc_config"`
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if config.RPCConfig == nil {
		return nil, fmt.Errorf("配置文件中未找到rpc_config部分")
	}

	// 处理环境变量替换
	for key, value := range config.RPCConfig {
		config.RPCConfig[key] = expandEnvVariables(value)
	}

	return config.RPCConfig, nil
}

// expandEnvVariables 展开环境变量
func expandEnvVariables(value string) string {
	// 简单的环境变量替换，支持 ${VAR_NAME} 格式
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
		// 如果系统环境变量中没有，尝试从.env文件读取
		if envValue := getEnvFromFile(envVar); envValue != "" {
			return envValue
		}
	}
	return value
}

// getEnvFromFile 从.env文件读取环境变量
func getEnvFromFile(envVar string) string {
	// 尝试从常见的.env文件读取
	envFiles := []string{".env", "configs/.env", "configs/env.example"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			continue
		}

		content, err := ioutil.ReadFile(envFile)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == envVar {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
