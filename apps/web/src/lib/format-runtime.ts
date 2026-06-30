export function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '—'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return s > 0 ? `${m}分${s}秒` : `${m}分钟`
  }
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (m > 0) return `${h}时${m}分`
  return `${h}小时`
}

export function formatMemoryBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '—'
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(1)}GB`
  if (bytes >= 1024 ** 2) return `${Math.round(bytes / 1024 ** 2)}MB`
  if (bytes >= 1024) return `${Math.round(bytes / 1024)}KB`
  return `${bytes}B`
}

export function formatListenAddr(addr: string): string {
  const trimmed = addr.trim()
  if (!trimmed) return '—'
  return trimmed
}

export type ProcessRuntimeMetric = {
  key: string
  label: string
  value: string
}

export function buildProcessRuntimeMetrics(snapshot?: {
  pid?: number
  uptime_seconds?: number
  memory_bytes?: number
  memory_source?: string
  listen_addr?: string
}): ProcessRuntimeMetric[] {
  if (!snapshot) return []
  const metrics: ProcessRuntimeMetric[] = []
  if (snapshot.pid != null && snapshot.pid > 0) {
    metrics.push({ key: 'pid', label: 'PID', value: String(snapshot.pid) })
  }
  if (snapshot.uptime_seconds != null && snapshot.uptime_seconds > 0) {
    metrics.push({
      key: 'uptime',
      label: '运行',
      value: formatUptime(snapshot.uptime_seconds),
    })
  }
  if (snapshot.memory_bytes != null && snapshot.memory_bytes > 0) {
    const label = snapshot.memory_source === 'heap' ? '堆内存' : '内存'
    metrics.push({
      key: 'memory',
      label,
      value: formatMemoryBytes(snapshot.memory_bytes),
    })
  }
  if (snapshot.listen_addr?.trim()) {
    metrics.push({
      key: 'listen',
      label: '监听',
      value: formatListenAddr(snapshot.listen_addr),
    })
  }
  return metrics
}

export function formatProcessRuntime(snapshot?: {
  pid?: number
  uptime_seconds?: number
  memory_bytes?: number
  memory_source?: string
  listen_addr?: string
}): string | null {
  if (!snapshot) return null
  const parts: string[] = []
  if (snapshot.pid != null && snapshot.pid > 0) {
    parts.push(`PID${snapshot.pid}`)
  }
  if (snapshot.uptime_seconds != null && snapshot.uptime_seconds > 0) {
    parts.push(`运行${formatUptime(snapshot.uptime_seconds)}`)
  }
  if (snapshot.memory_bytes != null && snapshot.memory_bytes > 0) {
    const mem = formatMemoryBytes(snapshot.memory_bytes)
    const label = snapshot.memory_source === 'heap' ? '堆内存' : '内存'
    parts.push(`${label}${mem}`)
  }
  if (snapshot.listen_addr?.trim()) {
    parts.push(`监听${formatListenAddr(snapshot.listen_addr)}`)
  }
  return parts.length > 0 ? parts.join('·') : null
}
