# Mini Chain - 轻量级区块链实现

Mini Chain 是一个用 Go 语言编写的教育性区块链项目，展示了区块链技术的核心概念和实现原理。该项目包含了多种不同的实现方式，适合学习和研究区块链技术。

## 项目概述

本项目旨在提供一个完整但简化的区块链实现，涵盖以下核心概念：
- 分布式账本技术
- 工作量证明共识机制
- P2P 网络通信
- 密码学基础
- UTXO 交易模型
- REST API 接口

详细的内容介绍全在微信公众号中。干货持续更新，敬请关注「代码扳手」微信公众号：

<img width="430" height="430" alt="image" src="wx.jpg" />

## 目录结构

```
mini_chain/
├── cmd/                    # 命令行应用程序
│   └── node/              # 主节点应用程序
├── gossip/                # 基于 Gossip 协议的区块链实现
│   ├── core/              # 核心区块链功能模块
│   │   ├── README.md      # 核心模块文档
│   │   ├── blockchain.go  # 核心区块链实现
│   │   ├── blockchain_test.go  # 核心区块链测试
│   │   ├── example_test.go     # 示例测试
│   │   └── integration_test.go # 集成测试
│   └── main.go            # Gossip 主程序
├── internal/              # 内部包
│   ├── api/               # REST API 和 WebSocket 接口
│   │   ├── api.go         # API 实现
│   │   └── ws.go         # WebSocket 管理
│   ├── blockchain/        # 区块链核心实现
│   │   ├── block.go       # 区块数据结构
│   │   ├── blockchain.go  # 区块链管理
│   │   ├── mempool.go     # 内存池管理
│   │   ├── pow.go         # 工作量证明算法
│   │   ├── transaction.go # 交易处理
│   │   └── utxo.go       # UTXO 模型实现
│   ├── p2p/               # P2P 网络层
│   │   ├── discovery.go   # 节点发现
│   │   ├── message.go     # 消息格式
│   │   └── p2p.go        # P2P 网络实现
│   └── wallet/            # 钱包功能
│       ├── account.go     # 账户管理
│       ├── keystore.go    # 密钥存储
│       └── wallet.go      # 钱包工具
├── libp2p/                # 基于 libp2p 的实现
│   └── main.go            # libp2p 主程序
├── p2p/                   # 简单 TCP P2P 实现
│   ├── README.md          # P2P 实现文档
│   └── main.go            # P2P 主程序
├── TESTING.md             # 测试指南
├── main.go                # 主程序入口
├── start_nodes.bat        # Windows 多节点启动脚本
├── start_nodes.sh         # Linux/Mac 多节点启动脚本
└── test_network.py        # 网络测试脚本
```

## 核心功能

### 1. 多种区块链实现
- **Simple P2P**: 基于原始 TCP 连接的简单点对点网络
- **Gossip Protocol**: 基于 Gossip 协议的消息传播机制
- **libp2p**: 使用工业级 libp2p 库构建的现代 P2P 网络

### 2. 区块链核心技术
- **工作量证明 (PoW)**: 基于哈希难题的共识算法
- **UTXO 模型**: 比特币风格的未花费交易输出模型
- **密码学安全**: ECDSA 签名验证和 SHA256 哈希算法
- **内存池管理**: 交易缓存和去重机制
- **区块验证**: 完整的区块头和工作量证明验证

### 3. 网络功能
- **节点发现**: mDNS 局域网发现和 DHT 分布式发现
- **消息传播**: GossipSub 消息广播机制
- **数据同步**: 区块链状态同步和冲突解决
- **REST API**: HTTP 接口用于外部系统交互
- **WebSocket**: 实时事件推送和订阅

### 4. 钱包系统
- **密钥管理**: ECDSA 密钥对生成和管理
- **地址生成**: 公钥到地址的转换
- **交易签名**: 数字签名和验证功能
- **密钥存储**: 加密密钥存储机制

## 快速开始

### 环境要求
- Go 1.15 或更高版本
- Python 3.6+ (仅用于测试脚本)

### 安装和构建
```bash
# 克隆项目
git clone <repository-url>
cd mini_chain

# 安装依赖
go mod tidy
```

### 运行节点
```bash
# 运行主节点
go run main.go 3000 8080

# 运行多个节点进行测试
python test_network.py
```

### 使用脚本启动网络
```bash
# Windows
start_nodes.bat

# Linux/Mac
chmod +x start_nodes.sh
./start_nodes.sh
```

## 测试

### 单元测试
```bash
# 运行核心区块链测试
cd gossip/core
go test -v

# 查看测试覆盖率
go test -cover
```

### 性能测试
```bash
# 运行基准测试
cd gossip/core
go test -bench=.
```

### 网络测试
参考 [TESTING.md](file:///G:/wxblog/mini_chain/TESTING.md) 文件了解详细的网络测试方法。

## 开发指南

### 代码规范
- 遵循 Go 语言标准格式
- 提供详细的中文注释
- 保持完整的单元测试覆盖

### 扩展建议
1. 添加持久化存储（如 BadgerDB 或 LevelDB）
2. 实现更复杂的共识机制（如 PoS 或 DPoS）
3. 添加智能合约支持（类似 Ethereum）
4. 实现更完善的网络安全机制
5. 添加区块浏览器功能

## 项目架构

### 主要组件关系

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Wallet    │    │  Blockchain  │    │     P2P     │
│             │◄──►│              │◄──►│             │
│ - Accounts  │    │ - Blocks     │    │ - Network   │
│ - Keys      │    │ - UTXO Set   │    │ - Messages  │
│ - Signing   │    │ - Mempool    │    │ - Discovery │
└─────────────┘    └──────────────┘    └─────────────┘
                          ▲
                          │
                   ┌──────▼──────┐
                   │     API     │
                   │             │
                   │ - REST      │
                   │ - WebSocket │
                   └─────────────┘
```

### 数据流向

1. **交易创建**: 用户通过钱包创建并签名交易
2. **交易广播**: 交易通过 P2P 网络广播到所有节点
3. **交易池**: 节点将交易放入内存池等待处理
4. **区块挖掘**: 矿工从交易池中选择交易进行区块挖掘
5. **区块验证**: 其他节点验证新区块并添加到本地链
6. **状态同步**: 网络中的所有节点保持状态一致

## 学习资源

该项目适合作为以下学习目的：
- 区块链基础概念理解
- Go 语言分布式系统开发
- P2P 网络编程实践
- 密码学在区块链中的应用
- 共识算法实现原理

---

*注意：此项目仅供学习和研究使用，不应在生产环境中部署。*