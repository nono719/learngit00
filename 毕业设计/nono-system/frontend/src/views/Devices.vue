<template>
  <div class="devices-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>设备管理</span>
          <el-button type="primary" @click="showRegisterDialog = true">
            注册设备
          </el-button>
        </div>
      </template>

      <el-table :data="devices" v-loading="loading">
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
import { ArrowDown, CircleCheck, CircleClose, Warning } from '@element-plus/icons-vue'
import api from '../api'
import deviceApi from '../api/device'
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

const viewDevice = (device) => {
  // TODO: 实现查看设备详情
  ElMessage.info('查看设备详情功能待实现')
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
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

