# 预言机服务连接问题排查指南

## 问题：404 Not Found

如果遇到 `GET http://localhost:3000/oracle-api/status 404 (Not Found)` 错误，请按以下步骤排查：

## 1. 检查预言机服务是否运行

### 方法1：检查进程
```bash
# 检查端口9000是否被占用
lsof -i :9000

# 或使用
netstat -an | grep 9000
```

### 方法2：直接测试API
```bash
# 测试预言机健康检查接口
curl http://localhost:9000/api/v1/health

# 测试状态接口
curl http://localhost:9000/api/v1/status
```

如果返回JSON数据，说明服务正在运行。

## 2. 启动预言机服务

### 使用统一启动脚本（推荐）
```bash
# 在项目根目录
./bin/startall

# 或
go run ./cmd/startall
```

这会自动启动：
- 预言机服务（端口9000）
- 后端API服务（端口8080）
- 前端开发服务器（端口3000）

### 单独启动预言机服务
```bash
cd oracle
go run ./cmd/oracle
```

## 3. 检查Vite代理配置

确保 `frontend/vite.config.js` 中有以下配置：

```javascript
proxy: {
  '/oracle-api': {
    target: 'http://localhost:9000',
    changeOrigin: true,
    rewrite: (path) => path.replace(/^\/oracle-api/, '/api/v1'),
    secure: false,
  },
}
```

## 4. 重启前端服务

**重要**：修改 `vite.config.js` 后必须重启前端服务！

```bash
# 停止当前服务（Ctrl+C）
# 然后重新启动
cd frontend
npm run dev
```

## 5. 验证代理是否工作

在浏览器中直接访问：
```
http://localhost:3000/oracle-api/health
```

如果能看到JSON响应，说明代理配置正确。

## 6. 检查浏览器控制台

打开浏览器开发者工具（F12），查看：
- Network标签页：查看请求的详细信息
- Console标签页：查看错误信息

## 7. 常见问题

### 问题1：预言机服务启动失败
**原因**：配置文件路径错误或数据库未连接
**解决**：
- 检查 `oracle/config/config.yaml` 是否存在
- 确保PostgreSQL数据库正在运行
- 检查数据库连接配置

### 问题2：代理404错误
**原因**：前端服务未重启或代理配置错误
**解决**：
- 重启前端服务
- 检查 `vite.config.js` 中的代理配置
- 确认路径重写规则正确

### 问题3：CORS错误
**原因**：直接访问预言机API（已解决，已添加CORS支持）
**解决**：使用Vite代理访问（`/oracle-api`）

## 8. 调试步骤

1. **检查服务状态**
   ```bash
   curl http://localhost:9000/api/v1/health
   ```

2. **检查代理**
   在浏览器访问：`http://localhost:3000/oracle-api/health`

3. **查看Vite日志**
   前端服务启动时会显示代理配置信息

4. **检查网络请求**
   在浏览器开发者工具的Network标签页查看：
   - 请求URL
   - 响应状态码
   - 响应内容

## 9. 如果仍然无法连接

1. 确认所有服务都在运行：
   - 预言机服务（端口9000）
   - 后端服务（端口8080）
   - 前端服务（端口3000）

2. 检查防火墙设置

3. 查看服务日志：
   - 预言机服务日志（终端输出）
   - 前端服务日志（终端输出）

4. 尝试直接访问：
   ```bash
   # 在浏览器中直接访问
   http://localhost:9000/api/v1/health
   ```

如果直接访问可以，但通过代理不行，说明是代理配置问题。
