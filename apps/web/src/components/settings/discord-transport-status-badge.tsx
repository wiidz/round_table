import type { DiscordTransportPhase } from '@/types/settings'
import { cn } from '@/lib/utils'

export function DiscordTransportStatusBadge({
  phase,
  pid,
  readyAt,
  className,
}: {
  phase: DiscordTransportPhase
  pid?: number
  readyAt?: string
  className?: string
}) {
  if (phase === 'ready') {
    return (
      <span
        className={cn(
          'inline-flex items-center gap-2 rounded-full bg-success-soft px-2.5 py-1 text-xs font-medium text-success ring-1 ring-success/20',
          className,
        )}
      >
        <span className="relative flex size-2 shrink-0">
          <span className="absolute inline-flex size-full animate-ping rounded-full bg-success/50 opacity-75" />
          <span className="relative inline-flex size-2 rounded-full bg-success" />
        </span>
        运行中
        {pid != null && pid > 0 && (
          <span className="font-mono text-[11px] font-normal tabular-nums text-success/85">
            PID {pid}
          </span>
        )}
        {readyAt && (
          <span className="font-mono text-[11px] font-normal tabular-nums text-success/75">
            {formatReadyAt(readyAt)}
          </span>
        )}
      </span>
    )
  }

  if (phase === 'starting') {
    return (
      <span
        className={cn(
          'inline-flex items-center gap-2 rounded-full bg-warning-soft px-2.5 py-1 text-xs font-medium text-warning ring-1 ring-warning/25',
          className,
        )}
      >
        <span className="relative flex size-2 shrink-0">
          <span className="absolute inline-flex size-full animate-ping rounded-full bg-warning/50 opacity-75" />
          <span className="relative inline-flex size-2 rounded-full bg-warning" />
        </span>
        启动中
        {pid != null && pid > 0 && (
          <span className="font-mono text-[11px] font-normal tabular-nums text-warning/85">
            PID {pid}
          </span>
        )}
      </span>
    )
  }

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full bg-black/[0.04] px-2.5 py-1 text-xs text-text-tertiary ring-1 ring-black/[0.06]',
        className,
      )}
    >
      <span aria-hidden className="size-1.5 rounded-full bg-text-tertiary/40" />
      未启动
    </span>
  )
}

function formatReadyAt(raw: string): string {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  const h = String(d.getHours()).padStart(2, '0')
  const m = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s} 就绪`
}
