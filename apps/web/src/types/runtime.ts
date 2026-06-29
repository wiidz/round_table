export interface ProcessSnapshot {
  pid?: number
  uptime_seconds?: number
  memory_bytes?: number
  memory_source?: 'rss' | 'heap' | string
  /** HTTP 监听地址（如 :7777） */
  listen_addr?: string
}

export interface RuntimeResponse {
  server: ProcessSnapshot
  discord_transport?: ProcessSnapshot
}
