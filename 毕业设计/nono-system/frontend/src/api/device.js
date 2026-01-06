import api from './index'

export default {
  // 注册设备
  register(data) {
    return api.post('/devices', data)
  },

  // 获取设备信息
  get(did) {
    // 对 DID 进行 URL 编码，确保特殊字符（如 :）被正确处理
    const encodedDid = encodeURIComponent(did)
    return api.get(`/devices/${encodedDid}`)
  },

  // 列出设备
  list(params) {
    return api.get('/devices', { params })
  },

  // 更新设备状态
  updateStatus(did, status) {
    // 对 DID 进行 URL 编码
    const encodedDid = encodeURIComponent(did)
    return api.put(`/devices/${encodedDid}/status`, { status })
  },

  // 吊销设备
  revoke(did) {
    // 对 DID 进行 URL 编码
    const encodedDid = encodeURIComponent(did)
    return api.delete(`/devices/${encodedDid}`)
  },
}

