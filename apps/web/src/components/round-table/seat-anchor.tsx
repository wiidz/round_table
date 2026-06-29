import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { LiveBubble } from '@/components/round-table/live-bubble'
import { seatBubbleTailClass, type SeatLayout } from '@/lib/round-table-layout'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface SeatAnchorProps {
  seat: SeatLayout
  liveMessage?: ChatMessage | null
  highlighted?: boolean
  dimmed?: boolean
  focused?: boolean
  hasSpoken?: boolean
  onLiveClick?: (message: ChatMessage) => void
  className?: string
}

function contentLayoutClass(seat: SeatLayout, hasLive: boolean): string {
  if (!hasLive) return 'flex-col items-center gap-1'
  if (seat.kind === 'moderator') return 'flex-col items-center gap-1'
  if (seat.kind === 'principal') return 'flex-col-reverse items-center gap-1'
  if (seat.x < 50) return 'flex-row items-start gap-1.5'
  return 'flex-row-reverse items-start gap-1.5'
}

export function SeatAnchor({
  seat,
  liveMessage,
  highlighted = false,
  dimmed = false,
  focused = false,
  hasSpoken = false,
  onLiveClick,
  className,
}: SeatAnchorProps) {
  const hasLive = liveMessage != null
  const tail = seatBubbleTailClass(seat)

  return (
    <div
      className={cn('absolute', className)}
      style={{
        left: `${seat.x}%`,
        top: `${seat.y}%`,
        transform: 'translate(-50%, -50%)',
      }}
    >
      <div className={cn('flex', contentLayoutClass(seat, hasLive))}>
        <div
          className={cn(
            'relative shrink-0 rounded-xl transition-all duration-200',
            highlighted && 'ring-2 ring-ai ring-offset-2 ring-offset-surface',
            focused && !highlighted && 'ring-2 ring-brand ring-offset-2 ring-offset-surface',
            !highlighted && !focused && !hasLive && hasSpoken && 'opacity-90',
            !highlighted && !focused && !hasLive && !hasSpoken && 'opacity-75',
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
          {hasSpoken && !hasLive && (
            <span
              className="absolute -bottom-0.5 -right-0.5 size-2 rounded-full bg-ai ring-2 ring-surface"
              aria-hidden
            />
          )}
        </div>

        {hasLive && (
          <LiveBubble
            message={liveMessage}
            tail={tail}
            highlighted={highlighted}
            dimmed={dimmed}
            onClick={() => onLiveClick?.(liveMessage)}
          />
        )}
      </div>

      <p
        className={cn(
          'mx-auto mt-1 max-w-[5rem] truncate text-center text-[10px] font-medium',
          highlighted ? 'text-ai' : focused ? 'text-brand' : 'text-text-secondary',
          dimmed && 'text-text-tertiary',
        )}
      >
        {seat.label}
      </p>
    </div>
  )
}
