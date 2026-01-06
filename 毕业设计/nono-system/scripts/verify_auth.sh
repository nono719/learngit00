#!/bin/bash

# 跨链认证验证脚本
# 用法: ./scripts/verify_auth.sh [device_did] [source_domain] [target_domain]

API_URL="${API_URL:-http://localhost:8080/api/v1}"
DEVICE_DID="${1:-did:example:test123}"
SOURCE_DOMAIN="${2:-domain1}"
TARGET_DOMAIN="${3:-domain2}"

echo "=========================================="
echo "跨链认证验证脚本"
echo "=========================================="
echo "API地址: $API_URL"
echo "设备DID: $DEVICE_DID"
echo "源域: $SOURCE_DOMAIN"
echo "目标域: $TARGET_DOMAIN"
echo ""

# 1. 发起跨域认证请求
echo "步骤1: 发起跨域认证请求..."
RESPONSE=$(curl -s -X POST "$API_URL/auth/cross-domain" \
  -H "Content-Type: application/json" \
  -d "{
    \"device_did\": \"$DEVICE_DID\",
    \"source_domain\": \"$SOURCE_DOMAIN\",
    \"target_domain\": \"$TARGET_DOMAIN\"
  }")

if [ $? -ne 0 ]; then
  echo "❌ 请求失败，请检查后端服务是否运行"
  exit 1
fi

echo "响应: $RESPONSE"
echo ""

# 提取字段
AUTHORIZED=$(echo "$RESPONSE" | grep -o '"authorized":[^,}]*' | cut -d':' -f2 | tr -d ' ')
RECORD_ID=$(echo "$RESPONSE" | grep -o '"record_id":[^,}]*' | cut -d':' -f2 | tr -d ' ')
TX_HASH=$(echo "$RESPONSE" | grep -o '"tx_hash":"[^"]*"' | cut -d'"' -f4)

# 2. 检查交易哈希
echo "步骤2: 检查交易哈希..."
if [ -z "$TX_HASH" ] || [ "$TX_HASH" == "null" ]; then
  echo "⚠️  未生成交易哈希"
  echo "可能原因:"
  echo "  - 区块链客户端未初始化"
  echo "  - 区块链节点未连接"
  echo "  - 配置文件中未设置 private_key"
  echo ""
  echo "认证结果: $AUTHORIZED (仅基于本地状态)"
  exit 0
fi

echo "✅ 交易哈希: $TX_HASH"
echo ""

# 3. 等待交易确认
echo "步骤3: 等待交易确认（3秒）..."
sleep 3

# 4. 验证交易
echo "步骤4: 验证交易状态..."
VERIFY_RESPONSE=$(curl -s "$API_URL/auth/verify/$TX_HASH")

if [ $? -ne 0 ]; then
  echo "❌ 验证请求失败"
  exit 1
fi

echo "验证响应: $VERIFY_RESPONSE"
echo ""

# 提取验证结果
STATUS=$(echo "$VERIFY_RESPONSE" | grep -o '"status":[^,}]*' | cut -d':' -f2 | tr -d ' ')
BLOCK_NUMBER=$(echo "$VERIFY_RESPONSE" | grep -o '"block_number":[^,}]*' | cut -d':' -f2 | tr -d ' ')

# 5. 显示结果
echo "=========================================="
echo "验证结果"
echo "=========================================="
echo "交易哈希: $TX_HASH"
echo "认证结果: $AUTHORIZED"
echo "交易状态: $STATUS"
echo "区块号: $BLOCK_NUMBER"
echo ""

if [ "$STATUS" == "true" ] || [ "$STATUS" == "1" ]; then
  echo "✅ 跨链认证成功！"
  echo "   - 交易已成功上链"
  echo "   - 区块号: $BLOCK_NUMBER"
  echo "   - 认证结果: $AUTHORIZED"
else
  echo "❌ 跨链认证失败"
  echo "   - 交易状态: $STATUS"
fi

# 6. 查询认证记录
echo ""
echo "步骤5: 查询认证记录..."
RECORDS=$(curl -s "$API_URL/auth/records/$DEVICE_DID")
echo "认证记录: $RECORDS"

