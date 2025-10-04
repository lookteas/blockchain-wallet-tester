# 环境变量配置说明

## 概述

Transfer Tool 支持通过环境变量进行全局配置，程序启动时会自动加载 `.env` 文件，无需通过命令行参数指定。

## 自动加载机制

程序启动时会按以下顺序自动查找并加载环境变量文件：

1. `.env` (根目录)
2. `configs/.env`
3. `configs/env.example`

> **特别注意**：找到第一个存在的文件后就会加载，不会继续查找其他文件。

## 环境变量配置

### 私钥配置

```env
# 私钥配置（必需）
PRIVATE_KEYS=your_private_key_1,your_private_key_2,your_private_key_3
```

### RPC节点配置

```env
# 测试网
SEPOLIA_RPC_URL=https://sepolia.drpc.org
GOERLI_RPC_URL=https://rpc.ankr.com/eth_goerli

# 主网
MAINNET_RPC_URL=https://ethereum.publicnode.com

# 其他链
BNB_RPC_URL=https://bsc-dataseed.binance.org/
POLYGON_RPC_URL=https://polygon-rpc.com/
```

### API密钥配置

```env
# RPC服务商API密钥（可选）
INFURA_API_KEY=your_infura_api_key
ALCHEMY_API_KEY=your_alchemy_api_key
ANKR_API_KEY=your_ankr_api_key
```

## 配置文件中的环境变量

在YAML配置文件中，您可以使用环境变量：

```yaml
# 批量转账配置文件
transfer:
  token_address: ""

data_sources:
  recipients_xlsx: "./configs/recipients.xlsx"

# RPC节点配置（使用环境变量）
rpc_config:
  sepolia: "${SEPOLIA_RPC_URL}"
  mainnet: "${MAINNET_RPC_URL}"
  bnb: "${BNB_RPC_URL}"
  polygon: "${POLYGON_RPC_URL}"

# API密钥配置（从环境变量读取）
api_keys:
  infura: "${INFURA_API_KEY}"
  alchemy: "${ALCHEMY_API_KEY}"
  ankr: "${ANKR_API_KEY}"
```

## 配置优先级

1. **系统环境变量** (最高优先级)
2. **.env文件中的变量**
3. **默认配置** (最低优先级)

## 使用示例

### 1. 基本配置

```bash
# 复制示例文件
cp configs/env.example .env

# 编辑配置文件
notepad .env
```

### 2. 使用工具

```bash
# 查看余额（自动使用.env中的配置）
./transfer-tool balance

# 单笔转账
./transfer-tool send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.1

# 批量转账（使用环境变量配置）
./transfer-tool batch --config configs/config_with_env.yaml
```

### 3. 不同网络

```bash
# 使用Sepolia测试网
./transfer-tool balance

# 使用主网
./transfer-tool --network mainnet balance

# 使用BSC链
./transfer-tool --network bnb balance
```

## 高级配置

### 使用Infura API

```env
# .env文件
PRIVATE_KEYS=your_private_key_1,your_private_key_2
INFURA_API_KEY=your_infura_project_id

# 在配置文件中使用
# configs/config_with_env.yaml
rpc_config:
  sepolia: "https://sepolia.infura.io/v3/${INFURA_API_KEY}"
  mainnet: "https://mainnet.infura.io/v3/${INFURA_API_KEY}"
```

### 使用Alchemy API

```env
# .env文件
PRIVATE_KEYS=your_private_key_1,your_private_key_2
ALCHEMY_API_KEY=your_alchemy_api_key

# 在配置文件中使用
rpc_config:
  sepolia: "https://eth-sepolia.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
  mainnet: "https://eth-mainnet.g.alchemy.com/v2/${ALCHEMY_API_KEY}"
```

## 安全建议

1. **不要提交.env文件**: 将 `.env` 添加到 `.gitignore`
2. **使用示例文件**: 提交 `configs/env.example` 作为模板
3. **保护API密钥**: 不要将API密钥硬编码在配置文件中
4. **定期轮换**: 定期更换API密钥和私钥

## 故障排除

### 环境变量未加载

- 检查 `.env` 文件是否存在
- 检查文件格式是否正确（KEY=VALUE）
- 检查是否有语法错误

### RPC连接失败

- 检查RPC URL是否正确
- 检查API密钥是否有效
- 尝试使用其他RPC提供商

### 配置文件错误

- 检查YAML语法
- 检查环境变量名称是否正确
- 检查变量是否已定义

## 示例文件

参考以下文件了解完整配置：

- `configs/env.example` - 环境变量示例
- `configs/config.example.yaml` - 配置文件示例
- `configs/config_with_env.yaml` - 使用环境变量的配置示例
