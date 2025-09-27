# Wallet Transfer 基础使用示例

本文档提供了 Wallet Transfer 的基础使用示例，帮助您快速上手。

## 环境准备

### 1. 设置私钥

```bash
# 方式1：环境变量
export PRIVATE_KEY="your_private_key_here"

# 方式2：多个私钥（逗号分隔）
export PRIVATE_KEY="key1,key2,key3"
```

### 2. 获取测试代币

在测试网络上，您需要获取一些测试代币：
- Sepolia: https://sepoliafaucet.com/
- Goerli: https://goerlifaucet.com/
- Mumbai: https://faucet.polygon.technology/

## 基础操作

### 1. 查看帮助信息

```bash
./wallet-transfer --help
./wallet-transfer transfer --help
./wallet-transfer balance --help
```

### 2. 基础余额查询

```bash
# 查询当前私钥对应地址的余额
./wallet-transfer balance --network sepolia

# 查询指定地址的余额
./wallet-transfer balance \
  --addresses 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --network sepolia

# 以JSON格式输出余额
./wallet-transfer balance --network sepolia --output json
```

## 转账操作示例

### 1. 一对一转账 (One-to-One)

最基础的转账模式，一个发送方向一个接收方转账

```bash
# 基础一对一转账
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia

# 指定Gas价格的转账
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --gas-price 20000000000 \
  --network sepolia
```

### 2. 一对多转账 (One-to-Many)

一个发送方向多个接收方转账，常用于空投测试

```bash
# 向多个地址转账相同金额
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount 0.01 \
  --network sepolia

# 使用随机金额范围
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount-range 0.001-0.01 \
  --network sepolia
```

### 3. 多对一转账 (Many-to-One)

多个发送方向一个接收方转账，常用于资金归集

```bash
# 多个钱包向一个地址转账
./wallet-transfer transfer \
  --mode many-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia

# 转账所有余额（保留Gas费用）
./wallet-transfer transfer \
  --mode many-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount-range 0.001-0.01 \
  --network sepolia
```

### 4. 多对多转账 (Many-to-Many)

多个发送方向多个接收方转账，最复杂的转账模式

```bash
# 多对多转账
./wallet-transfer transfer \
  --mode many-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount-range 0.001-0.01 \
  --network sepolia
```

## 高级功能

### 1. 批量处理

#### 从文件读取地址列表

创建一个 `recipients.txt` 文件：
```
0x742d35Cc6634C0532925a3b8D4C9db96590c6C87
0x8ba1f109551bD432803012645Hac136c22C177e9
0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5
```

然后执行批量转账：
```bash
# 读取文件中的地址列表
RECIPIENTS=$(cat recipients.txt | tr '\n' ',' | sed 's/,$//')
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients $RECIPIENTS \
  --amount 0.001 \
  --network sepolia
```

### 2. 不同输出格式

```bash
# JSON格式输出
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --output json \
  --network sepolia

# CSV格式输出
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --output csv \
  --network sepolia

# 表格格式输出（默认）
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --output table \
  --network sepolia
```

### 3. 配置文件使用

创建配置文件 `config/config.yaml`：
```yaml
network: sepolia
workers: 10
timeout: 300
concurrent: true
output: json
```

使用配置文件：
```bash
./wallet-transfer transfer \
  --config ./config/config.yaml \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01
```

## 实用脚本示例

### 1. 空投脚本

```bash
#!/bin/bash
# airdrop.sh - 批量空投脚本

NETWORK="sepolia"
AMOUNT="0.001"
RECIPIENTS_FILE="airdrop_list.txt"

echo "开始空投..."
./wallet-transfer balance --network $NETWORK

RECIPIENTS=$(cat $RECIPIENTS_FILE | tr '\n' ',' | sed 's/,$//')
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients $RECIPIENTS \
  --amount $AMOUNT \
  --network $NETWORK \
  --output json

echo "空投完成！"
./wallet-transfer balance --network $NETWORK --output table
```

### 2. 资金归集脚本

```bash
#!/bin/bash
# collect.sh - 资金归集脚本

NETWORK="sepolia"
COLLECTOR_ADDRESS="0x742d35Cc6634C0532925a3b8D4C9db96590c6C87"

echo "开始资金归集..."
./wallet-transfer balance --network $NETWORK --output table

./wallet-transfer transfer \
  --mode many-to-one \
  --recipients $COLLECTOR_ADDRESS \
  --amount-range 0.001-0.01 \
  --network $NETWORK \
  --concurrent \
  --workers 20

echo "归集完成！"
```

### 3. 压力测试脚本

```bash
#!/bin/bash
# stress_test.sh - 网络压力测试

NETWORK="sepolia"
DURATION=300  # 5分钟
RECIPIENTS="0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9"

echo "开始压力测试，持续时间: ${DURATION}秒"

timeout $DURATION ./wallet-transfer transfer \
  --mode many-to-many \
  --recipients $RECIPIENTS \
  --amount-range 0.0001-0.001 \
  --network $NETWORK \
  --concurrent \
  --workers 50 \
  --output json

echo "压力测试完成！"
```

## 监控和分析

### 1. 实时余额监控

```bash
# 实时监控余额变化
./wallet-transfer balance --network sepolia --output json

# 监控多个地址
ADDRESSES="0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9"
./wallet-transfer balance --addresses $ADDRESSES --network sepolia

# 生成余额报告
./wallet-transfer transfer \
  --mode one-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --output json \
  --network sepolia

# 转账后余额检查
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount 0.001 \
  --network sepolia

# 批量余额查询
./wallet-transfer transfer \
  --mode many-to-one \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87 \
  --amount 0.01 \
  --network sepolia

# 复杂转账场景
./wallet-transfer transfer \
  --mode many-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount-range 0.001-0.01 \
  --network sepolia \
  --concurrent \
  --workers 10

### 2. 持续监控

```bash
# 每5秒检查一次余额
watch -n 5 './wallet-transfer balance --network sepolia'
```

### 3. 交易历史分析

```bash
# 执行转账并保存结果
./wallet-transfer transfer \
  --mode one-to-many \
  --recipients 0x742d35Cc6634C0532925a3b8D4C9db96590c6C87,0x8ba1f109551bD432803012645Hac136c22C177e9 \
  --amount 0.001 \
  --output json \
  --network sepolia > transfer_history.json
```

## 故障排除

### 常见问题

1. **余额不足**
   - 检查账户余额是否足够支付转账金额和Gas费用
   - 使用测试网络水龙头获取测试代币

2. **Gas价格过低**
   - 网络拥堵时适当提高Gas价格
   - 使用 `--auto-gas` 参数自动估算

3. **网络超时**
   - 增加 `--timeout` 参数值
   - 检查网络连接和RPC节点状态

4. **私钥格式错误**
   - 确保私钥格式正确（64位十六进制字符串）
   - 检查环境变量设置

通过这些基础示例，您可以快速掌握 Wallet Transfer 的核心功能。建议从简单的余额查询开始，逐步尝试各种转账模式，最后再进行复杂的批量操作。