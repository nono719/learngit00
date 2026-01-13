<template>
  <div class="dashboard-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card class="stat-card stat-card-primary" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.devices?.total || 0 }}</div>
              <div class="stat-label">设备总数</div>
            </div>
          </div>
          <div class="stat-footer">
            <span class="stat-change positive">
              <el-icon><ArrowUp /></el-icon>
              {{ statistics.recent_activity?.devices_registered_7d || 0 }} 本周新增
            </span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card stat-card-success" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.devices?.active || 0 }}</div>
              <div class="stat-label">活跃设备</div>
            </div>
          </div>
          <div class="stat-footer">
            <el-progress 
              :percentage="getPercentage(statistics.devices?.active, statistics.devices?.total)" 
              :color="customColors"
              :show-text="false"
              style="margin-top: 10px"
            />
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card stat-card-warning" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Lock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.authentication?.total || 0 }}</div>
              <div class="stat-label">认证总数</div>
            </div>
          </div>
          <div class="stat-footer">
            <span class="stat-change">
              成功率: {{ statistics.authentication?.success_rate?.toFixed(1) || 0 }}%
            </span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card stat-card-info" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Link /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.authentication?.on_chain || 0 }}</div>
              <div class="stat-label">上链认证</div>
            </div>
          </div>
          <div class="stat-footer">
            <span class="stat-change">
              上链率: {{ statistics.authentication?.on_chain_rate?.toFixed(1) || 0 }}%
            </span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20">
      <!-- 设备状态分布 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>设备状态分布</span>
            </div>
          </template>
          <div class="chart-container">
            <div class="pie-chart-wrapper">
              <div class="pie-chart" :style="getPieChartStyle()">
                <div class="pie-center">
                  <div class="pie-center-value">{{ statistics.devices?.total || 0 }}</div>
                  <div class="pie-center-label">设备总数</div>
                </div>
              </div>
            </div>
            <div class="chart-legend">
              <div class="legend-item">
                <span class="legend-color active"></span>
                <span class="legend-text">
                  <strong>活跃:</strong> {{ statistics.devices?.active || 0 }} 
                  ({{ getPercentage(statistics.devices?.active, statistics.devices?.total) }}%)
                </span>
              </div>
              <div class="legend-item">
                <span class="legend-color suspicious"></span>
                <span class="legend-text">
                  <strong>可疑:</strong> {{ statistics.devices?.suspicious || 0 }} 
                  ({{ getPercentage(statistics.devices?.suspicious, statistics.devices?.total) }}%)
                </span>
              </div>
              <div class="legend-item">
                <span class="legend-color revoked"></span>
                <span class="legend-text">
                  <strong>已吊销:</strong> {{ statistics.devices?.revoked || 0 }} 
                  ({{ getPercentage(statistics.devices?.revoked, statistics.devices?.total) }}%)
                </span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 认证统计 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>认证统计</span>
            </div>
          </template>
          <div class="chart-container">
            <div class="bar-chart">
              <div class="bar-item">
                <div class="bar-label">成功</div>
                <div class="bar-wrapper">
                  <div 
                    class="bar-fill success" 
                    :style="{ width: getBarWidth('success') + '%' }"
                  >
                    <span class="bar-value">{{ statistics.authentication?.successful || 0 }}</span>
                  </div>
                </div>
              </div>
              <div class="bar-item">
                <div class="bar-label">失败</div>
                <div class="bar-wrapper">
                  <div 
                    class="bar-fill danger" 
                    :style="{ width: getBarWidth('failed') + '%' }"
                  >
                    <span class="bar-value">{{ statistics.authentication?.failed || 0 }}</span>
                  </div>
                </div>
              </div>
              <div class="bar-item">
                <div class="bar-label">上链</div>
                <div class="bar-wrapper">
                  <div 
                    class="bar-fill info" 
                    :style="{ width: getBarWidth('on_chain') + '%' }"
                  >
                    <span class="bar-value">{{ statistics.authentication?.on_chain || 0 }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 域设备分布 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>各域设备分布</span>
            </div>
          </template>
          <div class="domain-chart">
            <div 
              v-for="domain in statistics.domains?.details || []" 
              :key="domain.domain"
              class="domain-item"
            >
              <div class="domain-header">
                <span class="domain-name">{{ domain.domain }}</span>
                <span class="domain-count">共 {{ domain.total }} 台设备</span>
              </div>
              <div class="domain-bars">
                <div class="domain-bar-item">
                  <span class="bar-label-small">活跃</span>
                  <div class="domain-bar-wrapper">
                    <div 
                      class="domain-bar-fill active" 
                      :style="{ width: getDomainBarWidth(domain.active, domain.total) + '%' }"
                    >
                      {{ domain.active }}
                    </div>
                  </div>
                </div>
                <div class="domain-bar-item">
                  <span class="bar-label-small">可疑</span>
                  <div class="domain-bar-wrapper">
                    <div 
                      class="domain-bar-fill suspicious" 
                      :style="{ width: getDomainBarWidth(domain.suspicious, domain.total) + '%' }"
                    >
                      {{ domain.suspicious }}
                    </div>
                  </div>
                </div>
                <div class="domain-bar-item">
                  <span class="bar-label-small">已吊销</span>
                  <div class="domain-bar-wrapper">
                    <div 
                      class="domain-bar-fill revoked" 
                      :style="{ width: getDomainBarWidth(domain.revoked, domain.total) + '%' }"
                    >
                      {{ domain.revoked }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-if="!statistics.domains?.details?.length" description="暂无域数据" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Monitor, CircleCheck, Lock, Link, ArrowUp } from '@element-plus/icons-vue'
import api from '../api'

const statistics = ref({})

const customColors = [
  { color: '#67c23a', percentage: 100 },
  { color: '#e6a23c', percentage: 80 },
  { color: '#f56c6c', percentage: 60 },
]

const loadStatistics = async () => {
  try {
    statistics.value = await api.get('/statistics')
  } catch (error) {
    ElMessage.error('加载统计数据失败: ' + error.message)
  }
}

const getPercentage = (value, total) => {
  if (!total || total === 0) return 0
  return Math.round((value / total) * 100)
}

const getPieChartStyle = () => {
  const total = statistics.value.devices?.total || 1
  const active = statistics.value.devices?.active || 0
  const suspicious = statistics.value.devices?.suspicious || 0
  const revoked = statistics.value.devices?.revoked || 0
  
  const activePercent = (active / total) * 100
  const suspiciousPercent = (suspicious / total) * 100
  const revokedPercent = (revoked / total) * 100
  
  // 计算角度
  const activeDeg = activePercent * 3.6
  const suspiciousDeg = suspiciousPercent * 3.6
  const revokedDeg = revokedPercent * 3.6
  
  // 构建conic-gradient
  let gradient = ''
  let currentDeg = 0
  
  if (active > 0) {
    gradient += `#67c23a ${currentDeg}deg ${currentDeg + activeDeg}deg`
    currentDeg += activeDeg
  }
  
  if (suspicious > 0) {
    if (gradient) gradient += ', '
    gradient += `#e6a23c ${currentDeg}deg ${currentDeg + suspiciousDeg}deg`
    currentDeg += suspiciousDeg
  }
  
  if (revoked > 0) {
    if (gradient) gradient += ', '
    gradient += `#f56c6c ${currentDeg}deg ${currentDeg + revokedDeg}deg`
    currentDeg += revokedDeg
  }
  
  // 填充剩余部分为灰色
  if (currentDeg < 360) {
    if (gradient) gradient += ', '
    gradient += `#e4e7ed ${currentDeg}deg 360deg`
  }
  
  // 如果没有数据，显示灰色
  if (total === 0 || (!active && !suspicious && !revoked)) {
    gradient = '#e4e7ed 0deg 360deg'
  }
  
  return {
    background: `conic-gradient(${gradient})`,
  }
}

const getBarWidth = (type) => {
  const total = statistics.value.authentication?.total || 0
  if (total === 0) return 0
  
  let value = 0
  if (type === 'success') {
    value = statistics.value.authentication?.successful || 0
  } else if (type === 'failed') {
    value = statistics.value.authentication?.failed || 0
  } else if (type === 'on_chain') {
    value = statistics.value.authentication?.on_chain || 0
  }
  
  const percentage = (value / total) * 100
  // 确保最小宽度，如果值大于0但百分比很小，至少显示一点
  if (value > 0 && percentage < 1) {
    return 1
  }
  return Math.max(0, Math.min(100, percentage))
}

const getDomainBarWidth = (value, total) => {
  if (!total || total === 0) return 0
  return (value / total) * 100
}

onMounted(() => {
  loadStatistics()
  // 每30秒刷新一次
  setInterval(loadStatistics, 30000)
})
</script>

<style scoped>
.dashboard-page {
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

/* 统计卡片 */
.stat-card {
  border-radius: 12px;
  transition: all 0.3s ease;
  overflow: hidden;
  position: relative;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
}

.stat-card-success::before {
  background: linear-gradient(90deg, #67c23a 0%, #85ce61 100%);
}

.stat-card-warning::before {
  background: linear-gradient(90deg, #e6a23c 0%, #ebb563 100%);
}

.stat-card-info::before {
  background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
  color: #667eea;
}

.stat-card-success .stat-icon {
  background: linear-gradient(135deg, #67c23a15 0%, #85ce6115 100%);
  color: #67c23a;
}

.stat-card-warning .stat-icon {
  background: linear-gradient(135deg, #e6a23c15 0%, #ebb56315 100%);
  color: #e6a23c;
}

.stat-card-info .stat-icon {
  background: linear-gradient(135deg, #409eff15 0%, #66b1ff15 100%);
  color: #409eff;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 8px;
  animation: countUp 1s ease;
}

@keyframes countUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stat-label {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.stat-footer {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #ebeef5;
}

.stat-change {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
}

.stat-change.positive {
  color: #67c23a;
}

/* 图表容器 */
.chart-container {
  padding: 20px;
  min-height: 300px;
}

/* 饼图 */
.pie-chart-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 30px;
  min-height: 250px;
}

.pie-chart {
  width: 200px;
  height: 200px;
  border-radius: 50%;
  position: relative;
  animation: pieGrow 1s ease;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

@keyframes pieGrow {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

.pie-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  z-index: 1;
  background: #ffffff;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.pie-center-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
  margin-bottom: 5px;
}

.pie-center-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

.chart-legend {
  display: flex;
  justify-content: center;
  gap: 30px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.legend-item:hover {
  background: #f0f2f5;
  transform: translateX(5px);
}

.legend-text {
  flex: 1;
}

.legend-text strong {
  color: #303133;
  font-weight: 600;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 4px;
}

.legend-color.active {
  background: #67c23a;
}

.legend-color.suspicious {
  background: #e6a23c;
}

.legend-color.revoked {
  background: #f56c6c;
}

/* 柱状图 */
.bar-chart {
  padding: 20px 0;
}

.bar-item {
  margin-bottom: 25px;
}

.bar-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.bar-label-value {
  font-size: 12px;
  color: #909399;
  font-weight: 400;
}

.bar-wrapper {
  height: 40px;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: visible;
  position: relative;
  width: 100%;
}

.bar-fill {
  height: 100%;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 15px;
  color: #ffffff;
  font-weight: 600;
  font-size: 14px;
  min-width: 0;
  transition: width 0.8s ease;
  position: relative;
  overflow: hidden;
}

/* 移除动画，使用transition实现平滑效果 */

.bar-fill.success {
  background: linear-gradient(90deg, #67c23a 0%, #85ce61 100%);
}

.bar-fill.danger {
  background: linear-gradient(90deg, #f56c6c 0%, #f78989 100%);
}

.bar-fill.info {
  background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
}

/* 域分布图 */
.domain-chart {
  padding: 20px 0;
}

.domain-item {
  margin-bottom: 30px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.domain-item:hover {
  background: #f0f2f5;
  transform: translateX(5px);
}

.domain-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.domain-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.domain-count {
  font-size: 14px;
  color: #909399;
}

.domain-bars {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.domain-bar-item {
  display: flex;
  align-items: center;
  gap: 15px;
}

.bar-label-small {
  width: 60px;
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

.domain-bar-wrapper {
  flex: 1;
  height: 28px;
  background: #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
}

.domain-bar-fill {
  height: 100%;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  animation: barGrow 1s ease;
  transition: width 0.3s ease;
}

.domain-bar-fill.active {
  background: linear-gradient(90deg, #67c23a 0%, #85ce61 100%);
}

.domain-bar-fill.suspicious {
  background: linear-gradient(90deg, #e6a23c 0%, #ebb563 100%);
}

.domain-bar-fill.revoked {
  background: linear-gradient(90deg, #f56c6c 0%, #f78989 100%);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: #303133;
}
</style>
