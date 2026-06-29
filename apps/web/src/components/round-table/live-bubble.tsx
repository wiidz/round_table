import { condenseMessage } from '@/lib/condense-message'
import type { BubbleTail } from '@/lib/round-table-layout'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface LiveBubbleProps {
  message: ChatMessage
  tail: BubbleTail
  highlighted?: boolean
  dimmed?: boolean
  compact?: boolean
  sequence?: number | null
  onClick?: () => void
  className?: string
}

function bubbleToneClass(message: ChatMessage): string {
  if (message.role === 'moderator') {
    return 'chat-bubble chat-bubble--moderator'
  }
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

export function LiveBubble({
  message,
  tail,
  highlighted = false,
  dimmed = false,
  compact = false,
  sequence,
  onClick,
  className,
}: LiveBubbleProps) {
  const { summary } = condenseMessage(message.content, compact ? 72 : 96)
  const sequenceNo = sequence ?? message.turn ?? null

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group relative block w-full text-left transition-all duration-200',
        compact
          ? 'max-w-[11rem] sm:max-w-[12rem]'
          : 'max-w-[min(100%,15rem)] sm:max-w-[min(100%,16rem)]',
        highlighted && 'z-20 scale-100 opacity-100',
        dimmed && 'z-0 scale-[0.97] opacity-45 saturate-[0.55]',
        !highlighted && !dimmed && 'z-10 opacity-80',
        className,
      )}
    >
      {sequenceNo != null && (
        <span
          className={cn(
            'absolute -top-2 z-10 rounded-md px-2 py-0.5 font-mono text-[11px] font-semibold tabular-nums shadow-sm',
            highlighted
              ? 'bg-ai text-white ring-1 ring-ai/30'
              : 'bg-surface text-text-secondary ring-1 ring-black/[0.08]',
          )}
        >
          #{sequenceNo}
        </span>
      )}

      <div
        className={cn(
          bubbleToneClass(message),
          tailClass(tail),
          'px-4 py-3',
          highlighted && 'chat-bubble--live-highlight',
          dimmed && 'chat-bubble--live-dimmed',
        )}
      >
        <p
          className={cn(
            compact ? 'line-clamp-2 text-[12px] leading-snug' : 'line-clamp-4 text-[13px] leading-relaxed sm:text-[14px]',
          )}
        >
          {summary || '…'}
        </p>
        <span className="mt-1.5 block text-[11px] text-text-tertiary opacity-0 transition-opacity group-hover:opacity-100">
          点击查看全文
        </span>
      </div>
    </button>
  )
}
