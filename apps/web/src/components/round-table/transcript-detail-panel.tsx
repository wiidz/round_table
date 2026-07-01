import { X } from 'lucide-react'

import { MarkdownDocument } from '@/components/markdown/markdown-document'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import { speakerId } from '@/lib/chat-display'
import { formatChatTime, formatDateTimeYMDHMS } from '@/lib/format-date'
import { TranscriptEmptyState, TranscriptPanelHeader } from '@/components/round-table/transcript-empty-state'
import { hePanelShell } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface TranscriptDetailPanelProps {
  message: ChatMessage | null
  sequence?: number | null
  onClear?: () => void
  className?: string
}

export function TranscriptDetailPanel({
  message,
  sequence,
  onClear,
  className,
}: TranscriptDetailPanelProps) {
  const { locale, t, messageLabel } = useI18n()

  return (
    <aside
      className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}
      aria-label={t('transcript.detail.title')}
    >
      {message ? (
        <>
          <div className="flex shrink-0 items-start justify-between gap-3 border-b border-black/[0.06] px-4 py-4 sm:px-5">
            <div className="flex min-w-0 items-start gap-3">
              <ProfileAvatar
                id={speakerId(message)}
                name={messageLabel(message)}
                size="sm"
              />
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  {(sequence ?? message.turn) != null && (
                    <span className="rounded-md bg-brand-soft px-2 py-0.5 font-mono text-[12px] font-semibold tabular-nums text-brand">
                      #{(sequence ?? message.turn)!}
                    </span>
                  )}
                  <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">
                    {messageLabel(message)}
                  </h2>
                </div>
                <p className="mt-1 text-[12px] text-text-tertiary tabular-nums">
                  {formatDateTimeYMDHMS(new Date(message.createdAt).toISOString()) ||
                    formatChatTime(message.createdAt, locale)}
                </p>
              </div>
            </div>
            {onClear && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="shrink-0 px-2"
                onClick={onClear}
              >
                <X className="size-4" aria-hidden />
                <span className="sr-only">{t('transcript.detail.clearSelection')}</span>
              </Button>
            )}
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5 sm:py-5">
            <MarkdownDocument
              content={message.content}
              constrained={false}
              className="prose-sm max-w-none text-text-primary"
            />
          </div>
        </>
      ) : (
        <>
          <TranscriptPanelHeader title={t('transcript.detail.title')} />
          <TranscriptEmptyState
            variant="detail"
            title={t('transcript.detail.emptyTitle')}
            description={t('transcript.detail.emptyDescription')}
            className="flex-1"
          />
        </>
      )}
    </aside>
  )
}
