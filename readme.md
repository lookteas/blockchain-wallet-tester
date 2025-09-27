# Wallet Transfer - 区块链钱包测试工具

Wallet Transfer 是一个功能强大的区块链钱包转账工具，专为以太坊及其兼容网络设计。它支持批量转账操作、余额查询、并发执行和安全的私钥管理。

## 🚀 主要功能

- **多种区块链网络支持**：Ethereum、BSC、Polygon、Goerli、Sepolia、Mumbai
- **批量转账操作**：支持一对一、一对多、多对一、多对多转账模式
- **并发执行**：可配置的工作线程数和速率控制
- **安全私钥管理**：支持环境变量、文件和交互式输入
- **余额查询**：批量查询钱包余额
- **灵活配置**：支持配置文件和命令行参数
- **多种输出格式**：表格、JSON、CSV格式输出

## 📦 安装

### 从源码编译

```bash
git clone <repository-url>
cd gotester
go build -o gotester main.go
```

### 系统要求

- Go 1.19 或更高版本
- 网络连接（用于访问区块链RPC节点）

## 🔧 配置

### 配置文件

创建 `config/config.yaml` 文件：

```yaml
# 网络配置
networks:
  ethereum:
    name: "Ethereum Mainnet"
    chain_id: 1
    rpc_url: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
  sepolia:
    name: "Sepolia Testnet"
    chain_id: 11155111
    rpc_url: "https://sepolia.infura.io/v3/YOUR_PROJECT_ID"

# 默认设置
defaults:
  network: "sepolia"
  concurrent: true
  workers: 10
  timeout: 300
  confirmations: 1
```

### 环境变量

设置私钥环境变量：

```bash
# Windows
set PRIVATE_KEYS=0x1234...,0x5678...

# Linux/Mac
export PRIVATE_KEYS=0x1234...,0x5678...
```

## 📖 使用指南

### 基本命令

```bash
# 查看帮助
./gotester --help

# 查看转账命令帮助
./gotester transfer --help

# 查看余额命令帮助
./gotester balance --help
```

### 余额查询

```bash
# 查询钱包余额（从环境变量读取私钥）
./gotester balance --network sepolia

# 查询指定地址余额
./gotester balance --addresses 0x1234...,0x5678... --network sepolia

# 以JSON格式输出
./gotester balance --output json --network sepolia

# 以ETH为单位显示
./gotester balance --unit ether --network sepolia
```

### 转账操作

#### 一对一转账

```bash
./gotester transfer \
  --mode one-to-one \
  --recipients 0x1234...,0x5678... \
  --amount 0.01 \
  --unit ether \
  --network sepolia
```

#### 一对多转账

```bash
./gotester transfer \
  --mode one-to-many \
  --recipients 0x1234...,0x5678...,0x9abc... \
  --amount 0.005 \
  --unit ether \
  --network sepolia
```

#### 多对一转账

```bash
./gotester transfer \
  --mode many-to-one \
  --recipients 0x1234... \
  --amount 0.01 \
  --unit ether \
  --network sepolia
```

#### 多对多转账

```bash
./gotester transfer \
  --mode many-to-many \
  --recipients 0x1234...,0x5678... \
  --amount-range 0.001-0.01 \
  --unit ether \
  --network sepolia
```

### 高级选项

```bash
# 启用并发执行，设置工作线程数
./gotester transfer \
  --mode one-to-many \
  --recipients 0x1234... \
  --amount 0.01 \
  --concurrent \
  --workers 20 \
  --network sepolia

# 自定义Gas设置
./gotester transfer \
  --mode one-to-one \
  --recipients 0x1234... \
  --amount 0.01 \
  --gas-limit 25000 \
  --gas-price 20000000000 \
  --network sepolia

# 设置确认数和超时时间
./gotester transfer \
  --mode one-to-one \
  --recipients 0x1234... \
  --amount 0.01 \
  --confirmations 3 \
  --timeout 600 \
  --network sepolia
```

## 🔒 安全指南

### 私钥管理

1. **环境变量方式**（推荐用于开发环境）：
   ```bash
   export PRIVATE_KEYS=0x1234...,0x5678...
   ```

2. **文件方式**：
   创建 `private_keys.txt` 文件，每行一个私钥
   ```
   0x1234567890abcdef...
   0xfedcba0987654321...
   ```

3. **交互式输入**（最安全）：
   ```bash
   ./gotester transfer --private-keys interactive
   ```

### 安全建议

- ⚠️ **永远不要在生产环境中使用明文私钥**
- 🔐 使用硬件钱包或安全的密钥管理服务
- 🧪 在测试网络上充分测试后再在主网使用
- 💰 转账前确认余额充足（包括Gas费用）
- 🔍 仔细检查收款地址的正确性
- 📊 使用小额测试验证配置正确性

## 🛠️ 故障排除

### 常见问题

1. **连接RPC节点失败**
   - 检查网络连接
   - 验证RPC URL是否正确
   - 确认API密钥有效（如使用Infura等服务）

2. **私钥格式错误**
   - 确保私钥以 `0x` 开头
   - 验证私钥长度为64个十六进制字符

3. **余额不足**
   - 检查钱包ETH余额是否足够支付Gas费用
   - 验证转账金额设置是否正确

4. **Gas费用过高**
   - 使用 `--auto-gas` 自动估算Gas
   - 手动设置合适的 `--gas-price`

### 调试模式

```bash
# 启用详细日志
./gotester transfer --mode one-to-one --recipients 0x... --amount 0.01 --verbose

# 输出为JSON格式便于分析
./gotester transfer --mode one-to-one --recipients 0x... --amount 0.01 --output json
```

## 📊 输出格式

### 表格格式（默认）
```
=== 转账结果摘要 ===
+----------+-------+
|   指标   |  值   |
+----------+-------+
| 总任务数 |   5   |
| 成功     |   4   |
| 失败     |   1   |
+----------+-------+
```

### JSON格式
```json
{
  "total_tasks": 5,
  "successful": 4,
  "failed": 1,
  "total_amount": "50000000000000000",
  "total_fees": "1050000000000000",
  "duration": "45.2s",
  "tasks": [...]
}
```

### CSV格式
```csv
TaskID,From,To,Amount,Status,TxHash,Error,Duration
task-1,0x1234...,0x5678...,10000000000000000,completed,0xabcd...,,"2.1s"
```

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个项目。

## 📄 许可证

本项目采用MIT许可证。详见 [LICENSE](LICENSE) 文件。

## ⚠️ 免责声明

本工具仅用于测试目的。使用者需要自行承担使用风险，开发者不对任何损失负责。在主网使用前请务必在测试网络上充分测试。