#!/bin/bash

# 启动脚本
# 用于快速启动所有服务

set -e

echo "========================================="
echo "  物联网设备跨域身份认证系统启动脚本"
echo "========================================="

# 检查依赖
check_dependencies() {
    echo "检查依赖..."
    
    if ! command -v go &> /dev/null; then
        echo "错误: 未安装Go，请先安装Go 1.19+"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        echo "错误: 未安装Node.js，请先安装Node.js 16+"
        exit 1
    fi
    
    if ! command -v psql &> /dev/null; then
        echo "警告: 未找到psql，请确保PostgreSQL已安装"
    fi
    
    echo "依赖检查完成"
}

# 检查服务是否运行
check_service() {
    local port=$1
    local name=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        echo "警告: $name 已在端口 $port 运行"
        return 1
    fi
    return 0
}

# 启动PostgreSQL（如果使用Docker）
start_postgres() {
    if ! docker ps | grep -q nono-postgres; then
        echo "启动PostgreSQL容器..."
        docker run -d --name nono-postgres \
          -e POSTGRES_DB=nono_system \
          -e POSTGRES_USER=nono \
          -e POSTGRES_PASSWORD=nono123 \
          -p 5432:5432 \
          postgres:15-alpine || true
        sleep 3
    fi
}

# 启动Redis（如果使用Docker）
start_redis() {
    if ! docker ps | grep -q nono-redis; then
        echo "启动Redis容器..."
        docker run -d --name nono-redis \
          -p 6379:6379 \
          redis:7-alpine || true
        sleep 2
    fi
}

# 主函数
main() {
    check_dependencies
    
    # 检查端口占用
    check_service 8080 "后端服务"
    check_service 9000 "预言机服务"
    check_service 3000 "前端服务"
    
    # 启动依赖服务
    if command -v docker &> /dev/null; then
        start_postgres
        start_redis
    fi
    
    echo ""
    echo "准备启动服务..."
    echo "请在新终端窗口中运行以下命令："
    echo ""
    echo "终端1 - 预言机服务:"
    echo "  cd oracle && go run ./cmd/oracle"
    echo ""
    echo "终端2 - 后端API服务:"
    echo "  cd backend && go run ./cmd/server"
    echo ""
    echo "终端3 - 前端应用:"
    echo "  cd frontend && npm run dev"
    echo ""
    echo "或使用 make 命令:"
    echo "  make oracle    # 启动预言机"
    echo "  make backend   # 启动后端"
    echo ""
}

main

