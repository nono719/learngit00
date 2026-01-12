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
# 方式1: 统一启动（推荐，一键启动所有服务）
go run ./cmd/startall

# 或编译后运行
go build -o bin/startall ./cmd/startall
./bin/startall

# 方式2: 使用启动脚本
./scripts/start.sh

# 方式3: 使用Makefile（分别启动）
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

### 联盟链部署与观察

- **Ganache 配置**：请参考 [docs/GANACHE_SETUP.md](docs/GANACHE_SETUP.md) 了解如何在 Ganache 上部署和查看交易
- **联盟链观察**：请参考 [docs/BLOCKCHAIN_OBSERVATION.md](docs/BLOCKCHAIN_OBSERVATION.md) 了解如何在联盟链上观察功能实现
- **验证指南**：请参考 [docs/VERIFY_BLOCKCHAIN.md](docs/VERIFY_BLOCKCHAIN.md) 了解如何验证跨链认证
- **前端上链**：请参考 [docs/FRONTEND_ONCHAIN.md](docs/FRONTEND_ONCHAIN.md) 了解如何在前端直接进行上链操作

## 功能特性

### 核心功能
- ✅ 设备去中心化身份注册（DID）
- ✅ 设备状态实时更新
- ✅ 跨域身份认证
- ✅ 身份吊销机制
- ✅ 可信预言机服务
- ✅ 多数据源聚合验证

### 新增功能
- ✅ **系统统计和仪表板** - 提供系统整体统计信息
- ✅ **设备操作历史** - 记录设备所有操作历史，支持审计追踪
- ✅ **批量操作** - 支持批量注册和批量更新设备状态
- ✅ **高级搜索** - 灵活的多条件搜索和过滤功能
- ✅ **数据导出** - 支持将设备和认证记录导出为 CSV 格式
- ✅ **用户权限管理** - 完整的基于角色的访问控制（RBAC）系统

### 权限管理
系统支持5种角色，每种角色具有不同的权限和数据访问范围：
- **系统管理员** - 全域数据权限，可操作所有功能
- **系统操作人员** - 域级数据权限，仅可操作所属域设备
- **预言机节点** - 受限数据权限，仅可更新设备状态
- **管理/审计人员** - 只读数据权限，可查询所有记录
- **普通用户** - 受限只读权限，仅可查看公开信息

详细权限说明请参考 [docs/PERMISSIONS.md](docs/PERMISSIONS.md)

详细功能说明请参考 [docs/NEW_FEATURES.md](docs/NEW_FEATURES.md)

## 许可证

MIT License

