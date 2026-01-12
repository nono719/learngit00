#!/bin/bash

# 初始化管理员账号脚本
# 用法: ./scripts/init_admin.sh [username] [password]

API_URL="${API_URL:-http://localhost:8080/api/v1}"
USERNAME="${1:-admin}"
PASSWORD="${2:-admin123}"
EMAIL="${3:-admin@example.com}"

echo "=========================================="
echo "初始化管理员账号"
echo "=========================================="
echo "API地址: $API_URL"
echo "用户名: $USERNAME"
echo "密码: $PASSWORD"
echo ""

# 检查后端服务是否运行
if ! curl -s "$API_URL/../health" > /dev/null 2>&1; then
  echo "❌ 后端服务未运行，请先启动后端服务"
  echo "   执行: cd backend && go run ./cmd/server"
  exit 1
fi

echo "步骤1: 注册管理员账号..."
RESPONSE=$(curl -s -X POST "$API_URL/users/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"email\": \"$EMAIL\",
    \"role\": \"admin\"
  }")

if echo "$RESPONSE" | grep -q "error"; then
  if echo "$RESPONSE" | grep -q "already exists"; then
    echo "⚠️  用户已存在，尝试登录..."
  else
    echo "❌ 注册失败: $RESPONSE"
    exit 1
  fi
else
  echo "✅ 管理员账号创建成功"
  echo ""
fi

echo "步骤2: 登录获取 token..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "error"; then
  echo "❌ 登录失败: $LOGIN_RESPONSE"
  exit 1
fi

# 提取 token（简化处理）
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":[^,}]*' | cut -d':' -f2 | tr -d ' "')

if [ -z "$TOKEN" ]; then
  echo "❌ 无法获取 token"
  echo "响应: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ 登录成功"
echo ""
echo "=========================================="
echo "登录信息"
echo "=========================================="
echo "Token: $TOKEN"
echo "用户名: $USERNAME"
echo "密码: $PASSWORD"
echo ""
echo "使用方法："
echo "1. 在前端登录页面使用以上账号登录"
echo "2. 或使用 API 时添加请求头："
echo "   Authorization: Bearer $TOKEN"
echo ""
echo "示例："
echo "curl http://localhost:8080/api/v1/devices \\"
echo "  -H \"Authorization: Bearer $TOKEN\""
echo ""

