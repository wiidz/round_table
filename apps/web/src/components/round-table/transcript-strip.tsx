import { useEffect, useRef, useState } from 'react'

import { condenseMessage } from '@/lib/condense-message'
import { messageLabel } from '@/lib/chat-display'
import { formatChatTime } from '@/lib/format-date'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface TranscriptStripProps {
  messages: ChatMessage[]
  /** Latest turn message id (current speaker). */
  activeMessageId?: string | null
  selectedId?: string | null
  onSelect: (message: ChatMessage) => void
  className?: string
}

export function TranscriptStrip({
  messages,
  activeMessageId,
  selectedId,
  onSelect,
  className,
}: TranscriptStripProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [pinnedBottom, setPinnedBottom] = useState(true)
  const [newBelow, setNewBelow] = useState(0)

  useEffect(() => {
    const el = scrollRef.current
    if (!el || !pinnedBottom) return
    el.scrollTop = el.scrollHeight
    setNewBelow(0)
  }, [messages, pinnedBottom])

  const handleScroll = () => {
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24
    setPinnedBottom(atBottom)
    if (atBottom) setNewBelow(0)
  }

  useEffect(() => {
    if (pinnedBottom) return
    setNewBelow((n) => n + 1)
  }, [messages.length, pinnedBottom])

  const scrollToBottom = () => {
    const el = scrollRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
    setPinnedBottom(true)
    setNewBelow(0)
  }

  return (
    <div
      className={cn(
        'relative flex shrink-0 flex-col border-t border-black/[0.06] bg-black/[0.015]',
        className ?? 'h-36',
      )}
    >
      <div className="flex shrink-0 items-center justify-between px-5 py-2">
        <p className="text-[11px] font-medium uppercase tracking-[0.12em] text-text-tertiary">
          发言记录 · {messages.length} 条
        </p>
      </div>

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 space-y-1 overflow-y-auto overscroll-contain px-3 pb-3"
      >
        {messages.length === 0 && (
          <p className="px-2 py-3 text-center text-[12px] text-text-tertiary">暂无消息</p>
        )}
        {messages.map((message) => {
          const label = messageLabel(message)
          const { summary, truncated } = condenseMessage(message.content)
          const isCurrentTurn = activeMessageId != null && message.id === activeMessageId

          return (
            <button
              key={message.id}
              type="button"
              onClick={() => onSelect(message)}
              className={cn(
                'flex w-full items-start gap-2 rounded-lg px-2 py-2 text-left transition-colors',
                'hover:bg-black/[0.04]',
                selectedId === message.id && 'bg-brand-soft/80 ring-1 ring-brand/20',
                isCurrentTurn && 'ring-1 ring-ai/25 bg-ai-soft/50',
              )}
            >
              <span className="shrink-0 pt-0.5 font-mono text-[11px] tabular-nums text-text-tertiary">
                {message.turn != null ? `#${message.turn}` : '·'}
              </span>
              <span className="min-w-0 flex-1">
                <span className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                  <span className="text-[12px] font-medium text-text-secondary">{label}</span>
                  <span className="text-[11px] tabular-nums text-text-tertiary">
                    {formatChatTime(message.createdAt)}
                  </span>
                </span>
                <span
                  className={cn(
                    'mt-0.5 block text-[13px] leading-snug text-text-primary',
                    truncated ? 'line-clamp-1' : 'line-clamp-2',
                  )}
                >
                  {summary || '（空）'}
                </span>
              </span>
              {truncated && (
                <span className="shrink-0 pt-0.5 text-[11px] text-brand">详情</span>
              )}
            </button>
          )
        })}
      </div>

      {!pinnedBottom && newBelow > 0 && (
        <button
          type="button"
          onClick={scrollToBottom}
          className="absolute bottom-3 left-1/2 z-10 -translate-x-1/2 rounded-full bg-surface px-3 py-1 text-[12px] text-brand shadow-md ring-1 ring-black/[0.08]"
        >
          ↓ 新消息
        </button>
      )}
    </div>
  )
}
