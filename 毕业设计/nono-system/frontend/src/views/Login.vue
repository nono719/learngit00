<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <h2>物联网设备身份认证系统</h2>
          <p>请登录以继续</p>
        </div>
      </template>

      <el-form :model="loginForm" :rules="rules" ref="loginFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            @keyup.enter="handleLogin"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading" style="width: 100%">
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider>或</el-divider>

      <div class="quick-login">
        <p>快速登录（测试账号）：</p>
        <el-space wrap>
          <el-button size="small" @click="quickLogin('admin', 'admin123')">
            系统管理员
          </el-button>
          <el-button size="small" @click="quickLogin('operator1', 'operator123')">
            操作人员
          </el-button>
          <el-button size="small" @click="quickLogin('auditor1', 'auditor123')">
            审计人员
          </el-button>
        </el-space>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import authApi from '../api/auth'

const router = useRouter()

const loginForm = reactive({
  username: '',
  password: '',
})

const loginFormRef = ref(null)
const loading = ref(false)

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const result = await authApi.login(loginForm)
      
      // 保存 token 和用户信息
      localStorage.setItem('token', result.token)
      localStorage.setItem('user', JSON.stringify(result.user))
      
      ElMessage.success('登录成功')
      
      // 跳转到首页
      router.push('/')
    } catch (error) {
      ElMessage.error(error.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}

const quickLogin = async (username, password) => {
  loginForm.username = username
  loginForm.password = password
  await handleLogin()
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 100%;
  max-width: 450px;
}

.login-header {
  text-align: center;
}

.login-header h2 {
  margin: 0 0 10px 0;
  color: #303133;
}

.login-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.quick-login {
  text-align: center;
  padding-top: 20px;
}

.quick-login p {
  margin-bottom: 10px;
  color: #909399;
  font-size: 14px;
}
</style>

