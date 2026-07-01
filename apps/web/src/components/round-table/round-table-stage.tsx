import { RoundTableEmptyHint } from '@/components/round-table/round-table-empty-hint'
import { useI18n } from '@/hooks/use-i18n'
import { resolveLiveBubbleVariant } from '@/lib/live-bubble-variant'
import type { SeatLayout } from '@/lib/round-table-layout'
import { cn } from '@/lib/utils'
import type { ChatMessage, TypingStates } from '@/types/chat'

import { SeatAnchor } from './seat-anchor'

interface RoundTableStageProps {
  seats: SeatLayout[]
  latestBySeat: Map<string, ChatMessage>
  highlightMessageId: string | null
  referenceTurn: number | null
  focusedSeatId?: string | null
  typingStates?: TypingStates
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

function defaultCenterTitle(t: (key: string) => string, turnCount: number): string {
  if (turnCount > 0) return t('roundTable.stage.inProgress')
  return t('roundTable.stage.waitingTopic')
}

export function RoundTableStage({
  seats,
  latestBySeat,
  highlightMessageId,
  referenceTurn,
  focusedSeatId = null,
  typingStates,
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
  const { t } = useI18n()
  const spokenSeats = new Set<string>()
  for (const [seatId] of latestBySeat) {
    spokenSeats.add(seatId)
  }

  const title = centerTitle ?? defaultCenterTitle(t, turnCount)
  const subtitle =
    centerSubtitle ??
    (turnCount > 0
      ? t('roundTable.stage.turnCount', { count: turnCount })
      : t('roundTable.stage.seatHint'))

  return (
    <div className={cn('relative min-h-0 flex-1 overflow-hidden', className)}>
      {/* 背景光晕 */}
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_72%_62%_at_50%_50%,var(--ai-soft)_0%,transparent_68%)] opacity-55" />

      {/* 圆桌中心卡片 */}
      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-[15] -translate-x-1/2 -translate-y-1/2"
        aria-hidden
      >
        <div className="flex h-[8rem] w-[12rem] flex-col items-center justify-center rounded-[50%] bg-surface/90 px-4 text-center shadow-md ring-1 ring-black/[0.07] sm:h-[9rem] sm:w-[13rem]">
          <p className="line-clamp-2 text-[13px] font-semibold leading-snug text-text-primary">{title}</p>
          <p className="mt-1.5 line-clamp-2 text-[11px] leading-relaxed text-text-tertiary">{subtitle}</p>
        </div>
      </div>

      {/* 椭圆虚线轮廓 */}
      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-0 -translate-x-1/2 -translate-y-1/2 rounded-[50%] border border-dashed border-black/[0.06]"
        style={{ width: '92%', height: '82%' }}
        aria-hidden
      />

      {/* 席位层 */}
      <div className="relative z-10 h-full min-h-[20rem] w-full px-3 pb-8 pt-4 sm:px-6 md:px-8 lg:px-10">
        {seats.map((seat) => {
          const liveMessage = latestBySeat.get(seat.id) ?? null
          const highlighted = liveMessage != null && liveMessage.id === highlightMessageId
          const bubbleVariant = liveMessage
            ? resolveLiveBubbleVariant(liveMessage, highlightMessageId, referenceTurn)
            : 'before'
          const focused = focusedSeatId === seat.id && !highlighted
          const typing = typingStates?.get(seat.id) ?? null

          return (
            <SeatAnchor
              key={seat.id}
              seat={seat}
              liveMessage={liveMessage}
              bubbleVariant={bubbleVariant}
              typing={typing}
              highlighted={highlighted}
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
