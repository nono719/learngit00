import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截器 - 添加认证 token
api.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    // 处理 401 未授权错误
    if (error.response?.status === 401) {
      // 清除 token 并跳转到登录页
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // 如果不在登录页，则跳转
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
        ElMessage.error('登录已过期，请重新登录')
      }
    }
    const message = error.response?.data?.error || error.message || '请求失败'
    return Promise.reject(new Error(message))
  }
)

export default api

