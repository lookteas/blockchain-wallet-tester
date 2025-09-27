# Wallet Transfer

Wallet Transfer 是一个区块链钱包转账工具，支持以太坊及其兼容网络（如 Sepolia 等测试网络）。它支持多种转账模式、余额查询、并发操作和私钥管理。

## 🚀 主要功能

- **多种区块链网络支持**：Ethereum、BSC、Polygon、Goerli、Sepolia、Mumbai
- **批量转账操作**：支持一对一、一对多、多对一、多对多转账模式
- **并发执行**：可配置的工作线程数和速率控制
- **安全私钥管理**：支持环境变量、文件和交互式输入
- **余额查询**：批量查询钱包余额
- **灵活配置**：支持配置文件和命令行参数
- **多种输出格式**：表格、JSON、CSV格式输出

## 快速开始

### 1. 下载和安装

```bash
# 下载最新版本
git clone https://github.com/your-username/wallet-transfer.git
cd wallet-transfer

# 编译
go build -o wallet-transfer main.go
```

### 2. 基础使用

```bash
# 查看帮助
./wallet-transfer --help

# 查询余额
./wallet-transfer balance --network sepolia

# 简单转账
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia
```

## 安装

### 方式一：从源码编译

```bash
# 克隆仓库
git clone https://github.com/your-username/wallet-transfer.git
cd wallet-transfer

# 安装依赖
go mod tidy

# 编译
go build -o wallet-transfer main.go

# 运行
./wallet-transfer --help
```

### 方式二：直接下载

从 [Releases](https://github.com/your-username/wallet-transfer/releases) 页面下载预编译的二进制文件。

## 使用示例

### 余额查询

```bash
# 查询所有钱包余额
./wallet-transfer balance --network sepolia

# 查询指定地址余额
./wallet-transfer balance \
  --addresses 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --network sepolia
```

### 转账操作

```bash
# 一对一转账
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia

# 一对多转账
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients 0xAddr1,0xAddr2,0xAddr3 \
  --amount 0.005 \
  --network sepolia

# 多对一转账（资金汇总）
./wallet-transfer transfer \
  --mode many-to-one \
  --recipients 0xMainAddress \
  --amount 0.1 \
  --network sepolia

# 多对多转账
./wallet-transfer transfer \
  --mode many-to-many \
  --recipients 0xAddr1,0xAddr2 \
  --amount-range 0.001-0.01 \
  --network sepolia
```

### 高性能并发

```bash
# 启用并发执行
./wallet-transfer transfer \
  --mode many-to-many \
  --recipients 0xAddr1,0xAddr2,0xAddr3 \
  --amount 0.01 \
  --concurrent \
  --workers 20 \
  --network sepolia
```

## 配置

### 环境变量

```bash
# 设置私钥
export PRIVATE_KEYS="0x1234...,0x5678..."

# 设置网络
export WALLET_TRANSFER_NETWORK="sepolia"
export WALLET_TRANSFER_RPC_URL="https://sepolia.infura.io/v3/YOUR_PROJECT_ID"

# 性能配置
export WALLET_TRANSFER_WORKERS="20"
export WALLET_TRANSFER_TIMEOUT="600"
```

### 配置文件

创建 `config/config.yaml`：

```yaml
# 网络配置
networks:
  sepolia:
    name: "Sepolia Testnet"
    chain_id: 11155111
    rpc_url: "https://sepolia.infura.io/v3/YOUR_PROJECT_ID"
    explorer_url: "https://sepolia.etherscan.io"

# 默认设置
defaults:
  network: "sepolia"
  concurrent: true
  workers: 10
  timeout: 300
```

## 输出格式

支持多种输出格式：

```bash
# 表格格式（默认）
./wallet-transfer balance --network sepolia

# JSON格式
./wallet-transfer balance --network sepolia --output json

# CSV格式
./wallet-transfer balance --network sepolia --output csv
```

## 性能特性

- **高并发**：支持多线程并发执行，显著提升处理速度
- **智能重试**：内置重试机制，处理网络异常和临时故障
- **速率限制**：防止过度请求，保护RPC节点
- **断路器**：自动检测和处理持续性故障
- **内存优化**：高效的内存使用，支持大规模操作

## 安全特性

- **私钥保护**：支持环境变量和文件存储，避免硬编码
- **网络验证**：自动验证网络配置和Chain ID
- **金额限制**：可配置的转账金额限制
- **地址验证**：严格的地址格式验证
- **审计日志**：详细的操作日志记录

## 故障排除

### 常见问题

1. **insufficient funds**
   ```bash
   ./wallet-transfer balance --network sepolia
   ```

2. **connection timeout**
   ```bash
   ./wallet-transfer transfer --timeout 600 --rpc-url https://alternative-rpc.com
   ```

3. **gas price too low**
   ```bash
   ./wallet-transfer transfer --auto-gas
   ```

## 贡献

欢迎贡献代码！请查看 [贡献指南](CONTRIBUTING.md) 了解详细信息。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 支持

- 📖 [使用文档](docs/USAGE.md)
- 🔒 [安全指南](docs/SECURITY.md)
- 📚 [API文档](docs/API.md)
- 💡 [基础示例](examples/basic_usage.md)

如有问题，请提交 [Issue](https://github.com/your-username/wallet-transfer/issues)。