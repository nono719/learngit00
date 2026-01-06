package blockchain

// DeviceIdentityABI 设备身份合约ABI（简化版）
// 实际使用时应该从编译后的合约文件加载
const DeviceIdentityABI = `[
	{
		"inputs": [
			{"internalType": "string", "name": "_did", "type": "string"},
			{"internalType": "uint8", "name": "_status", "type": "uint8"}
		],
		"name": "updateDeviceStatus",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "string", "name": "", "type": "string"}
		],
		"name": "devices",
		"outputs": [
			{"internalType": "string", "name": "did", "type": "string"},
			{"internalType": "string", "name": "metadata", "type": "string"},
			{"internalType": "uint8", "name": "status", "type": "uint8"},
			{"internalType": "address", "name": "owner", "type": "address"},
			{"internalType": "uint256", "name": "registeredAt", "type": "uint256"},
			{"internalType": "uint256", "name": "lastUpdated", "type": "uint256"},
			{"internalType": "bool", "name": "exists", "type": "bool"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

