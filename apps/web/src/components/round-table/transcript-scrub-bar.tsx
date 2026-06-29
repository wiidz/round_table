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

  return (
    <div
      className={cn(
        'flex shrink-0 items-center gap-3 border-t border-black/[0.05] bg-black/[0.02] px-5 py-2.5',
        className,
      )}
    >
      <button
        type="button"
        onClick={() => onScrubTurnChange(null)}
        className={cn(
          'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
          isLive
            ? 'bg-ai-soft text-ai ring-1 ring-ai/25'
            : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
        )}
      >
        Live
      </button>

      <label className="flex min-w-0 flex-1 items-center gap-2">
        <span className="shrink-0 font-mono text-[11px] tabular-nums text-text-tertiary">
          #{value}
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
          / {maxTurn}
        </span>
      </label>
    </div>
  )
}
