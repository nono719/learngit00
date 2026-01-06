# Ganache 配置和使用指南

本文档说明如何在 Ganache 上部署智能合约并查看跨链认证记录。

## 1. 启动 Ganache

### 步骤 1：打开 Ganache

1. 打开 Ganache 应用程序
2. 点击 "New Workspace" 或 "Quickstart"
3. 如果是新工作区，可以设置：
   - **Workspace Name**: `nono-system`
   - **Network ID**: `5777` (默认)
   - **Port**: `8545` (默认)
   - **Hostname**: `127.0.0.1` 或 `localhost`

### 步骤 2：记录账户信息

启动后，Ganache 会显示：
- 10 个测试账户（每个账户有 100 ETH）
- 每个账户的私钥
- RPC Server 地址（通常是 `http://127.0.0.1:8545`）

**重要：** 复制第一个账户的私钥，稍后需要在后端配置中使用。

## 2. 部署智能合约

### 方法一：使用 Remix IDE（推荐）

1. **打开 Remix IDE**
   - 访问 https://remix.ethereum.org
   - 或使用本地 Remix 应用

2. **创建合约文件**
   - 在 `contracts` 文件夹中创建 `DeviceIdentity.sol`
   - 复制项目中的 `contracts/DeviceIdentity.sol` 内容

3. **编译合约**
   - 选择 Solidity 编译器版本（建议 0.8.0 或更高）
   - 点击 "Compile DeviceIdentity.sol"

4. **部署到 Ganache**
   - 切换到 "Deploy & Run Transactions" 标签
   - **Environment**: 选择 "Injected Provider" 或 "Web3 Provider"
   - **Web3 Provider Endpoint**: 输入 `http://127.0.0.1:8545`
   - 点击 "Deploy"
   - **复制合约地址**（例如：`0x1234...`）

### 方法二：使用 Truffle（可选）

如果需要使用 Truffle，需要先安装：

```bash
npm install -g truffle
```

然后创建 Truffle 项目并配置。

## 3. 配置后端连接 Ganache

### 步骤 1：更新后端配置文件

编辑 `backend/config/config.yaml`：

```yaml
blockchain:
  rpc_url: "http://127.0.0.1:8545"  # Ganache 的 RPC 地址
  chain_id: 5777  # Ganache 默认的链 ID
  contract_addr: "0x..."  # 替换为部署的合约地址
  private_key: "your_private_key_here"  # Ganache 第一个账户的私钥（不含0x前缀）
```

**重要提示：**
- `private_key` 是 Ganache 中第一个账户的私钥
- 私钥格式：直接复制 Ganache 显示的私钥，去掉 `0x` 前缀（如果有）
- 例如：如果 Ganache 显示 `0x1234abcd...`，配置时使用 `1234abcd...`

### 步骤 2：重启后端服务

```bash
cd /Users/chenminggang/毕业设计/nono-system/backend
go run ./cmd/server
```

启动时应该看到：
```
Blockchain client initialized successfully
```

## 4. 在 Ganache 上查看交易记录

### 方法一：使用 Ganache 界面

1. **查看交易列表**
   - 在 Ganache 主界面，点击 "Transactions" 标签
   - 可以看到所有交易记录
   - 每笔交易显示：
     - Transaction Hash（交易哈希）
     - From（发送方地址）
     - To（接收方地址，合约地址）
     - Value（交易金额）
     - Gas Used（Gas 使用量）

2. **查看交易详情**
   - 点击任意交易，查看详细信息
   - 可以看到：
     - Input Data（调用数据，包含函数名和参数）
     - Events（事件日志）

3. **查看事件日志**
   - 在交易详情中，查找 "Events" 部分
   - 应该能看到 `CrossDomainAuthRequested` 和 `CrossDomainAuthCompleted` 事件
   - 事件包含：
     - `did`: 设备 DID
     - `sourceDomain`: 源域
     - `targetDomain`: 目标域
     - `authorized`: 是否授权

### 方法二：使用后端 API 验证

