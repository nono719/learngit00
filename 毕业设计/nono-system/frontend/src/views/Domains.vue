<template>
  <div class="domains-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>域管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            创建域
          </el-button>
        </div>
      </template>

      <el-table :data="domains" v-loading="loading">
        <el-table-column prop="name" label="域名" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="owner" label="所有者" />
        <el-table-column prop="created_at" label="创建时间" />
      </el-table>
    </el-card>

    <!-- 创建域对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建域" width="500px">
      <el-form :model="domainForm" label-width="100px">
        <el-form-item label="域名" required>
          <el-input v-model="domainForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="domainForm.description" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createDomain">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const domains = ref([])
const loading = ref(false)
const showCreateDialog = ref(false)
const domainForm = ref({
  name: '',
  description: '',
})

const loadDomains = async () => {
  loading.value = true
  try {
    domains.value = await api.get('/domains')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}

const createDomain = async () => {
  try {
    await api.post('/domains', domainForm.value)
    ElMessage.success('域创建成功')
    showCreateDialog.value = false
    loadDomains()
  } catch (error) {
    ElMessage.error(error.message)
  }
}

onMounted(() => {
  loadDomains()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

