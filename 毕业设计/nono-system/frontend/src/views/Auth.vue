<template>
  <div class="auth-page">
    <!-- 区块链连接配置 -->
    <el-card style="margin-bottom: 20px">
      <template #header>
        <div class="card-header">
          <span>区块链连接配置</span>
          <el-tag v-if="blockchainConnected" type="success" effect="dark">
            <el-icon><CircleCheck /></el-icon>
            已连接
          </el-tag>
          <el-tag v-else type="info" effect="dark">
            <el-icon><CircleClose /></el-icon>
            未连接
          </el-tag>
        </div>
      </template>

      <el-tabs v-model="connectionTab" type="border-card">
        <!-- Ganache 连接 -->
        <el-tab-pane label="Ganache 连接" name="ganache">
          <el-form :model="ganacheConfig" label-width="140px" style="max-width: 700px">
            <el-form-item label="RPC URL">
              <el-input 
                v-model="ganacheConfig.rpcUrl" 
                placeholder="http://127.0.0.1:8545"
              />
            </el-form-item>
            <el-form-item label="合约地址" required>
              <el-input 
                v-model="ganacheConfig.contractAddress" 
                placeholder="0x..."
              />
            </el-form-item>
            <el-form-item label="私钥" required>
              <el-input 
                v-model="ganacheConfig.privateKey" 
                type="password"
                placeholder="不含0x前缀的私钥"
                show-password
              />
              <div style="font-size: 12px; color: #909399; margin-top: 5px">
                从 Ganache 复制账户私钥（去掉 0x 前缀）
              </div>
            </el-form-item>
            <el-form-item>
              <el-button 
                type="primary" 
                @click="connectGanache" 
                :loading="connecting"
                :disabled="blockchainConnected"
              >
                连接 Ganache
              </el-button>
              <el-button 
                v-if="blockchainConnected" 
                @click="disconnectBlockchain"
              >
                断开连接
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- MetaMask 连接 -->
        <el-tab-pane label="MetaMask 连接" name="metamask">
          <el-form :model="metamaskConfig" label-width="140px" style="max-width: 700px">
            <el-form-item label="合约地址" required>
              <el-input 
                v-model="metamaskConfig.contractAddress" 
                placeholder="0x..."
              />
            </el-form-item>
            <el-form-item>
              <el-button 
                type="primary" 
                @click="connectMetaMask" 
                :loading="connecting"
                :disabled="blockchainConnected"
              >
                连接 MetaMask
              </el-button>
              <el-button 
                v-if="blockchainConnected" 
                @click="disconnectBlockchain"
              >
                断开连接
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <!-- 连接信息 -->
      <div v-if="blockchainConnected && connectionInfo" style="margin-top: 20px; padding: 15px; background: #f5f7fa; border-radius: 4px">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="账户地址">
            <span style="font-family: monospace; font-size: 12px">{{ connectionInfo.address }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="网络">
            {{ connectionInfo.network.name }} (Chain ID: {{ connectionInfo.network.chainId }})
          </el-descriptions-item>
          <el-descriptions-item label="余额" :span="2">
            <span style="font-weight: bold; color: #409eff">{{ accountBalance }} ETH</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <!-- 跨域认证 -->
    <el-card>
      <template #header>
        <span>跨域认证</span>
      </template>

      <el-tabs v-model="authTab" type="border-card">
        <!-- 交易历史记录 -->
        <el-tab-pane label="交易历史" name="history">
          <div style="margin-top: 20px">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px">
              <h3>跨域认证交易历史 ({{ transactionHistory.length }})</h3>
              <el-button 
                type="danger" 
                size="small" 
                @click="clearHistory"
                :disabled="transactionHistory.length === 0"
              >
                清空历史
              </el-button>
            </div>
            
            <el-table 
              v-if="transactionHistory.length > 0" 
              :data="transactionHistory" 
              border
              style="margin-top: 10px"
            >
              <el-table-column prop="device_did" label="设备DID" width="150" />
              <el-table-column prop="source_domain" label="源域" width="100" />
              <el-table-column prop="target_domain" label="目标域" width="100" />
              <el-table-column prop="authorized" label="授权状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.authorized ? 'success' : 'danger'">
                    {{ row.authorized ? '已授权' : '未授权' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="tx_hash" label="交易哈希" min-width="200">
                <template #default="{ row }">
                  <div v-if="row.tx_hash">
                    <el-link 
                      type="primary" 
                      @click="goToVerify(row.tx_hash)"
                      :underline="false"
                    >
                      {{ formatTxHash(row.tx_hash) }}
                    </el-link>
                    <el-button 
                      type="text" 
                      size="small" 
                      @click="copyText(row.tx_hash)"
                      style="margin-left: 5px"
                    >
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                  </div>
                  <span v-else style="color: #909399">未上链</span>
                </template>
              </el-table-column>
              <el-table-column prop="block_number" label="区块号" width="120">
                <template #default="{ row }">
                  <span v-if="row.block_number" style="font-family: monospace">{{ row.block_number }}</span>
                  <span v-else style="color: #909399">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="timestamp" label="时间" width="180">
                <template #default="{ row }">
                  {{ formatTime(row.timestamp) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="120">
                <template #default="{ row }">
                  <el-button 
                    v-if="row.tx_hash"
                    type="primary" 
                    size="small" 
                    @click="goToVerify(row.tx_hash)"
                  >
                    验证
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            
            <el-empty 
              v-else 
              description="暂无交易历史记录" 
              :image-size="100"
              style="margin-top: 40px"
            />
          </div>
        </el-tab-pane>

        <!-- 后端API方式 -->
        <el-tab-pane label="后端API方式" name="api">
          <el-form :model="authForm" label-width="120px" style="max-width: 600px">
            <el-form-item label="设备DID" required>
              <el-input v-model="authForm.device_did" placeholder="did:example:device001" />
            </el-form-item>
            <el-form-item label="源域" required>
              <el-input v-model="authForm.source_domain" placeholder="smart_home" />
            </el-form-item>
            <el-form-item label="目标域" required>
              <el-input v-model="authForm.target_domain" placeholder="industrial_iot" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="requestAuth" :loading="loading">
                发起认证请求（后端）
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="authResult">
            <el-alert
              :title="authResult.authorized ? '认证成功' : '认证失败'"
              :type="authResult.authorized ? 'success' : 'error'"
              :closable="false"
              show-icon
              style="margin-top: 20px"
            >
              <template #default>
                <p>设备DID: {{ authResult.device_did }}</p>
                <p>源域: {{ authResult.source_domain }}</p>
                <p>目标域: {{ authResult.target_domain }}</p>
                <p>授权状态: {{ authResult.authorized ? '已授权' : '未授权' }}</p>
                <div v-if="authResult.tx_hash">
                  <p><strong>交易哈希:</strong> 
                    <el-link 
                      type="primary" 
                      @click="goToVerify(authResult.tx_hash)"
                      :underline="false"
                    >
                      {{ authResult.tx_hash }}
                    </el-link>
                  </p>
                </div>
                <div v-else style="margin-top: 10px">
                  <el-alert
                    title="未上链"
                    type="warning"
                    :closable="false"
                    show-icon
                  >
                    <template #default>
                      <p>后端区块链未连接，此认证未上链</p>
                      <p style="margin-top: 5px">
                        <el-button 
                          type="primary" 
                          size="small" 
                          @click="switchToOnchainAuth"
                        >
                          使用前端直接上链
                        </el-button>
                      </p>
                    </template>
                  </el-alert>
                </div>
              </template>
            </el-alert>
          </div>
        </el-tab-pane>

        <!-- 前端直接上链 -->
        <el-tab-pane label="前端直接上链" name="onchain">
          <el-alert
            v-if="!blockchainConnected"
            title="请先连接区块链"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 20px"
          />

          <!-- 设备注册 -->
          <el-card v-if="blockchainConnected" style="margin-bottom: 20px">
            <template #header>
              <span>设备注册（上链前必须先注册设备）</span>
            </template>
            <el-form :model="registerForm" label-width="120px" style="max-width: 600px">
              <el-form-item label="设备DID" required>
                <el-input v-model="registerForm.device_did" placeholder="1234124" />
              </el-form-item>
              <el-form-item label="设备元数据">
                <el-input 
                  v-model="registerForm.metadata" 
                  type="textarea" 
                  :rows="3"
                  placeholder='{"type": "sensor", "location": "room1"}'
                />
                <div style="font-size: 12px; color: #909399; margin-top: 5px">
                  可选，JSON格式字符串，默认为空对象 {}
                </div>
              </el-form-item>
              <el-form-item>
                <el-button 
                  type="success" 
                  @click="registerDeviceOnchain" 
                  :loading="registerLoading"
                >
                  <el-icon><Plus /></el-icon>
                  注册设备到区块链
                </el-button>
                <el-button 
                  @click="checkDeviceExists"
                  :loading="checkingDevice"
                >
                  <el-icon><Search /></el-icon>
                  检查设备是否存在
                </el-button>
              </el-form-item>
            </el-form>
            <div v-if="deviceCheckResult !== null" style="margin-top: 15px">
              <el-alert
                :title="deviceCheckResult.exists ? '设备已注册' : '设备未注册'"
                :type="deviceCheckResult.exists ? 'success' : 'warning'"
                :closable="false"
                show-icon
              >
                <template #default v-if="deviceCheckResult.exists">
                  <p><strong>设备DID:</strong> {{ deviceCheckResult.did }}</p>
                  <p v-if="deviceCheckResult.device_id"><strong>设备ID:</strong> {{ deviceCheckResult.device_id }}</p>
                  <p v-if="deviceCheckResult.domain && deviceCheckResult.domain !== '-'"><strong>所属域:</strong> {{ deviceCheckResult.domain }}</p>
                  <p><strong>状态（数据库）:</strong> 
                    <el-tag :type="getStatusType(deviceCheckResult.statusCode !== undefined ? deviceCheckResult.statusCode : (deviceCheckResult.status === 'active' ? 0 : deviceCheckResult.status === 'suspicious' ? 1 : 2))">
                      {{ getStatusText(deviceCheckResult.statusCode !== undefined ? deviceCheckResult.statusCode : (deviceCheckResult.status === 'active' ? 0 : deviceCheckResult.status === 'suspicious' ? 1 : 2)) }}
                    </el-tag>
                    <span v-if="deviceCheckResult.status === 'revoked' || deviceCheckResult.statusCode === 2" style="color: #f56c6c; margin-left: 10px; font-size: 12px;">
                      ⚠️ 设备已吊销
                    </span>
                  </p>
                  <p v-if="deviceCheckResult.blockchainStatus !== null && deviceCheckResult.blockchainStatus !== undefined">
                    <strong>状态（区块链）:</strong> 
                    <el-tag :type="getStatusType(deviceCheckResult.blockchainStatus)" size="small">
                      {{ getStatusText(deviceCheckResult.blockchainStatus) }}
                    </el-tag>
                    <span v-if="deviceCheckResult.statusConsistent === false" style="color: #e6a23c; margin-left: 10px; font-size: 12px;">
                      ⚠️ 状态不一致
                    </span>
                  </p>
                  <p v-if="deviceCheckResult.dbNotExists" style="color: #909399; font-size: 12px;">
                    ⓘ 提示：该设备仅在区块链上存在，数据库中未注册
                  </p>
                  <p v-if="deviceCheckResult.owner && deviceCheckResult.owner !== '-'"><strong>所有者:</strong> {{ deviceCheckResult.owner }}</p>
                </template>
              </el-alert>
            </div>
            <div v-if="registerResult" style="margin-top: 15px">
              <el-alert
                :title="registerResult.success ? '注册成功' : '注册失败'"
                :type="registerResult.success ? 'success' : 'error'"
                :closable="false"
                show-icon
              >
                <template #default v-if="registerResult.success">
                  <p><strong>交易哈希:</strong> 
                    <el-link 
                      type="primary" 
                      @click="copyText(registerResult.txHash)"
                      :underline="false"
                    >
                      {{ registerResult.txHash }}
                    </el-link>
                  </p>
                  <p><strong>区块号:</strong> {{ registerResult.blockNumber }}</p>
                </template>
                <template #default v-else>
                  <p><strong>错误:</strong> {{ registerResult.error }}</p>
                </template>
              </el-alert>
            </div>
          </el-card>

          <!-- 跨域认证 -->
          <el-form :model="onchainForm" label-width="120px" style="max-width: 600px" :disabled="!blockchainConnected">
            <el-form-item label="设备DID" required>
              <el-input v-model="onchainForm.device_did" placeholder="1234124" />
              <div style="font-size: 12px; color: #909399; margin-top: 5px">
                提示：设备必须先注册到区块链才能进行跨域认证
              </div>
            </el-form-item>
            <el-form-item label="源域" required>
              <el-input v-model="onchainForm.source_domain" placeholder="设备" />
            </el-form-item>
            <el-form-item label="目标域" required>
              <el-input v-model="onchainForm.target_domain" placeholder="能源" />
            </el-form-item>
            <el-form-item>
              <el-button 
                type="primary" 
                @click="requestOnchainAuth" 
                :loading="onchainLoading"
                :disabled="!blockchainConnected"
              >
                <el-icon><Link /></el-icon>
                直接上链认证
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 上链结果 -->
          <div v-if="onchainResult" style="margin-top: 20px">
            <el-alert
              :title="onchainResult.success ? '上链成功' : '上链失败'"
              :type="onchainResult.success ? 'success' : 'error'"
              :closable="false"
              show-icon
            >
              <template #default>
                <div v-if="onchainResult.success">
                  <p><strong>交易哈希:</strong> 
                    <el-link 
                      type="primary" 
                      @click="copyText(onchainResult.txHash)"
                      :underline="false"
                    >
                      {{ onchainResult.txHash }}
                    </el-link>
                    <el-button 
                      type="text" 
                      size="small" 
                      @click="copyText(onchainResult.txHash)"
                      style="margin-left: 5px"
                    >
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                  </p>
                  <p><strong>区块号:</strong> {{ onchainResult.blockNumber }}</p>
                  <p><strong>Gas 使用:</strong> {{ onchainResult.gasUsed }}</p>
                  <p><strong>授权状态:</strong> 
                    <el-tag :type="onchainResult.authorized ? 'success' : 'danger'">
                      {{ onchainResult.authorized ? '已授权' : '未授权' }}
                    </el-tag>
                  </p>
                  <div v-if="onchainResult.events && onchainResult.events.length > 0" style="margin-top: 10px">
                    <p><strong>事件:</strong></p>
                    <ul style="margin: 5px 0; padding-left: 20px">
                      <li v-for="(event, index) in onchainResult.events" :key="index">
                        {{ event.name }}
                      </li>
                    </ul>
                  </div>
                </div>
                <div v-else>
                  <p><strong>错误信息:</strong> {{ onchainResult.error }}</p>
                </div>
              </template>
            </el-alert>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleCheck, CircleClose, Link, DocumentCopy, Plus, Search } from '@element-plus/icons-vue'
import api from '../api'
import authApi from '../api/auth'
import deviceApi from '../api/device'
import blockchainService from '../services/blockchain'

// 标签页
const connectionTab = ref('ganache')
const authTab = ref('history')
const router = useRouter()

// 交易历史记录
const transactionHistory = ref([])

// Ganache 配置
const ganacheConfig = ref({
  rpcUrl: 'http://127.0.0.1:8545',
  contractAddress: '',
  privateKey: ''
})

// MetaMask 配置
const metamaskConfig = ref({
  contractAddress: ''
})

// 连接状态
const blockchainConnected = ref(false)
const connecting = ref(false)
const connectionInfo = ref(null)
const accountBalance = ref('0')

// 认证表单
const authForm = ref({
  device_did: '',
  source_domain: '',
  target_domain: '',
})

const authResult = ref(null)
const loading = ref(false)

// 设备注册表单
const registerForm = ref({
  device_did: '',
  metadata: '{}'
})

const registerLoading = ref(false)
const registerResult = ref(null)

// 设备检查
const checkingDevice = ref(false)
const deviceCheckResult = ref(null)

// 上链表单
const onchainForm = ref({
  device_did: '',
  source_domain: '',
  target_domain: '',
})

const onchainResult = ref(null)
const onchainLoading = ref(false)

// 从 localStorage 加载配置
onMounted(() => {
  const savedGanacheConfig = localStorage.getItem('ganacheConfig')
  if (savedGanacheConfig) {
    try {
      ganacheConfig.value = { ...ganacheConfig.value, ...JSON.parse(savedGanacheConfig) }
    } catch (e) {
      console.error('加载配置失败:', e)
    }
  }

  const savedMetamaskConfig = localStorage.getItem('metamaskConfig')
  if (savedMetamaskConfig) {
    try {
      metamaskConfig.value = { ...metamaskConfig.value, ...JSON.parse(savedMetamaskConfig) }
    } catch (e) {
      console.error('加载配置失败:', e)
    }
  }

  // 检查是否已连接
  if (blockchainService.isConnected()) {
    blockchainConnected.value = true
    updateConnectionInfo()
  }

  // 加载交易历史记录
  loadTransactionHistory()
})

// 加载交易历史记录
const loadTransactionHistory = () => {
  try {
    const history = localStorage.getItem('transactionHistory')
    if (history) {
      transactionHistory.value = JSON.parse(history)
      // 按时间倒序排列
      transactionHistory.value.sort((a, b) => {
        return new Date(b.timestamp) - new Date(a.timestamp)
      })
    }
  } catch (error) {
    console.error('加载交易历史失败:', error)
    transactionHistory.value = []
  }
}

// 保存交易到历史记录
const saveToHistory = (data) => {
  try {
    const record = {
      device_did: data.device_did || data.did || '',
      source_domain: data.source_domain || '',
      target_domain: data.target_domain || '',
      authorized: data.authorized !== undefined ? data.authorized : true,
      tx_hash: data.tx_hash || data.txHash || null,
      block_number: data.block_number || data.blockNumber || null,
      gas_used: data.gas_used || data.gasUsed || null,
      timestamp: data.timestamp || new Date().toISOString()
    }
    
    transactionHistory.value.unshift(record)
    
    // 限制最多保存100条记录
    if (transactionHistory.value.length > 100) {
      transactionHistory.value = transactionHistory.value.slice(0, 100)
    }
    
    // 保存到 localStorage
    localStorage.setItem('transactionHistory', JSON.stringify(transactionHistory.value))
  } catch (error) {
    console.error('保存交易历史失败:', error)
  }
}

// 清空历史记录
const clearHistory = () => {
  ElMessageBox.confirm('确定要清空所有交易历史记录吗？', '提示', {
    type: 'warning',
  }).then(() => {
    transactionHistory.value = []
    localStorage.removeItem('transactionHistory')
    ElMessage.success('历史记录已清空')
  }).catch(() => {
    // 取消操作
  })
}

// 格式化交易哈希
const formatTxHash = (hash) => {
  if (!hash) return ''
  if (hash.length <= 20) return hash
  return `${hash.substring(0, 10)}...${hash.substring(hash.length - 8)}`
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  try {
    const date = new Date(timeStr)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return timeStr
  }
}

// 跳转到验证页面
const goToVerify = (txHash) => {
  router.push({
    path: '/verify',
    query: { txHash }
  })
}

// 连接 Ganache
const connectGanache = async () => {
  if (!ganacheConfig.value.contractAddress) {
    ElMessage.warning('请输入合约地址')
    return
  }
  if (!ganacheConfig.value.privateKey) {
    ElMessage.warning('请输入私钥')
    return
  }

  connecting.value = true
  try {
    const result = await blockchainService.connect({
      rpcUrl: ganacheConfig.value.rpcUrl,
      contractAddress: ganacheConfig.value.contractAddress,
      privateKey: ganacheConfig.value.privateKey,
      useMetaMask: false
    })

    blockchainConnected.value = true
    connectionInfo.value = result
    await updateConnectionInfo()
    
    // 保存配置
    localStorage.setItem('ganacheConfig', JSON.stringify(ganacheConfig.value))
    
    ElMessage.success('已连接到 Ganache')
  } catch (error) {
    ElMessage.error(`连接失败: ${error.message}`)
    console.error('连接失败:', error)
  } finally {
    connecting.value = false
  }
}

// 连接 MetaMask
const connectMetaMask = async () => {
  if (!metamaskConfig.value.contractAddress) {
    ElMessage.warning('请输入合约地址')
    return
  }

  connecting.value = true
  try {
    const result = await blockchainService.connect({
      contractAddress: metamaskConfig.value.contractAddress,
      useMetaMask: true
    })

    blockchainConnected.value = true
    connectionInfo.value = result
    await updateConnectionInfo()
    
    // 保存配置
    localStorage.setItem('metamaskConfig', JSON.stringify(metamaskConfig.value))
    
    ElMessage.success('已连接到 MetaMask')
  } catch (error) {
    ElMessage.error(`连接失败: ${error.message}`)
    console.error('连接失败:', error)
  } finally {
    connecting.value = false
  }
}

// 更新连接信息
const updateConnectionInfo = async () => {
  try {
    const balance = await blockchainService.getBalance()
    accountBalance.value = parseFloat(balance).toFixed(4)
  } catch (error) {
    console.error('获取余额失败:', error)
  }
}

// 断开连接
const disconnectBlockchain = () => {
  blockchainService.disconnect()
  blockchainConnected.value = false
  connectionInfo.value = null
  accountBalance.value = '0'
  ElMessage.info('已断开连接')
}

// 切换到前端上链认证
const switchToOnchainAuth = () => {
  // 填充前端上链表单
  onchainForm.value.device_did = authForm.value.device_did
  onchainForm.value.source_domain = authForm.value.source_domain
  onchainForm.value.target_domain = authForm.value.target_domain
  
  // 切换到前端上链标签页
  authTab.value = 'onchain'
  
  // 如果未连接区块链，提示连接
  if (!blockchainConnected.value) {
    ElMessage.warning('请先连接区块链')
    connectionTab.value = 'ganache'
  }
}

// 后端API方式认证
const requestAuth = async () => {
  if (!authForm.value.device_did || !authForm.value.source_domain || !authForm.value.target_domain) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  // 先检查设备状态和域
  try {
    const device = await deviceApi.get(authForm.value.device_did)
    
    if (device.status === 'revoked') {
      ElMessage.error('设备已被吊销，无法进行跨域认证。请在设备管理页面激活设备。')
      return
    } else if (device.status === 'suspicious') {
      ElMessage.warning('设备状态为可疑，可能无法通过认证。建议在设备管理页面将设备状态设置为激活。')
    } else if (device.status !== 'active') {
      ElMessage.error(`设备状态异常（${device.status}），无法进行认证。请在设备管理页面将设备状态设置为激活。`)
      return
    }
    
    // 验证设备的域是否与源域匹配
    if (device.domain && device.domain !== authForm.value.source_domain) {
      ElMessage.error('源域错误：设备所属域与源域不匹配，无法进行跨域认证。')
      return
    }
    
    // 验证目标域是否存在
    try {
      await api.get(`/domains/${encodeURIComponent(authForm.value.target_domain)}`)
    } catch (error) {
      if (error.response?.status === 404 || (error.message && error.message.includes('not found'))) {
        ElMessage.error('目标域不存在：指定的目标域在系统中不存在，请先在域管理页面创建该域。')
        return
      }
      // 其他错误继续，让后端验证
    }
  } catch (error) {
    // 如果获取设备信息失败，可能是设备不存在，继续尝试认证让后端返回错误
    console.warn('检查设备状态失败:', error)
    if (error.message && error.message.includes('not found')) {
      ElMessage.error('设备不存在，请先注册设备')
      return
    }
  }

  loading.value = true
  try {
    const result = await api.post('/auth/cross-domain', authForm.value)
    authResult.value = {
      ...authForm.value,
      authorized: result.authorized,
      tx_hash: result.tx_hash || null
    }
    
    // 保存到历史记录（即使没有交易哈希也保存）
    saveToHistory({
      ...authForm.value,
      authorized: result.authorized,
      tx_hash: result.tx_hash || null
    })
    
    if (result.tx_hash) {
      ElMessage.success('认证请求已提交并上链')
    } else {
      ElMessage.warning('认证请求已提交，但未上链（后端区块链未连接）')
    }
  } catch (error) {
    // 改进错误提示
    let errorMsg = error.message || '认证失败'
    const errorResponse = error.response?.data
    
    // 检查后端返回的具体错误类型
    if (errorResponse?.error === '源域错误') {
      errorMsg = '源域错误：' + (errorResponse.message || '设备所属域与源域不匹配')
    } else if (errorResponse?.error === '目标域不存在') {
      errorMsg = '目标域不存在：' + (errorResponse.message || '指定的目标域在系统中不存在')
    } else if (errorMsg.includes('Device is not active') || errorMsg.includes('not active')) {
      errorMsg = '设备状态不是激活状态，无法进行跨域认证。请在设备管理页面将设备状态设置为激活。'
    } else if (errorMsg.includes('Device not found')) {
      errorMsg = '设备不存在，请先注册设备'
    } else if (errorMsg.includes('Device domain does not match source domain') || errorMsg.includes('源域错误')) {
      errorMsg = '源域错误：设备所属域与源域不匹配'
    } else if (errorMsg.includes('目标域不存在') || errorMsg.includes('target domain') && errorMsg.includes('not found')) {
      errorMsg = '目标域不存在：指定的目标域在系统中不存在'
    }
    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}

// 注册设备到区块链
const registerDeviceOnchain = async () => {
  if (!registerForm.value.device_did) {
    ElMessage.warning('请输入设备DID')
    return
  }

  if (!blockchainService.isConnected()) {
    ElMessage.warning('请先连接区块链')
    return
  }

  registerLoading.value = true
  registerResult.value = null

  try {
    const metadata = registerForm.value.metadata.trim() || '{}'
    // 验证 JSON 格式
    try {
      JSON.parse(metadata)
    } catch (e) {
      ElMessage.error('设备元数据必须是有效的 JSON 格式')
      registerLoading.value = false
      return
    }

    const result = await blockchainService.registerDevice(
      registerForm.value.device_did,
      metadata
    )

    registerResult.value = {
      success: true,
      ...result
    }

    // 更新余额
    await updateConnectionInfo()

    ElMessage.success('设备注册成功！交易已确认')
    
    // 自动检查设备状态
    await checkDeviceExists()
  } catch (error) {
    console.error('注册失败:', error)
    registerResult.value = {
      success: false,
      error: error.message || '注册失败'
    }
    
    // 改进错误提示
    let errorMsg = error.message || '注册失败'
    if (errorMsg.includes('Device already exists')) {
      errorMsg = '设备已存在，请使用其他 DID'
    } else if (errorMsg.includes('Only authorized admin')) {
      errorMsg = '当前账户没有注册权限，请使用管理员账户'
    }
    
    ElMessage.error(`注册失败: ${errorMsg}`)
  } finally {
    registerLoading.value = false
  }
}

// 检查设备是否存在
const checkDeviceExists = async () => {
  if (!registerForm.value.device_did && !onchainForm.value.device_did) {
    ElMessage.warning('请输入设备DID')
    return
  }

  // 获取DID并去除前后空格
  let did = (registerForm.value.device_did || onchainForm.value.device_did)
  if (did) {
    did = did.trim()
  }
  
  if (!did) {
    ElMessage.warning('设备DID不能为空')
    return
  }

  checkingDevice.value = true
  deviceCheckResult.value = null

  // 状态映射表（在函数顶部定义，供所有分支使用）
  const statusMap = { 0: 'active', 1: 'suspicious', 2: 'revoked' }

  try {
    // 1. 先查询数据库中的设备状态（权威状态）
    let dbDevice = null
    let dbStatus = null
    try {
      dbDevice = await deviceApi.get(did)
      dbStatus = dbDevice.status
    } catch (error) {
      console.warn('查询数据库设备信息失败:', error)
      // 数据库中没有设备，继续检查区块链
    }

    // 2. 如果区块链已连接，查询区块链上的设备状态
    let blockchainDevice = null
    let blockchainStatus = null
    if (blockchainService.isConnected()) {
      try {
        const exists = await blockchainService.checkDeviceExists(did)
        if (exists) {
          blockchainDevice = await blockchainService.getDevice(did)
          blockchainStatus = blockchainDevice.status // 0: Active, 1: Suspicious, 2: Revoked
        }
      } catch (error) {
        console.warn('查询区块链设备信息失败:', error)
      }
    }

    // 3. 组合结果，优先使用数据库状态
    if (dbDevice) {
      // 数据库中有设备，使用数据库状态（权威状态）
      deviceCheckResult.value = {
        exists: true,
        did: dbDevice.did,
        status: dbStatus, // 数据库状态（字符串）
        statusCode: dbStatus === 'active' ? 0 : dbStatus === 'suspicious' ? 1 : 2, // 转换为数字状态码用于显示
        owner: blockchainDevice?.owner || dbDevice.owner || '-',
        domain: dbDevice.domain || '-',
        device_id: dbDevice.device_id || '-',
        // 区块链状态（如果有）
        blockchainStatus: blockchainStatus !== null ? blockchainStatus : null,
        blockchainStatusText: blockchainStatus !== null ? (statusMap[blockchainStatus] || 'unknown') : null,
        // 状态一致性
        statusConsistent: blockchainStatus !== null && 
          ((dbStatus === 'active' && blockchainStatus === 0) ||
           (dbStatus === 'suspicious' && blockchainStatus === 1) ||
           (dbStatus === 'revoked' && blockchainStatus === 2))
      }
    } else if (blockchainDevice) {
      // 只有区块链上有设备，使用区块链状态
      deviceCheckResult.value = {
        exists: true,
        did: blockchainDevice.did,
        status: blockchainDevice.status, // 数字状态码
        statusCode: blockchainDevice.status,
        owner: blockchainDevice.owner || '-',
        domain: '-',
        device_id: '-',
        blockchainStatus: blockchainDevice.status,
        blockchainStatusText: statusMap[blockchainDevice.status] || 'unknown',
        statusConsistent: true,
        // 提示：数据库中没有该设备
        dbNotExists: true
      }
    } else {
      // 数据库和区块链都没有设备
      deviceCheckResult.value = {
        exists: false
      }
    }
  } catch (error) {
    console.error('检查设备失败:', error)
    deviceCheckResult.value = {
      exists: false
    }
  } finally {
    checkingDevice.value = false
  }
}

// 获取状态类型
const getStatusType = (status) => {
  const map = {
    0: 'success', // Active
    1: 'warning', // Suspicious
    2: 'danger'   // Revoked
  }
  return map[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const map = {
    0: '活跃',
    1: '可疑',
    2: '已吊销'
  }
  return map[status] || '未知'
}

// 前端直接上链认证
const requestOnchainAuth = async () => {
  if (!onchainForm.value.device_did || !onchainForm.value.source_domain || !onchainForm.value.target_domain) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  if (!blockchainService.isConnected()) {
    ElMessage.warning('请先连接区块链')
    return
  }

  // 先检查设备是否存在（区块链）
  const exists = await blockchainService.checkDeviceExists(onchainForm.value.device_did)
  if (!exists) {
    ElMessage.error('设备未在区块链上注册，请先注册设备')
    // 自动切换到注册表单
    registerForm.value.device_did = onchainForm.value.device_did
    await checkDeviceExists()
    return
  }

  // 同时检查数据库和区块链的设备状态
  let dbDeviceStatus = null
  let blockchainDeviceStatus = null
  
  // 1. 检查数据库中的设备状态（权威状态）
  try {
    const dbDevice = await deviceApi.get(onchainForm.value.device_did)
    dbDeviceStatus = dbDevice.status
    
    if (dbDeviceStatus === 'revoked') {
      ElMessage.error('设备在数据库中已被吊销，无法进行跨域认证。请在设备管理页面激活设备。')
      return
    } else if (dbDeviceStatus === 'suspicious') {
      ElMessage.warning('设备在数据库中状态为可疑，可能无法通过认证。建议在设备管理页面将设备状态设置为激活。')
    } else if (dbDeviceStatus !== 'active') {
      ElMessage.error(`设备在数据库中状态异常（${dbDeviceStatus}），无法进行认证。请在设备管理页面将设备状态设置为激活。`)
      return
    }
    
    // 验证设备的域是否与源域匹配
    if (dbDevice.domain && dbDevice.domain !== onchainForm.value.source_domain) {
      ElMessage.error('源域错误：设备所属域与源域不匹配，无法进行跨域认证。')
      return
    }
    
    // 验证目标域是否存在
    try {
      await api.get(`/domains/${encodeURIComponent(onchainForm.value.target_domain)}`)
    } catch (error) {
      if (error.response?.status === 404 || (error.message && error.message.includes('not found'))) {
        ElMessage.error('目标域不存在：指定的目标域在系统中不存在，请先在域管理页面创建该域。')
        return
      }
      // 其他错误继续，让后端验证
    }
  } catch (error) {
    console.warn('检查数据库设备状态失败:', error)
    if (error.message && error.message.includes('not found')) {
      ElMessage.error('设备在数据库中不存在，请先在设备管理页面注册设备')
      return
    }
    // 如果查询失败但不是"不存在"，继续检查区块链状态
  }

  // 2. 检查区块链上的设备状态
  try {
    const blockchainDevice = await blockchainService.getDevice(onchainForm.value.device_did)
    blockchainDeviceStatus = blockchainDevice.status // 0: Active, 1: Suspicious, 2: Revoked
    
    if (blockchainDeviceStatus === 2) {
      ElMessage.error('设备在区块链上已被吊销，无法进行跨域认证。请在设备管理页面激活设备。')
      return
    } else if (blockchainDeviceStatus === 1) {
      ElMessage.warning('设备在区块链上状态为可疑，可能无法通过认证。建议在设备管理页面将设备状态设置为激活。')
    } else if (blockchainDeviceStatus !== 0) {
      ElMessage.error(`设备在区块链上状态异常（状态码：${blockchainDeviceStatus}），无法进行认证`)
      return
    }
    
    // 如果数据库和区块链状态不一致，给出警告
    if (dbDeviceStatus !== null) {
      const statusMap = { 0: 'active', 1: 'suspicious', 2: 'revoked' }
      const blockchainStatusStr = statusMap[blockchainDeviceStatus] || 'unknown'
      
      if (dbDeviceStatus !== blockchainStatusStr) {
        ElMessage.warning({
          message: `警告：设备状态不一致。数据库：${dbDeviceStatus}，区块链：${blockchainStatusStr}。将以数据库状态为准。`,
          duration: 5000
        })
      }
    }
  } catch (error) {
    console.error('检查区块链设备状态失败:', error)
    // 如果数据库状态是active，即使区块链检查失败也继续
    if (dbDeviceStatus !== 'active') {
      ElMessage.error('无法检查区块链设备状态，且数据库状态不是激活，无法进行认证')
      return
    }
    ElMessage.warning('无法检查区块链设备状态，将以数据库状态为准继续认证')
  }

  onchainLoading.value = true
  onchainResult.value = null

  try {
    const result = await blockchainService.requestCrossDomainAuth(
      onchainForm.value.device_did,
      onchainForm.value.source_domain,
      onchainForm.value.target_domain
    )

    onchainResult.value = {
      success: true,
      ...result
    }

    // 保存到历史记录
    saveToHistory({
      device_did: onchainForm.value.device_did,
      source_domain: onchainForm.value.source_domain,
      target_domain: onchainForm.value.target_domain,
      authorized: result.authorized,
      tx_hash: result.txHash,
      block_number: result.blockNumber,
      gas_used: result.gasUsed
    })

    // 同步到后端数据库
    try {
      console.log('准备同步到数据库:', {
        device_did: onchainForm.value.device_did,
        source_domain: onchainForm.value.source_domain,
        target_domain: onchainForm.value.target_domain,
        tx_hash: result.txHash,
        authorized: result.authorized,
        block_number: result.blockNumber
      })
      
      const syncResult = await authApi.syncAuthRecord({
        device_did: onchainForm.value.device_did,
        source_domain: onchainForm.value.source_domain,
        target_domain: onchainForm.value.target_domain,
        tx_hash: result.txHash,
        authorized: result.authorized,
        block_number: result.blockNumber
      })
      
      console.log('同步结果:', syncResult)
      
      if (syncResult.message && syncResult.message.includes('already exists')) {
        ElMessage.success('上链成功！交易已确认（记录已存在）')
      } else {
        ElMessage.success('上链成功！交易已确认并同步到数据库')
      }
    } catch (error) {
      console.error('同步到数据库失败:', error)
      console.error('错误详情:', {
        message: error.message,
        response: error.response,
        status: error.response?.status,
        data: error.response?.data
      })
      
      // 改进错误提示
      let errorMsg = error.message || '未知错误'
      const status = error.response?.status
      
      if (status === 404) {
        errorMsg = 'API路由未找到，请确保后端服务已重启并包含最新的路由配置'
      } else if (status === 401 || errorMsg.includes('Unauthorized')) {
        errorMsg = '未登录或登录已过期，请重新登录'
      } else if (status === 403 || errorMsg.includes('Forbidden')) {
        errorMsg = '权限不足，当前用户没有同步认证记录的权限'
      } else if (errorMsg.includes('Device not found')) {
        errorMsg = '设备未在数据库中注册，请先在设备管理页面注册设备'
      } else if (errorMsg.includes('Network Error') || errorMsg.includes('Failed to fetch')) {
        errorMsg = '网络连接失败，请检查后端服务是否正常运行'
      }
      
      // 即使同步失败，也不影响上链成功，但给出明确提示
      ElMessage.warning({
        message: '上链成功，但同步到数据库失败：' + errorMsg,
        duration: 6000
      })
      
      // 根据错误类型给出不同的提示
      if (status === 404) {
        ElMessage.info({
          message: '请重启后端服务以加载最新的路由配置',
          duration: 5000
        })
      } else {
        ElMessage.info({
          message: '提示：可以在"认证验证"页面使用"同步缺失的交易哈希"功能手动同步',
          duration: 5000
        })
      }
    }

    // 更新余额
    await updateConnectionInfo()
  } catch (error) {
    console.error('上链失败:', error)
    
    // 改进错误提示
    let errorMsg = error.message || '上链失败'
    if (errorMsg.includes('Device does not exist')) {
      errorMsg = '设备未在区块链上注册，请先注册设备'
    } else if (errorMsg.includes('Device is not active')) {
      errorMsg = '设备状态不是活跃状态，无法进行跨域认证'
    }
    
    onchainResult.value = {
      success: false,
      error: errorMsg
    }
    ElMessage.error(`上链失败: ${errorMsg}`)
  } finally {
    onchainLoading.value = false
  }
}

// 复制文本
const copyText = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
.auth-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-descriptions__label) {
  font-weight: 600;
}
</style>

