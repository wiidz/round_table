import { useState } from 'react'
import { Loader2, Play, Square } from 'lucide-react'
import { toast } from 'sonner'

import { startDiscordTransport, stopDiscordTransport } from '@/api/settings'
import { DiscordTransportLogsDialog } from '@/components/settings/discord-transport-logs'
import { Button } from '@/components/ui/button'
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

function buttonLabel(phase: DiscordTransportPhase, togglingAction: ToggleAction | null): string {
  if (togglingAction === 'start') return '启动中…'
  if (togglingAction === 'stop') return '停止中…'
  if (phase === 'starting') return '启动中…'
  if (phase === 'ready') return '停止服务'
  return '启动服务'
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
  const [togglingAction, setTogglingAction] = useState<ToggleAction | null>(null)
  const resolvedPhase = phase ?? 'stopped'
  const running = resolvedPhase === 'starting' || resolvedPhase === 'ready'
  const style = resolveButtonStyle(resolvedPhase, togglingAction)
  const label = buttonLabel(resolvedPhase, togglingAction)
  const busy = loading || togglingAction != null

  async function handleToggle() {
    const action: ToggleAction = running ? 'stop' : 'start'
    setTogglingAction(action)
    try {
      const st = action === 'stop' ? await stopDiscordTransport() : await startDiscordTransport()
      const nextPhase = resolveDiscordTransportPhase(st)
      if (nextPhase === 'ready') {
        toast.success('Discord 服务已就绪')
      } else if (nextPhase === 'starting') {
        toast.success('Discord 服务启动中…')
      } else {
        toast.success('Discord 服务已停止')
      }
      onRefresh()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : action === 'stop' ? '停止失败' : '启动失败')
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
