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
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="viewDevice(row)">查看</el-button>
            <el-button
              size="small"
              type="danger"
              @click="revokeDevice(row)"
              :disabled="row.status === 'revoked'"
            >
              吊销
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 注册设备对话框 -->
    <el-dialog v-model="showRegisterDialog" title="注册设备" width="500px">
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
      </el-form>
      <template #footer>
        <el-button @click="showRegisterDialog = false">取消</el-button>
        <el-button type="primary" @click="registerDevice">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'
import deviceApi from '../api/device'

const devices = ref([])
const loading = ref(false)
const showRegisterDialog = ref(false)
const deviceForm = ref({
  did: '',
  device_id: '',
  device_type: '',
  domain: '',
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
  try {
    await deviceApi.register(deviceForm.value)
    ElMessage.success('设备注册成功')
    showRegisterDialog.value = false
    loadDevices()
  } catch (error) {
    ElMessage.error(error.message)
  }
}

const viewDevice = (device) => {
  // TODO: 实现查看设备详情
  ElMessage.info('查看设备详情功能待实现')
}

const revokeDevice = async (device) => {
  try {
    await ElMessageBox.confirm('确定要吊销该设备吗？', '提示', {
      type: 'warning',
    })
    await deviceApi.revoke(device.did)
    ElMessage.success('设备已吊销')
    loadDevices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message)
    }
  }
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

