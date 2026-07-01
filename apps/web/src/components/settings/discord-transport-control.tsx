import { useState } from 'react'
import { Loader2, Play, Square } from 'lucide-react'
import { toast } from 'sonner'

import { startDiscordTransport, stopDiscordTransport } from '@/api/settings'
import { DiscordTransportLogsDialog } from '@/components/settings/discord-transport-logs'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import { hePressable } from '@/lib/highend-styles'
import { resolveDiscordTransportPhase } from '@/lib/discord-transport-phase'
import { cn } from '@/lib/utils'
import type { DiscordTransportPhase } from '@/types/settings'

type ToggleAction = 'start' | 'stop'
type ButtonStyle = 'idle-start' | 'starting' | 'stop'

const STYLE_CLASS: Record<ButtonStyle, string> = {
  'idle-start':
    'border-0 bg-success text-white shadow-none hover:!bg-success/90 hover:!text-white focus-visible:ring-success',
  starting:
    'border-0 bg-warning text-white shadow-none hover:!bg-warning/90 hover:!text-white focus-visible:ring-warning',
  stop:
    'border-0 bg-destructive text-white shadow-none hover:!bg-destructive/90 hover:!text-white focus-visible:ring-destructive',
}

function resolveButtonStyle(
  phase: DiscordTransportPhase,
  togglingAction: ToggleAction | null,
): ButtonStyle {
  if (togglingAction === 'start') return 'starting'
  if (togglingAction === 'stop') return 'stop'
  if (phase === 'ready') return 'stop'
  if (phase === 'starting') return 'starting'
  return 'idle-start'
}

function ButtonIcon({
  style,
  busy,
  togglingAction,
}: {
  style: ButtonStyle
  busy: boolean
  togglingAction: ToggleAction | null
}) {
  if (busy && togglingAction != null) {
    return <Loader2 className="size-3.5 animate-spin" />
  }
  if (style === 'idle-start') {
    return <Play className="size-3.5" />
  }
  if (style === 'starting') {
    return <Loader2 className="size-3.5 animate-spin" />
  }
  return <Square className="size-3 fill-current stroke-none" aria-hidden />
}

export function DiscordTransportControl({
  phase,
  loading,
  onRefresh,
}: {
  phase: DiscordTransportPhase | null
  loading: boolean
  onRefresh: () => void
}) {
  const { t } = useI18n()
  const [togglingAction, setTogglingAction] = useState<ToggleAction | null>(null)
  const resolvedPhase = phase ?? 'stopped'
  const running = resolvedPhase === 'starting' || resolvedPhase === 'ready'
  const style = resolveButtonStyle(resolvedPhase, togglingAction)
  const busy = loading || togglingAction != null

  const label =
    togglingAction === 'start'
      ? t('settings.discord.transportStarting')
      : togglingAction === 'stop'
        ? t('settings.discord.transportStopping')
        : resolvedPhase === 'starting'
          ? t('settings.discord.transportStarting')
          : resolvedPhase === 'ready'
            ? t('settings.discord.transportStop')
            : t('settings.discord.transportStart')

  async function handleToggle() {
    const action: ToggleAction = running ? 'stop' : 'start'
    setTogglingAction(action)
    try {
      const st = action === 'stop' ? await stopDiscordTransport() : await startDiscordTransport()
      const nextPhase = resolveDiscordTransportPhase(st)
      if (nextPhase === 'ready') {
        toast.success(t('settings.discord.transportReady'))
      } else if (nextPhase === 'starting') {
        toast.success(t('settings.discord.transportStartingToast'))
      } else {
        toast.success(t('settings.discord.transportStopped'))
      }
      onRefresh()
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : action === 'stop'
            ? t('settings.discord.transportStopFailed')
            : t('settings.discord.transportStartFailed'),
      )
    } finally {
      setTogglingAction(null)
    }
  }

  return (
    <div className="flex shrink-0 items-center gap-2">
      <DiscordTransportLogsDialog />
      <Button
        type="button"
        variant="ghost"
        size="sm"
        disabled={busy}
        className={cn(hePressable, 'gap-1.5 rounded-xl', STYLE_CLASS[style])}
        onClick={() => void handleToggle()}
      >
        <ButtonIcon style={style} busy={busy} togglingAction={togglingAction} />
        {label}
      </Button>
    </div>
  )
}
