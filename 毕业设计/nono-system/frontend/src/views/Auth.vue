<template>
  <div class="auth-page">
    <el-card>
      <template #header>
        <span>跨域认证</span>
      </template>

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
            发起认证请求
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider />

      <div v-if="authResult">
        <el-alert
          :title="authResult.authorized ? '认证成功' : '认证失败'"
          :type="authResult.authorized ? 'success' : 'error'"
          :closable="false"
          show-icon
        >
          <template #default>
            <p>设备DID: {{ authResult.device_did }}</p>
            <p>源域: {{ authResult.source_domain }}</p>
            <p>目标域: {{ authResult.target_domain }}</p>
            <p>授权状态: {{ authResult.authorized ? '已授权' : '未授权' }}</p>
          </template>
        </el-alert>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const authForm = ref({
  device_did: '',
  source_domain: '',
  target_domain: '',
})

const authResult = ref(null)
const loading = ref(false)

const requestAuth = async () => {
  if (!authForm.value.device_did || !authForm.value.source_domain || !authForm.value.target_domain) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  loading.value = true
  try {
    const result = await api.post('/auth/cross-domain', authForm.value)
    authResult.value = {
      ...authForm.value,
      authorized: result.authorized,
    }
    ElMessage.success('认证请求已提交')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  padding: 20px;
}
</style>

