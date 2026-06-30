import { ChevronLeft, ChevronRight } from 'lucide-react'

import { cn } from '@/lib/utils'

interface TranscriptScrubBarProps {
  maxTurn: number
  scrubTurn: number | null
  onScrubTurnChange: (turn: number | null) => void
  className?: string
}

export function TranscriptScrubBar({
  maxTurn,
  scrubTurn,
  onScrubTurnChange,
  className,
}: TranscriptScrubBarProps) {
  if (maxTurn < 2) return null

  const value = scrubTurn ?? maxTurn
  const isLive = scrubTurn == null

  const stepPrev = () => {
    const next = value - 1
    if (next < 1) return
    onScrubTurnChange(next)
  }

  const stepNext = () => {
    const next = value + 1
    if (next >= maxTurn) {
      onScrubTurnChange(null)
    } else {
      onScrubTurnChange(next)
    }
  }

  return (
    <div
      className={cn(
        'flex shrink-0 items-center gap-2.5 border-t border-black/[0.05] bg-black/[0.02] px-4 py-2.5',
        className,
      )}
    >
      {/* Live toggle */}
      <button
        type="button"
        onClick={() => onScrubTurnChange(null)}
        title="跳到最新（Live 模式）"
        className={cn(
          'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
          isLive
            ? 'bg-ai-soft text-ai ring-1 ring-ai/25'
            : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
        )}
      >
        Live
      </button>

      {/* Step prev / next */}
      <button
        type="button"
        onClick={stepPrev}
        disabled={value <= 1}
        title="上一步"
        aria-label="上一步"
        className={cn(
          'inline-flex shrink-0 items-center gap-1 rounded-lg px-2.5 py-1.5',
          'bg-surface text-[12px] font-medium text-text-secondary',
          'ring-1 ring-inset ring-black/[0.08] shadow-sm',
          'transition-[color,background-color,box-shadow,opacity] duration-200',
          'hover:bg-black/[0.03] hover:text-text-primary hover:ring-black/[0.12]',
          'disabled:cursor-not-allowed disabled:opacity-35 disabled:hover:bg-surface',
        )}
      >
        <ChevronLeft className="size-4 shrink-0" strokeWidth={2} aria-hidden />
        <span className="hidden sm:inline">上一步</span>
      </button>

      {/* Slider */}
      <label className="flex min-w-0 flex-1 items-center gap-1.5">
        <span className="shrink-0 font-mono text-[11px] tabular-nums text-text-tertiary">
          {value}
        </span>
        <input
          type="range"
          min={1}
          max={maxTurn}
          value={value}
          onChange={(event) => {
            const next = Number.parseInt(event.target.value, 10)
            if (next >= maxTurn) {
              onScrubTurnChange(null)
            } else {
              onScrubTurnChange(next)
            }
          }}
          className="h-1.5 min-w-0 flex-1 cursor-pointer accent-brand"
          aria-label="发言回放"
        />
        <span className="shrink-0 font-mono text-[11px] tabular-nums text-text-tertiary">
          {maxTurn}
        </span>
      </label>

      <button
        type="button"
        onClick={stepNext}
        disabled={isLive}
        title="下一步"
        aria-label="下一步"
        className={cn(
          'inline-flex shrink-0 items-center gap-1 rounded-lg px-2.5 py-1.5',
          'bg-surface text-[12px] font-medium text-text-secondary',
          'ring-1 ring-inset ring-black/[0.08] shadow-sm',
          'transition-[color,background-color,box-shadow,opacity] duration-200',
          'hover:bg-black/[0.03] hover:text-text-primary hover:ring-black/[0.12]',
          'disabled:cursor-not-allowed disabled:opacity-35 disabled:hover:bg-surface',
        )}
      >
        <span className="hidden sm:inline">下一步</span>
        <ChevronRight className="size-4 shrink-0" strokeWidth={2} aria-hidden />
      </button>
    </div>
  )
}
