import { useEffect } from 'react'
import { X } from 'lucide-react'

import { MarkdownDocument } from '@/components/markdown/markdown-document'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { Button } from '@/components/ui/button'
import { messageLabel, speakerId } from '@/lib/chat-display'
import { formatChatTime, formatDateTimeYMDHMS } from '@/lib/format-date'
import { heSubsectionTitleNeutral } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface TranscriptDrawerProps {
  message: ChatMessage | null
  sequence?: number | null
  onClose: () => void
}

function drawerAvatar(message: ChatMessage): { id: string; name: string } {
  const label = messageLabel(message)
  return { id: speakerId(message), name: label }
}

export function TranscriptDrawer({ message, sequence, onClose }: TranscriptDrawerProps) {
  const open = message != null

  useEffect(() => {
    if (!open) return
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open, onClose])

  if (!message) return null

  const avatar = drawerAvatar(message)
  const label = messageLabel(message)
  const at = formatDateTimeYMDHMS(new Date(message.createdAt).toISOString())

  return (
    <div className="fixed inset-0 z-50 flex justify-end" role="presentation">
      <button
        type="button"
        className="absolute inset-0 bg-black/25 backdrop-blur-[1px]"
        aria-label="关闭详情"
        onClick={onClose}
      />

      <aside
        className={cn(
          'relative flex h-full w-full max-w-lg flex-col border-l border-black/[0.06] bg-surface shadow-[var(--shadow-overlay)]',
        )}
        role="dialog"
        aria-modal="true"
        aria-labelledby="transcript-drawer-title"
      >
        <div className="flex shrink-0 items-start justify-between gap-3 border-b border-black/[0.06] px-5 py-4">
          <div className="flex min-w-0 items-start gap-3">
            <ProfileAvatar id={avatar.id} name={avatar.name} size="sm" />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                {(sequence ?? message.turn) != null && (
                  <span className="rounded-md bg-brand-soft px-2 py-0.5 font-mono text-[12px] font-semibold tabular-nums text-brand">
                    #{(sequence ?? message.turn)!}
                  </span>
                )}
                <h2 id="transcript-drawer-title" className={heSubsectionTitleNeutral}>
                  {label}
                </h2>
              </div>
              <p className="mt-1 text-[12px] text-text-tertiary tabular-nums">
                {at || formatChatTime(message.createdAt)}
              </p>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" className="shrink-0 px-2" onClick={onClose}>
            <X className="size-4" aria-hidden />
            <span className="sr-only">关闭</span>
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
          <MarkdownDocument
            content={message.content}
            constrained={false}
            className="prose-sm max-w-none text-text-primary"
          />
        </div>
      </aside>
    </div>
  )
}
