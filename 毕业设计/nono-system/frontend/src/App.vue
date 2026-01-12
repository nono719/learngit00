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
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor, Connection, Lock, Document, User, ArrowDown, Search, DataAnalysis } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const activeMenu = computed(() => route.path)

const currentUser = ref(null)
const isLoggedIn = computed(() => {
  const token = localStorage.getItem('token')
  if (token && !currentUser.value) {
    // 从 localStorage 加载用户信息
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
    }
  }
  return !!token
})

onMounted(() => {
  // 加载用户信息
  const userStr = localStorage.getItem('user')
  if (userStr) {
    currentUser.value = JSON.parse(userStr)
  }
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
.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 20px;
}

.header-content h1 {
  margin: 0;
  color: #409eff;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
}

.user-dropdown .el-icon {
  margin-right: 5px;
}

.sidebar-menu {
  height: 100%;
}
</style>

