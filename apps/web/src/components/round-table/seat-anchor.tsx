import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { LiveBubble, TypingBubble, type LiveBubbleVariant } from '@/components/round-table/live-bubble'
import {
  liveBubbleSlotClass,
  seatAnchorTransform,
  seatBubbleTailClass,
  type SeatLayout,
} from '@/lib/round-table-layout'
import { transcriptSpeakerLabelClass } from '@/lib/transcript-speaker-label'
import { cn } from '@/lib/utils'
import type { ChatMessage, TypingState } from '@/types/chat'

interface SeatAnchorProps {
  seat: SeatLayout
  liveMessage?: ChatMessage | null
  bubbleVariant?: LiveBubbleVariant
  typing?: TypingState | null
  highlighted?: boolean
  focused?: boolean
  hasSpoken?: boolean
  onLiveClick?: (message: ChatMessage) => void
  className?: string
}

export function SeatAnchor({
  seat,
  liveMessage,
  bubbleVariant = 'before',
  typing = null,
  highlighted = false,
  focused = false,
  hasSpoken = false,
  onLiveClick,
  className,
}: SeatAnchorProps) {
  const hasLive = liveMessage != null
  const isTyping = typing != null && !hasLive
  const showBubble = hasLive || isTyping
  const tail = seatBubbleTailClass(seat)
  const expanded = bubbleVariant === 'active'
  const labelClass = transcriptSpeakerLabelClass({
    highlighted,
    focused,
    muted: bubbleVariant === 'after' && !highlighted,
  })

  return (
    <div
      className={cn('absolute', className)}
      style={{
        left: `${seat.x}%`,
        top: `${seat.y}%`,
        transform: seatAnchorTransform(seat, showBubble),
      }}
    >
      <div className="animate-seat-in relative">
        <div
          className={cn(
            'relative z-10 shrink-0 rounded-xl transition-[box-shadow] duration-300',
            highlighted && 'ring-2 ring-ai ring-offset-2 ring-offset-surface animate-speaker-pulse',
            focused && !highlighted && 'ring-2 ring-brand ring-offset-2 ring-offset-surface',
            !highlighted && !focused && !showBubble && hasSpoken && 'opacity-90',
            !highlighted && !focused && !showBubble && !hasSpoken && 'opacity-75',
          )}
        >
          <ProfileAvatar
            id={seat.id}
            name={seat.label}
            size="sm"
            className={cn(
              highlighted && 'shadow-[0_0_0_3px_var(--ai-soft)]',
              focused && !highlighted && 'shadow-[0_0_0_3px_var(--brand-soft)]',
            )}
          />
          {hasSpoken && !showBubble && (
            <span
              className="absolute -bottom-0.5 -right-0.5 size-2 rounded-full bg-ai ring-2 ring-surface"
              aria-hidden
            />
          )}
        </div>

        {hasLive && seat.kind === 'moderator' ? (
          <div
            className={cn(
              liveBubbleSlotClass(seat, expanded),
              'flex flex-col items-center gap-1.5',
            )}
          >
            <LiveBubble
              message={liveMessage}
              tail={tail}
              variant={bubbleVariant}
              highlighted={highlighted}
              onClick={() => onLiveClick?.(liveMessage)}
              className="!max-w-none w-full"
            />
            <p className={labelClass}>{seat.label}</p>
          </div>
        ) : null}

        {hasLive && seat.kind !== 'moderator' && (
          <div className={liveBubbleSlotClass(seat, expanded)}>
            <LiveBubble
              message={liveMessage}
              tail={tail}
              variant={bubbleVariant}
              highlighted={highlighted}
              onClick={() => onLiveClick?.(liveMessage)}
              className="!max-w-none w-full"
            />
          </div>
        )}

        {isTyping && (
          <div className={liveBubbleSlotClass(seat, false)}>
            <TypingBubble
              tail={tail}
              role={typing.role === 'user' ? 'user' : typing.role === 'moderator' ? 'moderator' : 'participant'}
              className="!max-w-none w-full"
            />
          </div>
        )}

        {!(seat.kind === 'moderator' && hasLive) && (
          <p
            className={cn(
              labelClass,
              'absolute left-1/2 top-[calc(100%+0.375rem)] z-10 -translate-x-1/2',
            )}
          >
            {seat.label}
          </p>
        )}
      </div>
    </div>
  )
}
