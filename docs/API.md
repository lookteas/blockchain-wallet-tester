# Wallet Transfer API 文档

本文档详细介绍了 Wallet Transfer 的 API 接口和使用方法。

## 导入包

```go
import (
    "wallet-transfer/pkg/wallet"
    "wallet-transfer/pkg/transfer"
    "wallet-transfer/pkg/config"
    "wallet-transfer/pkg/crypto"
)
```

## 核心接口

### 钱包管理

```go
// 钱包管理器接口
type WalletManager interface {
    LoadWallets(privateKeys []string) error
    GetWallets() []*Wallet
    GetWallet(address string) (*Wallet, error)
    GetBalance(address string, network string) (*big.Int, error)
}

// 创建钱包管理器
manager := wallet.NewManager()
```

### 转账管理

```go
// 转账管理器接口
type TransferManager interface {
    ExecuteTransfer(config *TransferConfig) (*TransferResult, error)
    GetTransferStatus(taskID string) (*TransferStatus, error)
    CancelTransfer(taskID string) error
}

// 创建转账管理器
transferManager := transfer.NewManager(walletManager, networkConfig)
```

## 使用示例

### 基础钱包操作

```go
package main

import (
    "fmt"
    "log"
    "wallet-transfer/pkg/wallet"
    "wallet-transfer/pkg/config"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 创建钱包管理器
    manager := wallet.NewManager()
    
    // 加载私钥
    privateKeys := []string{
        "0x1234567890abcdef...",
        "0xfedcba0987654321...",
    }
    
    err = manager.LoadWallets(privateKeys)
    if err != nil {
        log.Fatal(err)
    }

    // 查询余额
    wallets := manager.GetWallets()
    for _, w := range wallets {
        balance, err := manager.GetBalance(w.Address, "sepolia")
        if err != nil {
            log.Printf("Error getting balance for %s: %v", w.Address, err)
            continue
        }
        fmt.Printf("Address: %s, Balance: %s ETH\n", w.Address, balance.String())
    }
}
```

### 转账操作

```go
package main

import (
    "log"
    "math/big"
    "wallet-transfer/pkg/transfer"
    "wallet-transfer/pkg/wallet"
    "wallet-transfer/pkg/config"
)

func main() {
    // 初始化组件
    cfg, _ := config.LoadConfig("config/config.yaml")
    walletManager := wallet.NewManager()
    transferManager := transfer.NewManager(walletManager, cfg.Networks["sepolia"])

    // 加载钱包
    privateKeys := []string{"0x1234..."}
    walletManager.LoadWallets(privateKeys)

    // 配置转账
    transferConfig := &transfer.Config{
        Mode:       transfer.OneToOne,
        Recipients: []string{"0x742d35Cc6634C0532925a3b8D4C9db96590c6C87"},
        Amount:     big.NewInt(10000000000000000), // 0.01 ETH
        Network:    "sepolia",
        Concurrent: true,
        Workers:    10,
    }

    // 执行转账
    result, err := transferManager.ExecuteTransfer(transferConfig)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Transfer completed: %d successful, %d failed\n", 
        result.Successful, result.Failed)
}
```

## 配置管理

### 网络配置

```go
// 网络配置结构
type NetworkConfig struct {
    Name        string `yaml:"name"`
    ChainID     int64  `yaml:"chain_id"`
    RPCURL      string `yaml:"rpc_url"`
    ExplorerURL string `yaml:"explorer_url"`
}

// 加载网络配置
networks := map[string]*NetworkConfig{
    "sepolia": {
        Name:        "Sepolia Testnet",
        ChainID:     11155111,
        RPCURL:      "https://sepolia.infura.io/v3/YOUR_PROJECT_ID",
        ExplorerURL: "https://sepolia.etherscan.io",
    },
}
```

### 应用配置

```go
// 应用配置结构
type Config struct {
    Networks map[string]*NetworkConfig `yaml:"networks"`
    Defaults *DefaultConfig            `yaml:"defaults"`
    Security *SecurityConfig           `yaml:"security"`
}

// 从文件加载配置
config, err := config.LoadConfig("config/config.yaml")
if err != nil {
    log.Fatal(err)
}
```

## 错误处理

### 错误类型

```go
// 钱包相关错误
var (
    ErrInvalidPrivateKey = errors.New("invalid private key format")
    ErrWalletNotFound    = errors.New("wallet not found")
    ErrInsufficientFunds = errors.New("insufficient funds")
)

// 转账相关错误
var (
    ErrInvalidRecipient = errors.New("invalid recipient address")
    ErrTransferFailed   = errors.New("transfer failed")
    ErrNetworkError     = errors.New("network error")
)
```

### 错误处理示例

