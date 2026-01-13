<template>
  <div id="app">
    <el-container v-if="isLoggedIn">
      <el-header>
        <div class="header-content">
          <h1>物联网设备跨域身份认证系统</h1>
          <div class="user-info">
            <el-dropdown @command="handleCommand">
              <span class="user-dropdown">
                <el-icon><User /></el-icon>
                {{ currentUser?.username || '用户' }}
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item disabled>
                    <span>角色: {{ getRoleName(currentUser?.role) }}</span>
                  </el-dropdown-item>
                  <el-dropdown-item v-if="currentUser?.domain" disabled>
                    <span>域: {{ currentUser.domain }}</span>
                  </el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-header>
      <el-container>
        <el-aside width="200px">
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
          >
            <el-menu-item index="/dashboard">
              <el-icon><DataAnalysis /></el-icon>
              <span>仪表板</span>
            </el-menu-item>
            <el-menu-item index="/devices">
              <el-icon><Monitor /></el-icon>
              <span>设备管理</span>
            </el-menu-item>
            <el-menu-item index="/domains">
              <el-icon><Connection /></el-icon>
              <span>域管理</span>
            </el-menu-item>
            <el-menu-item index="/auth">
              <el-icon><Lock /></el-icon>
              <span>跨域认证</span>
            </el-menu-item>
            <el-menu-item index="/logs">
              <el-icon><Document /></el-icon>
              <span>认证日志</span>
            </el-menu-item>
            <el-menu-item index="/verify">
              <el-icon><Search /></el-icon>
              <span>认证验证</span>
            </el-menu-item>
            <el-menu-item index="/oracle">
              <el-icon><DataAnalysis /></el-icon>
              <span>预言机服务</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
    <router-view v-else />
  </div>
</template>

<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor, Connection, Lock, Document, User, ArrowDown, Search, DataAnalysis } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const activeMenu = computed(() => route.path)

const currentUser = ref(null)
const isLoggedIn = ref(false)

// 检查登录状态的函数
const checkLoginStatus = () => {
  const token = localStorage.getItem('token')
  isLoggedIn.value = !!token
  
  if (token) {
    // 从 localStorage 加载用户信息
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        currentUser.value = JSON.parse(userStr)
      } catch (e) {
        console.error('解析用户信息失败:', e)
        currentUser.value = null
      }
    }
  } else {
    currentUser.value = null
  }
}

// 监听路由变化，更新登录状态
watch(() => route.path, () => {
  checkLoginStatus()
}, { immediate: true })

// 监听 localStorage 变化（跨标签页同步）
window.addEventListener('storage', (e) => {
  if (e.key === 'token' || e.key === 'user') {
    checkLoginStatus()
  }
})

// 监听登录状态变化事件（同标签页内）
window.addEventListener('login-status-changed', () => {
  checkLoginStatus()
})

onMounted(() => {
  // 初始化时检查登录状态
  checkLoginStatus()
})

const getRoleName = (role) => {
  const roleMap = {
    'admin': '系统管理员',
    'operator': '系统操作人员',
    'oracle': '预言机节点',
    'auditor': '管理/审计人员',
    'user': '普通用户',
  }
  return roleMap[role] || role
}

const handleCommand = (command) => {
  if (command === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    currentUser.value = null
    ElMessage.success('已退出登录')
    router.push('/login')
  }
}
</script>

<style scoped>
#app {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

:deep(.el-header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 0;
  height: 64px !important;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 30px;
}

.header-content h1 {
  margin: 0;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.user-info {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #ffffff;
  padding: 8px 16px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.user-dropdown:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.user-dropdown .el-icon {
  margin-right: 8px;
  font-size: 18px;
}

:deep(.el-aside) {
  background: #ffffff;
  box-shadow: 2px 0 8px 0 rgba(0, 0, 0, 0.05);
  border-right: 1px solid #e4e7ed;
}

.sidebar-menu {
  height: 100%;
  border-right: none;
  background: transparent;
}

:deep(.el-menu-item) {
  margin: 8px 12px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

:deep(.el-menu-item:hover) {
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
  transform: translateX(4px);
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  font-weight: 500;
}

:deep(.el-menu-item.is-active .el-icon) {
  color: #ffffff;
}

:deep(.el-main) {
  padding: 24px;
  background: transparent;
}

:deep(.el-card) {
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: none;
  transition: all 0.3s ease;
}

:deep(.el-card:hover) {
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

:deep(.el-card__header) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 1px solid #e9ecef;
  padding: 20px;
  font-weight: 600;
  color: #2c3e50;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

