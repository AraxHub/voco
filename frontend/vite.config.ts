import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': (() => {
        // Автоопределение окружения:
        // - локально: бэк на http://localhost:8080
        // - docker-compose: бэк доступен как http://voco-backend:8080 (service name)
        const override = process.env.VITE_PROXY_TARGET?.toString().trim()
        if (override) return override

        let inDocker = false
        try {
          inDocker = fs.existsSync('/.dockerenv')
        } catch {
          // ignore
        }

        return inDocker ? 'http://voco-backend:8080' : 'http://localhost:8080'
      })(),
    },
  },
})
