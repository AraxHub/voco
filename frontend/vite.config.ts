import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import fs from 'node:fs'

function apiProxyTarget(): string {
  const override = process.env.VITE_PROXY_TARGET?.toString().trim()
  if (override) return override
  let inDocker = false
  try {
    inDocker = fs.existsSync('/.dockerenv')
  } catch {
    // ignore
  }
  return inDocker ? 'http://voco-backend:8080' : 'http://localhost:8080'
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': {
        target: apiProxyTarget(),
        changeOrigin: true,
        // Required for chat/call realtime (otherwise conversation list & incoming call lag).
        ws: true,
      },
    },
  },
})
