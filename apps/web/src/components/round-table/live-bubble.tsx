import { TranscriptSequenceBadge, sequenceBadgeToneFromLiveBubble } from '@/components/round-table/transcript-sequence-badge'
import { condenseMessage } from '@/lib/condense-message'
import { formatChatTime } from '@/lib/format-date'
import type { BubbleTail } from '@/lib/round-table-layout'
import { transcriptBubbleBadgePaddingTop } from '@/lib/transcript-speaker-label'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

export type LiveBubbleVariant = 'active' | 'before' | 'after'

const LIVE_BUBBLE_VARIANT = {
  active: {
    maxChars: 200,
    maxWidth: 'max-w-[min(100%,24rem)] sm:max-w-[min(100%,26rem)]',
    lineClamp: 'line-clamp-[8]',
    textSize: 'text-[13px] leading-relaxed sm:text-[14px] sm:leading-relaxed',
    padding: '!px-5 !pt-4 !pb-6',
    timeInset: 'left-5 bottom-1',
    shellClass: '',
  },
  before: {
    maxChars: 44,
    maxWidth: 'max-w-[9.5rem] sm:max-w-[10.5rem]',
    lineClamp: 'line-clamp-2',
    textSize: 'text-[11px] leading-snug',
    padding: '!px-2.5 !pt-2 !pb-2.5',
    paddingWithBadge: `!px-2.5 ${transcriptBubbleBadgePaddingTop} !pb-2.5`,
    shellClass: 'opacity-75',
  },
  after: {
    maxChars: 44,
    maxWidth: 'max-w-[9.5rem] sm:max-w-[10.5rem]',
    lineClamp: 'line-clamp-2',
    textSize: 'text-[11px] leading-snug text-text-tertiary',
    padding: '!px-2.5 !pt-2 !pb-2.5',
    paddingWithBadge: `!px-2.5 ${transcriptBubbleBadgePaddingTop} !pb-2.5`,
    shellClass: 'opacity-50 saturate-[0.65]',
  },
} as const

interface LiveBubbleProps {
  message: ChatMessage
  tail: BubbleTail
  variant?: LiveBubbleVariant
  highlighted?: boolean
  /** @deprecated use variant="before" */
  compact?: boolean
  sequence?: number | null
  onClick?: () => void
  className?: string
}

function bubbleToneClass(message: ChatMessage): string {
  if (message.role === 'moderator') return 'chat-bubble chat-bubble--moderator'
  if (message.role === 'user') return 'chat-bubble chat-bubble--user'
  return 'chat-bubble chat-bubble--participant'
}

function tailClass(tail: BubbleTail): string {
  switch (tail) {
    case 'top':
      return 'chat-bubble--tail-top'
    case 'bottom':
      return 'chat-bubble--tail-bottom'
    case 'right':
      return 'chat-bubble--tail-right'
    default:
      return 'chat-bubble--tail-left'
  }
}

function bubbleTimeClass(message: ChatMessage, variant: LiveBubbleVariant): string {
  if (message.role === 'user') return 'text-white/45'
  if (variant === 'after') return 'text-text-tertiary/55'
  return 'text-text-tertiary/70'
}

export function LiveBubble({
  message,
  tail,
  variant = 'active',
  highlighted = false,
  compact = false,
  sequence,
  onClick,
  className,
}: LiveBubbleProps) {
  const resolvedVariant: LiveBubbleVariant = compact ? 'before' : variant
  const config = LIVE_BUBBLE_VARIANT[resolvedVariant]
  const { summary } = condenseMessage(message.content, config.maxChars)
  const sequenceNo = sequence ?? message.turn ?? null
  const timeLabel = formatChatTime(message.createdAt)
  const shellPadding =
    sequenceNo != null && resolvedVariant !== 'active'
      ? LIVE_BUBBLE_VARIANT[resolvedVariant].paddingWithBadge
      : config.padding
  const isCompact = resolvedVariant !== 'active'

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group relative block w-full text-left transition-all duration-200 animate-bubble-in',
        config.maxWidth,
        highlighted && 'z-20 scale-100 opacity-100',
        !highlighted && resolvedVariant === 'before' && 'z-[8]',
        !highlighted && resolvedVariant === 'after' && 'z-[5]',
        className,
      )}
    >
      {sequenceNo != null && (
        <TranscriptSequenceBadge
          sequence={sequenceNo}
          tone={sequenceBadgeToneFromLiveBubble(highlighted, resolvedVariant)}
        />
      )}

      <div
        className={cn(
          bubbleToneClass(message),
          tailClass(tail),
          shellPadding,
          config.shellClass,
          'chat-bubble--live chat-bubble--interactive relative',
          resolvedVariant === 'after' && !highlighted && 'ring-1 ring-dashed ring-black/[0.08]',
          highlighted && 'chat-bubble--live-highlight',
        )}
      >
        <p className={cn(config.lineClamp, config.textSize)}>
          {summary || '…'}
        </p>
        {timeLabel &&
          (isCompact ? (
            <span
              className={cn(
                'block text-[9px] font-normal tabular-nums',
                bubbleTimeClass(message, resolvedVariant),
              )}
            >
              {timeLabel}
            </span>
          ) : (
            <span
              className={cn(
                'pointer-events-none absolute text-[9px] font-normal tabular-nums',
                config.timeInset,
                bubbleTimeClass(message, resolvedVariant),
              )}
            >
              {timeLabel}
            </span>
          ))}
      </div>
    </button>
  )
}

// ── Typing Bubble ────────────────────────────────────────────────────────────

interface TypingBubbleProps {
  tail: BubbleTail
  role: 'moderator' | 'participant' | 'user'
  compact?: boolean
  className?: string
}

export function TypingBubble({ tail, role, compact = false, className }: TypingBubbleProps) {
  const toneClass =
    role === 'moderator'
      ? 'chat-bubble chat-bubble--moderator'
      : role === 'user'
        ? 'chat-bubble chat-bubble--user'
        : 'chat-bubble chat-bubble--participant'

  return (
    <div
      className={cn(
        'relative block animate-bubble-in',
        compact ? 'max-w-[9rem]' : 'max-w-[8rem]',
        'z-10 opacity-90',
        className,
      )}
    >
      <div className={cn(toneClass, tailClass(tail), 'chat-bubble--live', '!px-4 !pt-3 !pb-4')}>
        <span className="flex items-center gap-1.5">
          {[0, 1, 2].map((i) => (
            <span
              key={i}
              className="inline-block size-1.5 rounded-full bg-current opacity-60"
              style={{ animation: `typing-dot 1.2s ease-in-out ${i * 0.18}s infinite` }}
            />
          ))}
        </span>
      </div>
    </div>
  )
}
