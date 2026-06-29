import { MarkdownDocument } from '@/components/markdown/markdown-document'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { assignsTurn, messageAvatar, messageLabel } from '@/lib/chat-display'
import { formatChatTime } from '@/lib/format-date'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

export function bubbleShellClass(message: ChatMessage, isUser: boolean): string {
  if (isUser) {
    return 'chat-bubble chat-bubble--user chat-bubble--tail-right'
  }
  if (message.role === 'system') {
    const tone = message.error ? 'chat-bubble--system-error' : 'chat-bubble--system'
    return cn('chat-bubble chat-bubble--tail-left', tone)
  }
  if (message.role === 'participant') {
    return 'chat-bubble chat-bubble--participant chat-bubble--tail-left'
  }
  return 'chat-bubble chat-bubble--moderator chat-bubble--tail-left'
}

export function ChatBubble({
  message,
  sequence,
}: {
  message: ChatMessage
  sequence?: number | null
}) {
  const isUser = message.role === 'user'
  const label = messageLabel(message)
  const avatar = messageAvatar(message)
  const timeLabel = formatChatTime(message.createdAt)
  const sequenceNo = sequence ?? message.turn ?? null
  const showSequence = assignsTurn(message.role) && sequenceNo != null

  return (
    <div className={cn('flex w-full', isUser ? 'justify-end' : 'justify-start')}>
      <div className={cn('flex max-w-[min(100%,42rem)] flex-col gap-1', isUser && 'items-end')}>
        {label && (
          <p className="px-1 text-[11px] font-medium text-text-tertiary">
            {showSequence ? `#${sequenceNo} · ${label}` : label}
          </p>
        )}

        <div className={cn('flex items-start gap-2.5', isUser && 'flex-row-reverse')}>
          <ProfileAvatar id={avatar.id} name={avatar.name} size="sm" className="shrink-0" />

          <div className={bubbleShellClass(message, isUser)}>
            {isUser ? (
              <p className="whitespace-pre-wrap">{message.content}</p>
            ) : (
              <MarkdownDocument
                content={message.content}
                constrained={false}
                className="prose-sm max-w-none"
              />
            )}

            {timeLabel && (
              <p
                className={cn(
                  'mt-1 text-[11px] tabular-nums',
                  isUser ? 'text-right text-white/70' : 'text-text-tertiary',
                )}
              >
                {timeLabel}
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
