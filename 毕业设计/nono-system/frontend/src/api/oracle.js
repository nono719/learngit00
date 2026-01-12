import axios from 'axios'

// 预言机服务API
// 优先使用Vite代理（/oracle-api），如果代理不工作，可以改为直接访问
const ORACLE_BASE_URL = import.meta.env.VITE_ORACLE_API_URL || '/oracle-api'
const oracleApi = axios.create({
  baseURL: ORACLE_BASE_URL,
  timeout: 10000,
})

console.log('Oracle API baseURL:', ORACLE_BASE_URL)

// 响应拦截器 - 处理响应数据
oracleApi.interceptors.response.use(
  (response) => {
    // 直接返回响应数据
    return response.data || response
  },
  (error) => {
    // 处理错误
    const errorInfo = {
      url: error.config?.url,
      baseURL: error.config?.baseURL,
      fullURL: error.config?.baseURL + error.config?.url,
      status: error.response?.status,
      statusText: error.response?.statusText,
      message: error.message,
      code: error.code
    }
    console.error('Oracle API Error:', errorInfo)
    
    // 如果是404，提供更详细的错误信息
    if (error.response?.status === 404) {
      error.message = `预言机API未找到 (404): ${errorInfo.fullURL || errorInfo.url}。请确保：1. 预言机服务正在运行 2. 已重启前端服务`
    } else if (error.code === 'ECONNREFUSED' || error.message.includes('Failed to fetch')) {
      error.message = '无法连接到预言机服务（端口9000）。请确保服务正在运行并已重启前端服务'
    }
    
    return Promise.reject(error)
  }
)

export default {
  // 获取预言机状态
  getStatus() {
    return oracleApi.get('/status').catch(error => {
      console.error('getStatus error:', error)
      throw error
    })
  },

  // 获取数据源列表
  getDataSources() {
    return oracleApi.get('/datasources').catch(error => {
      console.error('getDataSources error:', error)
      throw error
    })
  },

  // 获取设备状态
  getDeviceStatus(did) {
    const encodedDid = encodeURIComponent(did)
    return oracleApi.get(`/device/${encodedDid}/status`)
  },

  // 获取所有设备状态
  getAllDevicesStatus() {
    return oracleApi.get('/devices/status')
  },

  // 获取设备共识状态
  getConsensusStatus(did) {
    const encodedDid = encodeURIComponent(did)
    return oracleApi.get(`/consensus/${encodedDid}`)
  },

  // 健康检查
  healthCheck() {
    return oracleApi.get('/health')
  },

  // 获取配置信息
  getConfig() {
    return oracleApi.get('/config')
  },
}
