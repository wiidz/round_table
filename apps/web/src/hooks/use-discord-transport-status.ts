import { useCallback, useEffect, useState } from 'react'

import { fetchDiscordTransportStatus } from '@/api/settings'
import { resolveDiscordTransportPhase } from '@/lib/discord-transport-phase'
import type { DiscordTransportStatus } from '@/types/settings'

export function useDiscordTransportStatus(enabled: boolean, intervalMs = 5000) {
  const [status, setStatus] = useState<DiscordTransportStatus | null>(null)

  const refresh = useCallback(async () => {
    try {
      const st = await fetchDiscordTransportStatus()
      setStatus(st)
    } catch {
      setStatus(null)
    }
  }, [])

  const phase = resolveDiscordTransportPhase(status)

  useEffect(() => {
    if (!enabled) return
    void refresh()
    const ms = phase === 'starting' ? Math.min(intervalMs, 2000) : intervalMs
    const timer = window.setInterval(() => void refresh(), ms)
    return () => window.clearInterval(timer)
  }, [enabled, intervalMs, refresh, phase])

  return {
    status,
    phase,
    running: status?.running ?? false,
    ready: phase === 'ready',
    refresh,
  }
}
