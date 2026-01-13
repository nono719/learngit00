<template>
  <div class="oracle-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>预言机服务</span>
          <el-button type="primary" @click="refreshAll" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <!-- 配置提示 -->
      <el-alert
        v-if="oracleStatus.data_sources === 0 || !oracleStatus.blockchain"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        <template #title>
          <div>
            <p><strong>配置提示：</strong></p>
            <ul style="margin: 10px 0; padding-left: 20px">
              <li v-if="oracleStatus.data_sources === 0">
                <strong>数据源未配置：</strong>请在 <code>oracle/config/config.yaml</code> 中配置数据源，或确保数据源服务正在运行
              </li>
              <li v-if="!oracleStatus.blockchain">
                <strong>区块链未连接：</strong>请在 <code>oracle/config/config.yaml</code> 中配置区块链连接信息（合约地址和私钥）
              </li>
            </ul>
            <p style="margin-top: 10px; font-size: 12px; color: #909399">
              配置文件路径：<code>oracle/config/config.yaml</code> | 修改后需要重启预言机服务
            </p>
          </div>
        </template>
      </el-alert>

      <!-- 预言机状态概览 -->
      <el-row :gutter="20" style="margin-bottom: 20px">
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="服务状态" :value="oracleStatus.status || 'unknown'">
              <template #prefix>
                <el-icon :style="{ color: oracleStatus.status === 'running' ? '#67c23a' : '#f56c6c' }">
                  <CircleCheck v-if="oracleStatus.status === 'running'" />
                  <CircleClose v-else />
                </el-icon>
              </template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="数据源数量" :value="oracleStatus.data_sources || 0">
              <template #prefix>
                <el-icon style="color: #409eff"><Connection /></el-icon>
              </template>
            </el-statistic>
            <div v-if="oracleStatus.data_sources === 0" style="margin-top: 10px; font-size: 12px; color: #e6a23c">
              ⚠️ 未加载数据源
              <el-tooltip content="数据源可能配置了但初始化失败，请查看'配置信息'标签页" placement="top">
                <el-icon style="margin-left: 5px; cursor: help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="区块链连接" :value="oracleStatus.blockchain ? '已连接' : '未连接'">
              <template #prefix>
                <el-icon :style="{ color: oracleStatus.blockchain ? '#67c23a' : '#f56c6c' }">
                  <Link v-if="oracleStatus.blockchain" />
                  <Unlock v-else />
                </el-icon>
              </template>
            </el-statistic>
            <div v-if="!oracleStatus.blockchain" style="margin-top: 10px; font-size: 12px; color: #e6a23c">
              <div v-if="oracleConfig && oracleConfig.blockchain && oracleConfig.blockchain.configured && oracleConfig.blockchain.private_key_configured">
                ⚠️ 已配置但连接失败
                <el-tooltip content="请检查：1. Ganache是否运行 2. RPC URL是否正确 3. 查看预言机服务日志" placement="top">
                  <el-icon style="margin-left: 5px; cursor: help"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
              <div v-else>
                ⚠️ 需要配置合约地址和私钥
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="采集间隔" :value="oracleStatus.interval ? oracleStatus.interval + '秒' : '-'">
              <template #prefix>
                <el-icon style="color: #e6a23c"><Clock /></el-icon>
              </template>
            </el-statistic>
            <div v-if="oracleStatus.min_consensus" style="margin-top: 10px; font-size: 12px; color: #909399">
              最小共识: {{ oracleStatus.min_consensus }}
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- API Key管理 -->
        <el-tab-pane label="API Key管理" name="apikey">
          <el-card style="margin-top: 20px">
            <template #header>
              <span>预言机账户与API Key管理</span>
            </template>
            
            <!-- 注册账户 -->
            <el-card shadow="never" style="margin-bottom: 20px">
              <template #header>
                <span>注册预言机账户</span>
              </template>
              <el-form :model="registerForm" :rules="registerRules" ref="registerFormRef" label-width="120px" style="max-width: 600px">
                <el-form-item label="用户名" prop="username">
                  <el-input v-model="registerForm.username" placeholder="oracle_service" />
                </el-form-item>
                <el-form-item label="密码" prop="password">
                  <el-input v-model="registerForm.password" type="password" show-password placeholder="至少6位" />
                </el-form-item>
                <el-form-item label="确认密码" prop="confirmPassword">
                  <el-input v-model="registerForm.confirmPassword" type="password" show-password placeholder="再次输入密码" />
                </el-form-item>
                <el-form-item label="邮箱" prop="email">
                  <el-input v-model="registerForm.email" placeholder="oracle@example.com" />
                </el-form-item>
                <el-form-item label="所属域" prop="domain">
                  <el-input v-model="registerForm.domain" placeholder="your_domain" />
                  <div style="font-size: 12px; color: #909399; margin-top: 5px">
                    预言机账户必须指定所属域
                  </div>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleRegister" :loading="registerLoading">
                    注册账户
                  </el-button>
                </el-form-item>
              </el-form>
            </el-card>

            <!-- 登录获取API Key -->
            <el-card shadow="never" style="margin-bottom: 20px">
              <template #header>
                <span>登录获取API Key</span>
              </template>
              <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef" label-width="120px" style="max-width: 600px">
                <el-form-item label="用户名" prop="username">
                  <el-input v-model="loginForm.username" placeholder="输入预言机账户用户名" />
                </el-form-item>
                <el-form-item label="密码" prop="password">
                  <el-input v-model="loginForm.password" type="password" show-password placeholder="输入密码" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleLogin" :loading="loginLoading">
                    登录获取API Key
                  </el-button>
                </el-form-item>
              </el-form>
            </el-card>

            <!-- 显示API Key -->
            <el-card v-if="apiKey" shadow="never" style="margin-bottom: 20px">
              <template #header>
                <span>当前API Key</span>
              </template>
              <el-alert type="success" :closable="false" style="margin-bottom: 15px">
                <template #title>
                  <div>
                    <p><strong>API Key已获取：</strong></p>
                    <el-input 
                      v-model="apiKey" 
                      readonly 
                      style="margin-top: 10px"
                    >
                      <template #append>
                        <el-button @click="copyApiKey">
                          <el-icon><CopyDocument /></el-icon>
                          复制
                        </el-button>
                      </template>
                    </el-input>
                  </div>
                </template>
              </el-alert>
              
              <el-card shadow="never" style="background-color: #f5f7fa">
                <template #header>
                  <span>配置说明</span>
                </template>
                <div>
                  <p><strong>步骤1：</strong>复制上面的API Key</p>
                  <p><strong>步骤2：</strong>编辑配置文件 <code>oracle/config/config.yaml</code></p>
                  <p><strong>步骤3：</strong>在数据源配置中添加API Key：</p>
                  <pre style="background-color: #fff; padding: 15px; border-radius: 4px; overflow-x: auto; margin-top: 10px;"><code>data_sources:
  - name: "monitoring_api"
    type: "monitoring"
    url: "http://localhost:8080/api/v1/devices/status"
    api_key: "{{ apiKey }}"  # 粘贴复制的API Key
    enabled: true</code></pre>
                  <p style="margin-top: 15px"><strong>步骤4：</strong>重启预言机服务使配置生效</p>
                  <el-alert type="warning" :closable="false" style="margin-top: 15px">
                    <template #title>
                      <div>
                        <p><strong>安全提示：</strong></p>
                        <ul style="margin: 5px 0; padding-left: 20px">
                          <li>不要将包含API Key的配置文件提交到Git</li>
                          <li>定期轮换API Key以提高安全性</li>
                          <li>只给预言机账户必要的权限</li>
                        </ul>
                      </div>
                    </template>
                  </el-alert>
                </div>
              </el-card>
            </el-card>

            <!-- 测试API Key -->
            <el-card v-if="apiKey" shadow="never">
              <template #header>
                <span>测试API Key</span>
              </template>
              <el-button type="success" @click="testApiKey" :loading="testLoading">
                测试API Key是否有效
              </el-button>
              <div v-if="testResult" style="margin-top: 15px">
                <el-alert :type="testResult.success ? 'success' : 'error'" :closable="false">
                  <template #title>
                    <div>
                      <p><strong>{{ testResult.success ? '测试成功' : '测试失败' }}</strong></p>
                      <p v-if="testResult.message">{{ testResult.message }}</p>
                      <p v-if="testResult.data" style="margin-top: 10px; font-size: 12px">
                        <strong>响应数据：</strong>
                        <pre style="background-color: #f5f7fa; padding: 10px; border-radius: 4px; overflow-x: auto; margin-top: 5px">{{ JSON.stringify(testResult.data, null, 2) }}</pre>
                      </p>
                    </div>
                  </template>
                </el-alert>
              </div>
            </el-card>
          </el-card>
        </el-tab-pane>

        <!-- 配置信息 -->
        <el-tab-pane label="配置信息" name="config">
          <el-card style="margin-top: 20px">
            <template #header>
              <span>当前配置</span>
            </template>
            <div v-if="oracleConfig">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="采集间隔">
                  {{ oracleConfig.oracle?.interval }} 秒
                </el-descriptions-item>
                <el-descriptions-item label="投票节点数">
                  {{ oracleConfig.oracle?.voting_nodes }}
                </el-descriptions-item>
                <el-descriptions-item label="最小共识数">
                  {{ oracleConfig.oracle?.min_consensus }}
                </el-descriptions-item>
                <el-descriptions-item label="区块链RPC">
                  {{ oracleConfig.blockchain?.rpc_url || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="合约地址">
                  <span v-if="oracleConfig.blockchain?.configured">
                    {{ oracleConfig.blockchain?.contract_addr }}
                  </span>
                  <span v-else style="color: #e6a23c">未配置（需要部署合约后配置）</span>
                </el-descriptions-item>
                <el-descriptions-item label="私钥配置">
                  <el-tag :type="oracleConfig.blockchain?.private_key_configured ? 'success' : 'warning'">
                    {{ oracleConfig.blockchain?.private_key_configured ? '已配置' : '未配置' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="配置的数据源数">
                  {{ oracleConfig.data_sources?.configured?.length || 0 }}
                </el-descriptions-item>
                <el-descriptions-item label="已加载的数据源数">
                  <el-tag :type="oracleConfig.data_sources?.loaded > 0 ? 'success' : 'warning'">
                    {{ oracleConfig.data_sources?.loaded || 0 }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>

              <!-- 配置的数据源列表 -->
              <div v-if="oracleConfig.data_sources?.configured?.length > 0" style="margin-top: 20px">
                <h4>配置的数据源 ({{ oracleConfig.data_sources.configured.length }} 个)</h4>
                <el-table :data="oracleConfig.data_sources.configured" border>
                  <el-table-column prop="name" label="名称" width="150" />
                  <el-table-column prop="type" label="类型" width="120" />
                  <el-table-column prop="url" label="URL" min-width="200" />
                  <el-table-column prop="enabled" label="启用状态" width="100">
                    <template #default="{ row }">
                      <el-tag :type="row.enabled ? 'success' : 'info'">
                        {{ row.enabled ? '已启用' : '已禁用' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="api_key_configured" label="API Key" width="120">
                    <template #default="{ row }">
                      <el-tag :type="row.api_key_configured ? 'success' : 'warning'">
                        {{ row.api_key_configured ? '已配置' : '未配置' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                </el-table>
                <el-alert
                  v-if="oracleConfig.data_sources.configured.length > 0 && oracleConfig.data_sources.loaded === 0"
                  type="error"
                  :closable="false"
                  style="margin-top: 15px"
                >
                  <template #title>
                    <div>
                      <p><strong>⚠️ 数据源配置了但未成功加载！</strong></p>
                      <p style="margin-top: 10px"><strong>已配置：</strong>{{ oracleConfig.data_sources.configured.length }} 个</p>
                      <p><strong>已加载：</strong>{{ oracleConfig.data_sources.loaded }} 个</p>
                      <p style="margin-top: 10px"><strong>可能的原因：</strong></p>
                      <ul style="margin: 5px 0; padding-left: 20px">
                        <li>数据源URL不正确或后端服务未运行</li>
                        <li>API Key无效或权限不足</li>
                        <li>数据源初始化时出错（检查预言机服务日志）</li>
                        <li>预言机服务未重启（修改配置后需要重启）</li>
                      </ul>
                      <p style="margin-top: 10px; font-size: 12px; color: #909399">
                        <strong>排查步骤：</strong><br>
                        1. 检查后端服务是否运行（端口8080）<br>
                        2. 检查API Key是否有效（在"API Key管理"标签页测试）<br>
                        3. 查看预言机服务启动日志，查找 "ERROR: failed to create data source" 错误信息<br>
                        4. 确保已重启预言机服务（修改配置后必须重启）
                      </p>
                    </div>
                  </template>
                </el-alert>
                <el-alert
                  v-else-if="oracleConfig.data_sources.configured.length > 0 && oracleConfig.data_sources.loaded > 0"
                  type="success"
                  :closable="false"
                  style="margin-top: 15px"
                >
                  <template #title>
                    <div>
                      <p><strong>✅ 数据源加载成功！</strong></p>
                      <p style="margin-top: 5px; font-size: 12px">
                        已配置 {{ oracleConfig.data_sources.configured.length }} 个，成功加载 {{ oracleConfig.data_sources.loaded }} 个
                      </p>
                    </div>
                  </template>
                </el-alert>
              </div>
              <div v-else style="margin-top: 20px; padding: 20px; text-align: center; background-color: #f5f7fa; border-radius: 4px;">
                <p>未配置数据源</p>
                <p style="font-size: 12px; color: #909399; margin-top: 10px">
                  请在 <code>oracle/config/config.yaml</code> 中配置数据源
                </p>
              </div>

              <!-- 区块链配置说明 -->
              <el-card style="margin-top: 20px">
                <template #header>
                  <span>区块链配置说明</span>
                </template>
                <p>区块链连接需要在 <code>oracle/config/config.yaml</code> 中配置：</p>
                <pre style="background-color: #f5f7fa; padding: 15px; border-radius: 4px; overflow-x: auto;"><code>blockchain:
  rpc_url: "http://localhost:8545"  # Ganache RPC地址
  chain_id: 1
  contract_addr: "0x..."  # 部署合约后的地址（替换默认值）
  private_key: "0x..."  # 预言机账户私钥（替换默认值）</code></pre>
                <p style="margin-top: 10px; color: #909399; font-size: 12px">
                  ⚠️ 修改配置后需要重启预言机服务才能生效
                </p>
              </el-card>
            </div>
            <div v-else style="text-align: center; padding: 40px">
              <el-alert type="info" :closable="false" style="margin-bottom: 20px">
                <template #title>
                  <div>
                    <p><strong>配置信息未加载</strong></p>
                    <p style="margin-top: 10px; font-size: 12px; color: #909399">
                      可能的原因：
                    </p>
                    <ul style="margin: 10px 0; padding-left: 20px; text-align: left; font-size: 12px; color: #909399">
                      <li>预言机服务未运行（端口9000）</li>
                      <li>前端服务未重启（修改vite.config.js后需要重启）</li>
                      <li>代理配置问题</li>
                    </ul>
                    <p style="margin-top: 10px; font-size: 12px; color: #909399">
                      请确保预言机服务正在运行，然后点击下方按钮重试
                    </p>
                  </div>
                </template>
              </el-alert>
              <el-button type="primary" @click="loadConfig" :loading="loading">
                重新加载配置信息
              </el-button>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- 数据源状态 -->
        <el-tab-pane label="数据源" name="datasources">
          <div v-if="dataSources.length === 0" style="margin-top: 20px; padding: 20px; text-align: center; background-color: #f5f7fa; border-radius: 4px;">
            <el-empty description="未配置数据源">
              <template #image>
                <el-icon style="font-size: 48px; color: #909399"><Connection /></el-icon>
              </template>
              <el-button type="primary" @click="showDataSourceConfig = true">
                查看配置说明
              </el-button>
            </el-empty>
          </div>
          <el-table v-else :data="dataSources" border style="margin-top: 20px">
            <el-table-column prop="name" label="数据源名称" width="200" />
            <el-table-column prop="healthy" label="健康状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.healthy ? 'success' : 'danger'">
                  {{ row.healthy ? '正常' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button size="small" @click="testDataSource(row.name)">
                  测试连接
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <el-card style="margin-top: 20px">
            <template #header>
              <span>数据源配置说明</span>
            </template>
            <p>数据源在 <code>oracle/config/config.yaml</code> 文件中配置：</p>
            <pre style="background-color: #f5f7fa; padding: 15px; border-radius: 4px; overflow-x: auto;"><code>data_sources:
  - name: "monitoring_api"
    type: "monitoring"
    url: "http://localhost:8000/api/devices/status"
    api_key: "your_api_key_here"  # 可选：如果API需要认证，填写API Key
    enabled: true
  - name: "certificate_service"
    type: "certificate"
    url: "http://localhost:8001/api/certificates"
    api_key: "your_api_key_here"  # 可选：如果API需要认证，填写API Key
    enabled: true</code></pre>
            <div style="margin-top: 15px; padding: 15px; background-color: #e6f7ff; border-left: 4px solid #409eff; border-radius: 4px;">
              <p style="margin: 0 0 10px 0;"><strong>API Key 配置说明：</strong></p>
              <ul style="margin: 0; padding-left: 20px; color: #606266">
                <li><strong>如果API不需要认证：</strong>可以留空或设置为空字符串 <code>""</code></li>
                <li><strong>如果API需要Bearer Token认证：</strong>填写完整的token，系统会自动添加 <code>Authorization: Bearer {api_key}</code> 请求头</li>
                <li><strong>API Key格式：</strong>可以是JWT token、API密钥字符串等，根据你的API服务要求填写</li>
                <li><strong>安全提示：</strong>不要将包含真实API Key的配置文件提交到版本控制系统</li>
              </ul>
            </div>
            <p style="margin-top: 10px; color: #909399; font-size: 12px">
              ⚠️ 修改配置后需要重启预言机服务才能生效
            </p>
          </el-card>
        </el-tab-pane>

        <!-- 设备状态监控 -->
        <el-tab-pane label="设备状态" name="devices">
          <el-form :inline="true" style="margin-top: 20px">
            <el-form-item label="设备DID">
              <el-input 
                v-model="deviceQueryForm.did" 
                placeholder="输入设备DID查询"
                clearable
                style="width: 300px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="queryDeviceStatus" :loading="deviceLoading">
                查询
              </el-button>
              <el-button @click="loadAllDevices" :loading="deviceLoading">
                查看所有设备
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 设备状态详情 -->
          <div v-if="deviceStatus" style="margin-top: 20px">
            <el-card>
              <template #header>
                <span>设备状态详情</span>
              </template>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="设备DID">{{ deviceStatus.did }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="getStatusType(deviceStatus.status)">
                    {{ deviceStatus.status }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="在线状态">
                  <el-tag :type="deviceStatus.online ? 'success' : 'danger'">
                    {{ deviceStatus.online ? '在线' : '离线' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="数据源">{{ deviceStatus.source }}</el-descriptions-item>
                <el-descriptions-item label="最后在线时间">
                  {{ formatTime(deviceStatus.last_seen) }}
                </el-descriptions-item>
                <el-descriptions-item label="采集时间">
                  {{ formatTime(deviceStatus.timestamp) }}
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </div>

          <!-- 所有设备状态列表 -->
          <el-table 
            v-if="allDevicesStatus && Object.keys(allDevicesStatus).length > 0" 
            :data="deviceList" 
            border 
            style="margin-top: 20px"
          >
            <el-table-column prop="did" label="设备DID" width="200" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="online" label="在线状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.online ? 'success' : 'danger'">
                  {{ row.online ? '在线' : '离线' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="source" label="数据源" width="150" />
            <el-table-column prop="timestamp" label="采集时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.timestamp) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button size="small" @click="viewConsensus(row.did)">
                  查看共识
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 共识机制 -->
        <el-tab-pane label="共识机制" name="consensus">
          <el-form :inline="true" style="margin-top: 20px">
            <el-form-item label="设备DID">
              <el-input 
                v-model="consensusQueryForm.did" 
                placeholder="输入设备DID查看共识详情"
                clearable
                style="width: 300px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="queryConsensus" :loading="consensusLoading">
                查询共识状态
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 共识详情 -->
          <div v-if="consensusStatus" style="margin-top: 20px">
            <el-card>
              <template #header>
                <span>共识详情 - {{ consensusStatus.device_did }}</span>
              </template>
              
              <el-descriptions :column="2" border style="margin-bottom: 20px">
                <el-descriptions-item label="总投票数">{{ consensusStatus.total_votes }}</el-descriptions-item>
                <el-descriptions-item label="最小共识数">{{ consensusStatus.min_consensus }}</el-descriptions-item>
                <el-descriptions-item label="是否达成共识">
                  <el-tag :type="consensusStatus.consensus ? 'success' : 'warning'">
                    {{ consensusStatus.consensus ? '已达成' : '未达成' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="共识状态" v-if="consensusStatus.consensus_status">
                  <el-tag :type="getStatusType(consensusStatus.consensus_status.status)">
                    {{ consensusStatus.consensus_status.status }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>

              <!-- 投票详情 -->
              <h4>投票详情</h4>
              <el-table :data="voteDetailsList" border style="margin-top: 10px">
                <el-table-column prop="status" label="状态" width="150">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)">
                      {{ row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="count" label="投票数" width="120" />
                <el-table-column prop="percentage" label="占比" width="120">
                  <template #default="{ row }">
                    {{ row.percentage }}%
                  </template>
                </el-table-column>
                <el-table-column label="是否达成共识" width="150">
                  <template #default="{ row }">
                    <el-tag :type="row.isConsensus ? 'success' : 'info'">
                      {{ row.isConsensus ? '是' : '否' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>

              <!-- 所有数据源状态 -->
              <h4 style="margin-top: 20px">所有数据源状态</h4>
              <el-table :data="allStatusesList" border style="margin-top: 10px">
                <el-table-column prop="source" label="数据源" width="150" />
                <el-table-column prop="status" label="状态" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)">
                      {{ row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="online" label="在线" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.online ? 'success' : 'danger'">
                      {{ row.online ? '是' : '否' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="timestamp" label="采集时间" width="180">
                  <template #default="{ row }">
                    {{ formatTime(row.timestamp) }}
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, CircleCheck, CircleClose, Connection, Link, Unlock, Clock, CopyDocument, QuestionFilled } from '@element-plus/icons-vue'
import oracleApi from '../api/oracle'
import authApi from '../api/auth'
import api from '../api/index'

const loading = ref(false)
const activeTab = ref('datasources')
const showDataSourceConfig = ref(false)

// API Key管理
const registerForm = ref({
  username: 'oracle_service',
  password: '',
  confirmPassword: '',
  email: 'oracle@example.com',
  domain: ''
})
const registerFormRef = ref(null)
const registerLoading = ref(false)

const loginForm = ref({
  username: '',
  password: ''
})
const loginFormRef = ref(null)
const loginLoading = ref(false)

const apiKey = ref('')
const testResult = ref(null)
const testLoading = ref(false)

// 注册表单验证规则
const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名长度不能少于3位', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerForm.value.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  domain: [
    { required: true, message: '请输入所属域', trigger: 'blur' }
  ]
}

// 登录表单验证规则
const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 预言机状态
const oracleStatus = ref({})
const dataSources = ref([])
const oracleConfig = ref(null)

// 设备查询
const deviceQueryForm = ref({ did: '' })
const deviceStatus = ref(null)
const allDevicesStatus = ref(null)
const deviceLoading = ref(false)

// 共识查询
const consensusQueryForm = ref({ did: '' })
const consensusStatus = ref(null)
const consensusLoading = ref(false)

// 计算属性
const deviceList = computed(() => {
  if (!allDevicesStatus.value) return []
  return Object.values(allDevicesStatus.value).map(status => ({
    did: status.did,
    status: status.status,
    online: status.online,
    source: status.source,
    timestamp: status.timestamp
  }))
})

const voteDetailsList = computed(() => {
  if (!consensusStatus.value || !consensusStatus.value.vote_details) return []
  
  const total = consensusStatus.value.total_votes
  const minConsensus = consensusStatus.value.min_consensus
  
  return Object.entries(consensusStatus.value.vote_details).map(([status, count]) => ({
    status,
    count,
    percentage: total > 0 ? ((count / total) * 100).toFixed(1) : 0,
    isConsensus: count >= minConsensus
  }))
})

const allStatusesList = computed(() => {
  if (!consensusStatus.value || !consensusStatus.value.all_statuses) return []
  return consensusStatus.value.all_statuses.map(status => ({
    source: status.source,
    status: status.status,
    online: status.online,
    timestamp: status.timestamp
  }))
})

// 加载预言机状态
const loadOracleStatus = async () => {
  try {
    const result = await oracleApi.getStatus()
    // oracleApi拦截器已经处理了response.data
    oracleStatus.value = result
    if (result.data_sources_detail) {
      dataSources.value = result.data_sources_detail
    }
  } catch (error) {
    console.error('加载预言机状态失败:', error)
    console.error('错误详情:', {
      message: error.message,
      response: error.response,
      status: error.response?.status,
      url: error.config?.url,
      baseURL: error.config?.baseURL,
      fullURL: error.config?.baseURL + error.config?.url
    })
    
    let errorMsg = error.message || '加载预言机状态失败'
    if (error.response?.status === 404) {
      errorMsg = `预言机API未找到（404）。请确保：\n1. 预言机服务正在运行（端口9000）\n2. 已重启前端服务（修改vite.config.js后需要重启）\n3. 检查浏览器控制台的代理日志`
    }
    
    ElMessage.error(errorMsg)
  }
}

// 加载数据源
const loadDataSources = async () => {
  try {
    const result = await oracleApi.getDataSources()
    // oracleApi拦截器已经处理了response.data
    dataSources.value = Array.isArray(result) ? result : []
  } catch (error) {
    console.error('加载数据源失败:', error)
    // 如果加载失败，不显示错误，因为可能服务未启动
  }
}

// 加载配置信息
const loadConfig = async () => {
  try {
    const result = await oracleApi.getConfig()
    oracleConfig.value = result
  } catch (error) {
    console.error('加载配置信息失败:', error)
    // 如果是404，可能是预言机服务未运行或代理配置问题
    if (error.response?.status === 404) {
      // 不显示错误消息，因为可能服务未启动
      console.warn('预言机配置API未找到，可能服务未运行')
      oracleConfig.value = null
    } else {
      ElMessage.error('加载配置信息失败: ' + (error.message || '未知错误'))
    }
  }
}

// 查询设备状态
const queryDeviceStatus = async () => {
  if (!deviceQueryForm.value.did) {
    ElMessage.warning('请输入设备DID')
    return
  }

  deviceLoading.value = true
  try {
    const result = await oracleApi.getDeviceStatus(deviceQueryForm.value.did)
    deviceStatus.value = result
    ElMessage.success('查询成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || error.message || '查询失败')
    deviceStatus.value = null
  } finally {
    deviceLoading.value = false
  }
}

// 加载所有设备状态
const loadAllDevices = async () => {
  deviceLoading.value = true
  try {
    const result = await oracleApi.getAllDevicesStatus()
    allDevicesStatus.value = result
    ElMessage.success(`加载了 ${Object.keys(result).length} 个设备的状态`)
  } catch (error) {
    ElMessage.error(error.message || error.response?.data?.error || '加载失败')
    allDevicesStatus.value = null
  } finally {
    deviceLoading.value = false
  }
}

// 查询共识状态
const queryConsensus = async () => {
  if (!consensusQueryForm.value.did) {
    ElMessage.warning('请输入设备DID')
    return
  }

  consensusLoading.value = true
  try {
    const result = await oracleApi.getConsensusStatus(consensusQueryForm.value.did)
    consensusStatus.value = result
    ElMessage.success('查询成功')
  } catch (error) {
    ElMessage.error(error.message || error.response?.data?.error || '查询失败')
    consensusStatus.value = null
  } finally {
    consensusLoading.value = false
  }
}

// 查看共识
const viewConsensus = (did) => {
  consensusQueryForm.value.did = did
  activeTab.value = 'consensus'
  queryConsensus()
}

// 注册预言机账户
const handleRegister = async () => {
  if (!registerFormRef.value) return
  
  await registerFormRef.value.validate(async (valid) => {
    if (!valid) return

    registerLoading.value = true
    try {
      const result = await authApi.register({
        username: registerForm.value.username,
        password: registerForm.value.password,
        email: registerForm.value.email,
        role: 'oracle',
        domain: registerForm.value.domain
      })
      
      ElMessage.success('注册成功！请使用该账户登录获取API Key')
      
      // 自动填充登录表单
      loginForm.value.username = registerForm.value.username
      loginForm.value.password = registerForm.value.password
      
      // 重置注册表单
      registerForm.value.password = ''
      registerForm.value.confirmPassword = ''
    } catch (error) {
      let errorMsg = error.message || '注册失败'
      if (error.response?.data?.error) {
        errorMsg = error.response.data.error
      }
      ElMessage.error(errorMsg)
    } finally {
      registerLoading.value = false
    }
  })
}

// 登录获取API Key
const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return

    loginLoading.value = true
    try {
      const result = await authApi.login({
        username: loginForm.value.username,
        password: loginForm.value.password
      })
      
      // 保存API Key（token）
      apiKey.value = String(result.token)
      testResult.value = null
      
      ElMessage.success('登录成功！API Key已获取')
      
      // 切换到API Key显示区域
      activeTab.value = 'apikey'
    } catch (error) {
      let errorMsg = error.message || '登录失败'
      if (error.response?.data?.error) {
        errorMsg = error.response.data.error
      }
      ElMessage.error(errorMsg)
    } finally {
      loginLoading.value = false
    }
  })
}

// 复制API Key
const copyApiKey = async () => {
  try {
    await navigator.clipboard.writeText(apiKey.value)
    ElMessage.success('API Key已复制到剪贴板')
  } catch (error) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = apiKey.value
    textArea.style.position = 'fixed'
    textArea.style.opacity = '0'
    document.body.appendChild(textArea)
    textArea.select()
    try {
      document.execCommand('copy')
      ElMessage.success('API Key已复制到剪贴板')
    } catch (err) {
      ElMessage.error('复制失败，请手动复制')
    }
    document.body.removeChild(textArea)
  }
}

// 测试API Key
const testApiKey = async () => {
  if (!apiKey.value) {
    ElMessage.warning('请先登录获取API Key')
    return
  }

  testLoading.value = true
  testResult.value = null

  try {
    // 使用API Key测试后端API（通过前端API客户端，自动处理认证）
    // 临时设置token到localStorage，让API客户端使用它
    const originalToken = localStorage.getItem('token')
    localStorage.setItem('token', apiKey.value)
    
    try {
      // 使用前端的API客户端测试
      const response = await api.get('/devices/status')
      
      // 恢复原始token
      if (originalToken) {
        localStorage.setItem('token', originalToken)
      } else {
        localStorage.removeItem('token')
      }
      
      const data = response || []
      
      testResult.value = {
        success: true,
        message: `API Key有效！成功获取 ${Array.isArray(data) ? data.length : 0} 条设备状态数据${Array.isArray(data) && data.length === 0 ? '（当前没有设备数据）' : ''}`,
        data: data
      }
    } catch (apiError) {
      // 恢复原始token
      if (originalToken) {
        localStorage.setItem('token', originalToken)
      } else {
        localStorage.removeItem('token')
      }
      
      throw apiError
    }
  } catch (error) {
    let errorMsg = error.message || '测试失败'
    let errorData = null
    
    if (error.response) {
      errorData = error.response.data || error.response
      errorMsg = errorData.error || errorMsg
    }
    
    // 如果是404，可能是路由问题或没有设备数据
    if (error.response?.status === 404) {
      errorMsg = 'API端点未找到或当前没有设备数据。请确保：1. 后端服务正在运行 2. 已注册至少一个设备'
    } else if (error.response?.status === 401) {
      errorMsg = 'API Key无效或已过期，请重新登录获取'
    } else if (error.response?.status === 403) {
      errorMsg = 'API Key权限不足，请使用具有设备查询权限的账户'
    }
    
    testResult.value = {
      success: false,
      message: errorMsg,
      data: errorData
    }
  } finally {
    testLoading.value = false
  }
}

// 测试数据源
const testDataSource = async (name) => {
  try {
    await oracleApi.healthCheck()
    ElMessage.success(`数据源 ${name} 连接正常`)
  } catch (error) {
    ElMessage.error(`数据源 ${name} 连接失败`)
  }
}

// 刷新所有
const refreshAll = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadOracleStatus(),
      loadDataSources(),
      loadConfig()
    ])
    ElMessage.success('刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    loading.value = false
  }
}

// 工具函数
const getStatusType = (status) => {
  const map = {
    active: 'success',
    suspicious: 'warning',
    revoked: 'danger',
  }
  return map[status] || 'info'
}

const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  try {
    const date = new Date(timeStr)
    return date.toLocaleString('zh-CN')
  } catch {
    return timeStr
  }
}

onMounted(() => {
  // 加载数据，但不显示错误（因为服务可能未启动）
  loadOracleStatus().catch(() => {
    // 静默失败，不显示错误
  })
  loadDataSources().catch(() => {
    // 静默失败，不显示错误
  })
  loadConfig().catch(() => {
    // 静默失败，不显示错误
  })
  
  // 每30秒自动刷新
  setInterval(() => {
    if (activeTab.value === 'datasources' || activeTab.value === 'config') {
      loadOracleStatus().catch(() => {})
      loadDataSources().catch(() => {})
      if (activeTab.value === 'config') {
        loadConfig().catch(() => {})
      }
    }
  }, 30000)
})
</script>

<style scoped>
.oracle-page {
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

:deep(.el-button--primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

:deep(.el-card--hover) {
  transition: all 0.3s ease;
}

:deep(.el-card--hover:hover) {
  transform: translateY(-4px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

:deep(.el-statistic) {
  text-align: center;
}

:deep(.el-statistic__head) {
  color: #909399;
  font-size: 14px;
  margin-bottom: 8px;
}

:deep(.el-statistic__number) {
  font-weight: 600;
  color: #303133;
}

:deep(.el-tabs--border-card) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-tabs__header) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  margin: 0;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-tabs__item.is-active) {
  color: #667eea;
  font-weight: 600;
}

:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-table th) {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  color: #606266;
  font-weight: 600;
}

:deep(.el-table tr:hover) {
  background: #f8f9fa;
}

:deep(.el-tag) {
  border-radius: 12px;
  font-weight: 500;
  padding: 4px 12px;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}
</style>
