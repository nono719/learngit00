package blockchain

// DeviceIdentityABI 设备身份合约ABI
const DeviceIdentityABI = `[
	{
		"inputs": [
			{"internalType": "string", "name": "_did", "type": "string"},
			{"internalType": "string", "name": "_sourceDomain", "type": "string"},
			{"internalType": "string", "name": "_targetDomain", "type": "string"}
		],
		"name": "requestCrossDomainAuth",
		"outputs": [
			{"internalType": "bool", "name": "", "type": "bool"}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
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
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "internalType": "string", "name": "did", "type": "string"},
			{"indexed": false, "internalType": "string", "name": "sourceDomain", "type": "string"},
			{"indexed": false, "internalType": "string", "name": "targetDomain", "type": "string"}
		],
		"name": "CrossDomainAuthRequested",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "internalType": "string", "name": "did", "type": "string"},
			{"indexed": false, "internalType": "string", "name": "sourceDomain", "type": "string"},
			{"indexed": false, "internalType": "string", "name": "targetDomain", "type": "string"},
			{"indexed": false, "internalType": "bool", "name": "authorized", "type": "bool"}
		],
		"name": "CrossDomainAuthCompleted",
		"type": "event"
	}
]`