1. **发起跨域认证**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/cross-domain \
     -H "Content-Type: application/json" \
     -d '{
       "device_did": "1234124",
       "source_domain": "设备",
       "target_domain": "能源"
     }'
   ```

2. **获取交易哈希**
   响应中会包含 `tx_hash` 字段：
   ```json
   {
     "authorized": true,
     "record_id": 1,
     "tx_hash": "0x1234567890abcdef..."
   }
   ```

3. **在 Ganache 中查找交易**
   - 复制 `tx_hash`
   - 在 Ganache 的 "Transactions" 标签中搜索
   - 或使用 "Search" 功能查找

4. **验证交易**
   ```bash
   curl http://localhost:8080/api/v1/auth/verify/0x1234567890abcdef...
   ```

## 5. 查看合约状态

### 使用 Ganache 的合约交互功能

1. **添加合约**
   - 在 Ganache 中，找到 "Contracts" 标签
   - 点击 "Add Contract"
   - 输入合约地址和 ABI

2. **查询设备信息**
   - 在合约界面，可以调用 `getDevice` 函数
   - 输入设备 DID，查看设备状态

### 使用 Remix IDE

1. **连接到 Ganache**
   - 在 Remix 的 "Deploy & Run" 标签
   - 选择 "At Address"
   - 输入合约地址

2. **调用合约函数**
   - 可以调用 `getDevice` 查询设备信息
   - 可以查看 `authRecords` 映射，查看认证记录

## 6. 常见问题

### Q1: 后端启动时显示 "Blockchain client not initialized"

**原因：**
- Ganache 未启动
- RPC URL 配置错误
- 合约地址或私钥配置错误

**解决方法：**
1. 确认 Ganache 正在运行
2. 检查 `rpc_url` 是否为 `http://127.0.0.1:8545`
3. 确认合约地址正确
4. 确认私钥格式正确（不含 0x 前缀）

### Q2: 交易失败或 Gas 不足

**原因：**
- Gas Limit 设置过低
- 账户余额不足

**解决方法：**
1. Ganache 账户默认有 100 ETH，应该足够
2. 如果还是失败，检查 Gas Limit 设置（代码中设置为 200000）

### Q3: 看不到事件日志

**原因：**
- 交易可能还在 pending 状态
- 事件解析失败

**解决方法：**
1. 等待几秒钟让交易确认
2. 在 Ganache 中查看交易的 "Logs" 部分
3. 使用后端 API 验证交易状态

### Q4: 如何查看历史交易

**方法：**
1. 在 Ganache 的 "Transactions" 标签中查看所有交易
2. 使用后端 API 查询认证记录：
   ```bash
   curl http://localhost:8080/api/v1/auth/records/1234124
   ```
3. 使用交易哈希查询：
   ```bash
   curl http://localhost:8080/api/v1/auth/verify/{tx_hash}
   ```

## 7. 验证跨链认证成功的标志

在 Ganache 上确认跨链认证成功的标志：

1. ✅ **交易出现在 Ganache 交易列表中**
   - 状态为 "Success"
   - To 地址是合约地址

2. ✅ **事件日志存在**
   - 包含 `CrossDomainAuthRequested` 事件
   - 包含 `CrossDomainAuthCompleted` 事件
   - 事件数据中包含正确的 DID 和域信息

3. ✅ **后端 API 返回交易哈希**
   - 响应中包含 `tx_hash` 字段
   - 验证接口返回 `status: true`

4. ✅ **数据库记录包含交易哈希**
   - `auth_records` 表中的 `tx_hash` 字段不为空

## 8. 测试脚本

创建一个完整的测试流程：

```bash
#!/bin/bash

# 1. 发起跨域认证
echo "发起跨域认证..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/cross-domain \
  -H "Content-Type: application/json" \
  -d '{
    "device_did": "1234124",
    "source_domain": "设备",
    "target_domain": "能源"
  }')

echo "响应: $RESPONSE"

# 2. 提取交易哈希
TX_HASH=$(echo $RESPONSE | grep -o '"tx_hash":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TX_HASH" ]; then
  echo "交易哈希: $TX_HASH"
  echo "请在 Ganache 中搜索此交易哈希查看详情"
  
  # 3. 验证交易
  sleep 2
  echo "验证交易..."
  curl -s http://localhost:8080/api/v1/auth/verify/$TX_HASH | jq .
else
  echo "未生成交易哈希，可能区块链未连接"
fi
```

## 总结

通过以上步骤，你可以在 Ganache 上：
- ✅ 查看所有跨链认证交易
- ✅ 查看交易详情和事件日志
- ✅ 验证交易是否成功上链
- ✅ 查询合约状态和设备信息

这样就可以完整地验证跨链认证功能是否正常工作。

