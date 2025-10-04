# Transfer Tool

基于 `urfave/cli/v2` 的多钱包转账与查询 CLI 工具

## 功能特性

- 🔐 **私钥安全**: 所有私钥仅从 `.env` 文件加载，绝不通过命令行暴露
- 🧭 **命令直观**: 日常操作命令接近自然语言
- 👀 **余额全览**: 显示所有钱包的余额状态
- ⚖️ **批量轮询**: 批量任务使用所有钱包，按 Round-Robin 均匀分配
- 🌐 **多网络支持**: 默认 Sepolia，支持 mainnet、polygon 等

## 快速开始

### 1. 编译工具
```bash
go build -o transfer-tool
```

### 2. 配置私钥
复制 `env.example` 为 `.env` 并填入您的私钥：
```bash
cp env.example .env
```

编辑 `.env` 文件：
```env
PRIVATE_KEYS=your_private_key_1,your_private_key_2,your_private_key_3
```

### 3. 基本使用

#### 查看所有钱包余额
```bash
./transfer-tool balance
```

#### 单笔转账
```bash
# 向指定地址转账 0.1 ETH
./transfer-tool send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.1

# 跳过确认提示
./transfer-tool send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.1 --yes

# 在主网转账（需要额外确认）
./transfer-tool --network mainnet send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 1.0
```

#### 批量转账
```bash
# 使用配置文件进行批量转账
./transfer-tool batch --config config.example.yaml
```

## 网络支持

- `sepolia` (默认) - 测试网
- `goerli` - 测试网  
- `mainnet` - 主网
- `bnb` - BSC链
- `polygon` - Polygon链

## 配置文件格式

### 批量转账配置 (config.yaml)
```yaml
transfer:
  token_address: ""  # 留空表示ETH转账

data_sources:
  recipients_xlsx: "./data/recipients.xlsx"  # 接收方Excel文件
```

### 接收方Excel文件格式
Excel文件必须包含以下列：
- `address`: 接收方以太坊地址
- `amount`: 转账金额（ETH）

示例：
| address | amount |
|---------|--------|
| 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 | 0.1 |
| 0x8ba1f109551bD432803012645Hac136c | 0.05 |

## 安全提醒

1. **私钥安全**: 私钥仅存储在 `.env` 文件中，绝不通过命令行暴露
2. **主网确认**: 主网操作需要输入 `MAINNET` 确认
3. **测试优先**: 建议先在测试网验证功能
4. **备份重要**: 请妥善保管私钥和配置文件

## 错误处理

- 余额不足时会显示具体需要的金额
- 网络连接失败时会显示错误信息
- 配置文件格式错误时会指出具体问题
- 所有操作都有详细的错误提示

## 报告生成

批量转账完成后会生成 JSON 格式的报告文件，包含：
- 每笔转账的详细状态
- 成功/失败统计
- 交易哈希和区块浏览器链接
- 错误信息（如有）

## 示例工作流

```bash
# 1. 查看所有钱包余额
./transfer-tool balance

# 2. 单笔测试转账
./transfer-tool send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.01

# 3. 批量空投
./transfer-tool batch --config airdrop.yaml

# 4. 查看主网余额（谨慎！）
./transfer-tool --network mainnet balance
```
