import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],

  build: {
    outDir: '../internal/web/dist',
    emptyOutDir: true,
  },

  server: {
    port: 5173,

    proxy: {
      '/api': 'http://localhost:4215',
      '/v1': 'http://localhost:4215',
      '/health': 'http://localhost:4215',
    },
  },
})