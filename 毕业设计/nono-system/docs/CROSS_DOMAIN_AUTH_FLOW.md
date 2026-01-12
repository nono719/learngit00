# 跨域认证完整流程

## 流程概述

完整的跨域认证流程包括以下步骤：

1. **设备身份注册** → 
2. **预言机报告设备状态** → 
3. **设备在另一个域发起访问请求** → 
4. **目标域向区块链验证该设备身份与状态** → 
5. **完成跨域认证并授权访问**

## 系统功能完整性检查

### ✅ 1. 设备身份注册

**实现状态：完全支持**

#### 前端功能
- **位置**：`frontend/src/views/Devices.vue`
- **功能**：
  - 设备注册到数据库
  - 可选：同时注册到区块链（通过 `blockchainService.registerDevice`）
  - 支持输入设备元数据（JSON格式）

#### 后端功能
- **位置**：`backend/internal/handlers/device.go` → `RegisterDevice`
- **功能**：
  - 接收设备注册请求
  - 保存设备信息到数据库
  - 记录设备状态为 `active`

#### 区块链功能
- **智能合约**：`contracts/DeviceIdentity.sol` → `registerDevice`
- **功能**：
  - 验证设备DID唯一性
  - 存储设备信息（DID、元数据、状态、所有者、时间戳）
  - 触发 `DeviceRegistered` 事件
  - 要求：调用者必须是授权管理员（`onlyAuthorizedAdmin`）

#### 前端区块链服务
- **位置**：`frontend/src/services/blockchain.js` → `registerDevice`
- **功能**：
  - 连接Ganache或MetaMask
  - 调用智能合约 `registerDevice` 函数
  - 等待交易确认
  - 解析并返回交易结果

---

### ✅ 2. 预言机报告设备状态

**实现状态：完全支持**

#### 预言机服务
- **位置**：`oracle/internal/oracle/oracle.go`
- **核心功能**：
  1. **数据采集**（`collectAndUpdate`）：
     - 从所有配置的数据源采集设备状态
     - 数据源类型：`monitoring`、`certificate`、`api`
     - 支持API Key认证
  
  2. **多数投票机制**（`majorityVote`）：
     - 统计每个设备在不同数据源中的状态
     - 选择出现次数最多的状态作为共识状态
     - 检查是否达到最小共识数量（`min_consensus`）
  
  3. **区块链更新**（`updateBlockchain`）：
     - 将共识状态转换为智能合约枚举值（0=Active, 1=Suspicious, 2=Revoked）
     - 调用智能合约 `updateDeviceStatus` 函数
     - 要求：预言机账户必须是授权预言机（`onlyAuthorizedOracle`）

#### 智能合约
- **位置**：`contracts/DeviceIdentity.sol` → `updateDeviceStatus`
- **功能**：
  - 验证设备存在
  - 更新设备状态
  - 更新 `lastUpdated` 时间戳
  - 触发 `DeviceStatusUpdated` 事件
  - 要求：调用者必须是授权预言机（`onlyAuthorizedOracle`）

#### 数据源配置
- **位置**：`oracle/config/config.yaml`
- **配置项**：
  - 数据源URL（如：`http://localhost:8080/api/v1/devices/status`）
  - API Key（用于后端认证）
  - 数据源类型和启用状态

#### 后端API
- **位置**：`backend/internal/handlers/device.go` → `GetDeviceStatuses`
- **功能**：
  - 返回所有设备的状态列表
  - 支持API Key认证（Bearer Token）
  - 格式化为预言机所需的数据格式

---

### ✅ 3. 设备在另一个域发起访问请求

**实现状态：完全支持**

#### 前端功能
- **位置**：`frontend/src/views/Auth.vue`
- **功能**：
  - **后端API方式**：
    - 输入设备DID、源域、目标域
    - 调用后端API `/api/v1/auth/request`
    - 显示认证结果和交易哈希
  
  - **前端直接上链**：
    - 连接区块链（Ganache/MetaMask）
    - 调用智能合约 `requestCrossDomainAuth` 函数
    - 等待交易确认
    - 同步交易哈希到后端数据库
  
  - **验证检查**：
    - 检查设备是否存在（数据库和区块链）
    - 检查设备状态是否为 `active`
    - 验证源域是否与设备所属域匹配
    - 验证目标域是否存在

#### 后端功能
- **位置**：`backend/internal/handlers/auth.go` → `RequestCrossDomainAuth`
- **功能**：
  - 验证设备存在和状态
  - 验证源域匹配
  - 验证目标域存在
  - 调用区块链合约（如果连接）
  - 记录认证记录和日志

#### 智能合约
- **位置**：`contracts/DeviceIdentity.sol` → `requestCrossDomainAuth`
- **功能**：
  - 验证设备存在（`devices[_did].exists`）
  - 验证设备状态为 `Active`（`devices[_did].status == DeviceStatus.Active`）
  - 创建跨域认证记录
  - 触发 `CrossDomainAuthRequested` 和 `CrossDomainAuthCompleted` 事件
  - 返回授权结果（`bool authorized`）

---

