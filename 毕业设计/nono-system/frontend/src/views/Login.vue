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
      
      // 确保 token 是字符串格式
      const token = String(result.token || result.user?.id || '')
      if (!token) {
        throw new Error('登录失败：未收到有效的 token')
      }
      
      // 保存 token 和用户信息
      localStorage.setItem('token', token)
      if (result.user) {
        localStorage.setItem('user', JSON.stringify(result.user))
      }
      
      // 触发自定义事件，通知 App 组件更新登录状态
      window.dispatchEvent(new Event('login-status-changed'))
      
      ElMessage.success('登录成功')
      
      // 跳转到首页
      router.push('/')
    } catch (error) {
      console.error('登录错误:', error)
      const errorMsg = error.message || '登录失败'
      if (errorMsg.includes('Invalid username or password')) {
        ElMessage.error('用户名或密码错误')
      } else if (errorMsg.includes('Unauthorized')) {
        ElMessage.error('用户名或密码错误')
      } else {
        ElMessage.error(errorMsg)
      }
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
  position: relative;
  overflow: hidden;
}

.login-container::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.1) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: move 20s linear infinite;
}

@keyframes move {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(50px, 50px);
  }
}

.login-card {
  width: 100%;
  max-width: 480px;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
  position: relative;
  z-index: 1;
  animation: fadeInUp 0.6s ease;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

:deep(.el-card__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px 16px 0 0;
  padding: 30px;
  border: none;
}

.login-header {
  text-align: center;
}

.login-header h2 {
  margin: 0 0 10px 0;
  color: #ffffff;
  font-size: 24px;
  font-weight: 600;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.login-header p {
  margin: 0;
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
}

:deep(.el-card__body) {
  padding: 40px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  height: 44px;
  font-weight: 500;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

:deep(.el-button--primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

:deep(.el-divider__text) {
  background: rgba(255, 255, 255, 0.95);
  color: #909399;
  font-size: 12px;
}

.quick-login {
  text-align: center;
  padding-top: 20px;
}

.quick-login p {
  margin-bottom: 15px;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-button--small) {
  border-radius: 6px;
  transition: all 0.3s ease;
}

:deep(.el-button--small:hover) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>

