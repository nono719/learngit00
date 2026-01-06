# 跨链认证验证指南

本文档说明如何确认跨域认证是否在联盟链上成功执行。

## 验证方法

### 方法一：通过API接口验证

#### 1. 查看认证记录（包含交易哈希）

```bash
# 获取设备的认证记录
curl http://localhost:8080/api/v1/auth/records/{device_did}
```

响应示例：
```json
[
  {
    "id": 1,
    "device_did": "did:example:device123",
    "source_domain": "domain1",
    "target_domain": "domain2",
    "authorized": true,
    "tx_hash": "0x1234567890abcdef...",
    "timestamp": "2025-01-02T10:30:00Z"
  }
]
```

**关键字段说明：**
- `tx_hash`: 区块链交易哈希，如果为空字符串，说明未调用区块链
- `authorized`: 认证结果（true/false）

#### 2. 验证交易状态

使用交易哈希查询区块链上的交易详情：

```bash
# 验证交易
curl http://localhost:8080/api/v1/auth/verify/{tx_hash}
```

响应示例：
```json
{
  "tx_hash": "0x1234567890abcdef...",
  "status": true,
  "block_number": 12345,
  "gas_used": 100000,
  "events": [
    {
      "address": "0xContractAddress...",
      "topics": ["0xEventSignature..."],
      "data": "0x...",
      "block_number": 12345,
      "tx_index": 0
    }
  ],
  "confirmations": 1
}
```

**关键字段说明：**
- `status`: true 表示交易成功，false 表示交易失败
- `block_number`: 交易所在的区块号
- `events`: 交易触发的事件列表（包含 `CrossDomainAuthCompleted` 事件）

### 方法二：直接查询区块链

#### 1. 使用 geth 控制台

```bash
# 连接到区块链节点
geth attach http://localhost:8545

# 查询交易收据
eth.getTransactionReceipt("0x交易哈希")

# 查询交易详情
eth.getTransaction("0x交易哈希")

# 查询区块信息
eth.getBlock(区块号, true)  # true 表示包含交易详情
```

#### 2. 使用 Web3.js 或 ethers.js

```javascript
// 使用 ethers.js 示例
const ethers = require('ethers');
const provider = new ethers.providers.JsonRpcProvider('http://localhost:8545');

// 获取交易收据
const receipt = await provider.getTransactionReceipt('0x交易哈希');
console.log('交易状态:', receipt.status === 1 ? '成功' : '失败');
console.log('区块号:', receipt.blockNumber);
console.log('Gas使用:', receipt.gasUsed.toString());

// 解析事件
const contract = new ethers.Contract(contractAddress, abi, provider);
const filter = contract.filters.CrossDomainAuthCompleted();
const events = await contract.queryFilter(filter, fromBlock, toBlock);
```

### 方法三：查看后端日志

后端服务在调用区块链时会输出日志：

```
Cross-domain auth transaction submitted: txHash=0x1234..., authorized=true
```

如果区块链未连接，会看到：
```
Blockchain not connected, using local device status
```

### 方法四：查询数据库

```sql
-- 查看认证记录（包含交易哈希）
SELECT id, device_did, source_domain, target_domain, authorized, tx_hash, timestamp 
FROM auth_records 
WHERE device_did = 'did:example:device123'
ORDER BY timestamp DESC;

-- 查看认证日志
SELECT * FROM auth_logs 
WHERE device_did = 'did:example:device123'
ORDER BY created_at DESC;
```

**判断标准：**
- `tx_hash` 字段不为空：说明已调用区块链
- `tx_hash` 字段为空：说明未调用区块链（可能是区块链未连接或配置错误）

## 验证步骤

### 完整验证流程

1. **发起跨域认证请求**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/cross-domain \
     -H "Content-Type: application/json" \
     -d '{
       "device_did": "did:example:device123",
       "source_domain": "domain1",
       "target_domain": "domain2"
     }'
   ```

2. **检查响应中的交易哈希**
   响应应包含 `tx_hash` 字段：
   ```json
   {
     "authorized": true,
     "record_id": 1,
     "tx_hash": "0x1234567890abcdef..."
   }
   ```

3. **验证交易状态**
   ```bash
   curl http://localhost:8080/api/v1/auth/verify/0x1234567890abcdef...
   ```

4. **确认交易已上链**
   - `status` 为 `true`
   - `block_number` 不为 0
   - `events` 中包含相关事件

## 常见问题

### Q1: 响应中没有 `tx_hash` 字段

**原因：**
- 区块链客户端未初始化
- 区块链节点未连接
- 配置文件中未设置 `private_key`

**解决方法：**
1. 检查后端配置 `backend/config/config.yaml`：
   ```yaml
   blockchain:
     rpc_url: "http://localhost:8545"
     contract_addr: "0x..."  # 确保已部署合约
     private_key: "your_private_key_without_0x"  # 必须设置
   ```

2. 检查后端日志，确认区块链客户端是否成功初始化

### Q2: 交易哈希存在但验证失败

**可能原因：**
- 交易还在待处理队列中（pending）
- 交易被拒绝（reverted）
- 区块链节点不同步

**解决方法：**
1. 等待一段时间后再次验证
2. 检查区块链节点的同步状态
3. 查看交易详情，确认失败原因

### Q3: 如何确认事件已正确触发

**方法：**
1. 使用验证接口查看 `events` 字段
2. 查找 `CrossDomainAuthCompleted` 事件
3. 解析事件数据，确认 `authorized` 字段的值

## 测试脚本

创建一个测试脚本来验证跨链认证：

```bash
#!/bin/bash

# 配置
API_URL="http://localhost:8080/api/v1"
DEVICE_DID="did:example:test123"
SOURCE_DOMAIN="domain1"
TARGET_DOMAIN="domain2"

# 发起跨域认证
echo "发起跨域认证请求..."
RESPONSE=$(curl -s -X POST "$API_URL/auth/cross-domain" \
  -H "Content-Type: application/json" \
  -d "{
    \"device_did\": \"$DEVICE_DID\",
    \"source_domain\": \"$SOURCE_DOMAIN\",
    \"target_domain\": \"$TARGET_DOMAIN\"
  }")

echo "响应: $RESPONSE"

# 提取交易哈希
TX_HASH=$(echo $RESPONSE | jq -r '.tx_hash')

if [ "$TX_HASH" != "null" ] && [ "$TX_HASH" != "" ]; then
  echo "交易哈希: $TX_HASH"
  
  # 等待交易确认
  sleep 3
  
  # 验证交易
  echo "验证交易..."
  VERIFY_RESPONSE=$(curl -s "$API_URL/auth/verify/$TX_HASH")
  echo "验证结果: $VERIFY_RESPONSE"
  
  STATUS=$(echo $VERIFY_RESPONSE | jq -r '.status')
  if [ "$STATUS" == "true" ]; then
    echo "✅ 跨链认证成功！"
  else
    echo "❌ 跨链认证失败"
  fi
else
  echo "⚠️  未生成交易哈希，可能区块链未连接"
fi
```

## 总结

确认跨链认证成功的标志：

1. ✅ API响应中包含 `tx_hash` 字段
2. ✅ 数据库 `auth_records` 表中 `tx_hash` 字段不为空
3. ✅ 使用验证接口查询交易，`status` 为 `true`
4. ✅ 交易已包含在区块中（`block_number` 不为 0）
5. ✅ 事件列表中包含 `CrossDomainAuthCompleted` 事件

如果以上条件都满足，说明跨链认证已成功在联盟链上执行。

