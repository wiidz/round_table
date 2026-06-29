import { condenseMessage } from '@/lib/condense-message'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface LiveBubbleProps {
  message: ChatMessage
  tail: 'left' | 'right'
  highlighted?: boolean
  dimmed?: boolean
  onClick?: () => void
  className?: string
}

function bubbleToneClass(message: ChatMessage): string {
  if (message.role === 'moderator') {
    return 'chat-bubble chat-bubble--moderator'
  }
  return 'chat-bubble chat-bubble--participant'
}

function tailClass(tail: 'left' | 'right'): string {
  return tail === 'left' ? 'chat-bubble--tail-left' : 'chat-bubble--tail-right'
}

export function LiveBubble({
  message,
  tail,
  highlighted = false,
  dimmed = false,
  onClick,
  className,
}: LiveBubbleProps) {
  const { summary } = condenseMessage(message.content, 72)

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group relative block max-w-[9.5rem] text-left transition-all duration-200 sm:max-w-[11rem]',
        highlighted && 'z-20 scale-100 opacity-100',
        dimmed && 'z-0 scale-[0.97] opacity-45 saturate-[0.55]',
        !highlighted && !dimmed && 'z-10 opacity-80',
        className,
      )}
    >
      {message.turn != null && (
        <span
          className={cn(
            'absolute -top-2 z-10 rounded-md px-1.5 py-0.5 font-mono text-[10px] font-semibold tabular-nums shadow-sm',
            highlighted
              ? 'bg-ai text-white ring-1 ring-ai/30'
              : 'bg-surface text-text-secondary ring-1 ring-black/[0.08]',
          )}
        >
          #{message.turn}
        </span>
      )}

      <div
        className={cn(
          bubbleToneClass(message),
          tailClass(tail),
          'px-3 py-2',
          highlighted && 'chat-bubble--live-highlight',
          dimmed && 'chat-bubble--live-dimmed',
        )}
      >
        <p className="line-clamp-3 text-[12px] leading-snug">{summary || '…'}</p>
        <span className="mt-1 block text-[10px] text-text-tertiary opacity-0 transition-opacity group-hover:opacity-100">
          点击查看全文
        </span>
      </div>
    </button>
  )
}
