import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // string shorthand: '/foo' -> 'http://localhost:4567/foo'
      // '/api': 'http://localhost:8888',
      // Proxying /api to the Go backend
      '/api': {
        target: 'http://localhost:8888', // Your Go backend address
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, ''), // Uncomment if your Go API doesn't expect /api prefix
      },
    },
  },
})
