import { TranscriptEmptyState } from '@/components/round-table/transcript-empty-state'
import { TranscriptHistoryList } from '@/components/round-table/transcript-history-list'
import { useI18n } from '@/hooks/use-i18n'
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
  const { t } = useI18n()
  const isEmpty = messages.length === 0

  return (
    <aside
      className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}
      aria-label={t('transcript.history.title')}
    >
      {isEmpty ? (
        <>
          <div className="shrink-0 border-b border-black/[0.06] px-4 py-4 sm:px-5">
            <div className="flex items-center justify-between gap-3">
              <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">
                {t('transcript.history.title')}
              </h2>
            </div>
          </div>
          <TranscriptEmptyState
            variant="list"
            title={t('transcript.history.emptyTitle')}
            description={t('transcript.history.emptyDescription')}
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
