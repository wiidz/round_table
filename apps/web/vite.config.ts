import path from 'node:path'
import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const deployEnvDir = path.resolve(__dirname, '../../deploy')

function envPort(raw: string | undefined, fallback: number): number {
  if (!raw?.trim()) return fallback
  const n = Number.parseInt(raw, 10)
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function apiProxyTarget(env: Record<string, string>): string {
  const explicit = env.ROUND_TABLE_API_PROXY?.trim()
  if (explicit) return explicit
  const httpPort = envPort(env.ROUND_TABLE_HTTP_PORT, 7777)
  return `http://127.0.0.1:${httpPort}`
}

function runtimeEnv(...layers: Record<string, string>[]): Record<string, string> {
  const fromProcess: Record<string, string> = {}
  for (const [key, val] of Object.entries(process.env)) {
    if (val == null) continue
    if (key.startsWith('ROUND_TABLE_') || key.startsWith('VITE_')) {
      fromProcess[key] = val
    }
  }
  return Object.assign({}, ...layers, fromProcess)
}

export default defineConfig(({ mode }) => {
  const env = runtimeEnv(loadEnv(mode, deployEnvDir, ''), loadEnv(mode, __dirname, ''))

  const port = envPort(env.ROUND_TABLE_WEB_PORT ?? env.VITE_DEV_PORT, 5173)
  const proxyTarget = apiProxyTarget(env)

  const proxy = {
    '/api': {
      target: proxyTarget,
      changeOrigin: true,
    },
  }

  return {
    plugins: [tailwindcss(), react()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
      dedupe: ['react', 'react-dom'],
    },
    server: {
      port,
      strictPort: true,
      proxy,
    },
    preview: {
      port,
      strictPort: true,
      proxy,
    },
  }
})
