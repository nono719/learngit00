import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path, // 保持路径不变
        secure: false,
      },
      '/oracle-api': {
        target: 'http://localhost:9000',
        changeOrigin: true,
        rewrite: (path) => {
          // /oracle-api/status -> /api/v1/status
          // /oracle-api/datasources -> /api/v1/datasources
          const newPath = path.replace(/^\/oracle-api/, '/api/v1')
          console.log('Oracle proxy rewrite:', path, '->', newPath)
          return newPath
        },
        secure: false,
        configure: (proxy, options) => {
          proxy.on('error', (err, req, res) => {
            console.error('Oracle proxy error:', err)
          })
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log('Oracle proxy request:', req.method, req.url, '->', proxyReq.path)
          })
        },
      },
    },
  },
})

