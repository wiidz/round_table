import { X } from 'lucide-react'

import { MarkdownDocument } from '@/components/markdown/markdown-document'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { Button } from '@/components/ui/button'
import { messageLabel, speakerId } from '@/lib/chat-display'
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

function detailAvatar(message: ChatMessage): { id: string; name: string } {
  const label = messageLabel(message)
  return { id: speakerId(message), name: label }
}

export function TranscriptDetailPanel({
  message,
  sequence,
  onClear,
  className,
}: TranscriptDetailPanelProps) {
  return (
    <aside
      className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}
      aria-label="发言详情"
    >
      {message ? (
        <>
          <div className="flex shrink-0 items-start justify-between gap-3 border-b border-black/[0.06] px-4 py-4 sm:px-5">
            <div className="flex min-w-0 items-start gap-3">
              <ProfileAvatar id={detailAvatar(message).id} name={detailAvatar(message).name} size="sm" />
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
                    formatChatTime(message.createdAt)}
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
                <span className="sr-only">清除选择</span>
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
          <TranscriptPanelHeader title="发言详情" />
          <TranscriptEmptyState
            variant="detail"
            title="选择一条发言"
            description="点击左侧记录或圆桌气泡，在此查看完整 Markdown 内容。"
            className="flex-1"
          />
        </>
      )}
    </aside>
  )
}