### ✅ 4. 目标域向区块链验证该设备身份与状态

**实现状态：完全支持**

#### 区块链查询功能
- **位置**：`backend/internal/blockchain/client.go` → `GetDeviceStatus`
- **功能**：
  - 调用智能合约 `devices` 映射查询设备信息
  - 返回设备状态（0=Active, 1=Suspicious, 2=Revoked）
  - 可以查询设备是否存在、状态、元数据、所有者、注册时间等

#### 智能合约查询
- **位置**：`contracts/DeviceIdentity.sol` → `devices` 映射（public）
- **功能**：
  - 公开映射，任何人都可以查询
  - 返回完整的设备信息结构体：
    - `did`：设备DID
    - `metadata`：设备元数据
    - `status`：设备状态（枚举）
    - `owner`：设备所有者地址
    - `registeredAt`：注册时间戳
    - `lastUpdated`：最后更新时间戳
    - `exists`：是否存在

#### 交易验证功能
- **位置**：`backend/internal/handlers/auth.go` → `VerifyTransaction`
- **功能**：
  - 根据交易哈希查询交易收据
  - 解析交易中的事件（`CrossDomainAuthCompleted`）
  - 返回认证结果和详细信息

#### 前端验证功能
- **位置**：`frontend/src/views/Verify.vue`
- **功能**：
  - 根据设备DID查询认证记录
  - 根据交易哈希验证交易
  - 显示设备状态（从数据库和区块链）
  - 显示认证记录详情

---

### ✅ 5. 完成跨域认证并授权访问

**实现状态：完全支持**

#### 认证结果处理
- **智能合约返回**：
  - `requestCrossDomainAuth` 函数返回 `bool authorized`
  - 触发 `CrossDomainAuthCompleted` 事件，包含授权结果

- **后端处理**：
  - 解析区块链交易结果或事件
  - 记录认证记录到数据库（`AuthRecord`）
  - 记录认证日志（`AuthLog`）
  - 返回认证结果给前端

- **前端显示**：
  - 显示认证成功/失败
  - 显示交易哈希（如果上链）
  - 显示认证记录ID
  - 保存到交易历史（localStorage）

#### 授权访问控制
- **设备状态检查**：
  - 只有 `active` 状态的设备才能通过认证
  - `suspicious` 和 `revoked` 状态的设备会被拒绝

- **域验证**：
  - 源域必须与设备所属域匹配
  - 目标域必须存在于系统中

- **权限控制**：
  - 操作人员只能发起自己域设备的跨域认证
  - 管理员可以发起任何域的跨域认证

---

## 完整流程示例

### 场景：设备从域A访问域B

1. **设备注册**（在域A）
   ```
   前端 → 后端API → 数据库
   前端 → 区块链服务 → 智能合约 → 区块链
   ```

2. **预言机报告状态**（定期执行）
   ```
   预言机 → 数据源API → 后端 `/api/v1/devices/status`
   预言机 → 多数投票 → 共识状态
   预言机 → 智能合约 `updateDeviceStatus` → 区块链
   ```

3. **设备发起跨域认证请求**（从域A到域B）
   ```
   前端（域A） → 后端API `/api/v1/auth/request`
   后端 → 验证设备存在和状态
   后端 → 智能合约 `requestCrossDomainAuth`
   智能合约 → 验证设备状态为 Active
   智能合约 → 返回授权结果
   后端 → 记录认证记录和日志
   前端 → 显示认证结果
   ```

4. **目标域验证**（域B验证设备）
   ```
   目标域后端 → 区块链客户端 `GetDeviceStatus`
   区块链客户端 → 智能合约 `devices` 映射
   智能合约 → 返回设备信息（包括状态）
   目标域后端 → 验证设备状态为 Active
   目标域后端 → 授权访问
   ```

5. **完成认证并授权**
   ```
   认证记录保存到数据库
   认证日志记录
   前端显示认证成功
   设备获得访问权限
   ```

---

## 系统功能完整性总结

| 流程步骤 | 实现状态 | 说明 |
|---------|---------|------|
| 1. 设备身份注册 | ✅ 完全支持 | 支持数据库和区块链双重注册 |
| 2. 预言机报告设备状态 | ✅ 完全支持 | 多数据源采集、多数投票、区块链更新 |
| 3. 设备发起访问请求 | ✅ 完全支持 | 支持后端API和前端直接上链两种方式 |
| 4. 目标域验证设备 | ✅ 完全支持 | 区块链查询设备状态和身份 |
| 5. 完成认证并授权 | ✅ 完全支持 | 完整的认证记录和日志系统 |

## 结论

**✅ 系统完全支持完整的跨域认证流程！**

所有5个步骤都已实现，包括：
- 设备身份注册（数据库+区块链）
- 预言机状态报告（多数据源+多数投票+区块链更新）
- 跨域认证请求（前端+后端+智能合约）
- 区块链验证（设备状态和身份查询）
- 认证授权（完整的记录和日志系统）

系统设计遵循了去中心化身份管理的原则，使用区块链作为可信的验证源，同时保持了良好的用户体验和系统可扩展性。
