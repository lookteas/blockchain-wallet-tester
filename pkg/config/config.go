package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 应用程序配置结构
type Config struct {
	Networks map[string]NetworkConfig `mapstructure:"networks"`
	Defaults DefaultConfig            `mapstructure:"defaults"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Name     string `mapstructure:"name"`
	ChainID  int64  `mapstructure:"chain_id"`
	RPCURL   string `mapstructure:"rpc_url"`
	Symbol   string `mapstructure:"symbol"`
	Decimals int    `mapstructure:"decimals"`
}

// DefaultConfig 默认配置
type DefaultConfig struct {
	Network           string `mapstructure:"network"`
	PrivateKeysSource string `mapstructure:"private_keys_source"`
	Concurrent        bool   `mapstructure:"concurrent"`
	Workers           int    `mapstructure:"workers"`
	Confirmations     int    `mapstructure:"confirmations"`
	Timeout           int    `mapstructure:"timeout"`
	Output            string `mapstructure:"output"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
		
		// 获取用户主目录
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".wallet-transfer"))
		}
	}

	// 设置环境变量前缀
	v.SetEnvPrefix("WALLET_TRANSFER")
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认配置
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 默认网络配置
	v.SetDefault("networks.ethereum.name", "Ethereum Mainnet")
	v.SetDefault("networks.ethereum.chain_id", 1)
	v.SetDefault("networks.ethereum.rpc_url", "https://mainnet.infura.io/v3/YOUR_INFURA_KEY")
	v.SetDefault("networks.ethereum.symbol", "ETH")
	v.SetDefault("networks.ethereum.decimals", 18)

	v.SetDefault("networks.goerli.name", "Ethereum Goerli Testnet")
	v.SetDefault("networks.goerli.chain_id", 5)
	v.SetDefault("networks.goerli.rpc_url", "https://goerli.infura.io/v3/YOUR_INFURA_KEY")
	v.SetDefault("networks.goerli.symbol", "ETH")
	v.SetDefault("networks.goerli.decimals", 18)

	v.SetDefault("networks.sepolia.name", "Ethereum Sepolia Testnet")
	v.SetDefault("networks.sepolia.chain_id", 11155111)
	v.SetDefault("networks.sepolia.rpc_url", "https://sepolia.infura.io/v3/YOUR_INFURA_KEY")
	v.SetDefault("networks.sepolia.symbol", "ETH")
	v.SetDefault("networks.sepolia.decimals", 18)

	v.SetDefault("networks.bsc.name", "BSC Mainnet")
	v.SetDefault("networks.bsc.chain_id", 56)
	v.SetDefault("networks.bsc.rpc_url", "https://bsc-dataseed.binance.org/")
	v.SetDefault("networks.bsc.symbol", "BNB")
	v.SetDefault("networks.bsc.decimals", 18)

	v.SetDefault("networks.bsc-testnet.name", "BSC Chapel Testnet")
	v.SetDefault("networks.bsc-testnet.chain_id", 97)
	v.SetDefault("networks.bsc-testnet.rpc_url", "https://data-seed-prebsc-1-s1.binance.org:8545/")
	v.SetDefault("networks.bsc-testnet.symbol", "BNB")
	v.SetDefault("networks.bsc-testnet.decimals", 18)

	v.SetDefault("networks.polygon.name", "Polygon Mainnet")
	v.SetDefault("networks.polygon.chain_id", 137)
	v.SetDefault("networks.polygon.rpc_url", "https://polygon-rpc.com/")
	v.SetDefault("networks.polygon.symbol", "MATIC")
	v.SetDefault("networks.polygon.decimals", 18)

	v.SetDefault("networks.mumbai.name", "Polygon Mumbai Testnet")
	v.SetDefault("networks.mumbai.chain_id", 80001)
	v.SetDefault("networks.mumbai.rpc_url", "https://rpc-mumbai.maticvigil.com/")
	v.SetDefault("networks.mumbai.symbol", "MATIC")
	v.SetDefault("networks.mumbai.decimals", 18)

	// 默认设置
	v.SetDefault("defaults.network", "sepolia")
	v.SetDefault("defaults.private_keys_source", "env")
	v.SetDefault("defaults.concurrent", false)
	v.SetDefault("defaults.workers", 10)
	v.SetDefault("defaults.confirmations", 1)
	v.SetDefault("defaults.timeout", 300)
	v.SetDefault("defaults.output", "table")
}

// GetNetworkConfig 获取网络配置
func (c *Config) GetNetworkConfig(network string) (NetworkConfig, error) {
	if netConfig, exists := c.Networks[network]; exists {
		return netConfig, nil
	}
	return NetworkConfig{}, fmt.Errorf("未找到网络配置: %s", network)
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	// 验证默认网络是否存在
	if _, exists := c.Networks[c.Defaults.Network]; !exists {
		return fmt.Errorf("默认网络 '%s' 不存在", c.Defaults.Network)
	}

	// 验证工作线程数
	if c.Defaults.Workers <= 0 {
		return fmt.Errorf("工作线程数必须大于0")
	}

	// 验证确认数
	if c.Defaults.Confirmations < 0 {
		return fmt.Errorf("确认数不能为负数")
	}

	// 验证超时时间
	if c.Defaults.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于0")
	}

	// 验证输出格式
	validOutputs := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}
	if !validOutputs[c.Defaults.Output] {
		return fmt.Errorf("无效的输出格式: %s", c.Defaults.Output)
	}

	// 验证私钥来源
	validSources := map[string]bool{
		"env":         true,
		"file":        true,
		"interactive": true,
	}
	if !validSources[c.Defaults.PrivateKeysSource] {
		return fmt.Errorf("无效的私钥来源: %s", c.Defaults.PrivateKeysSource)
	}

	return nil
}