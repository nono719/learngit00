import { ethers } from 'ethers'

// 合约 ABI（简化版，只包含需要的函数和事件）
const CONTRACT_ABI = [
  {
    inputs: [
      { internalType: 'string', name: '_did', type: 'string' },
      { internalType: 'string', name: '_metadata', type: 'string' }
    ],
    name: 'registerDevice',
    outputs: [],
    stateMutability: 'nonpayable',
    type: 'function'
  },
  {
    inputs: [
      { internalType: 'string', name: '_did', type: 'string' },
      { internalType: 'string', name: '_sourceDomain', type: 'string' },
      { internalType: 'string', name: '_targetDomain', type: 'string' }
    ],
    name: 'requestCrossDomainAuth',
    outputs: [{ internalType: 'bool', name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
    type: 'function'
  },
  {
    inputs: [{ internalType: 'string', name: '', type: 'string' }],
    name: 'devices',
    outputs: [
      { internalType: 'string', name: 'did', type: 'string' },
      { internalType: 'string', name: 'metadata', type: 'string' },
      { internalType: 'uint8', name: 'status', type: 'uint8' },
      { internalType: 'address', name: 'owner', type: 'address' },
      { internalType: 'uint256', name: 'registeredAt', type: 'uint256' },
      { internalType: 'uint256', name: 'lastUpdated', type: 'uint256' },
      { internalType: 'bool', name: 'exists', type: 'bool' }
    ],
    stateMutability: 'view',
    type: 'function'
  },
  {
    anonymous: false,
    inputs: [
      { indexed: true, internalType: 'string', name: 'did', type: 'string' },
      { indexed: false, internalType: 'string', name: 'sourceDomain', type: 'string' },
      { indexed: false, internalType: 'string', name: 'targetDomain', type: 'string' }
    ],
    name: 'CrossDomainAuthRequested',
    type: 'event'
  },
  {
    anonymous: false,
    inputs: [
      { indexed: true, internalType: 'string', name: 'did', type: 'string' },
      { indexed: false, internalType: 'string', name: 'sourceDomain', type: 'string' },
      { indexed: false, internalType: 'string', name: 'targetDomain', type: 'string' },
      { indexed: false, internalType: 'bool', name: 'authorized', type: 'bool' }
    ],
    name: 'CrossDomainAuthCompleted',
    type: 'event'
  },
  {
    anonymous: false,
    inputs: [
      { indexed: true, internalType: 'string', name: 'did', type: 'string' },
      { indexed: true, internalType: 'address', name: 'owner', type: 'address' },
      { indexed: false, internalType: 'uint256', name: 'timestamp', type: 'uint256' }
    ],
    name: 'DeviceRegistered',
    type: 'event'
  }
]

class BlockchainService {
  constructor() {
    this.provider = null
    this.signer = null
    this.contract = null
    this.contractAddress = null
    this.network = null
  }

  /**
   * 连接到 Ganache 或 MetaMask
   * @param {Object} config - 连接配置
   * @param {string} config.rpcUrl - RPC URL (可选，如果使用 MetaMask 则不需要)
   * @param {string} config.contractAddress - 合约地址
   * @param {string} config.privateKey - 私钥 (可选，如果使用 MetaMask 则不需要)
   * @param {boolean} config.useMetaMask - 是否使用 MetaMask
   */
  async connect(config = {}) {
    try {
      const {
        rpcUrl = 'http://127.0.0.1:8545',
        contractAddress,
        privateKey = null,
        useMetaMask = false
      } = config

      if (!contractAddress) {
        throw new Error('合约地址不能为空')
      }

      this.contractAddress = contractAddress

      // 使用 MetaMask
      if (useMetaMask) {
        if (typeof window.ethereum === 'undefined') {
          throw new Error('未检测到 MetaMask，请先安装 MetaMask 扩展')
        }

        // 请求账户访问权限
        await window.ethereum.request({ method: 'eth_requestAccounts' })
        // ethers v6 使用 BrowserProvider
        this.provider = new ethers.BrowserProvider(window.ethereum)
        this.signer = await this.provider.getSigner()
        
        // 获取网络信息
        const network = await this.provider.getNetwork()
        this.network = {
          chainId: Number(network.chainId),
          name: network.name
        }
      } else {
        // 使用 Ganache 或自定义 RPC
        if (!privateKey) {
          throw new Error('使用 RPC 连接时必须提供私钥')
        }

        this.provider = new ethers.JsonRpcProvider(rpcUrl)
        this.signer = new ethers.Wallet(privateKey, this.provider)
        
        // 获取网络信息
        const network = await this.provider.getNetwork()
        this.network = {
          chainId: Number(network.chainId),
          name: network.name || 'Local Network'
        }
      }

      // 创建合约实例
      this.contract = new ethers.Contract(
        contractAddress,
        CONTRACT_ABI,
        this.signer
      )

      // 获取当前账户地址
      const address = await this.signer.getAddress()
      
      return {
        success: true,
        address,
        network: this.network
      }
    } catch (error) {
      console.error('连接区块链失败:', error)
      throw error
    }
  }

  /**
   * 检查连接状态
   */
  isConnected() {
    return this.contract !== null && this.provider !== null
  }

  /**
   * 获取当前账户地址
   */
  async getAddress() {
    if (!this.signer) {
      throw new Error('未连接到区块链')
    }
    return await this.signer.getAddress()
  }

  /**
   * 获取账户余额
   */
  async getBalance() {
    if (!this.signer) {
      throw new Error('未连接到区块链')
    }
    const address = await this.signer.getAddress()
    const balance = await this.provider.getBalance(address)
    return ethers.formatEther(balance)
  }

  /**
   * 请求跨域认证
   * @param {string} did - 设备 DID
   * @param {string} sourceDomain - 源域
   * @param {string} targetDomain - 目标域
   * @returns {Promise<Object>} 交易结果
   */
  async requestCrossDomainAuth(did, sourceDomain, targetDomain) {
    if (!this.contract) {
      throw new Error('未连接到区块链，请先连接')
    }

    try {
      // 调用合约函数
      const tx = await this.contract.requestCrossDomainAuth(
        did,
        sourceDomain,
        targetDomain,
        {
          gasLimit: 200000 // 设置 gas limit
        }
      )

      // 等待交易确认
      const receipt = await tx.wait()

      // 解析事件
      const events = this.parseEvents(receipt)

      return {
        success: true,
        txHash: receipt.hash,
        blockNumber: receipt.blockNumber,
        gasUsed: receipt.gasUsed.toString(),
        events,
        authorized: this.getAuthorizedFromEvents(events)
      }
    } catch (error) {
      console.error('跨域认证失败:', error)
      throw error
    }
  }

  /**
   * 注册设备到区块链
   * @param {string} did - 设备 DID
   * @param {string} metadata - 设备元数据（JSON字符串）
   * @returns {Promise<Object>} 交易结果
   */
  async registerDevice(did, metadata = '{}') {
    if (!this.contract) {
      throw new Error('未连接到区块链，请先连接')
    }

    try {
      // 调用合约函数
      const tx = await this.contract.registerDevice(did, metadata, {
        gasLimit: 200000
      })

      // 等待交易确认
      const receipt = await tx.wait()

      // 解析事件
      const events = this.parseEvents(receipt)

      return {
        success: true,
        txHash: receipt.hash,
        blockNumber: receipt.blockNumber,
        gasUsed: receipt.gasUsed.toString(),
        events
      }
    } catch (error) {
      console.error('注册设备失败:', error)
      throw error
    }
  }

  /**
   * 查询设备信息
   * @param {string} did - 设备 DID
   */
  async getDevice(did) {
    if (!this.contract) {
      throw new Error('未连接到区块链，请先连接')
    }

    try {
      const device = await this.contract.devices(did)
      return {
        did: device.did,
        metadata: device.metadata,
        status: Number(device.status), // 0: Active, 1: Suspicious, 2: Revoked
        owner: device.owner,
        registeredAt: Number(device.registeredAt),
        lastUpdated: Number(device.lastUpdated),
        exists: device.exists
      }
    } catch (error) {
      console.error('查询设备失败:', error)
      throw error
    }
  }

  /**
   * 检查设备是否存在
   * @param {string} did - 设备 DID
   * @returns {Promise<boolean>} 设备是否存在
   */
  async checkDeviceExists(did) {
    try {
      const device = await this.getDevice(did)
      return device.exists
    } catch (error) {
      // 如果查询失败，可能是设备不存在
      return false
    }
  }

  /**
   * 解析交易事件
   */
  parseEvents(receipt) {
    const events = []
    
    if (receipt.logs && receipt.logs.length > 0) {
      for (const log of receipt.logs) {
        try {
          const parsedLog = this.contract.interface.parseLog(log)
          if (parsedLog) {
            events.push({
              name: parsedLog.name,
              args: parsedLog.args,
              address: log.address
            })
          }
        } catch (e) {
          // 忽略无法解析的日志
        }
      }
    }
    
    return events
  }

  /**
   * 从事件中获取授权结果
   */
  getAuthorizedFromEvents(events) {
    const completedEvent = events.find(e => e.name === 'CrossDomainAuthCompleted')
    if (completedEvent && completedEvent.args) {
      return completedEvent.args.authorized || false
    }
    return null
  }

  /**
   * 获取交易收据
   * @param {string} txHash - 交易哈希
   */
  async getTransactionReceipt(txHash) {
    if (!this.provider) {
      throw new Error('未连接到区块链')
    }
    return await this.provider.getTransactionReceipt(txHash)
  }

  /**
   * 断开连接
   */
  disconnect() {
    this.provider = null
    this.signer = null
    this.contract = null
    this.contractAddress = null
    this.network = null
  }
}

// 导出单例
export default new BlockchainService()

