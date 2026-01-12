// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title DeviceIdentity
 * @dev 设备去中心化身份管理智能合约
 * 功能包括：设备注册、状态更新、跨域认证、身份吊销
 */
contract DeviceIdentity {
    // 设备状态枚举
    enum DeviceStatus {
        Active,      // 活跃
        Suspicious,  // 可疑
        Revoked      // 已吊销
    }

    // 设备信息结构体
    struct Device {
        string did;              // 去中心化身份标识符
        string metadata;        // 设备元数据（JSON格式）
        DeviceStatus status;    // 设备状态
        address owner;          // 设备所有者地址
        uint256 registeredAt;   // 注册时间戳
        uint256 lastUpdated;    // 最后更新时间戳
        bool exists;            // 是否存在
    }

    // 跨域认证记录
    struct CrossDomainAuth {
        string sourceDomain;    // 源域
        string targetDomain;    // 目标域
        string deviceDid;       // 设备DID
        bool authorized;        // 是否授权
        uint256 timestamp;      // 认证时间戳
    }

    // 事件定义
    event DeviceRegistered(string indexed did, address indexed owner, uint256 timestamp);
    event DeviceStatusUpdated(string indexed did, DeviceStatus status, uint256 timestamp);
    event CrossDomainAuthRequested(string indexed did, string sourceDomain, string targetDomain);
    event CrossDomainAuthCompleted(string indexed did, string sourceDomain, string targetDomain, bool authorized);
    event DeviceRevoked(string indexed did, uint256 timestamp);

    // 存储映射
    mapping(string => Device) public devices;                    // DID => Device
    mapping(string => CrossDomainAuth[]) public authRecords;     // DID => AuthRecords
    mapping(address => bool) public authorizedOracles;           // 授权预言机地址
    mapping(address => bool) public authorizedAdmins;            // 授权管理员地址

    address public owner;  // 合约所有者

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    modifier onlyAuthorizedOracle() {
        require(authorizedOracles[msg.sender], "Only authorized oracle can call this function");
        _;
    }

    modifier onlyAuthorizedAdmin() {
        require(authorizedAdmins[msg.sender] || msg.sender == owner, "Only authorized admin can call this function");
        _;
    }

    constructor() {
        owner = msg.sender;
        authorizedAdmins[msg.sender] = true;
    }

    /**
     * @dev 注册新设备
     * @param _did 设备DID标识符
     * @param _metadata 设备元数据（JSON字符串）
     */
    function registerDevice(string memory _did, string memory _metadata) 
        public 
        onlyAuthorizedAdmin 
    {
        require(!devices[_did].exists, "Device already exists");
        require(bytes(_did).length > 0, "DID cannot be empty");

        devices[_did] = Device({
            did: _did,
            metadata: _metadata,
            status: DeviceStatus.Active,
            owner: msg.sender,
            registeredAt: block.timestamp,
            lastUpdated: block.timestamp,
            exists: true
        });

        emit DeviceRegistered(_did, msg.sender, block.timestamp);
    }

    /**
     * @dev 更新设备状态（由预言机调用）
     * @param _did 设备DID
     * @param _status 新状态
     */
    function updateDeviceStatus(string memory _did, DeviceStatus _status) 
        public 
        onlyAuthorizedOracle 
    {
        require(devices[_did].exists, "Device does not exist");
        
        devices[_did].status = _status;
        devices[_did].lastUpdated = block.timestamp;

        emit DeviceStatusUpdated(_did, _status, block.timestamp);
    }

    /**
     * @dev 跨域认证请求
     * @param _did 设备DID
     * @param _sourceDomain 源域
     * @param _targetDomain 目标域
     * @return 是否授权
     */
    function requestCrossDomainAuth(
        string memory _did,
        string memory _sourceDomain,
        string memory _targetDomain
    ) public returns (bool) {
        require(devices[_did].exists, "Device does not exist");
        require(devices[_did].status == DeviceStatus.Active, "Device is not active");

        bool authorized = devices[_did].status == DeviceStatus.Active;

        CrossDomainAuth memory auth = CrossDomainAuth({
            sourceDomain: _sourceDomain,
            targetDomain: _targetDomain,
            deviceDid: _did,
            authorized: authorized,
            timestamp: block.timestamp
        });

        authRecords[_did].push(auth);

        emit CrossDomainAuthRequested(_did, _sourceDomain, _targetDomain);
        emit CrossDomainAuthCompleted(_did, _sourceDomain, _targetDomain, authorized);

        return authorized;
    }

    /**
     * @dev 吊销设备身份
     * @param _did 设备DID
     */
    function revokeDevice(string memory _did) public onlyAuthorizedAdmin {
        require(devices[_did].exists, "Device does not exist");
        require(devices[_did].status != DeviceStatus.Revoked, "Device already revoked");

        devices[_did].status = DeviceStatus.Revoked;
        devices[_did].lastUpdated = block.timestamp;

        emit DeviceRevoked(_did, block.timestamp);
    }

    /**
     * @dev 查询设备信息
     * @param _did 设备DID
     * @return did 设备DID标识符
     * @return metadata 设备元数据
     * @return status 设备状态
     * @return deviceOwner 设备所有者地址
     * @return registeredAt 注册时间戳
     * @return lastUpdated 最后更新时间戳
     */
    function getDevice(string memory _did) public view returns (
        string memory did,
        string memory metadata,
        DeviceStatus status,
        address deviceOwner,
        uint256 registeredAt,
        uint256 lastUpdated
    ) {
        require(devices[_did].exists, "Device does not exist");
        Device memory device = devices[_did];
        return (
            device.did,
            device.metadata,
            device.status,
            device.owner,
            device.registeredAt,
            device.lastUpdated
        );
    }

    /**
     * @dev 查询设备认证记录
     * @param _did 设备DID
     * @return 认证记录数组
     */
    function getAuthRecords(string memory _did) public view returns (CrossDomainAuth[] memory) {
        return authRecords[_did];
    }

    /**
     * @dev 授权预言机地址
     * @param _oracle 预言机地址
     */
    function authorizeOracle(address _oracle) public onlyOwner {
        authorizedOracles[_oracle] = true;
    }

    /**
     * @dev 撤销预言机授权
     * @param _oracle 预言机地址
     */
    function revokeOracle(address _oracle) public onlyOwner {
        authorizedOracles[_oracle] = false;
    }

    /**
     * @dev 授权管理员地址
     * @param _admin 管理员地址
     */
    function authorizeAdmin(address _admin) public onlyOwner {
        authorizedAdmins[_admin] = true;
    }

    /**
     * @dev 撤销管理员授权
     * @param _admin 管理员地址
     */
    function revokeAdmin(address _admin) public onlyOwner {
        authorizedAdmins[_admin] = false;
    }
}

