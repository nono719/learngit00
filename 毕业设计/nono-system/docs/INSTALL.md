# 系统安装与启动指南

## 前置要求

### 必需软件
- **Go 1.19+** - [下载地址](https://golang.org/dl/)
- **Node.js 16+** - [下载地址](https://nodejs.org/)
- **PostgreSQL 15+** - [下载地址](https://www.postgresql.org/download/)
- **Redis 7+** - [下载地址](https://redis.io/download/)
- **Docker & Docker Compose** (可选，用于容器化部署) - [下载地址](https://www.docker.com/)

### 区块链环境
- **FISCO BCOS** 或 **Hyperledger Fabric** 或 **本地以太坊节点**
- 已部署的智能合约地址

## 安装步骤

### 1. 克隆项目

```bash
cd /Users/chenminggang/毕业设计/nono-system
```

### 2. 配置数据库

#### 创建PostgreSQL数据库

```bash
# 使用psql连接PostgreSQL
psql -U postgres

# 创建数据库和用户
CREATE DATABASE nono_system;
CREATE USER nono WITH PASSWORD 'nono123';
GRANT ALL PRIVILEGES ON DATABASE nono_system TO nono;
\q
```

#### 启动Redis

```bash
# 使用Docker启动Redis（推荐）
docker run -d -p 6379:6379 redis:7-alpine

# 或使用本地Redis
redis-server
```

### 3. 配置区块链

#### 选项A: 使用本地测试链（Ganache/Truffle）

```bash
# 安装Ganache CLI
npm install -g ganache-cli

# 启动本地测试链
ganache-cli -p 8545
```

#### 选项B: 使用FISCO BCOS

参考 [FISCO BCOS文档](https://fisco-bcos-documentation.readthedocs.io/) 搭建节点

### 4. 部署智能合约

```bash
# 进入合约目录
cd contracts

# 使用Truffle或Hardhat编译和部署
# 示例（需要根据实际使用的工具链调整）
truffle compile
truffle migrate --network development

# 记录部署后的合约地址，更新到配置文件中
```

### 5. 配置服务

#### 复制配置文件模板

```bash
# 预言机服务配置
cp config/oracle.config.example.yaml oracle/config/config.yaml

# 后端服务配置
cp config/backend.config.example.yaml backend/config/config.yaml
```

#### 编辑配置文件

**oracle/config/config.yaml:**
- 更新 `blockchain.contract_addr` 为部署的合约地址
- 更新 `blockchain.private_key` 为预言机账户私钥
- 更新 `blockchain.rpc_url` 为区块链节点地址

**backend/config/config.yaml:**
- 更新 `blockchain.contract_addr` 为部署的合约地址
- 更新数据库连接信息（如需要）

### 6. 安装依赖

#### Go依赖

```bash
# 在项目根目录
go mod download

# 安装预言机服务依赖
cd oracle
go mod download

# 安装后端服务依赖
cd ../backend
go mod download
```

#### 前端依赖

```bash
cd frontend
npm install
```

## 启动系统

### 方式一：使用Makefile（推荐）

```bash
# 在项目根目录

# 启动预言机服务（新终端窗口）
make oracle

# 启动后端API服务（新终端窗口）
make backend

# 启动前端应用（新终端窗口）
cd frontend && npm run dev
```

### 方式二：手动启动

#### 1. 启动预言机服务

```bash
cd oracle
go run ./cmd/oracle
```

服务将在 `http://localhost:9000` 启动

#### 2. 启动后端API服务

```bash
cd backend
go run ./cmd/server
```

服务将在 `http://localhost:8080` 启动

#### 3. 启动前端应用

```bash
cd frontend
npm run dev
```

前端将在 `http://localhost:3000` 启动

### 方式三：使用Docker Compose

```bash
# 在项目根目录
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 验证安装

### 1. 检查服务健康状态

```bash
# 检查后端服务
curl http://localhost:8080/health

# 检查预言机服务
curl http://localhost:9000/api/v1/health
```

### 2. 访问前端界面

打开浏览器访问：`http://localhost:3000`

### 3. 测试API

```bash
# 注册设备
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "did": "did:example:device001",
    "device_id": "device001",
    "device_type": "sensor",
    "domain": "smart_home",
    "manufacturer": "Test Corp"
  }'

# 查询设备
curl http://localhost:8080/api/v1/devices/did:example:device001
```

## 常见问题

### 1. 数据库连接失败

- 检查PostgreSQL是否运行：`pg_isready`
- 检查数据库配置是否正确
- 检查防火墙设置

### 2. 区块链连接失败

- 检查区块链节点是否运行
- 检查RPC URL是否正确
- 检查合约地址是否正确

### 3. 端口被占用

- 修改配置文件中的端口号
- 或使用 `lsof -i :8080` 查找占用进程并关闭

### 4. Go模块下载失败

```bash
# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

## 下一步

- 查看 [API文档](API.md) 了解API接口
- 查看 [部署文档](DEPLOY.md) 了解生产环境部署
- 查看 [开发文档](DEVELOPMENT.md) 了解开发指南

