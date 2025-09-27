# Wallet Transfer 使用指南

本文档提供了 Wallet Transfer 工具的详细使用说明和最佳实践。

## 目录

- [快速开始](#快速开始)
- [配置详解](#配置详解)
- [命令参考](#命令参考)
- [转账模式详解](#转账模式详解)
- [并发和性能优化](#并发和性能优化)
- [实际使用案例](#实际使用案例)
- [最佳实践](#最佳实践)

## 快速开始

### 1. 环境准备

```bash
# 设置私钥环境变量
export PRIVATE_KEYS="0x1234567890abcdef...,0xfedcba0987654321..."

# 或者创建私钥文件
echo "0x1234567890abcdef..." > private_keys.txt
echo "0xfedcba0987654321..." >> private_keys.txt
```

### 2. 基础余额查询

```bash
# 查询当前私钥对应地址的余额
./wallet-transfer balance --network sepolia

# 查询指定地址的余额
./wallet-transfer balance --addresses 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 --network sepolia
```

### 3. 基础转账操作

```bash
# 一对一转账
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia
```

## 配置详解

### 配置文件结构

创建 `config/config.yaml`：

```yaml
# 网络配置
networks:
  # 主网配置
  ethereum:
    name: "Ethereum Mainnet"
    chain_id: 1
    rpc_url: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
    explorer_url: "https://etherscan.io"
  
  # 测试网配置
  sepolia:
    name: "Sepolia Testnet"
    chain_id: 11155111
    rpc_url: "https://sepolia.infura.io/v3/YOUR_PROJECT_ID"
    explorer_url: "https://sepolia.etherscan.io"
  
  # BSC配置
  bsc:
    name: "Binance Smart Chain"
    chain_id: 56
    rpc_url: "https://bsc-dataseed1.binance.org"
    explorer_url: "https://bscscan.com"

# 默认设置
defaults:
  network: "sepolia"
  concurrent: true
  workers: 10
  timeout: 300
  confirmations: 1
  output: "table"
  
# Gas配置
gas:
  auto_gas: true
  gas_limit: 21000
  gas_price: "20000000000"  # 20 Gwei
  
# 安全设置
security:
  max_amount_per_tx: "1000000000000000000"  # 1 ETH
  require_confirmation: true
```

### 环境变量配置

```bash
# 必需的环境变量
export PRIVATE_KEYS="key1,key2,key3"

# 可选的环境变量
export GOTESTER_NETWORK="sepolia"
export GOTESTER_RPC_URL="https://sepolia.infura.io/v3/YOUR_PROJECT_ID"
export GOTESTER_WORKERS="20"
export GOTESTER_TIMEOUT="600"
```

## 命令参考

### 全局参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--config` | string | `./config/config.yaml` | 配置文件路径 |
| `--network` | string | `sepolia` | 区块链网络 |
| `--rpc-url` | string | - | 自定义RPC URL |
| `--private-keys` | string | `env` | 私钥来源 |
| `--concurrent` | bool | `false` | 启用并发执行 |
| `--workers` | int | `10` | 工作线程数 |
| `--timeout` | int | `300` | 超时时间（秒） |
| `--output` | string | `table` | 输出格式 |

### balance 命令

查询钱包余额。

```
wallet-transfer balance [flags]
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `-a, --addresses` | strings | 要查询的地址列表 |
| `-k, --private-keys` | strings | 私钥列表 |
| `-n, --network` | string | 网络名称 |
| `-r, --rpc-url` | string | 自定义RPC URL |
| `-u, --unit` | string | 显示单位（wei/gwei/ether） |
| `-o, --output` | string | 输出格式（table/json/csv） |

### transfer 命令

执行转账操作。

```
wallet-transfer transfer [flags]
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `--mode` | string | 转账模式 |
| `--recipients` | string | 收款地址列表 |
| `--amount` | string | 固定转账金额 |
| `--amount-range` | string | 转账金额范围 |
| `--unit` | string | 金额单位 |
| `--gas-limit` | uint | Gas限制 |
| `--gas-price` | string | Gas价格 |
| `--auto-gas` | bool | 自动估算Gas |
| `--data` | string | 交易数据 |

## 转账模式详解

### 1. one-to-one（一对一）

每个发送钱包对应一个接收地址。

```bash
# 示例：3个钱包分别向3个地址转账
./gotester transfer \
  --mode one-to-one \
  --recipients 0xAddr1,0xAddr2,0xAddr3 \
  --amount 0.01 \
  --unit ether
```

**适用场景**：
- 批量发放奖励
- 分散资金到多个地址
- 测试多个钱包功能

### 2. one-to-many（一对多）

一个发送钱包向多个接收地址转账。

```bash
# 示例：从第一个钱包向多个地址转账
./gotester transfer \
  --mode one-to-many \
  --recipients 0xAddr1,0xAddr2,0xAddr3,0xAddr4 \
  --amount 0.005 \
  --unit ether
```

**适用场景**：
- 空投代币
- 批量转账给多个用户
- 资金分发

### 3. many-to-one（多对一）

多个发送钱包向一个接收地址转账。

```bash
# 示例：多个钱包向一个地址汇总资金
./gotester transfer \
  --mode many-to-one \
  --recipients 0xMainAddress \
  --amount 0.1 \
  --unit ether
```

**适用场景**：
- 资金汇总
- 收集测试代币
- 集中管理资金

### 4. many-to-many（多对多）

多个发送钱包向多个接收地址转账（循环匹配）。

```bash
# 示例：多对多随机转账
./gotester transfer \
  --mode many-to-many \
  --recipients 0xAddr1,0xAddr2,0xAddr3 \
  --amount-range 0.001-0.01 \
  --unit ether
```

**适用场景**：
- 压力测试
- 模拟真实交易场景
- 网络性能测试

## 并发和性能优化

### 并发配置

```bash
# 启用并发，设置20个工作线程
./gotester transfer \
  --concurrent \
  --workers 20 \
  --mode many-to-many \
  --recipients 0xAddr1,0xAddr2 \
  --amount 0.01
```

### 性能调优参数

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| `--workers` | 10-50 | 根据网络性能调整 |
| `--timeout` | 300-600 | 网络较慢时增加 |
| `--confirmations` | 1-3 | 安全性要求高时增加 |

### 速率限制

```bash
# 配置文件中设置速率限制
rate_limit: 10  # 每秒最多10个交易
```

## 实际使用案例

### 案例1：测试网代币分发

```bash
# 1. 检查主钱包余额
./gotester balance --network sepolia

# 2. 向100个地址分发测试代币
./gotester transfer \
  --mode one-to-many \
  --recipients $(cat recipient_addresses.txt | tr '\n' ',') \
  --amount 0.1 \
  --unit ether \
  --network sepolia \
  --concurrent \
  --workers 20
```

### 案例2：压力测试

```bash
# 多对多高频转账测试
./gotester transfer \
  --mode many-to-many \
  --recipients 0xAddr1,0xAddr2,0xAddr3,0xAddr4 \
  --amount-range 0.001-0.005 \
  --unit ether \
  --network sepolia \
  --concurrent \
  --workers 50 \
  --timeout 600
```

### 案例3：资金汇总

```bash
# 1. 查看所有钱包余额
./gotester balance --network sepolia --output json > balances.json

# 2. 将资金汇总到主地址
./gotester transfer \
  --mode many-to-one \
  --recipients 0xMainAddress \
  --amount 0.95 \
  --unit ether \
  --network sepolia
```

## 最佳实践

### 1. 安全实践

- **测试先行**：在主网使用前，先在测试网充分测试
- **小额测试**：首次使用时进行小额转账测试
- **私钥安全**：使用环境变量或安全文件存储私钥
- **余额检查**：转账前确认余额充足

### 2. 性能优化

- **合理设置并发数**：根据网络性能调整workers数量
- **监控Gas价格**：使用合适的Gas价格避免交易失败
- **批量操作**：尽量使用批量操作提高效率

### 3. 错误处理

- **重试机制**：网络不稳定时启用重试
- **超时设置**：根据网络状况调整超时时间
- **日志记录**：保存操作日志便于问题排查

### 4. 监控和调试

```bash
# 使用JSON输出便于程序处理
./gotester transfer --output json > transfer_result.json

# 使用CSV输出便于Excel分析
./gotester transfer --output csv > transfer_result.csv

# 详细日志输出
./gotester transfer --verbose
```

### 5. 成本控制

- **Gas优化**：使用`--auto-gas`自动估算合适的Gas
- **批量操作**：减少单独交易的Gas开销
- **时间选择**：在网络拥堵较少时进行操作

## 故障排除

### 常见错误及解决方案

1. **insufficient funds**
   ```bash
   # 检查余额
   ./gotester balance --network sepolia
   ```

2. **nonce too low**
   ```bash
   # 等待之前的交易确认，或重启工具
   ```

3. **gas price too low**
   ```bash
   # 使用自动Gas估算
   ./gotester transfer --auto-gas
   ```

4. **connection timeout**
   ```bash
   # 增加超时时间或更换RPC节点
   ./gotester transfer --timeout 600 --rpc-url https://alternative-rpc.com
   ```

## 高级功能

### 自定义交易数据

```bash
# 发送带有数据的交易
./gotester transfer \
  --mode one-to-one \
  --recipients 0xContractAddress \
  --amount 0 \
  --data 0xa9059cbb000000000000000000000000742d35cc6634c0532925a3b8d4c9db96590c6c8700000000000000000000000000000000000000000000000016345785d8a0000
```

### 配置文件模板

```yaml
# 生产环境配置模板
production:
  networks:
    ethereum:
      rpc_url: "${ETHEREUM_RPC_URL}"
    bsc:
      rpc_url: "${BSC_RPC_URL}"
  
  security:
    max_amount_per_tx: "10000000000000000000"  # 10 ETH
    require_confirmation: true
    
  performance:
    workers: 5
    rate_limit: 2
    timeout: 900
```
