<template>
  <div class="logs-page">
    <el-card>
      <template #header>
        <span>认证日志</span>
      </template>

      <el-table :data="logs" v-loading="loading">
        <el-table-column prop="device_did" label="设备DID" width="200" />
        <el-table-column prop="source_domain" label="源域" />
        <el-table-column prop="target_domain" label="目标域" />
        <el-table-column prop="action" label="操作">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="消息" />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const logs = ref([])
const loading = ref(false)

const loadLogs = async () => {
  loading.value = true
  try {
    logs.value = await api.get('/auth/logs')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}

const getActionType = (action) => {
  const map = {
    success: 'success',
    failed: 'danger',
    request: 'info',
  }
  return map[action] || 'info'
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.logs-page {
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
</style>

