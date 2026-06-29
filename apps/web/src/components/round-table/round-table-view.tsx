import { RoundTableStage } from '@/components/round-table/round-table-stage'
import { TranscriptStrip } from '@/components/round-table/transcript-strip'
import type { SeatLayout } from '@/lib/round-table-layout'
import type { ChatMessage } from '@/types/chat'

interface RoundTableViewProps {
  seats: SeatLayout[]
  messages: ChatMessage[]
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  focusedSeatId: string | null
  turnCount: number
  activeMessageId: string | null
  selectedMessageId: string | null
  rosterLoading: boolean
  rosterFromApi: boolean
  participantCount: number
  onSelectMessage: (message: ChatMessage) => void
}

export function RoundTableView({
  seats,
  messages,
  latestBySeat,
  activeSpeakerId,
  focusedSeatId,
  turnCount,
  activeMessageId,
  selectedMessageId,
  rosterLoading,
  rosterFromApi,
  participantCount,
  onSelectMessage,
}: RoundTableViewProps) {
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <RoundTableStage
        seats={seats}
        latestBySeat={latestBySeat}
        activeSpeakerId={activeSpeakerId}
        focusedSeatId={focusedSeatId}
        turnCount={turnCount}
        rosterLoading={rosterLoading}
        rosterFromApi={rosterFromApi}
        participantCount={participantCount}
        onLiveMessageClick={onSelectMessage}
        className="min-h-0 flex-1"
      />

      <TranscriptStrip
        messages={messages}
        activeMessageId={activeMessageId}
        selectedId={selectedMessageId}
        onSelect={onSelectMessage}
        className="h-36 shrink-0"
      />
    </div>
  )
}
