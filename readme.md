# Transfer Tool

基于 `urfave/cli/v2` 的多钱包转账与查询 CLI 工具

## 项目结构

```
transfer-tool/
├── cmd/transfer-tool/          # 主程序入口
│   └── main.go
├── internal/                   # 内部包
│   ├── commands/              # 命令实现
│   │   ├── balance.go         # 余额查询
│   │   ├── batch.go           # 批量转账
│   │   └── send.go            # 单笔转账
│   ├── config/                # 配置管理
│   │   ├── app.go             # 应用配置
│   │   └── batch.go           # 批量转账配置
│   └── wallet/                # 钱包管理
│       └── manager.go
├── pkg/                       # 公共包
│   ├── ethereum/              # 以太坊相关
│   └── excel/                 # Excel处理
├── configs/                   # 配置文件
│   ├── config.example.yaml    # 配置示例
│   ├── config.yaml            # 实际配置
│   ├── env.example            # 环境变量示例
│   ├── recipients.example.xlsx # 接收方示例
│   └── recipients.xlsx        # 实际接收方数据
├── data/                      # 数据目录
│   ├── .gitkeep              # 保持目录结构
│   └── batch_report_*.md     # 批量转账报告
├── docs/                      # 文档
│   ├── README.md             # 使用说明
│   ├── ARCHITECTURE.md       # 架构文档
│   └── ...                   # 其他文档
├── go.mod                     # Go模块文件
├── go.sum                     # 依赖校验文件
├── transfer-tool.exe          # 编译后的可执行文件
└── README.md                  # 项目说明
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
- 📊 **Markdown报告**: 生成格式化的Markdown报告，保存在data目录

## 详细文档

- [使用说明](docs/README.md) - 详细的使用指南
- [环境变量配置](docs/ENV_CONFIG.md) - 环境变量配置说明
- [RPC配置](docs/RPC_CONFIG.md) - RPC节点配置说明
- [架构文档](docs/ARCHITECTURE.md) - 项目架构说明

## 安全提醒

- 私钥仅存储在 `.env` 文件中，绝不通过命令行暴露
- 主网操作需要额外确认
- 建议先在测试网验证功能
