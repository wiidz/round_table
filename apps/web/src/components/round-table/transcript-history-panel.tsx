import { TranscriptEmptyState } from '@/components/round-table/transcript-empty-state'
import { TranscriptHistoryList } from '@/components/round-table/transcript-history-list'
import { hePanelShell } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface TranscriptHistoryPanelProps {
  messages: ChatMessage[]
  activeMessageId?: string | null
  selectedId?: string | null
  onSelect: (message: ChatMessage) => void
  className?: string
}

/** Vertical transcript list in the left gutter beside max-w-6xl main content. */
export function TranscriptHistoryPanel({
  messages,
  activeMessageId,
  selectedId,
  onSelect,
  className,
}: TranscriptHistoryPanelProps) {
  const isEmpty = messages.length === 0

  return (
    <aside
      className={cn(hePanelShell, 'flex min-h-0 flex-col overflow-hidden', className)}
      aria-label="发言记录"
    >
      {isEmpty ? (
        <>
          <div className="shrink-0 border-b border-black/[0.06] px-4 py-4 sm:px-5">
            <div className="flex items-center justify-between gap-3">
              <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">
                发言记录
              </h2>
            </div>
          </div>
          <TranscriptEmptyState
            variant="list"
            title="暂无发言"
            description="会议开始后，司仪、专家与你的消息会按顺序出现在这里。"
          />
        </>
      ) : (
        <TranscriptHistoryList
          messages={messages}
          activeMessageId={activeMessageId}
          selectedId={selectedId}
          onSelect={onSelect}
          className="min-h-0 flex-1"
        />
      )}
    </aside>
  )
}
