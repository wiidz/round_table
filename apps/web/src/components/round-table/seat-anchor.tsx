import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { cn } from '@/lib/utils'
import type { SeatLayout } from '@/lib/round-table-layout'

interface SeatAnchorProps {
  seat: SeatLayout
  active?: boolean
  hasSpoken?: boolean
  className?: string
}

export function SeatAnchor({ seat, active, hasSpoken, className }: SeatAnchorProps) {
  return (
    <div
      className={cn('absolute flex flex-col items-center gap-1', className)}
      style={{
        left: `${seat.x}%`,
        top: `${seat.y}%`,
        transform: 'translate(-50%, -50%)',
      }}
    >
      <div
        className={cn(
          'relative rounded-xl transition-all duration-200',
          active && 'ring-2 ring-ai ring-offset-2 ring-offset-surface',
          !active && hasSpoken && 'opacity-90',
          !active && !hasSpoken && 'opacity-70',
        )}
      >
        <ProfileAvatar
          id={seat.id}
          name={seat.label}
          size="sm"
          className={cn(active && 'shadow-[0_0_0_3px_var(--ai-soft)]')}
        />
        {hasSpoken && !active && (
          <span
            className="absolute -bottom-0.5 -right-0.5 size-2 rounded-full bg-ai ring-2 ring-surface"
            aria-hidden
          />
        )}
      </div>
      <span
        className={cn(
          'max-w-[4.5rem] truncate text-center text-[11px] font-medium',
          active ? 'text-ai' : 'text-text-secondary',
        )}
      >
        {seat.label}
      </span>
    </div>
  )
}
