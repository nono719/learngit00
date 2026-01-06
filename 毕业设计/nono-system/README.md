# 基于区块链的物联网设备跨域身份认证系统

## 项目简介

本项目实现了一个基于区块链的去中心化物联网设备身份认证系统，支持跨域设备身份验证与授权。

## 系统架构

系统采用四层架构设计：

1. **设备层（Device Layer）**：物联网设备实体
2. **预言机网络层（Oracle Network Layer）**：可信数据源与链外数据桥接
3. **区块链核心层（Blockchain Core Layer）**：去中心化身份管理与智能合约
4. **应用服务层（Application Service Layer）**：业务逻辑与API服务

## 技术栈

- **区块链平台**：FISCO BCOS / Hyperledger Fabric
- **共识机制**：PoA (Proof of Authority) / Raft
- **后端语言**：Go
- **智能合约**：Solidity
- **前端框架**：React / Vue.js
- **数据库**：PostgreSQL / MySQL

## 项目结构

```
nono-system/
├── contracts/              # 智能合约
├── oracle/                 # 预言机服务
├── backend/                # 后端API服务
├── frontend/               # 前端Web应用
├── docs/                   # 项目文档
├── config/                 # 配置文件
├── scripts/                # 部署脚本
└── tests/                  # 测试文件
```

## 快速开始

### 前置要求

- Go 1.19+
- Node.js 16+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)
- 区块链节点环境

### 快速启动（5分钟）

1. **启动依赖服务**
```bash
# 使用Docker启动PostgreSQL和Redis
docker run -d --name nono-postgres \
  -e POSTGRES_DB=nono_system \
  -e POSTGRES_USER=nono \
  -e POSTGRES_PASSWORD=nono123 \
  -p 5432:5432 postgres:15-alpine

docker run -d --name nono-redis -p 6379:6379 redis:7-alpine
```

2. **配置服务**
```bash
# 复制配置文件
cp config/oracle.config.example.yaml oracle/config/config.yaml
cp config/backend.config.example.yaml backend/config/config.yaml
# 编辑配置文件，更新区块链连接信息
```

3. **启动服务**
```bash
# 方式1: 使用启动脚本
./scripts/start.sh

# 方式2: 使用Makefile（推荐）
make oracle    # 终端1
make backend   # 终端2
cd frontend && npm run dev  # 终端3
```

4. **访问系统**
- 前端界面: http://localhost:3000
- 后端API: http://localhost:8080
- 预言机服务: http://localhost:9000

### 详细安装说明

请参考 [docs/INSTALL.md](docs/INSTALL.md) 获取完整的安装和配置指南。

### 快速测试

请参考 [docs/QUICKSTART.md](docs/QUICKSTART.md) 获取快速测试指南。

## 功能特性

- ✅ 设备去中心化身份注册（DID）
- ✅ 设备状态实时更新
- ✅ 跨域身份认证
- ✅ 身份吊销机制
- ✅ 可信预言机服务
- ✅ 多数据源聚合验证

## 许可证

MIT License

