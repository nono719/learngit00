# 快速启动指南

## 5分钟快速启动

### 1. 启动依赖服务

```bash
# 启动PostgreSQL（使用Docker）
docker run -d --name nono-postgres \
  -e POSTGRES_DB=nono_system \
  -e POSTGRES_USER=nono \
  -e POSTGRES_PASSWORD=nono123 \
  -p 5432:5432 \
  postgres:15-alpine

# 启动Redis（使用Docker）
docker run -d --name nono-redis \
  -p 6379:6379 \
  redis:7-alpine

# 启动本地测试区块链（Ganache）
npx ganache-cli -p 8545 -d
```

### 2. 初始化数据库

```bash
# 连接PostgreSQL并创建数据库
psql -U postgres -c "CREATE DATABASE nono_system;"
psql -U postgres -c "CREATE USER nono WITH PASSWORD 'nono123';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE nono_system TO nono;"
```

### 3. 配置服务

```bash
# 复制配置文件
cp config/oracle.config.example.yaml oracle/config/config.yaml
cp config/backend.config.example.yaml backend/config/config.yaml

# 注意：需要先部署智能合约并更新合约地址
```

### 4. 启动服务

打开三个终端窗口：

**终端1 - 预言机服务：**
```bash
cd oracle
go run ./cmd/oracle
```

**终端2 - 后端API服务：**
```bash
cd backend
go run ./cmd/server
```

**终端3 - 前端应用：**
```bash
cd frontend
npm install  # 首次运行需要
npm run dev
```

### 5. 访问系统

- 前端界面：http://localhost:3000
- 后端API：http://localhost:8080
- 预言机服务：http://localhost:9000

## 测试系统功能

### 1. 创建域

```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{
    "name": "smart_home",
    "description": "智能家居域"
  }'
```

### 2. 注册设备

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "did": "did:example:device001",
    "device_id": "device001",
    "device_type": "sensor",
    "domain": "smart_home",
    "manufacturer": "Test Corp"
  }'
```

### 3. 跨域认证

```bash
curl -X POST http://localhost:8080/api/v1/auth/cross-domain \
  -H "Content-Type: application/json" \
  -d '{
    "device_did": "did:example:device001",
    "source_domain": "smart_home",
    "target_domain": "industrial_iot"
  }'
```

### 4. 查看认证记录

```bash
curl http://localhost:8080/api/v1/auth/records/did:example:device001
```

## 使用前端界面

1. 打开浏览器访问 http://localhost:3000
2. 在"设备管理"页面注册新设备
3. 在"域管理"页面创建域
4. 在"跨域认证"页面测试跨域认证功能
5. 在"认证日志"页面查看认证历史

