<template>
  <div class="devices-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card class="stat-mini-card" shadow="hover">
          <div class="stat-mini-content">
            <div class="stat-mini-icon total">
              <el-icon :size="24"><Monitor /></el-icon>
            </div>
            <div class="stat-mini-info">
              <div class="stat-mini-value">{{ devices.length }}</div>
              <div class="stat-mini-label">设备总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-mini-card" shadow="hover">
          <div class="stat-mini-content">
            <div class="stat-mini-icon active">
              <el-icon :size="24"><CircleCheck /></el-icon>
            </div>
            <div class="stat-mini-info">
              <div class="stat-mini-value">{{ activeCount }}</div>
              <div class="stat-mini-label">活跃设备</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-mini-card" shadow="hover">
          <div class="stat-mini-content">
            <div class="stat-mini-icon suspicious">
              <el-icon :size="24"><Warning /></el-icon>
            </div>
            <div class="stat-mini-info">
              <div class="stat-mini-value">{{ suspiciousCount }}</div>
              <div class="stat-mini-label">可疑设备</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-mini-card" shadow="hover">
          <div class="stat-mini-content">
            <div class="stat-mini-icon revoked">
              <el-icon :size="24"><CircleClose /></el-icon>
            </div>
            <div class="stat-mini-info">
              <div class="stat-mini-value">{{ revokedCount }}</div>
              <div class="stat-mini-label">已吊销</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div class="card-header">
          <span>设备管理</span>
          <el-button type="primary" @click="showRegisterDialog = true">
            <el-icon><Plus /></el-icon>
            注册设备
          </el-button>
        </div>
      </template>

      <el-table :data="devices" v-loading="loading" stripe>
        <el-table-column prop="did" label="DID" width="200" />
        <el-table-column prop="device_id" label="设备ID" />
        <el-table-column prop="device_type" label="设备类型" />
        <el-table-column prop="domain" label="所属域" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300">
          <template #default="{ row }">
            <el-button size="small" @click="viewDevice(row)">查看</el-button>
            <el-dropdown @command="(cmd) => handleStatusChange(row, cmd)" trigger="click">
              <el-button size="small" type="primary">
                状态管理 <el-icon><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item 
                    command="active" 
                    :disabled="row.status === 'active'"
                  >
                    <el-icon><CircleCheck /></el-icon> 激活
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="suspicious" 
                    :disabled="row.status === 'suspicious'"
                  >
                    <el-icon><Warning /></el-icon> 标记为可疑
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="revoked" 
                    :disabled="row.status === 'revoked'"
                    divided
                  >
                    <el-icon><CircleClose /></el-icon> 吊销
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 注册设备对话框 -->
    <el-dialog v-model="showRegisterDialog" title="注册设备" width="600px">
      <el-form :model="deviceForm" label-width="100px">
        <el-form-item label="DID" required>
          <el-input v-model="deviceForm.did" />
        </el-form-item>
        <el-form-item label="设备ID" required>
          <el-input v-model="deviceForm.device_id" />
        </el-form-item>
        <el-form-item label="设备类型">
          <el-input v-model="deviceForm.device_type" />
        </el-form-item>
        <el-form-item label="所属域" required>
          <el-input v-model="deviceForm.domain" />
        </el-form-item>
        
        <!-- 区块链注册选项 -->
        <el-divider>区块链注册（可选）</el-divider>
        <el-form-item label="同时上链">
          <el-switch 
            v-model="deviceForm.registerToBlockchain" 
            :disabled="!blockchainConnected"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 5px">
            <span v-if="blockchainConnected">
              已连接区块链，可以同时注册到区块链
            </span>
            <span v-else>
              未连接区块链，请先在"跨域认证"页面连接区块链
            </span>
          </div>
        </el-form-item>
        <el-form-item 
          v-if="deviceForm.registerToBlockchain && blockchainConnected" 
          label="设备元数据"
        >
          <el-input 
            v-model="deviceForm.blockchainMetadata" 
            type="textarea" 
            :rows="3"
            placeholder='{"type": "sensor", "location": "room1"}'
          />
          <div style="font-size: 12px; color: #909399; margin-top: 5px">
            可选，JSON格式字符串，将存储到区块链
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRegisterDialog = false">取消</el-button>
        <el-button type="primary" @click="registerDevice" :loading="registerLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 设备详情对话框 -->
    <el-dialog 
      v-model="showDeviceDetail" 
      title="设备详情" 
      width="900px"
      :close-on-click-modal="false"
    >
      <div v-if="loadingDetail" style="text-align: center; padding: 40px">
        <el-icon class="is-loading" :size="40"><Loading /></el-icon>
        <p style="margin-top: 20px; color: #909399">加载中...</p>
      </div>
      
      <div v-else-if="deviceDetail">
        <el-tabs v-model="detailTab" type="border-card">
          <!-- 基本信息 -->
          <el-tab-pane label="基本信息" name="info">
            <el-descriptions :column="2" border style="margin-top: 20px">
              <el-descriptions-item label="设备DID" :span="2">
                <span style="font-family: monospace; font-weight: 600">{{ deviceDetail.did }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="设备ID">
                {{ deviceDetail.device_id || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="设备类型">
                {{ deviceDetail.device_type || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="制造商">
                {{ deviceDetail.manufacturer || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="型号">
                {{ deviceDetail.model || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="固件版本">
                {{ deviceDetail.firmware || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="所属域">
                <el-tag type="info">{{ deviceDetail.domain || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="设备状态">
                <el-tag :type="getStatusType(deviceDetail.status)">
                  {{ getStatusText(deviceDetail.status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="所有者">
                {{ deviceDetail.owner || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="注册时间">
                {{ formatDate(deviceDetail.registered_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="最后更新">
                {{ formatDate(deviceDetail.last_updated) }}
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ formatDate(deviceDetail.created_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="元数据" :span="2">
                <pre v-if="deviceDetail.metadata" style="background: #f5f7fa; padding: 10px; border-radius: 4px; max-height: 200px; overflow: auto; font-size: 12px">{{ 
                  JSON.stringify(JSON.parse(deviceDetail.metadata || '{}'), null, 2) 
                }}</pre>
                <span v-else style="color: #909399">无</span>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 区块链信息 -->
          <el-tab-pane label="区块链信息" name="blockchain">
            <div v-if="loadingBlockchain" style="text-align: center; padding: 40px">
              <el-icon class="is-loading" :size="30"><Loading /></el-icon>
              <p style="margin-top: 15px; color: #909399">加载中...</p>
            </div>
            <div v-else-if="blockchainDeviceInfo" style="margin-top: 20px">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="设备DID" :span="2">
                  <span style="font-family: monospace">{{ blockchainDeviceInfo.did }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="设备状态">
                  <el-tag :type="blockchainDeviceInfo.status === 0 ? 'success' : blockchainDeviceInfo.status === 1 ? 'warning' : 'danger'">
                    {{ blockchainDeviceInfo.status === 0 ? '活跃' : blockchainDeviceInfo.status === 1 ? '可疑' : '已吊销' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="所有者地址">
                  <span style="font-family: monospace; font-size: 12px">{{ blockchainDeviceInfo.owner }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="注册时间">
                  {{ new Date(blockchainDeviceInfo.registeredAt * 1000).toLocaleString('zh-CN') }}
                </el-descriptions-item>
                <el-descriptions-item label="最后更新">
                  {{ new Date(blockchainDeviceInfo.lastUpdated * 1000).toLocaleString('zh-CN') }}
                </el-descriptions-item>
                <el-descriptions-item label="元数据" :span="2">
                  <pre style="background: #f5f7fa; padding: 10px; border-radius: 4px; max-height: 200px; overflow: auto; font-size: 12px">{{ 
                    JSON.stringify(JSON.parse(blockchainDeviceInfo.metadata || '{}'), null, 2) 
                  }}</pre>
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <el-empty v-else description="设备未在区块链上注册或区块链未连接" />
          </el-tab-pane>

          <!-- 操作历史 -->
          <el-tab-pane label="操作历史" name="history">
            <div v-if="loadingHistory" style="text-align: center; padding: 40px">
              <el-icon class="is-loading" :size="30"><Loading /></el-icon>
              <p style="margin-top: 15px; color: #909399">加载中...</p>
            </div>
            <div v-else-if="deviceHistory.length > 0" style="margin-top: 20px">
              <el-timeline>
                <el-timeline-item
                  v-for="(item, index) in deviceHistory"
                  :key="index"
                  :timestamp="formatDate(item.created_at)"
                  placement="top"
                  :type="item.action === 'revoke' ? 'danger' : item.action === 'register' ? 'success' : 'primary'"
                >
                  <el-card shadow="hover">
                    <div style="display: flex; justify-content: space-between; align-items: start">
                      <div>
                        <h4 style="margin: 0 0 10px 0; color: #303133">
                          <el-icon style="vertical-align: middle; margin-right: 5px">
                            <Document v-if="item.action === 'register'" />
                            <Warning v-else-if="item.action === 'status_change'" />
                            <CircleClose v-else-if="item.action === 'revoke'" />
                            <Edit v-else />
                          </el-icon>
                          {{ getActionText(item.action) }}
                        </h4>
                        <p style="margin: 5px 0; color: #606266; font-size: 13px">
                          {{ item.description || '无描述' }}
                        </p>
                        <div v-if="item.old_value && item.new_value" style="margin-top: 10px; font-size: 12px; color: #909399">
                          <div>变更前: <code>{{ item.old_value }}</code></div>
                          <div style="margin-top: 5px">变更后: <code>{{ item.new_value }}</code></div>
                        </div>
                      </div>
                      <div style="text-align: right">
                        <el-tag size="small" type="info">{{ item.changed_by || '系统' }}</el-tag>
                        <div v-if="item.tx_hash" style="margin-top: 5px">
                          <el-link 
                            type="primary" 
                            :underline="false" 
                            style="font-size: 11px; font-family: monospace"
                            @click="copyText(item.tx_hash)"
                          >
                            <el-icon><Link /></el-icon>
                            交易哈希
                          </el-link>
                        </div>
                      </div>
                    </div>
                  </el-card>
                </el-timeline-item>
              </el-timeline>
            </div>
            <el-empty v-else description="暂无操作历史" />
          </el-tab-pane>

          <!-- 认证记录 -->
          <el-tab-pane label="认证记录" name="auth">
            <div v-if="loadingAuthRecords" style="text-align: center; padding: 40px">
              <el-icon class="is-loading" :size="30"><Loading /></el-icon>
              <p style="margin-top: 15px; color: #909399">加载中...</p>
            </div>
            <div v-else-if="deviceAuthRecords.length > 0" style="margin-top: 20px">
              <el-table :data="deviceAuthRecords" stripe border>
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
                    <div v-if="row.tx_hash && row.tx_hash !== '未上链'">
                      <el-link 
                        type="primary" 
                        :underline="false" 
                        style="font-family: monospace; font-size: 12px"
                        @click="copyText(row.tx_hash)"
                      >
                        {{ row.tx_hash.substring(0, 20) }}...
                      </el-link>
                    </div>
                    <span v-else style="color: #909399">未上链</span>
                  </template>
                </el-table-column>
                <el-table-column prop="timestamp" label="认证时间" width="180">
                  <template #default="{ row }">
                    {{ formatDate(row.timestamp) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
            <el-empty v-else description="暂无认证记录" />
          </el-tab-pane>
        </el-tabs>
      </div>
      
      <template #footer>
        <el-button @click="showDeviceDetail = false">关闭</el-button>
        <el-button 
          v-if="deviceDetail" 
          type="primary" 
          @click="viewDevice(deviceDetail); loadDeviceHistory(deviceDetail.did); loadDeviceAuthRecords(deviceDetail.did)"
        >
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </template>
    </el-dialog>

    <!-- 区块链注册结果 -->
    <el-dialog v-model="showBlockchainResult" title="区块链注册结果" width="500px">
      <div v-if="blockchainRegisterResult">
        <el-alert
          :title="blockchainRegisterResult.success ? '区块链注册成功' : '区块链注册失败'"
          :type="blockchainRegisterResult.success ? 'success' : 'error'"
          :closable="false"
          show-icon
        >
          <template #default v-if="blockchainRegisterResult.success">
            <p><strong>交易哈希:</strong> 
              <el-link 
                type="primary" 
                @click="copyText(blockchainRegisterResult.txHash)"
                :underline="false"
              >
                {{ blockchainRegisterResult.txHash }}
              </el-link>
            </p>
            <p><strong>区块号:</strong> {{ blockchainRegisterResult.blockNumber }}</p>
          </template>
          <template #default v-else>
            <p><strong>错误:</strong> {{ blockchainRegisterResult.error }}</p>
          </template>
        </el-alert>
      </div>
      <template #footer>
        <el-button type="primary" @click="showBlockchainResult = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, CircleCheck, CircleClose, Warning, Plus, Monitor, Link, Document, Loading, Edit, Refresh } from '@element-plus/icons-vue'
import api from '../api'
import deviceApi from '../api/device'
import authApi from '../api/auth'
import blockchainService from '../services/blockchain'

const devices = ref([])
const loading = ref(false)
const showRegisterDialog = ref(false)
const registerLoading = ref(false)
const showBlockchainResult = ref(false)
const blockchainRegisterResult = ref(null)

const deviceForm = ref({
  did: '',
  device_id: '',
  device_type: '',
  domain: '',
  registerToBlockchain: false,
  blockchainMetadata: '{}'
})

// 检查区块链连接状态
const blockchainConnected = computed(() => {
  return blockchainService.isConnected()
})

// 统计不同状态的设备数量
const activeCount = computed(() => {
  return devices.value.filter(d => d.status === 'active').length
})

const suspiciousCount = computed(() => {
  return devices.value.filter(d => d.status === 'suspicious').length
})

const revokedCount = computed(() => {
  return devices.value.filter(d => d.status === 'revoked').length
})

const loadDevices = async () => {
  loading.value = true
  try {
    devices.value = await deviceApi.list()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}

const registerDevice = async () => {
  if (!deviceForm.value.did || !deviceForm.value.device_id || !deviceForm.value.domain) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  registerLoading.value = true
  blockchainRegisterResult.value = null

  try {
    // 1. 先注册到数据库
    await deviceApi.register({
      did: deviceForm.value.did,
      device_id: deviceForm.value.device_id,
      device_type: deviceForm.value.device_type,
      domain: deviceForm.value.domain
    })
    
    ElMessage.success('设备已注册到数据库')

    // 2. 如果选择了注册到区块链，则注册到区块链
    if (deviceForm.value.registerToBlockchain && blockchainService.isConnected()) {
      try {
        let metadata = deviceForm.value.blockchainMetadata.trim() || '{}'
        
        // 验证 JSON 格式
        try {
          JSON.parse(metadata)
        } catch (e) {
          ElMessage.warning('设备元数据格式错误，使用默认值')
          metadata = '{}'
        }

        const result = await blockchainService.registerDevice(
          deviceForm.value.did,
          metadata
        )

        blockchainRegisterResult.value = {
          success: true,
          ...result
        }
        
        showBlockchainResult.value = true
        ElMessage.success('设备已同时注册到区块链')
      } catch (error) {
        console.error('区块链注册失败:', error)
        
        let errorMsg = error.message || '区块链注册失败'
        if (errorMsg.includes('Device already exists')) {
          errorMsg = '设备已在区块链上存在'
        } else if (errorMsg.includes('Only authorized admin')) {
          errorMsg = '当前账户没有区块链注册权限，请使用管理员账户'
        }
        
        blockchainRegisterResult.value = {
          success: false,
          error: errorMsg
        }
        showBlockchainResult.value = true
        ElMessage.warning(`数据库注册成功，但区块链注册失败: ${errorMsg}`)
      }
    }

    // 3. 刷新设备列表
    showRegisterDialog.value = false
    loadDevices()
    
    // 4. 重置表单
    deviceForm.value = {
      did: '',
      device_id: '',
      device_type: '',
      domain: '',
      registerToBlockchain: false,
      blockchainMetadata: '{}'
    }
  } catch (error) {
    ElMessage.error(error.message || '设备注册失败')
  } finally {
    registerLoading.value = false
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

const showDeviceDetail = ref(false)
const deviceDetail = ref(null)
const deviceHistory = ref([])
const deviceAuthRecords = ref([])
const loadingDetail = ref(false)
const loadingHistory = ref(false)
const loadingAuthRecords = ref(false)
const blockchainDeviceInfo = ref(null)
const loadingBlockchain = ref(false)
const detailTab = ref('info')

const viewDevice = async (device) => {
  showDeviceDetail.value = true
  deviceDetail.value = null
  deviceHistory.value = []
  deviceAuthRecords.value = []
  blockchainDeviceInfo.value = null
  
  // 加载设备详情
  loadingDetail.value = true
  try {
    deviceDetail.value = await deviceApi.get(device.did)
  } catch (error) {
    ElMessage.error('加载设备详情失败: ' + error.message)
    loadingDetail.value = false
    return
  } finally {
    loadingDetail.value = false
  }
  
  // 并行加载历史记录和认证记录
  loadDeviceHistory(device.did)
  loadDeviceAuthRecords(device.did)
  
  // 如果区块链已连接，加载区块链信息
  if (blockchainService.isConnected()) {
    loadBlockchainDeviceInfo(device.did)
  }
}

const loadDeviceHistory = async (did) => {
  loadingHistory.value = true
  try {
    const response = await deviceApi.getHistory(did, { page: 1, page_size: 20 })
    deviceHistory.value = response.history || []
  } catch (error) {
    console.error('加载设备历史失败:', error)
    deviceHistory.value = []
  } finally {
    loadingHistory.value = false
  }
}

const loadDeviceAuthRecords = async (did) => {
  loadingAuthRecords.value = true
  try {
    deviceAuthRecords.value = await authApi.getAuthRecords(did)
  } catch (error) {
    console.error('加载认证记录失败:', error)
    deviceAuthRecords.value = []
  } finally {
    loadingAuthRecords.value = false
  }
}

const loadBlockchainDeviceInfo = async (did) => {
  loadingBlockchain.value = true
  try {
    const device = await blockchainService.getDevice(did)
    blockchainDeviceInfo.value = device
  } catch (error) {
    console.error('加载区块链设备信息失败:', error)
    blockchainDeviceInfo.value = null
  } finally {
    loadingBlockchain.value = false
  }
}

const formatDate = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getStatusText = (status) => {
  const map = {
    active: '活跃',
    suspicious: '可疑',
    revoked: '已吊销'
  }
  return map[status] || status
}

const getActionText = (action) => {
  const map = {
    register: '注册',
    status_change: '状态变更',
    revoke: '吊销',
    update: '更新'
  }
  return map[action] || action
}

// 处理状态变更
const handleStatusChange = async (device, newStatus) => {
  const statusMap = {
    active: '激活',
    suspicious: '标记为可疑',
    revoked: '吊销'
  }
  
  const statusText = statusMap[newStatus] || newStatus
  
  try {
    let confirmMessage = ''
    if (newStatus === 'revoked') {
      confirmMessage = `确定要吊销该设备吗？吊销后设备将失去所有认证特权，无法进行跨域认证。`
    } else if (newStatus === 'active') {
      confirmMessage = `确定要激活该设备吗？激活后设备可以正常进行跨域认证。`
    } else if (newStatus === 'suspicious') {
      confirmMessage = `确定要将该设备标记为可疑吗？标记后设备将暂时失去认证特权。`
    }
    
    await ElMessageBox.confirm(confirmMessage, `确认${statusText}`, {
      type: newStatus === 'revoked' ? 'warning' : 'info',
    })
    
    if (newStatus === 'revoked') {
      // 使用吊销接口
      await deviceApi.revoke(device.did)
      ElMessage.success('设备已吊销，将失去所有认证特权')
    } else {
      // 使用状态更新接口
      await deviceApi.updateStatus(device.did, newStatus)
      ElMessage.success(`设备状态已更新为：${statusText}`)
    }
    
    loadDevices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '状态更新失败')
    }
  }
}

const revokeDevice = async (device) => {
  await handleStatusChange(device, 'revoked')
}

const getStatusType = (status) => {
  const map = {
    active: 'success',
    suspicious: 'warning',
    revoked: 'danger',
  }
  return map[status] || 'info'
}

onMounted(() => {
  loadDevices()
})
</script>

<style scoped>
.devices-page {
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
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

:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-table th) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  color: #606266;
  font-weight: 600;
}

:deep(.el-table tr:hover) {
  background: #f8f9fa;
}

:deep(.el-tag) {
  border-radius: 12px;
  font-weight: 500;
  padding: 4px 12px;
}

:deep(.el-dialog) {
  border-radius: 12px;
}

:deep(.el-dialog__header) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  padding: 20px;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}

/* 统计卡片 */
.stat-mini-card {
  border-radius: 12px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.stat-mini-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

.stat-mini-content {
  display: flex;
  align-items: center;
  padding: 5px 0;
}

.stat-mini-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  transition: all 0.3s ease;
}

.stat-mini-icon.total {
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
  color: #667eea;
}

.stat-mini-icon.active {
  background: linear-gradient(135deg, #67c23a15 0%, #85ce6115 100%);
  color: #67c23a;
}

.stat-mini-icon.suspicious {
  background: linear-gradient(135deg, #e6a23c15 0%, #ebb56315 100%);
  color: #e6a23c;
}

.stat-mini-icon.revoked {
  background: linear-gradient(135deg, #f56c6c15 0%, #f7898915 100%);
  color: #f56c6c;
}

.stat-mini-card:hover .stat-mini-icon {
  transform: scale(1.1);
}

.stat-mini-info {
  flex: 1;
}

.stat-mini-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 5px;
}

.stat-mini-label {
  font-size: 13px;
  color: #909399;
  font-weight: 500;
}

:deep(.devices-table) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.devices-table .el-table__row) {
  transition: all 0.3s ease;
}

:deep(.devices-table .el-table__row:hover) {
  background: #f8f9fa;
  transform: scale(1.01);
}
</style>

