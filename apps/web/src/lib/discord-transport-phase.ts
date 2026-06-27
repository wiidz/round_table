import type { DiscordTransportPhase, DiscordTransportStatus } from '@/types/settings'

export function resolveDiscordTransportPhase(
  status: DiscordTransportStatus | null | undefined,
): DiscordTransportPhase {
  if (!status) return 'stopped'
  if (status.phase === 'ready' || status.phase === 'starting' || status.phase === 'stopped') {
    return status.phase
  }
  return status.running ? 'starting' : 'stopped'
}
