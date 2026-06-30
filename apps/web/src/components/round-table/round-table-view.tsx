import { RoundTableStage } from '@/components/round-table/round-table-stage'
import { TranscriptScrubBar } from '@/components/round-table/transcript-scrub-bar'
import { TranscriptStrip } from '@/components/round-table/transcript-strip'
import type { SeatLayout } from '@/lib/round-table-layout'
import type { ChatMessage, TypingStates } from '@/types/chat'

interface RoundTableViewProps {
  seats: SeatLayout[]
  messages: ChatMessage[]
  latestBySeat: Map<string, ChatMessage>
  focusedSeatId: string | null
  typingStates?: TypingStates
  turnCount: number
  maxTurn: number
  scrubTurn: number | null
  onScrubTurnChange: (turn: number | null) => void
  activeMessageId: string | null
  selectedMessageId: string | null
  highlightMessageId: string | null
  referenceTurn: number | null
  rosterLoading: boolean
  rosterFromApi: boolean
  rosterTotal: number
  seatedExpertCount: number
  centerTitle?: string
  centerSubtitle?: string
  onSelectMessage: (message: ChatMessage) => void
  /** Hide bottom strip when history lives in the left gutter rail. */
  showTranscriptStrip?: boolean
}

export function RoundTableView({
  seats,
  messages,
  latestBySeat,
  focusedSeatId,
  typingStates,
  turnCount,
  maxTurn,
  scrubTurn,
  onScrubTurnChange,
  activeMessageId,
  selectedMessageId,
  highlightMessageId,
  referenceTurn,
  rosterLoading,
  rosterFromApi,
  rosterTotal,
  seatedExpertCount,
  centerTitle,
  centerSubtitle,
  onSelectMessage,
  showTranscriptStrip = true,
}: RoundTableViewProps) {
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <RoundTableStage
        seats={seats}
        latestBySeat={latestBySeat}
        highlightMessageId={highlightMessageId}
        referenceTurn={referenceTurn}
        focusedSeatId={focusedSeatId}
        typingStates={typingStates}
        turnCount={turnCount}
        rosterLoading={rosterLoading}
        rosterFromApi={rosterFromApi}
        rosterTotal={rosterTotal}
        seatedExpertCount={seatedExpertCount}
        centerTitle={centerTitle}
        centerSubtitle={centerSubtitle}
        onLiveMessageClick={onSelectMessage}
        className="min-h-0 flex-1"
      />

      <TranscriptScrubBar
        maxTurn={maxTurn}
        scrubTurn={scrubTurn}
        onScrubTurnChange={onScrubTurnChange}
      />

      {showTranscriptStrip && (
        <TranscriptStrip
          messages={messages}
          activeMessageId={activeMessageId}
          selectedId={selectedMessageId}
          onSelect={onSelectMessage}
          className="h-36 shrink-0"
        />
      )}
    </div>
  )
}