```go
result, err := transferManager.ExecuteTransfer(config)
if err != nil {
    switch {
    case errors.Is(err, transfer.ErrInsufficientFunds):
        log.Println("Insufficient funds for transfer")
    case errors.Is(err, transfer.ErrNetworkError):
        log.Println("Network connection error")
    default:
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

## 高级功能

### 并发控制

```go
// 并发配置
type ConcurrencyConfig struct {
    Enabled     bool `yaml:"enabled"`
    Workers     int  `yaml:"workers"`
    RateLimit   int  `yaml:"rate_limit"`
    Timeout     int  `yaml:"timeout"`
}

// 使用并发执行
config := &transfer.Config{
    Concurrent: true,
    Workers:    20,
    RateLimit:  10, // 每秒10个请求
    Timeout:    300, // 5分钟超时
}
```

### 重试机制

```go
// 重试配置
type RetryConfig struct {
    MaxRetries int           `yaml:"max_retries"`
    Delay      time.Duration `yaml:"delay"`
    Backoff    float64       `yaml:"backoff"`
}

// 配置重试
retryConfig := &RetryConfig{
    MaxRetries: 3,
    Delay:      time.Second * 2,
    Backoff:    2.0,
}
```

## 监控和日志

### 性能监控

```go
// 性能指标
type Metrics struct {
    TotalTransactions int64         `json:"total_transactions"`
    SuccessfulTx      int64         `json:"successful_tx"`
    FailedTx          int64         `json:"failed_tx"`
    AverageGasUsed    *big.Int      `json:"average_gas_used"`
    TotalDuration     time.Duration `json:"total_duration"`
}

// 获取性能指标
metrics := transferManager.GetMetrics()
fmt.Printf("Success rate: %.2f%%\n", 
    float64(metrics.SuccessfulTx)/float64(metrics.TotalTransactions)*100)
```

### 日志配置

```go
// 日志配置
type LogConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
    Output string `yaml:"output"`
}

// 配置日志
logConfig := &LogConfig{
    Level:  "info",
    Format: "json",
    Output: "stdout",
}
```

## 配置管理

### 网络配置

```go
// 网络配置结构
type NetworkConfig struct {
    Name        string `yaml:"name"`
    ChainID     int64  `yaml:"chain_id"`
    RPCURL      string `yaml:"rpc_url"`
    ExplorerURL string `yaml:"explorer_url"`
}

// 加载网络配置
networks := map[string]*NetworkConfig{
    "sepolia": {
        Name:        "Sepolia Testnet",
        ChainID:     11155111,
        RPCURL:      "https://sepolia.infura.io/v3/YOUR_PROJECT_ID",
        ExplorerURL: "https://sepolia.etherscan.io",
    },
}
```

### 应用配置

```go
// 应用配置结构
type Config struct {
    Networks map[string]*NetworkConfig `yaml:"networks"`
    Defaults *DefaultConfig            `yaml:"defaults"`
    Security *SecurityConfig           `yaml:"security"`
}

// 从文件加载配置
config, err := config.LoadConfig("config/config.yaml")
if err != nil {
    log.Fatal(err)
}
```

## 错误处理

### 自定义错误类型

```go
// 转账相关错误
type TransferError struct {
    Code    int
    Message string
    Cause   error
}

func (e *TransferError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("转账错误 [%d]: %s, 原因: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("转账错误 [%d]: %s", e.Code, e.Message)
}

// 错误代码常量
const (
    ErrCodeInsufficientBalance = 1001
    ErrCodeInvalidAddress      = 1002
    ErrCodeNetworkError        = 1003
    ErrCodeGasEstimationFailed = 1004
    ErrCodeTransactionFailed   = 1005
)

// 创建特定错误
func NewInsufficientBalanceError(address common.Address, required, available *big.Int) *TransferError {
    return &TransferError{
        Code: ErrCodeInsufficientBalance,
        Message: fmt.Sprintf("地址 %s 余额不足: 需要 %s, 可用 %s", 
            address.Hex(), required.String(), available.String()),
    }
}
```

## 性能优化

### 连接池管理

```go
type ConnectionPool struct {
    clients chan *ethclient.Client
    factory func() (*ethclient.Client, error)
    maxSize int
}

func NewConnectionPool(rpcURL string, maxSize int) *ConnectionPool {
    pool := &ConnectionPool{
        clients: make(chan *ethclient.Client, maxSize),
        maxSize: maxSize,
        factory: func() (*ethclient.Client, error) {
            return ethclient.Dial(rpcURL)
        },
    }
    
    // 预创建连接
    for i := 0; i < maxSize; i++ {
        if client, err := pool.factory(); err == nil {
            pool.clients <- client
        }
    }
    
    return pool
}

func (p *ConnectionPool) Get() (*ethclient.Client, error) {
    select {
    case client := <-p.clients:
        return client, nil
    default:
        return p.factory()
    }
}

func (p *ConnectionPool) Put(client *ethclient.Client) {
    select {
    case p.clients <- client:
    default:
        client.Close()
    }
}
```

