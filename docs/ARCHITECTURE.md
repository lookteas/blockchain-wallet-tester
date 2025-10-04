# 项目架构说明

## 目录结构

```
transfer-tool/
├── cmd/                           # 应用程序入口点
│   └── transfer-tool/             # 主程序
│       └── main.go               # CLI应用入口，定义命令和全局选项
├── internal/                      # 内部包（不对外暴露）
│   ├── wallet/                   # 钱包管理模块
│   │   └── manager.go           # 钱包管理器，私钥加载，地址推导，以太坊交互
│   ├── commands/                 # 命令处理模块
│   │   ├── send.go              # 单笔转账命令实现
│   │   ├── balance.go           # 余额查询命令实现
│   │   └── batch.go             # 批量转账命令实现
│   └── config/                   # 配置管理模块
│       └── batch.go             # 批量转账配置，Excel处理，报告生成
├── pkg/                          # 可复用的包（预留）
│   ├── ethereum/                 # 以太坊相关工具包
│   └── excel/                    # Excel处理工具包
├── configs/                      # 配置文件目录
│   ├── config.example.yaml      # 批量转账配置模板
│   ├── env.example              # 环境变量模板
│   ├── recipients.xlsx          # 接收方数据文件
│   └── recipients.example.xlsx  # 接收方数据模板
├── docs/                         # 文档目录
│   ├── README.md                # 详细项目说明
│   ├── USAGE.md                 # 使用指南
│   ├── prd.md                   # 原始需求文档
│   └── ARCHITECTURE.md          # 本架构说明文档
├── go.mod                        # Go模块定义
├── go.sum                        # Go模块校验和
└── README.md                     # 项目根目录说明
```

## 模块职责

### cmd/transfer-tool/
- **职责**: 应用程序入口点
- **内容**: CLI框架定义，命令注册，全局选项配置
- **特点**: 只负责组装各个模块，不包含业务逻辑

### internal/wallet/
- **职责**: 钱包管理和以太坊交互
- **功能**:
  - 私钥加载和地址推导
  - 网络连接管理
  - 余额查询
  - Gas估算
  - 交易构建和发送
- **特点**: 封装所有以太坊相关操作

### internal/commands/
- **职责**: 命令处理逻辑
- **模块**:
  - `send.go`: 单笔转账命令
  - `balance.go`: 余额查询命令  
  - `batch.go`: 批量转账命令
- **特点**: 每个命令独立文件，职责清晰

### internal/config/
- **职责**: 配置管理和数据处理
- **功能**:
  - YAML配置文件解析
  - Excel文件处理
  - 批量转账报告生成
- **特点**: 统一管理所有配置相关操作

### pkg/
- **职责**: 可复用的工具包（预留）
- **规划**: 未来可提取通用功能到此目录

### configs/
- **职责**: 配置文件和示例数据
- **内容**: 各种配置模板和示例文件
- **特点**: 用户可直接复制使用

### docs/
- **职责**: 项目文档
- **内容**: 使用说明、架构文档、需求文档
- **特点**: 完整的文档体系

## 设计原则

### 1. 分层架构
- **表现层**: `cmd/` - CLI接口
- **业务层**: `internal/commands/` - 业务逻辑
- **数据层**: `internal/wallet/`, `internal/config/` - 数据操作
- **工具层**: `pkg/` - 通用工具

### 2. 模块化设计
- 每个模块职责单一
- 模块间依赖关系清晰
- 便于测试和维护

### 3. 配置分离
- 配置文件独立存放
- 支持多种配置格式
- 提供完整示例

### 4. 文档完整
- 架构文档
- 使用文档
- 需求文档
- 代码注释

## 依赖关系

```
cmd/transfer-tool/
    ↓
internal/commands/
    ↓
internal/wallet/ + internal/config/
    ↓
pkg/ (预留)
```

## 扩展性

### 添加新命令
1. 在 `internal/commands/` 创建新文件
2. 在 `cmd/transfer-tool/main.go` 注册命令
3. 实现命令逻辑

### 添加新网络
1. 在 `internal/wallet/manager.go` 添加网络配置
2. 更新网络选择逻辑

### 添加新配置格式
1. 在 `internal/config/` 添加解析器
2. 更新配置加载逻辑

## 安全考虑

- 私钥仅在 `internal/wallet/` 中处理
- 敏感信息不写入日志
- 配置文件路径验证
- 输入参数严格校验
