import { cn } from '@/lib/utils'

export type TranscriptSequenceBadgeTone = 'highlight' | 'muted' | 'default'

export function TranscriptSequenceBadge({
  sequence,
  tone = 'default',
  className,
}: {
  sequence: number
  tone?: TranscriptSequenceBadgeTone
  className?: string
}) {
  return (
    <span
      className={cn(
        'absolute -top-2 left-2.5 z-10 rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tabular-nums shadow-sm ring-1',
        tone === 'highlight' && 'bg-ai text-white ring-ai/30',
        tone === 'muted' && 'bg-surface text-text-tertiary ring-black/[0.08]',
        tone === 'default' && 'bg-surface text-text-secondary ring-black/[0.10]',
        className,
      )}
    >
      #{sequence}
    </span>
  )
}

export function sequenceBadgeToneFromLiveBubble(
  highlighted: boolean,
  variant: 'active' | 'before' | 'after',
): TranscriptSequenceBadgeTone {
  if (highlighted) return 'highlight'
  if (variant === 'after') return 'muted'
  return 'default'
}

export function sequenceBadgeToneFromHistoryItem(selected: boolean, active: boolean): TranscriptSequenceBadgeTone {
  if (selected) return 'highlight'
  if (active) return 'default'
  return 'default'
}
