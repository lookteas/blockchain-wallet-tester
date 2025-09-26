# Blockchain Wallet Transfer Tester

一个用于区块链多钱包转账测试的 Go 工具，支持批量转账、并发执行、多种私钥加载方式，确保私钥安全。

## 功能特性

- 🔒 **安全私钥管理**：支持环境变量、加密配置文件、交互式输入
- ⚡ **批量转账**：支持多个钱包向单个或多个地址转账
- 🚀 **并发执行**：可选择并发或顺序执行转账
- 📊 **余额监控**：转账前后自动检查钱包余额
- 🔄 **交易确认**：可等待交易确认后再继续
- 🌐 **多网络支持**：支持 Ethereum、BSC、Polygon 等 EVM 兼容链

## 目录结构



blockchain-wallet-tester/
├── README.md
├── go.mod
├── go.sum
├── main.go
├── config/
│   ├── config.go
│   └── config.example.json
├── wallet/
│   ├── wallet.go
│   └── loader.go
├── blockchain/
│   └── client.go
├── transfer/
│   └── batch.go
└── .gitignore

## 安装

### 前提条件

- Go 1.20+
- 区块链节点 RPC URL（本地测试网或测试网）

### 使用说明

1. **克隆项目**：

   ```
   bash
   
   git clone https://github.com/your-username/blockchain-wallet-tester.git
   
   cd blockchain-wallet-tester
   ```

   

2. **安装依赖**：

   ```
   bash
   
   go mod tidy
   ```

   

3. **编译**：

   ```
   bash
   
   go build -o wallet-tester
   ```

   

4. **配置**（选择一种方式）：

   - 环境变量方式：创建 `.env` 文件

   - 配置文件方式：复制并编辑 `config/config.json`

     

5. **运行**：

   ```
   # 使用配置文件
   ./wallet-tester --config config/config.json
   
   # 交互式模式
   ./wallet-tester --interactive
   
   # 仅查看余额
   ./wallet-tester --balance-only
   ```

   

​	



## 编译

go build -o wallet-tester



## 配置



### 1. 环境变量配置（推荐用于开发）

创建 `.env` 文件：

cp .env.example .env



编辑 `.env` 文件：

```
# 区块链 RPC URL
RPC_URL=http://localhost:8545

# 转账金额（单位：wei）
TRANSFER_AMOUNT=10000000000000000

# 目标地址（多个地址用逗号分隔）
TARGET_ADDRESSES=0x742d35Cc6634C0532925a3b8D4C9db9653fE8b01,0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199

# 私钥（多个私钥用逗号分隔，仅用于开发环境）
WALLET_PRIVATE_KEYS=your_private_key_1,your_private_key_2
```





### 2. 配置文件方式（推荐用于生产）

复制配置文件模板：

```
cp config/config.example.json config/config.json
```



编辑 `config/config.json`：



```
{
  "rpc_url": "http://localhost:8545",
  "transfer_amount": "10000000000000000",
  "target_addresses": [
    "0x742d35Cc6634C0532925a3b8D4C9db9653fE8b01",
    "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199"
  ],
  "private_keys": [
    "your_private_key_1",
    "your_private_key_2"
  ],
  "concurrent": true,
  "wait_confirmations": true,
  "confirmations": 1
}
```



## 使用方法

### 1. 基本使用



```
# 使用环境变量
./wallet-tester

# 使用配置文件
./wallet-tester --config config/config.json

# 交互式模式（私钥不会保存在任何文件中）
./wallet-tester --interactive
```



### 2. 命令行参数

```
# 查看帮助
./wallet-tester --help

# 指定配置文件
./wallet-tester --config /path/to/config.json

# 启用交互式模式
./wallet-tester --interactive

# 仅显示余额（不执行转账）
./wallet-tester --balance-only

# 指定并发数
./wallet-tester --concurrent
```



### 3. 环境变量优先级

环境变量 > 配置文件 > 默认值



### 4. 安全注意事项

⚠️ **重要安全提醒**：

1. **永远不要**将私钥提交到版本控制系统
2. **测试网络**：仅在测试网络或本地开发网络使用
3. **权限控制**：确保配置文件权限设置正确（`chmod 600 config.json`）
4. **生产环境**：生产环境应使用专业的密钥管理服务



## 示例场景

### 场景1：多个钱包向同一个地址转账

env

```
TARGET_ADDRESSES=0x742d35Cc6634C0532925a3b8D4C9db9653fE8b01
```



### 场景2：多个钱包向多个地址轮询转账

env

```
TARGET_ADDRESSES=0x742d35Cc6634C0532925a3b8D4C9db9653fE8b01,0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199
```



### 场景3：并发转账测试

json

```
{
  "concurrent": true
}
```

### 输出示例



```
Starting batch transfer with 3 wallets to 2 addresses
Wallet 0x123...456 balance: 1000000000000000000 wei
Wallet 0x789...012 balance: 1000000000000000000 wei
Wallet 0x345...678 balance: 1000000000000000000 wei

Sending transactions...
Sent transaction from 0x123...456 to 0x742...8b01, tx hash: 0xabc...def
Sent transaction from 0x789...012 to 0x862...1199, tx hash: 0xghi...jkl
Sent transaction from 0x345...678 to 0x742...8b01, tx hash: 0xmnop...qrst

Waiting for transaction confirmations...
Transaction 0xabc...def confirmed
Transaction 0xghi...jkl confirmed
Transaction 0xmnop...qrst confirmed

Final balances:
Address 0x123...456: 990000000000000000 wei
Address 0x789...012: 990000000000000000 wei
Address 0x345...678: 990000000000000000 wei
```

