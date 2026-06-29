import { RoundTableEmptyHint } from '@/components/round-table/round-table-empty-hint'
import type { SeatLayout } from '@/lib/round-table-layout'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

import { SeatAnchor } from './seat-anchor'

interface RoundTableStageProps {
  seats: SeatLayout[]
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  focusedSeatId?: string | null
  turnCount: number
  rosterLoading?: boolean
  rosterFromApi?: boolean
  rosterTotal?: number
  seatedExpertCount?: number
  centerTitle?: string
  centerSubtitle?: string
  onLiveMessageClick?: (message: ChatMessage) => void
  className?: string
}

function defaultCenterTitle(turnCount: number): string {
  if (turnCount > 0) return '会议进行中'
  return '等待议题'
}

export function RoundTableStage({
  seats,
  latestBySeat,
  activeSpeakerId,
  focusedSeatId = null,
  turnCount,
  rosterLoading = false,
  rosterFromApi = false,
  rosterTotal = 0,
  seatedExpertCount = 0,
  centerTitle,
  centerSubtitle,
  onLiveMessageClick,
  className,
}: RoundTableStageProps) {
  const spokenSeats = new Set<string>()
  for (const [seatId] of latestBySeat) {
    spokenSeats.add(seatId)
  }

  const title = centerTitle ?? defaultCenterTitle(turnCount)
  const subtitle =
    centerSubtitle ??
    (turnCount > 0 ? `第 ${turnCount} 轮发言` : '发起会议后专家将入座')

  return (
    <div className={cn('relative min-h-0 flex-1 overflow-hidden', className)}>
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_70%_60%_at_50%_50%,var(--ai-soft)_0%,transparent_70%)] opacity-60" />

      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-[15] -translate-x-1/2 -translate-y-1/2"
        aria-hidden
      >
        <div className="flex h-[7rem] w-[10.5rem] flex-col items-center justify-center rounded-[50%] bg-surface px-4 text-center shadow-sm ring-1 ring-black/[0.08] sm:h-[7.5rem] sm:w-[11.5rem]">
          <p className="line-clamp-2 text-[13px] font-semibold leading-snug text-text-primary">{title}</p>
          <p className="mt-1.5 line-clamp-2 text-[11px] leading-relaxed text-text-tertiary">{subtitle}</p>
        </div>
      </div>

      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-0 -translate-x-1/2 -translate-y-1/2 rounded-[50%] border border-dashed border-black/[0.07]"
        style={{ width: '80%', height: '68%' }}
        aria-hidden
      />

      <div className="relative z-10 h-full min-h-[14rem] w-full px-3 pb-6 pt-4 sm:px-6 md:px-8">
        {seats.map((seat) => {
          const liveMessage = latestBySeat.get(seat.id) ?? null
          const highlighted = activeSpeakerId === seat.id && liveMessage != null
          const dimmed =
            activeSpeakerId != null &&
            activeSpeakerId !== seat.id &&
            liveMessage != null
          const focused = focusedSeatId === seat.id && !highlighted

          return (
            <SeatAnchor
              key={seat.id}
              seat={seat}
              liveMessage={liveMessage}
              highlighted={highlighted}
              dimmed={dimmed}
              focused={focused}
              hasSpoken={spokenSeats.has(seat.id)}
              onLiveClick={onLiveMessageClick}
            />
          )
        })}
      </div>

      <RoundTableEmptyHint
        loading={rosterLoading}
        rosterFromApi={rosterFromApi}
        rosterTotal={rosterTotal}
        seatedExpertCount={seatedExpertCount}
      />
    </div>
  )
}
