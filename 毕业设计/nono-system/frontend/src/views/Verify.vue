<template>
  <div class="verify-page">
    <el-card>
      <template #header>
        <span>跨域认证验证</span>
      </template>

      <!-- 查询区域 -->
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 按设备DID查询 -->
        <el-tab-pane label="按设备查询" name="device">
          <el-form :model="deviceQueryForm" label-width="120px" style="max-width: 600px; margin-top: 20px">
            <el-form-item label="设备DID" required>
              <el-input 
                v-model="deviceQueryForm.device_did" 
                placeholder="123456 或 did:example:device001"
                clearable
              />
              <div style="font-size: 12px; color: #909399; margin-top: 5px">
                提示：如果查询不到，可以尝试查看所有记录或从交易历史中查找
              </div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="queryByDevice" :loading="deviceLoading">
                查询认证记录
              </el-button>
              <el-button @click="queryAllRecords" :loading="deviceLoading">
                查看所有记录
              </el-button>
              <el-button @click="queryFromHistory" :loading="deviceLoading">
                从交易历史查找
              </el-button>
              <el-button @click="clearDeviceResults">清空</el-button>
            </el-form-item>
          </el-form>

          <!-- 设备状态信息 -->
          <div v-if="deviceStatusInfo" style="margin-top: 20px; padding: 15px; background-color: #f5f7fa; border-radius: 4px;">
            <h4 style="margin: 0 0 10px 0;">设备信息（数据库状态）</h4>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="设备DID">{{ deviceStatusInfo.did }}</el-descriptions-item>
              <el-descriptions-item label="设备ID">{{ deviceStatusInfo.device_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="所属域">{{ deviceStatusInfo.domain || '-' }}</el-descriptions-item>
              <el-descriptions-item label="设备状态">
                <el-tag :type="getStatusType(deviceStatusInfo.status)" size="small">
                  {{ getStatusText(deviceStatusInfo.status) }}
                </el-tag>
                <span v-if="deviceStatusInfo.status === 'revoked'" style="color: #f56c6c; margin-left: 10px; font-size: 12px;">
                  ⚠️ 设备已吊销，无法进行跨域认证
                </span>
                <span v-else-if="deviceStatusInfo.status === 'suspicious'" style="color: #e6a23c; margin-left: 10px; font-size: 12px;">
                  ⚠️ 设备状态可疑，可能无法通过认证
                </span>
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- 统计信息 -->
          <div v-if="deviceRecords.length > 0" style="margin-top: 30px">
            <el-row :gutter="20">
              <el-col :span="6">
                <el-statistic title="总记录数" :value="deviceRecords.length">
                  <template #prefix>
                    <el-icon style="vertical-align: -0.125em"><Document /></el-icon>
                  </template>
                </el-statistic>
              </el-col>
              <el-col :span="6">
                <el-statistic title="成功认证" :value="successCount">
                  <template #prefix>
                    <el-icon style="vertical-align: -0.125em; color: #67c23a"><CircleCheck /></el-icon>
                  </template>
                </el-statistic>
              </el-col>
              <el-col :span="6">
                <el-statistic title="失败认证" :value="failedCount">
                  <template #prefix>
                    <el-icon style="vertical-align: -0.125em; color: #f56c6c"><CircleClose /></el-icon>
                  </template>
                </el-statistic>
              </el-col>
              <el-col :span="6">
                <el-statistic title="已上链" :value="onChainCount">
                  <template #prefix>
                    <el-icon style="vertical-align: -0.125em; color: #409eff"><Link /></el-icon>
                  </template>
                </el-statistic>
              </el-col>
            </el-row>
          </div>

          <!-- 认证记录列表 -->
          <div v-if="deviceRecords.length > 0" style="margin-top: 30px">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px">
              <h3>认证记录 ({{ deviceRecords.length }})</h3>
              <el-button 
                type="primary" 
                size="small" 
                @click="syncMissingTxHashes"
                :loading="syncingTxHashes"
              >
                <el-icon><Refresh /></el-icon>
                同步缺失的交易哈希
              </el-button>
            </div>
            <el-table :data="deviceRecords" border style="margin-top: 10px">
              <el-table-column prop="device_did" label="设备DID" width="200" />
              <el-table-column prop="source_domain" label="源域" width="120" />
              <el-table-column prop="target_domain" label="目标域" width="120" />
              <el-table-column prop="authorized" label="授权状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.authorized ? 'success' : 'danger'">
                    {{ row.authorized ? '已授权' : '未授权' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="tx_hash" label="交易哈希" min-width="200">
                <template #default="{ row }">
                  <div v-if="row.tx_hash && row.tx_hash.trim()">
                    <el-link 
                      type="primary" 
                      @click="verifyTransaction(row.tx_hash)"
                      :underline="false"
                    >
                      {{ formatTxHash(row.tx_hash) }}
                    </el-link>
                    <el-button 
                      type="text" 
                      size="small" 
                      @click="copyTxHash(row.tx_hash)"
                      style="margin-left: 5px"
                    >
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                  </div>
                  <span v-else style="color: #909399">未上链</span>
                </template>
              </el-table-column>
              <el-table-column prop="timestamp" label="时间" width="180">
                <template #default="{ row }">
                  {{ formatTime(row.timestamp) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200">
                <template #default="{ row }">
                  <el-button 
                    v-if="row.tx_hash && row.tx_hash.trim()"
                    type="primary" 
                    size="small" 
                    @click="verifyTransaction(row.tx_hash)"
                  >
                    验证交易
                  </el-button>
                  <el-button 
                    v-else
                    type="warning" 
                    size="small" 
                    @click="showUpdateTxHashDialog(row)"
                  >
                    添加交易哈希
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <!-- 时间线视图 -->
            <div style="margin-top: 30px">
              <h3>时间线视图</h3>
              <el-timeline style="margin-top: 20px">
                <el-timeline-item
                  v-for="(record, index) in deviceRecords"
                  :key="index"
                  :timestamp="formatTime(record.timestamp)"
                  placement="top"
                  :type="record.authorized ? 'success' : 'danger'"
                  :icon="record.authorized ? CircleCheck : CircleClose"
                >
                  <el-card>
                    <h4>{{ record.source_domain }} → {{ record.target_domain }}</h4>
                    <p>
                      <el-tag :type="record.authorized ? 'success' : 'danger'" size="small">
                        {{ record.authorized ? '已授权' : '未授权' }}
                      </el-tag>
                      <span v-if="record.tx_hash && record.tx_hash.trim()" style="margin-left: 10px">
                        <el-link type="primary" @click="verifyTransaction(record.tx_hash)" :underline="false">
                          交易: {{ formatTxHash(record.tx_hash) }}
                        </el-link>
                      </span>
                      <span v-else style="margin-left: 10px; color: #909399">未上链</span>
                    </p>
                  </el-card>
                </el-timeline-item>
              </el-timeline>
            </div>
          </div>
          <div v-else style="margin-top: 30px; text-align: center; padding: 40px">
            <el-empty description="请输入设备DID查询认证记录" :image-size="100" />
          </div>
        </el-tab-pane>

        <!-- 按交易哈希查询 -->
        <el-tab-pane label="按交易哈希查询" name="transaction">
          <el-form :model="txQueryForm" label-width="120px" style="max-width: 600px; margin-top: 20px">
            <el-form-item label="交易哈希" required>
              <el-input 
                v-model="txQueryForm.tx_hash" 
                placeholder="0x..."
                clearable
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="verifyTransaction(txQueryForm.tx_hash)" :loading="txLoading">
                验证交易
              </el-button>
              <el-button @click="clearTxResults">清空</el-button>
            </el-form-item>
          </el-form>
          <div v-if="!txResult" style="margin-top: 30px; text-align: center; padding: 40px">
            <el-empty description="请输入交易哈希进行验证" :image-size="100" />
          </div>
        </el-tab-pane>
      </el-tabs>

      <!-- 交易验证结果 -->
      <el-card v-if="txResult" style="margin-top: 30px">
        <template #header>
          <div class="card-header">
            <span>交易验证结果</span>
            <el-button type="text" @click="txResult = null">关闭</el-button>
          </div>
        </template>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="交易哈希" :span="2">
            <div style="display: flex; align-items: center">
              <span style="font-family: monospace; word-break: break-all">{{ txResult.tx_hash }}</span>
              <el-button 
                type="text" 
                size="small" 
                @click="copyTxHash(txResult.tx_hash)"
                style="margin-left: 5px"
              >
                <el-icon><DocumentCopy /></el-icon>
              </el-button>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="交易状态">
            <el-tag :type="txResult.status ? 'success' : 'danger'" size="large">
              <el-icon v-if="txResult.status"><CircleCheck /></el-icon>
              <el-icon v-else><CircleClose /></el-icon>
              {{ txResult.status ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="区块号">
            <span style="font-family: monospace">{{ txResult.block_number || 'N/A' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="Gas使用">
            <span style="font-family: monospace">{{ txResult.gas_used || 'N/A' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="确认数">
            <span>{{ txResult.confirmations || 0 }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 事件列表 -->
        <div v-if="txResult.events && txResult.events.length > 0" style="margin-top: 20px">
          <h4>事件列表 ({{ txResult.events.length }})</h4>
          <el-table :data="txResult.events" border style="margin-top: 10px">
            <el-table-column prop="address" label="合约地址" min-width="200">
              <template #default="{ row }">
                <span style="font-family: monospace; font-size: 12px">{{ row.address }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="block_number" label="区块号" width="120">
              <template #default="{ row }">
                <span style="font-family: monospace">{{ row.block_number }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="tx_index" label="交易索引" width="100">
              <template #default="{ row }">
                <span style="font-family: monospace">{{ row.tx_index }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="topics" label="主题" min-width="300">
              <template #default="{ row }">
                <div v-if="row.topics && row.topics.length > 0">
                  <el-tag 
                    v-for="(topic, index) in row.topics" 
                    :key="index"
                    size="small"
                    style="margin: 2px; font-family: monospace; font-size: 10px"
                  >
                    {{ formatTopic(topic) }}
                  </el-tag>
                </div>
                <span v-else style="color: #909399">无</span>
              </template>
            </el-table-column>
            <el-table-column prop="data" label="数据" min-width="200">
              <template #default="{ row }">
                <el-popover placement="top" :width="400" trigger="hover">
                  <template #reference>
                    <span style="font-family: monospace; font-size: 11px; cursor: pointer">
                      {{ formatData(row.data) }}
                    </span>
                  </template>
                  <div style="font-family: monospace; font-size: 11px; word-break: break-all">
                    {{ row.data }}
                  </div>
                </el-popover>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <div v-else style="margin-top: 20px; color: #909399">
          <el-empty description="无事件" :image-size="80" />
        </div>
      </el-card>
    </el-card>

    <!-- 更新交易哈希对话框 -->
    <el-dialog v-model="showUpdateDialog" title="添加/更新交易哈希" width="500px">
      <el-form :model="updateTxHashForm" label-width="120px">
        <el-form-item label="设备DID">
          <el-input v-model="updateTxHashForm.device_did" disabled />
        </el-form-item>
        <el-form-item label="源域">
          <el-input v-model="updateTxHashForm.source_domain" disabled />
        </el-form-item>
        <el-form-item label="目标域">
          <el-input v-model="updateTxHashForm.target_domain" disabled />
        </el-form-item>
        <el-form-item label="交易哈希" required>
          <el-input 
            v-model="updateTxHashForm.tx_hash" 
            placeholder="0x..."
            clearable
          />
          <div style="font-size: 12px; color: #909399; margin-top: 5px">
            输入区块链交易哈希，将从交易历史中自动匹配
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpdateDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="updateTxHash"
          :loading="updatingTxHash"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, CircleCheck, CircleClose, Document, Link, Refresh } from '@element-plus/icons-vue'
import authApi from '../api/auth'
import blockchainService from '../services/blockchain'

const activeTab = ref('device')

// 设备查询
const deviceQueryForm = ref({
  device_did: '',
})
const deviceRecords = ref([])
const deviceLoading = ref(false)

// 交易查询
const txQueryForm = ref({
  tx_hash: '',
})
const txResult = ref(null)
const txLoading = ref(false)
const syncingTxHashes = ref(false)

// 更新交易哈希
const showUpdateDialog = ref(false)
const updatingTxHash = ref(false)
const updateTxHashForm = ref({
  device_did: '',
  source_domain: '',
  target_domain: '',
  tx_hash: ''
})

// 设备状态信息
const deviceStatusInfo = ref(null)

// 按设备查询认证记录
const queryByDevice = async () => {
  if (!deviceQueryForm.value.device_did) {
    ElMessage.warning('请输入设备DID')
    return
  }

  deviceLoading.value = true
  deviceStatusInfo.value = null
  
  try {
    // 先查询设备信息（数据库状态）
    try {
      const deviceApi = (await import('../api/device')).default
      const device = await deviceApi.get(deviceQueryForm.value.device_did)
      deviceStatusInfo.value = {
        did: device.did,
        status: device.status,
        domain: device.domain,
        device_id: device.device_id
      }
    } catch (error) {
      console.warn('查询设备信息失败:', error)
      // 设备不存在不影响查询认证记录
    }
    
    // 查询认证记录
    deviceRecords.value = await authApi.getAuthRecords(deviceQueryForm.value.device_did)
    if (deviceRecords.value.length === 0) {
      ElMessage.warning('未找到认证记录，建议：1. 检查DID是否正确 2. 查看所有记录 3. 从交易历史中查找')
    } else {
      ElMessage.success(`找到 ${deviceRecords.value.length} 条认证记录`)
    }
  } catch (error) {
    ElMessage.error(error.message || '查询失败')
    deviceRecords.value = []
  } finally {
    deviceLoading.value = false
  }
}

// 获取状态类型
const getStatusType = (status) => {
  const map = {
    active: 'success',
    suspicious: 'warning',
    revoked: 'danger',
  }
  return map[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const map = {
    active: '活跃',
    suspicious: '可疑',
    revoked: '已吊销',
  }
  return map[status] || status || '未知'
}

// 查询所有认证记录
const queryAllRecords = async () => {
  deviceLoading.value = true
  try {
    // 通过查询日志来获取所有记录（因为日志接口支持不传参数查询所有）
    const logs = await authApi.getAuthLogs({})
    
    if (logs.length === 0) {
      ElMessage.info('数据库中暂无认证记录')
      deviceRecords.value = []
      return
    }
    
    // 从日志中提取唯一的设备DID
    const uniqueDIDs = [...new Set(logs.map(log => log.device_did).filter(did => did))]
    
    if (uniqueDIDs.length === 0) {
      ElMessage.info('未找到有效的设备DID')
      deviceRecords.value = []
      return
    }
    
    // 为每个DID查询认证记录
    const allRecords = []
    for (const did of uniqueDIDs) {
      try {
        const records = await authApi.getAuthRecords(did)
        allRecords.push(...records)
      } catch (error) {
        console.warn(`查询设备 ${did} 的记录失败:`, error)
      }
    }
    
    // 按时间倒序排列
    allRecords.sort((a, b) => {
      return new Date(b.timestamp || b.created_at) - new Date(a.timestamp || a.created_at)
    })
    
    deviceRecords.value = allRecords
    
    if (deviceRecords.value.length === 0) {
      ElMessage.warning('找到了设备DID，但没有找到对应的认证记录')
    } else {
      ElMessage.success(`找到 ${deviceRecords.value.length} 条认证记录（所有设备）`)
      // 清空查询条件，显示所有记录
      deviceQueryForm.value.device_did = ''
    }
  } catch (error) {
    ElMessage.error('查询失败: ' + (error.message || '未知错误'))
    deviceRecords.value = []
  } finally {
    deviceLoading.value = false
  }
}

// 从交易历史中查找
const queryFromHistory = async () => {
  deviceLoading.value = true
  try {
    // 从 localStorage 加载交易历史
    const history = localStorage.getItem('transactionHistory')
    if (!history) {
      ElMessage.warning('没有找到交易历史记录')
      deviceRecords.value = []
      return
    }

    const transactionHistory = JSON.parse(history)
    
    if (transactionHistory.length === 0) {
      ElMessage.warning('交易历史为空')
      deviceRecords.value = []
      return
    }

    // 如果指定了设备DID，只查找该设备的记录
    if (deviceQueryForm.value.device_did) {
      const filtered = transactionHistory.filter(h => 
        h.device_did === deviceQueryForm.value.device_did
      )
      
      if (filtered.length === 0) {
        ElMessage.warning(`在交易历史中未找到设备 ${deviceQueryForm.value.device_did} 的记录`)
        deviceRecords.value = []
        return
      }
      
      // 转换为认证记录格式
      deviceRecords.value = filtered.map(h => ({
        device_did: h.device_did,
        source_domain: h.source_domain,
        target_domain: h.target_domain,
        authorized: h.authorized,
        tx_hash: h.tx_hash,
        block_number: h.block_number,
        timestamp: h.timestamp
      }))
      
      ElMessage.success(`从交易历史中找到 ${deviceRecords.value.length} 条记录`)
    } else {
      // 显示所有交易历史
      deviceRecords.value = transactionHistory.map(h => ({
        device_did: h.device_did,
        source_domain: h.source_domain,
        target_domain: h.target_domain,
        authorized: h.authorized,
        tx_hash: h.tx_hash,
        block_number: h.block_number,
        timestamp: h.timestamp
      }))
      
      ElMessage.success(`从交易历史中找到 ${deviceRecords.value.length} 条记录`)
    }
    
    // 按时间倒序排列
    deviceRecords.value.sort((a, b) => {
      return new Date(b.timestamp) - new Date(a.timestamp)
    })
  } catch (error) {
    console.error('从交易历史查询失败:', error)
    ElMessage.error('查询失败: ' + (error.message || '未知错误'))
    deviceRecords.value = []
  } finally {
    deviceLoading.value = false
  }
}

// 验证交易
const verifyTransaction = async (txHash) => {
  if (!txHash || !txHash.trim()) {
    ElMessage.warning('请输入交易哈希')
    return
  }

  txLoading.value = true
  try {
    // 优先使用前端区块链服务验证（如果已连接）
    if (blockchainService.isConnected()) {
      try {
        const receipt = await blockchainService.getTransactionReceipt(txHash.trim())
        
        // 解析事件（ethers v6 格式）
        const events = []
        if (receipt.logs && receipt.logs.length > 0) {
          for (const log of receipt.logs) {
            events.push({
              address: typeof log.address === 'string' ? log.address : log.address?.toString() || '',
              topics: log.topics || [],
              data: typeof log.data === 'string' ? log.data : log.data?.toString() || '',
              block_number: typeof receipt.blockNumber === 'bigint' ? Number(receipt.blockNumber) : receipt.blockNumber,
              tx_index: receipt.index || receipt.transactionIndex || 0
            })
          }
        }
        
        // 处理状态（ethers v6 中 status 可能是 1 或 null）
        const status = receipt.status === 1 || receipt.status === '0x1' || (receipt.status !== null && receipt.status !== 0)
        
        txResult.value = {
          tx_hash: txHash.trim(),
          status: status,
          block_number: typeof receipt.blockNumber === 'bigint' ? Number(receipt.blockNumber) : receipt.blockNumber,
          gas_used: receipt.gasUsed ? (typeof receipt.gasUsed === 'bigint' ? receipt.gasUsed.toString() : String(receipt.gasUsed)) : '0',
          events: events,
          confirmations: 1
        }
        
        ElMessage.success('交易验证完成（前端验证）')
        
        // 如果是从设备记录中点击的，切换到交易标签页
        if (activeTab.value === 'device') {
          activeTab.value = 'transaction'
        }
        return
      } catch (error) {
        console.warn('前端验证失败，尝试后端验证:', error)
        // 如果前端验证失败，继续尝试后端验证
      }
    }
    
    // 使用后端API验证
    txResult.value = await authApi.verifyTransaction(txHash.trim())
    ElMessage.success('交易验证完成（后端验证）')
    
    // 如果是从设备记录中点击的，切换到交易标签页
    if (activeTab.value === 'device') {
      activeTab.value = 'transaction'
    }
  } catch (error) {
    console.error('验证失败:', error)
    
    // 改进错误提示
    let errorMsg = error.message || '验证失败'
    if (errorMsg.includes('503') || errorMsg.includes('Service Unavailable')) {
      errorMsg = '后端区块链客户端未连接，请确保后端已配置并连接区块链'
    } else if (errorMsg.includes('Blockchain client not available')) {
      errorMsg = '区块链客户端不可用，请检查后端配置'
    } else if (errorMsg.includes('Transaction not found')) {
      errorMsg = '交易未找到，请确认交易哈希是否正确'
    }
    
    ElMessage.error(errorMsg)
    txResult.value = null
  } finally {
    txLoading.value = false
  }
}

// 清空设备查询结果
const clearDeviceResults = () => {
  deviceRecords.value = []
  deviceQueryForm.value.device_did = ''
}

// 清空交易查询结果
const clearTxResults = () => {
  txResult.value = null
  txQueryForm.value.tx_hash = ''
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

// 格式化主题
const formatTopic = (topic) => {
  if (!topic) return ''
  if (typeof topic === 'string') {
    return topic.length > 20 ? `${topic.substring(0, 10)}...${topic.substring(topic.length - 8)}` : topic
  }
  return String(topic)
}

// 格式化数据
const formatData = (data) => {
  if (!data) return ''
  if (typeof data === 'string') {
    return data.length > 30 ? `${data.substring(0, 15)}...${data.substring(data.length - 10)}` : data
  }
  return String(data)
}

// 复制交易哈希
const copyTxHash = async (hash) => {
  try {
    await navigator.clipboard.writeText(hash)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 页面加载时检查 URL 参数
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const txHash = urlParams.get('txHash')
  if (txHash) {
    txQueryForm.value.tx_hash = txHash
    activeTab.value = 'transaction'
    // 延迟自动验证，确保页面已完全加载
    setTimeout(() => {
      verifyTransaction(txHash)
    }, 500)
  }
})

// 显示更新交易哈希对话框
const showUpdateTxHashDialog = (row) => {
  updateTxHashForm.value = {
    device_did: row.device_did,
    source_domain: row.source_domain,
    target_domain: row.target_domain,
    tx_hash: ''
  }
  
  // 尝试从交易历史中自动匹配
  try {
    const history = localStorage.getItem('transactionHistory')
    if (history) {
      const transactionHistory = JSON.parse(history)
      const match = transactionHistory.find(h => {
        return h.device_did === row.device_did &&
               h.source_domain === row.source_domain &&
               h.target_domain === row.target_domain &&
               h.tx_hash && h.tx_hash.trim()
      })
      if (match) {
        updateTxHashForm.value.tx_hash = match.tx_hash
      }
    }
  } catch (error) {
    console.error('加载交易历史失败:', error)
  }
  
  showUpdateDialog.value = true
}

// 更新交易哈希
const updateTxHash = async () => {
  if (!updateTxHashForm.value.tx_hash || !updateTxHashForm.value.tx_hash.trim()) {
    ElMessage.warning('请输入交易哈希')
    return
  }

  updatingTxHash.value = true
  try {
    await authApi.syncAuthRecord({
      device_did: updateTxHashForm.value.device_did,
      source_domain: updateTxHashForm.value.source_domain,
      target_domain: updateTxHashForm.value.target_domain,
      tx_hash: updateTxHashForm.value.tx_hash.trim(),
      authorized: true // 默认授权，实际应该从交易中解析
    })
    
    ElMessage.success('交易哈希已更新')
    showUpdateDialog.value = false
    
    // 重新查询以刷新数据
    await queryByDevice()
  } catch (error) {
    ElMessage.error('更新失败: ' + (error.message || '未知错误'))
  } finally {
    updatingTxHash.value = false
  }
}

// 同步缺失的交易哈希
const syncMissingTxHashes = async () => {
  if (deviceRecords.value.length === 0) {
    ElMessage.warning('没有认证记录需要同步')
    return
  }

  // 从 localStorage 加载交易历史
  let transactionHistory = []
  try {
    const history = localStorage.getItem('transactionHistory')
    if (history) {
      transactionHistory = JSON.parse(history)
    }
  } catch (error) {
    console.error('加载交易历史失败:', error)
  }

  if (transactionHistory.length === 0) {
    ElMessage.warning('没有找到交易历史记录')
    return
  }

  syncingTxHashes.value = true
  let syncedCount = 0
  let failedCount = 0

  try {
    // 找出需要同步的记录（有设备DID匹配但没有交易哈希的记录）
    const recordsToSync = deviceRecords.value.filter(record => {
      return !record.tx_hash || !record.tx_hash.trim()
    })

    if (recordsToSync.length === 0) {
      ElMessage.info('所有记录都已包含交易哈希')
      return
    }

    // 为每条记录查找匹配的交易历史
    for (const record of recordsToSync) {
      // 在交易历史中查找匹配的记录
      const historyRecord = transactionHistory.find(h => {
        return h.device_did === record.device_did &&
               h.source_domain === record.source_domain &&
               h.target_domain === record.target_domain &&
               h.tx_hash && h.tx_hash.trim()
      })

      if (historyRecord && historyRecord.tx_hash) {
        try {
          // 同步到数据库
          await authApi.syncAuthRecord({
            device_did: record.device_did,
            source_domain: record.source_domain,
            target_domain: record.target_domain,
            tx_hash: historyRecord.tx_hash,
            authorized: historyRecord.authorized !== undefined ? historyRecord.authorized : record.authorized,
            block_number: historyRecord.block_number || null
          })
          
          // 更新本地记录
          record.tx_hash = historyRecord.tx_hash
          syncedCount++
        } catch (error) {
          console.error('同步失败:', error)
          failedCount++
        }
      }
    }

    if (syncedCount > 0) {
      ElMessage.success(`成功同步 ${syncedCount} 条记录的交易哈希`)
      // 重新查询以刷新数据
      await queryByDevice()
    } else if (failedCount > 0) {
      ElMessage.warning(`同步完成，但有 ${failedCount} 条记录同步失败`)
    } else {
      ElMessage.info('未找到匹配的交易历史记录')
    }
  } catch (error) {
    ElMessage.error('同步过程出错: ' + (error.message || '未知错误'))
  } finally {
    syncingTxHashes.value = false
  }
}

// 计算统计信息
const successCount = computed(() => {
  return deviceRecords.value.filter(r => r.authorized).length
})

const failedCount = computed(() => {
  return deviceRecords.value.filter(r => !r.authorized).length
})

const onChainCount = computed(() => {
  return deviceRecords.value.filter(r => r.tx_hash && r.tx_hash.trim()).length
})
</script>

<style scoped>
.verify-page {
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

:deep(.el-button--primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

:deep(.el-tabs--border-card) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-tabs__header) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  margin: 0;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-tabs__item.is-active) {
  color: #667eea;
  font-weight: 600;
}

:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
  font-size: 13px;
}

:deep(.el-table th) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  color: #606266;
  font-weight: 600;
}

:deep(.el-table tr:hover) {
  background: #f8f9fa;
}

:deep(.el-statistic) {
  text-align: center;
}

:deep(.el-statistic__head) {
  color: #909399;
  font-size: 14px;
  margin-bottom: 8px;
}

:deep(.el-statistic__number) {
  font-weight: 600;
  color: #303133;
}

:deep(.el-tag) {
  border-radius: 12px;
  font-weight: 500;
  padding: 4px 12px;
}

:deep(.el-descriptions__label) {
  font-weight: 600;
  color: #606266;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}
</style>

