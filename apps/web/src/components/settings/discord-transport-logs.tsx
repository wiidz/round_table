import { useCallback, useEffect, useRef, useState } from 'react'
import { RefreshCw, ScrollText, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'

import {
  clearDiscordTransportLogs,
  fetchDiscordTransportLogs,
  fetchDiscordTransportStatus,
} from '@/api/settings'
import { DiscordTransportStatusBadge } from '@/components/settings/discord-transport-status-badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import { useI18n } from '@/hooks/use-i18n'
import { resolveDiscordTransportPhase } from '@/lib/discord-transport-phase'
import {
  heFieldHint,
  heFormEmbed,
  hePressable,
  heSectionTitle,
  heSpring,
} from '@/lib/highend-styles'
import type { DiscordTransportPhase } from '@/types/settings'
import { cn } from '@/lib/utils'

type RefreshMode = 'live' | 'manual'

const LIVE_INTERVAL_MS = 3000

function formatTimeHMS(date: Date): string {
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

function RefreshModeControl({
  mode,
  loading,
  updatedAt,
  onModeChange,
  onRefresh,
}: {
  mode: RefreshMode
  loading: boolean
  updatedAt: Date | null
  onModeChange: (mode: RefreshMode) => void
  onRefresh: () => void
}) {
  const { t } = useI18n()

  return (
    <div className="flex flex-wrap items-center gap-2">
      <div
        className="inline-flex rounded-lg bg-black/[0.04] p-0.5 ring-1 ring-black/[0.06]"
        role="group"
        aria-label={t('settings.discord.logsRefreshMode')}
      >
        {(['live', 'manual'] as const).map((item) => {
          const active = mode === item
          return (
            <button
              key={item}
              type="button"
              aria-pressed={active}
              onClick={() => onModeChange(item)}
              className={cn(
                'rounded-md px-2.5 py-1 text-xs',
                heSpring,
                active && item === 'live' && 'bg-success-soft font-medium text-success ring-1 ring-success/20',
                active &&
                  item === 'manual' &&
                  'bg-surface font-medium text-text-primary shadow-sm ring-1 ring-black/[0.06]',
                !active && 'text-text-tertiary hover:text-text-secondary',
              )}
            >
              {item === 'live' ? t('settings.discord.logsLive') : t('settings.discord.logsManual')}
            </button>
          )
        })}
      </div>

      {mode === 'live' ? (
        <button
          type="button"
          disabled={loading}
          title={t('settings.discord.logsRefreshNow')}
          onClick={onRefresh}
          className={cn(
            'inline-flex items-center gap-1.5 text-xs tabular-nums text-text-tertiary',
            heSpring,
            'hover:text-text-secondary disabled:opacity-60',
          )}
        >
          <RefreshCw className={cn('size-3 shrink-0', loading && 'animate-spin')} />
          {updatedAt
            ? t('settings.discord.logsUpdatedAt', { time: formatTimeHMS(updatedAt) })
            : t('settings.discord.logsWaitingFirst')}
        </button>
      ) : (
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={loading}
          className={cn(hePressable, 'h-7 gap-1.5 rounded-lg px-2.5 text-xs')}
          onClick={onRefresh}
        >
          <RefreshCw className={cn('size-3.5', loading && 'animate-spin')} />
          {t('common.refresh')}
        </Button>
      )}
    </div>
  )
}

function ServiceStatus({
  phase,
  pid,
  readyAt,
}: {
  phase: DiscordTransportPhase
  pid?: number
  readyAt?: string
}) {
  return <DiscordTransportStatusBadge phase={phase} pid={pid} readyAt={readyAt} />
}

export function DiscordTransportLogsDialog() {
  const { t } = useI18n()
  const preRef = useRef<HTMLPreElement>(null)
  const [open, setOpen] = useState(false)
  const [refreshMode, setRefreshMode] = useState<RefreshMode>('manual')
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null)
  const [phase, setPhase] = useState<DiscordTransportPhase>('stopped')
  const [pid, setPid] = useState<number | undefined>()
  const [readyAt, setReadyAt] = useState('')
  const [lines, setLines] = useState('')
  const [logPath, setLogPath] = useState('')
  const [lastExit, setLastExit] = useState('')
  const [loading, setLoading] = useState(false)
  const [clearing, setClearing] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const [logs, status] = await Promise.all([
        fetchDiscordTransportLogs(500),
        fetchDiscordTransportStatus(),
      ])
      setLines(logs.lines)
      setLogPath(logs.path || status.log_path || '')
      setLastExit(status.last_exit ?? '')
      setPhase(resolveDiscordTransportPhase(status))
      setPid(status.pid)
      setReadyAt(status.ready_at ?? '')
      setUpdatedAt(new Date())
    } catch {
      setLines('')
      setPhase('stopped')
      setPid(undefined)
      setReadyAt('')
    } finally {
      setLoading(false)
    }
  }, [])

  const handleClose = useCallback(() => {
    setOpen(false)
  }, [])

  const handleClear = useCallback(async () => {
    setClearing(true)
    try {
      const logs = await clearDiscordTransportLogs()
      setLines(logs.lines)
      setLogPath(logs.path)
      setUpdatedAt(new Date())
      toast.success(t('settings.discord.logsClearSuccess'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('settings.discord.logsClearFailed'))
    } finally {
      setClearing(false)
    }
  }, [t])

  useEffect(() => {
    if (!open) return
    void refresh()
  }, [open, refresh])

  useEffect(() => {
    if (!open || refreshMode !== 'live') return
    const timer = window.setInterval(() => void refresh(), LIVE_INTERVAL_MS)
    return () => window.clearInterval(timer)
  }, [open, refreshMode, refresh])

  useEffect(() => {
    if (!open) return
    const el = preRef.current
    if (el) {
      el.scrollTop = el.scrollHeight
    }
  }, [lines, open])

  useEffect(() => {
    if (!open) return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        handleClose()
      }
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [open, handleClose])

  const lineCount = lines ? lines.split('\n').length : 0

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className={cn(
          hePressable,
          'h-8 gap-1.5 rounded-xs border-0 bg-transparent px-2.5 text-sm font-normal text-text-secondary shadow-none',
          'hover:!bg-black/[0.04] hover:!text-text-primary',
          'active:!bg-black/[0.06]',
        )}
        onClick={() => setOpen(true)}
      >
        <ScrollText className="size-3.5 text-text-tertiary" />
        {t('settings.discord.logsOpen')}
      </Button>

      <Dialog open={open} onClose={handleClose}>
        <DialogContent
          size="lg"
          padded={false}
          className="flex h-[min(48rem,calc(100vh-3rem))] flex-col"
          aria-labelledby="discord-logs-title"
        >
              <div className="flex items-center justify-between gap-4 px-6 py-4">
                <h2 id="discord-logs-title" className={heSectionTitle}>
                  {t('settings.discord.logsTitle')}
                </h2>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className={cn(hePressable, 'size-8 rounded-xl text-text-tertiary hover:text-text-secondary')}
                  aria-label={t('common.close')}
                  onClick={handleClose}
                >
                  <X className="size-4" />
                </Button>
              </div>

              <div className="flex flex-wrap items-center justify-between gap-3 border-y border-border-subtle bg-canvas/30 px-6 py-2.5">
                <div className="flex flex-wrap items-center gap-4">
                  <ServiceStatus phase={phase} pid={pid} readyAt={readyAt || undefined} />
                  <RefreshModeControl
                    mode={refreshMode}
                    loading={loading}
                    updatedAt={updatedAt}
                    onModeChange={setRefreshMode}
                    onRefresh={() => void refresh()}
                  />
                  {lineCount > 0 && (
                    <span className="text-xs tabular-nums text-text-tertiary">
                      {t('settings.discord.logsLineCount', { count: lineCount })}
                    </span>
                  )}
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={clearing || loading}
                  title={t('settings.discord.logsClearTitle')}
                  className={cn(
                    hePressable,
                    'gap-1.5 rounded-lg text-destructive hover:bg-destructive hover:text-white',
                  )}
                  onClick={() => void handleClear()}
                >
                  <Trash2 className={cn('size-3.5', clearing && 'animate-pulse')} />
                  {t('settings.discord.logsClear')}
                </Button>
              </div>

              {lastExit && (
                <div className="border-b border-destructive/15 bg-destructive/5 px-6 py-2 text-xs text-destructive">
                  {t('settings.discord.logsLastExit', { message: lastExit })}
                </div>
              )}

              <div className="min-h-0 flex-1 p-4">
                <div className={cn(heFormEmbed, 'flex h-full min-h-0 flex-col overflow-hidden')}>
                  <pre
                    ref={preRef}
                    className={cn(
                      'min-h-0 flex-1 overflow-auto p-4',
                      'font-mono text-[13px] leading-[1.65] text-text-secondary',
                    )}
                  >
                    {loading && !lines
                      ? t('settings.discord.logsLoading')
                      : lines || t('settings.discord.logsEmpty')}
                  </pre>
                </div>
              </div>

              {logPath && (
                <div className="border-t border-border-subtle px-6 py-2.5">
                  <p className={cn(heFieldHint, 'truncate font-mono text-[11px]')} title={logPath}>
                    {logPath}
                  </p>
                </div>
              )}
            </DialogContent>
      </Dialog>
    </>
  )
}
