# Transfer Tool

基于 `urfave/cli/v2` 的多钱包转账与查询 CLI 工具

## 项目结构

```
transfer-tool/
├── cmd/                    # 应用程序入口
│   └── transfer-tool/      # 主程序
├── internal/               # 内部包（不对外暴露）
│   ├── wallet/            # 钱包管理
│   ├── commands/          # 命令处理
│   └── config/            # 配置管理
├── pkg/                   # 可复用的包
│   ├── ethereum/          # 以太坊相关工具
│   └── excel/             # Excel处理工具
├── configs/               # 配置文件
│   ├── config.example.yaml
│   ├── env.example
│   └── recipients.xlsx
├── docs/                  # 文档
│   ├── README.md          # 详细说明
│   ├── USAGE.md           # 使用指南
│   └── prd.md             # 需求文档
└── go.mod                 # Go模块定义
```

## 快速开始

### 1. 编译工具

```bash
go build -o transfer-tool ./cmd/transfer-tool
```

### 2. 配置环境变量

```bash
cp configs/env.example .env
# 编辑 .env 文件，填入您的私钥和RPC配置
```

### 3. 基本使用

```bash
# 查看所有钱包余额
./transfer-tool balance

# 单笔转账
./transfer-tool send 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 0.1

# 批量转账
./transfer-tool batch --config configs/config.example.yaml
```

## 功能特性

- 🔐 **私钥安全**: 所有私钥仅从 `.env` 文件加载
- 🧭 **命令直观**: 接近自然语言的命令语法
- 👀 **余额全览**: 显示所有钱包的余额状态
- ⚖️ **批量轮询**: 自动轮询使用所有钱包
- 🌐 **多网络支持**: Sepolia、Goerli、Mainnet、BNB、Polygon
- ⚙️ **RPC配置**: 支持自定义RPC节点，提高连接稳定性
- 📊 **审计留痕**: 完整的操作记录和报告

## 详细文档

- [使用说明](docs/README.md) - 详细的使用指南
- [环境变量配置](docs/ENV_CONFIG.md) - 环境变量配置说明
- [RPC配置](docs/RPC_CONFIG.md) - RPC节点配置说明
- [架构文档](docs/ARCHITECTURE.md) - 项目架构说明

## 安全提醒

- 私钥仅存储在 `.env` 文件中，绝不通过命令行暴露
- 主网操作需要额外确认
- 建议先在测试网验证功能
