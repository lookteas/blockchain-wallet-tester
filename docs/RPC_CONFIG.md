# RPC节点配置说明

## 概述

Transfer Tool 支持自定义RPC节点配置，您可以根据需要配置不同网络的RPC端点，以获得更好的连接稳定性和性能。

## 配置方式

### 1. 全局RPC配置

在 `configs/rpc.yaml` 文件中配置所有网络的RPC节点：

```yaml
# 测试网
sepolia: "https://sepolia.drpc.org"
goerli: "https://rpc.ankr.com/eth_goerli"

# 主网
mainnet: "https://ethereum.publicnode.com"

# 其他链
bnb: "https://bsc-dataseed.binance.org/"
polygon: "https://polygon-rpc.com/"
```

### 2. 批量转账专用RPC配置

在批量转账配置文件 `configs/config.example.yaml` 中也可以配置RPC：

```yaml
transfer:
  token_address: ""

data_sources:
  recipients_xlsx: "./configs/recipients.xlsx"

# RPC节点配置（可选）
rpc_config:
  sepolia: "https://your-custom-sepolia-rpc.com"
  mainnet: "https://your-custom-mainnet-rpc.com"
```

## 支持的RPC提供商

### 免费公共节点
- **Ethereum Public Node**: `https://ethereum.publicnode.com`
- **Sepolia Public Node**: `https://sepolia.drpc.org`
- **BSC**: `https://bsc-dataseed.binance.org/`
- **Polygon**: `https://polygon-rpc.com/`

### 付费服务（需要注册）
- **Infura**: https://infura.io/
- **Alchemy**: https://www.alchemy.com/
- **Ankr**: https://www.ankr.com/rpc/

## 配置优先级

1. **批量转账配置文件中的RPC配置** (最高优先级)
2. **全局RPC配置文件** (`configs/rpc.yaml`)
3. **默认内置RPC节点** (最低优先级)

## 使用示例

### 查看余额（使用全局RPC配置）
```bash
./transfer-tool.exe balance
./transfer-tool.exe --network mainnet balance
```

### 单笔转账（使用全局RPC配置）
```bash
./transfer-tool.exe send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.1
```

### 批量转账（使用配置文件中的RPC）
```bash
./transfer-tool.exe batch --config configs/config.example.yaml
```

## 自定义RPC节点

### 1. 修改全局配置
编辑 `configs/rpc.yaml` 文件，将URL替换为您的RPC节点：

```yaml
sepolia: "https://your-sepolia-rpc.com"
mainnet: "https://your-mainnet-rpc.com"
```

### 2. 修改批量转账配置
编辑您的批量转账配置文件，添加RPC配置：

```yaml
rpc_config:
  sepolia: "https://your-sepolia-rpc.com"
  mainnet: "https://your-mainnet-rpc.com"
```

## 故障排除

### 连接超时
如果遇到连接超时，尝试：
1. 更换RPC节点
2. 检查网络连接
3. 使用更稳定的RPC提供商

### 认证错误
如果遇到认证错误（如Ankr需要API密钥）：
1. 注册RPC服务商账号
2. 获取API密钥
3. 在URL中包含API密钥：`https://rpc.ankr.com/eth/YOUR_API_KEY`

### 速率限制
如果遇到速率限制：
1. 升级RPC服务计划
2. 使用多个RPC节点轮询
3. 降低请求频率

## 最佳实践

1. **测试网优先**: 先在测试网验证功能
2. **备用节点**: 配置多个RPC节点作为备用
3. **监控使用量**: 注意RPC服务的使用限制
4. **安全性**: 不要将API密钥提交到版本控制系统

## 配置验证

您可以通过以下命令验证RPC配置是否正常工作：

```bash
# 测试Sepolia网络
./transfer-tool.exe --network sepolia balance

# 测试主网
./transfer-tool.exe --network mainnet balance

# 测试BSC
./transfer-tool.exe --network bnb balance
```

如果命令执行成功且能查询到余额，说明RPC配置正确。

