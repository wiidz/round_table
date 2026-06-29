import { useEffect, useRef, useState } from 'react'
import { Loader2, SendHorizonal, Wifi, WifiOff } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { MarkdownDocument } from '@/components/markdown/markdown-document'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { formatChatTime } from '@/lib/format-date'
import {
  heFormEmbed,
  hePanelShell,
  hePressable,
  heSpring,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { ChatConnectionState, ChatMessage } from '@/types/chat'

function ConnectionBadge({ state }: { state: ChatConnectionState }) {
  const label =
    state === 'open'
      ? '已连接'
      : state === 'connecting'
        ? '连接中'
        : state === 'error'
          ? '连接异常'
          : '已断开'

  const tone =
    state === 'open' ? 'success' : state === 'connecting' ? 'warning' : 'neutral'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset',
        tone === 'success' && 'bg-success-soft text-success ring-success/20',
        tone === 'warning' && 'bg-warning-soft text-warning ring-warning/25',
        tone === 'neutral' && 'bg-black/[0.04] text-text-tertiary ring-black/[0.06]',
      )}
    >
      {state === 'open' ? (
        <Wifi className="size-3" aria-hidden />
      ) : state === 'connecting' ? (
        <Loader2 className="size-3 animate-spin" aria-hidden />
      ) : (
        <WifiOff className="size-3" aria-hidden />
      )}
      {label}
    </span>
  )
}

function messageLabel(message: ChatMessage): string | null {
  if (message.role === 'user') return '我'
  if (message.role === 'system') return '系统'
  if (message.role === 'participant') {
    return message.authorName?.trim() || message.authorId || '专家'
  }
  return message.authorName?.trim() || '司仪'
}

function messageAvatar(message: ChatMessage): { id: string; name: string } {
  if (message.role === 'user') {
    return { id: 'user', name: '我' }
  }
  if (message.role === 'system') {
    return { id: 'system', name: '系统' }
  }
  if (message.role === 'participant') {
    return {
      id: message.authorId?.trim() || 'participant',
      name: message.authorName?.trim() || message.authorId || '专家',
    }
  }
  return { id: 'moderator', name: message.authorName?.trim() || '司仪' }
}

function bubbleShellClass(message: ChatMessage, isUser: boolean): string {
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

function ChatBubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === 'user'
  const label = messageLabel(message)
  const avatar = messageAvatar(message)
  const timeLabel = formatChatTime(message.createdAt)

  return (
    <div className={cn('flex w-full', isUser ? 'justify-end' : 'justify-start')}>
      <div className={cn('flex max-w-[min(100%,42rem)] flex-col gap-1', isUser && 'items-end')}>
        {label && (
          <p className="px-1 text-[11px] font-medium text-text-tertiary">{label}</p>
        )}

        <div className={cn('flex items-start gap-2.5', isUser && 'flex-row-reverse')}>
          <ProfileAvatar id={avatar.id} name={avatar.name} size="sm" className="shrink-0" />

          <div className={bubbleShellClass(message, isUser)}>
            {isUser ? (
              <p className="whitespace-pre-wrap">{message.content}</p>
            ) : (
              <MarkdownDocument content={message.content} constrained={false} className="prose-sm max-w-none" />
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

interface ChatWindowProps {
  className?: string
  connectionState: ChatConnectionState
  messages: ChatMessage[]
  sessionId: string | null
  lastError: string | null
  onSend: (content: string) => boolean
  onReconnect: () => void
}

export function ChatWindow({
  className,
  connectionState,
  messages,
  sessionId,
  lastError,
  onSend,
  onReconnect,
}: ChatWindowProps) {
  const [draft, setDraft] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)
  const canSend = connectionState === 'open'

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [messages])

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!canSend) return
    if (onSend(draft)) {
      setDraft('')
    }
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (!canSend) return
      if (onSend(draft)) {
        setDraft('')
      }
    }
  }

  return (
    <div className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}>
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-black/[0.05] px-5 py-4">
        <div>
          <h2 className={heSubsectionTitleNeutral}>与司仪对话</h2>
          <p className="mt-1 text-[12px] text-text-tertiary">
            浏览器 Transport · 无需 Principal · 可发起会议
          </p>
        </div>
        <div className="flex items-center gap-2">
          <ConnectionBadge state={connectionState} />
          {connectionState !== 'open' && (
            <Button type="button" variant="outline" size="sm" onClick={onReconnect}>
              重连
            </Button>
          )}
        </div>
      </div>

      {sessionId && connectionState === 'open' && (
        <p className="shrink-0 border-b border-black/[0.04] px-5 py-2 font-mono text-[11px] text-text-tertiary">
          会话 {sessionId.slice(0, 8)}…
        </p>
      )}

      <div ref={scrollRef} className="min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain px-5 py-5">
        {messages.length === 0 && (
          <p className="text-center text-[13px] text-text-tertiary">
            发送「会议状态」或「开个会」，或直接提问。
          </p>
        )}
        {messages.map((message) => (
          <ChatBubble key={message.id} message={message} />
        ))}
      </div>

      {lastError && connectionState === 'error' && (
        <p className="shrink-0 px-5 pb-2 text-[12px] text-danger">{lastError}</p>
      )}

      <form
        onSubmit={handleSubmit}
        className="shrink-0 border-t border-black/[0.05] bg-black/[0.015] px-5 py-4"
      >
        <div className={cn(heFormEmbed, 'flex items-end gap-3 p-3')}>
          <textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={2}
            placeholder={canSend ? '输入消息，Enter 发送，Shift+Enter 换行' : '连接中…'}
            disabled={!canSend}
            className={cn(
              'min-h-[3rem] flex-1 resize-none border-0 bg-transparent px-1 py-1 text-[14px]',
              'text-text-primary placeholder:text-text-tertiary focus:outline-none',
              'disabled:cursor-not-allowed disabled:opacity-60',
            )}
          />
          <Button
            type="submit"
            disabled={!canSend || !draft.trim()}
            className={cn(hePressable, heSpring, 'shrink-0 gap-1.5 rounded-xs px-4')}
          >
            <SendHorizonal className="size-4" aria-hidden />
            发送
          </Button>
        </div>
      </form>
    </div>
  )
}
