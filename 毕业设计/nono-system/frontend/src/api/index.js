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
      const errorMsg = error.response?.data?.error || '未授权'
      
      // 清除 token 和用户信息
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      
      // 如果不在登录页，则跳转
      if (window.location.pathname !== '/login') {
        // 延迟跳转，确保消息能显示
        setTimeout(() => {
          window.location.href = '/login'
        }, 100)
        
        // 根据错误类型显示不同的提示
        if (errorMsg.includes('Authorization header required')) {
          ElMessage.error('未登录，请先登录')
        } else if (errorMsg.includes('Invalid token')) {
          ElMessage.error('登录已过期或无效，请重新登录')
        } else {
          ElMessage.error('登录已过期，请重新登录')
        }
      } else {
        // 在登录页时也显示错误
        if (errorMsg.includes('Invalid username or password')) {
          // 登录页的错误由登录组件自己处理
        } else {
          ElMessage.warning('请先登录')
        }
      }
    }
    
    const message = error.response?.data?.error || error.message || '请求失败'
    return Promise.reject(new Error(message))
  }
)

export default api

