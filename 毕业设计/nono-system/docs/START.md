# 系统启动指南

## 启动方式概览

系统提供了三种启动方式：
1. **手动启动** - 适合开发和调试
2. **使用Makefile** - 推荐方式，简单快捷
3. **使用Docker Compose** - 适合生产环境

## 方式一：手动启动（推荐用于开发）

### 步骤1：启动依赖服务

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

### 步骤2：配置服务

```bash
# 复制配置文件
cp config/oracle.config.example.yaml oracle/config/config.yaml
cp config/backend.config.example.yaml backend/config/config.yaml

# 编辑配置文件，更新：
# - 区块链RPC地址
# - 合约地址（部署合约后）
# - 数据库连接信息（如需要）
```

### 步骤3：初始化数据库

```bash
# 创建数据库（如果尚未创建）
psql -U postgres -c "CREATE DATABASE nono_system;"
psql -U postgres -c "CREATE USER nono WITH PASSWORD 'nono123';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE nono_system TO nono;"
```

### 步骤4：启动服务（需要3个终端窗口）

**终端1 - 预言机服务：**
```bash
cd oracle
go mod download  # 首次运行需要
go run ./cmd/oracle
```

**终端2 - 后端API服务：**
```bash
cd backend
go mod download  # 首次运行需要
go run ./cmd/server
```

**终端3 - 前端应用：**
```bash
cd frontend
npm install  # 首次运行需要
npm run dev
```

## 方式二：使用Makefile（推荐）

### 安装依赖

```bash
# 在项目根目录
go mod download
cd oracle && go mod download
cd ../backend && go mod download
cd ../frontend && npm install
```

### 启动服务

```bash
# 终端1 - 预言机服务
make oracle

# 终端2 - 后端API服务
make backend

# 终端3 - 前端应用
cd frontend && npm run dev
```

## 方式三：使用Docker Compose

### 启动所有服务

```bash
# 在项目根目录
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 查看服务状态

```bash
docker-compose ps
```

## 验证系统运行

### 1. 检查服务健康状态

```bash
# 后端服务
curl http://localhost:8080/health

# 预言机服务
curl http://localhost:9000/api/v1/health
```

### 2. 访问前端界面

打开浏览器访问：http://localhost:3000

### 3. 测试API

```bash
# 创建域
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{"name": "smart_home", "description": "智能家居域"}'

# 注册设备
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "did": "did:example:device001",
    "device_id": "device001",
    "device_type": "sensor",
    "domain": "smart_home"
  }'

# 查询设备
curl http://localhost:8080/api/v1/devices/did:example:device001
```

## 服务端口说明

- **前端应用**: 3000
- **后端API**: 8080
- **预言机服务**: 9000
- **PostgreSQL**: 5432
- **Redis**: 6379
- **区块链节点**: 8545

## 常见问题排查

### 问题1：端口被占用

```bash
# 查找占用端口的进程
lsof -i :8080
lsof -i :9000
lsof -i :3000

# 杀死进程
kill -9 <PID>
```

### 问题2：数据库连接失败

```bash
# 检查PostgreSQL是否运行
docker ps | grep postgres
# 或
pg_isready

# 检查连接
psql -U nono -d nono_system -h localhost
```

### 问题3：Go模块下载失败

```bash
# 设置Go代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 清理模块缓存
go clean -modcache
go mod download
```

### 问题4：前端依赖安装失败

```bash
# 使用国内镜像
npm config set registry https://registry.npmmirror.com

# 清理缓存
npm cache clean --force
npm install
```

## 停止服务

### 手动启动的服务

按 `Ctrl+C` 停止各个终端中的服务

### Docker Compose启动的服务

```bash
docker-compose down
```

## 下一步

- 查看 [API文档](API.md) 了解API接口
- 查看 [开发文档](DEVELOPMENT.md) 了解开发指南
- 查看 [部署文档](DEPLOY.md) 了解生产环境部署

