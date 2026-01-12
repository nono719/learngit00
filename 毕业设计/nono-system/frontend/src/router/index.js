import { createRouter, createWebHistory } from 'vue-router'
import Devices from '../views/Devices.vue'
import Domains from '../views/Domains.vue'
import Auth from '../views/Auth.vue'
import Logs from '../views/Logs.vue'
import Verify from '../views/Verify.vue'
import Oracle from '../views/Oracle.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    redirect: '/devices',
  },
  {
    path: '/devices',
    name: 'Devices',
    component: Devices,
    meta: { requiresAuth: true },
  },
  {
    path: '/domains',
    name: 'Domains',
    component: Domains,
    meta: { requiresAuth: true },
  },
  {
    path: '/auth',
    name: 'Auth',
    component: Auth,
    meta: { requiresAuth: true },
  },
  {
    path: '/logs',
    name: 'Logs',
    component: Logs,
    meta: { requiresAuth: true },
  },
  {
    path: '/verify',
    name: 'Verify',
    component: Verify,
    meta: { requiresAuth: true },
  },
  {
    path: '/oracle',
    name: 'Oracle',
    component: Oracle,
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫 - 检查认证
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  // 如果访问登录页且已登录，跳转到首页
  if (to.path === '/login' && token) {
    next('/')
    return
  }
  
  // 如果需要认证但未登录，跳转到登录页
  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }
  
  next()
})

export default router

