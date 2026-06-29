import { TranscriptScrubBar } from '@/components/round-table/transcript-scrub-bar'
import { TranscriptStrip } from '@/components/round-table/transcript-strip'
import type { ChatMessage } from '@/types/chat'

interface StripOnlyViewProps {
  messages: ChatMessage[]
  maxTurn: number
  scrubTurn: number | null
  onScrubTurnChange: (turn: number | null) => void
  activeMessageId: string | null
  selectedMessageId: string | null
  onSelectMessage: (message: ChatMessage) => void
}

/** Narrow-screen meeting view: expanded transcript strip without round table stage. */
export function StripOnlyView({
  messages,
  maxTurn,
  scrubTurn,
  onScrubTurnChange,
  activeMessageId,
  selectedMessageId,
  onSelectMessage,
}: StripOnlyViewProps) {
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <p className="shrink-0 border-b border-black/[0.04] bg-black/[0.02] px-5 py-2 text-[11px] text-text-tertiary">
        窄屏模式：发言记录列表。宽屏可切换圆桌视图。
      </p>
      <TranscriptScrubBar
        maxTurn={maxTurn}
        scrubTurn={scrubTurn}
        onScrubTurnChange={onScrubTurnChange}
      />
      <TranscriptStrip
        messages={messages}
        activeMessageId={activeMessageId}
        selectedId={selectedMessageId}
        onSelect={onSelectMessage}
        className="min-h-0 flex-1"
      />
    </div>
  )
}
