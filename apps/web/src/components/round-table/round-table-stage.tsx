import { speakerId } from '@/lib/chat-display'
import type { SeatLayout } from '@/lib/round-table-layout'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

import { SeatAnchor } from './seat-anchor'

interface RoundTableStageProps {
  seats: SeatLayout[]
  messages: ChatMessage[]
  activeSpeakerId: string | null
  turnCount: number
  centerTitle?: string
  centerSubtitle?: string
  className?: string
}

function defaultCenterTitle(turnCount: number): string {
  if (turnCount > 0) return '会议进行中'
  return '等待议题'
}

export function RoundTableStage({
  seats,
  messages,
  activeSpeakerId,
  turnCount,
  centerTitle,
  centerSubtitle,
  className,
}: RoundTableStageProps) {
  const spokenSeats = new Set<string>()
  for (const message of messages) {
    if (message.turn == null) continue
    spokenSeats.add(speakerId(message))
  }

  const title = centerTitle ?? defaultCenterTitle(turnCount)
  const subtitle =
    centerSubtitle ??
    (turnCount > 0 ? `已记录 ${turnCount} 轮发言 · Live 气泡 M3` : '发起会议后专家将入座')

  return (
    <div className={cn('relative min-h-0 flex-1 overflow-hidden', className)}>
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_70%_60%_at_50%_50%,var(--ai-soft)_0%,transparent_70%)] opacity-60" />

      {/* Table surface */}
      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-0 -translate-x-1/2 -translate-y-1/2"
        aria-hidden
      >
        <div className="flex h-[7.5rem] w-[11rem] flex-col items-center justify-center rounded-[50%] bg-surface/90 px-4 text-center shadow-sm ring-1 ring-black/[0.08] sm:h-[8.5rem] sm:w-[12.5rem]">
          <p className="text-[13px] font-semibold leading-snug text-text-primary">{title}</p>
          <p className="mt-1.5 text-[11px] leading-relaxed text-text-tertiary">{subtitle}</p>
        </div>
      </div>

      {/* Ellipse guide (decorative) */}
      <div
        className="pointer-events-none absolute left-1/2 top-1/2 z-0 -translate-x-1/2 -translate-y-1/2 rounded-[50%] border border-dashed border-black/[0.07]"
        style={{ width: '80%', height: '68%' }}
        aria-hidden
      />

      <div className="relative z-10 h-full min-h-[14rem] w-full">
        {seats.map((seat) => (
          <SeatAnchor
            key={seat.id}
            seat={seat}
            active={activeSpeakerId === seat.id}
            hasSpoken={spokenSeats.has(seat.id)}
          />
        ))}
      </div>

      {seats.filter((s) => s.kind === 'participant').length === 0 && (
        <p className="absolute bottom-3 left-0 right-0 text-center text-[11px] text-text-tertiary">
          专家名录加载中或 roster 为空；发言后将按 author 入座
        </p>
      )}
    </div>
  )
}
