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

/** Ports blocked by Chromium (ERR_UNSAFE_PORT) — includes 6665–6669. */
const CHROME_BLOCKED_PORTS = new Set([
  6665, 6666, 6667, 6668, 6669, 6000, 4045, 3659, 2049,
])

function warnIfChromeBlockedPort(port: number) {
  if (CHROME_BLOCKED_PORTS.has(port)) {
    console.warn(
      `[vite] port ${port} is blocked by Chrome (ERR_UNSAFE_PORT). ` +
        `Use ROUND_TABLE_WEB_PORT=5173 in deploy/.env`,
    )
  }
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

function devHost(env: Record<string, string>): string | boolean {
  const raw = env.ROUND_TABLE_WEB_HOST?.trim()
  if (!raw || raw === 'true' || raw === '0.0.0.0') return true
  if (raw === 'false' || raw === '127.0.0.1') return '127.0.0.1'
  return raw
}

export default defineConfig(({ mode }) => {
  const env = runtimeEnv(loadEnv(mode, deployEnvDir, ''), loadEnv(mode, __dirname, ''))

  const port = envPort(env.ROUND_TABLE_WEB_PORT ?? env.VITE_DEV_PORT, 5173)
  warnIfChromeBlockedPort(port)
  const proxyTarget = apiProxyTarget(env)
  const host = devHost(env)

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
      host,
      strictPort: true,
      proxy,
    },
    preview: {
      port,
      host,
      strictPort: true,
      proxy,
    },
  }
})
