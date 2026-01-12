import api from './index'

export default {
  // 用户登录
  login(data) {
    return api.post('/users/login', data)
  },

  // 用户注册
  register(data) {
    return api.post('/users/register', data)
  },

  // 发起跨域认证请求
  requestCrossDomain(data) {
    return api.post('/auth/cross-domain', data)
  },

  // 获取设备的认证记录
  getAuthRecords(did) {
    const encodedDid = encodeURIComponent(did)
    return api.get(`/auth/records/${encodedDid}`)
  },

  // 获取认证日志
  getAuthLogs(params) {
    return api.get('/auth/logs', { params })
  },

  // 验证交易
  verifyTransaction(txHash) {
    const encodedTxHash = encodeURIComponent(txHash)
    return api.get(`/auth/verify/${encodedTxHash}`)
  },

  // 同步前端上链的认证记录到数据库
  syncAuthRecord(data) {
    return api.post('/auth/sync', data)
  },
}
