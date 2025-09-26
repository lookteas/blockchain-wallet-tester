package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

type Config struct {
	RPCURL            string           `mapstructure:"rpc_url"`
	TransferAmount    string           `mapstructure:"transfer_amount"`
	TargetAddresses   []common.Address `mapstructure:"target_addresses"`
	PrivateKeys       []string         `mapstructure:"private_keys"`
	Concurrent        bool             `mapstructure:"concurrent"`
	WaitConfirmations bool             `mapstructure:"wait_confirmations"`
	Confirmations     uint64           `mapstructure:"confirmations"`
}

func LoadConfig(configFile string) (*Config, error) {
	// 设置默认值
	viper.SetDefault("rpc_url", "http://localhost:8545")
	viper.SetDefault("transfer_amount", "10000000000000000") // 0.01 ETH
	viper.SetDefault("concurrent", false)
	viper.SetDefault("wait_confirmations", true)
	viper.SetDefault("confirmations", 1)

	// 从环境变量读取
	viper.AutomaticEnv()

	// 如果指定了配置文件，从文件读取
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	}

	// 从环境变量读取私钥（如果存在）
	if envKeys := os.Getenv("WALLET_PRIVATE_KEYS"); envKeys != "" {
		viper.Set("private_keys", strings.Split(envKeys, ","))
	}

	// 从环境变量读取目标地址（如果存在）
	if envTargets := os.Getenv("TARGET_ADDRESSES"); envTargets != "" {
		targetStrs := strings.Split(envTargets, ",")
		var targets []common.Address
		for _, target := range targetStrs {
			target = strings.TrimSpace(target)
			if common.IsHexAddress(target) {
				targets = append(targets, common.HexToAddress(target))
			}
		}
		viper.Set("target_addresses", targets)
	}

	// 从环境变量读取转账金额
	if envAmount := os.Getenv("TRANSFER_AMOUNT"); envAmount != "" {
		viper.Set("transfer_amount", envAmount)
	}

	// 从环境变量读取 RPC URL
	if envRPC := os.Getenv("RPC_URL"); envRPC != "" {
		viper.Set("rpc_url", envRPC)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.RPCURL == "" {
		return fmt.Errorf("RPC URL is required")
	}

	if len(cfg.TargetAddresses) == 0 {
		return fmt.Errorf("at least one target address is required")
	}

	for _, addr := range cfg.TargetAddresses {
		if addr == (common.Address{}) {
			return fmt.Errorf("invalid target address")
		}
	}

	if cfg.TransferAmount == "" {
		return fmt.Errorf("transfer amount is required")
	}

	if cfg.Confirmations == 0 {
		cfg.Confirmations = 1
	}

	return nil
}

// LoadConfigFromFile 从文件直接加载配置（不使用 viper）
func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// 验证地址格式
	var validatedAddresses []common.Address
	for _, addrStr := range cfg.TargetAddresses {
		if !common.IsHexAddress(addrStr.Hex()) {
			return nil, fmt.Errorf("invalid address format: %s", addrStr.Hex())
		}
		validatedAddresses = append(validatedAddresses, addrStr)
	}
	cfg.TargetAddresses = validatedAddresses

	return &cfg, nil
}