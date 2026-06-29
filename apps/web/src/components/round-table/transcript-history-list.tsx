import { useEffect, useMemo, useRef, useState } from 'react'

import { bubbleShellClass } from '@/components/chat/chat-bubble'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { condenseMessage } from '@/lib/condense-message'
import { assignsTurn, messageAvatar, messageLabel } from '@/lib/chat-display'
import { formatChatTime } from '@/lib/format-date'
import {
  buildMessageSequenceMap,
  messageSequenceNumber,
} from '@/lib/message-sequence'
import {
  filterTranscriptBySpeaker,
  listTranscriptSpeakers,
} from '@/lib/transcript-speakers'
import { cn } from '@/lib/utils'
import { heScrollbar } from '@/lib/highend-styles'
import type { ChatMessage } from '@/types/chat'

interface TranscriptHistoryListProps {
  messages: ChatMessage[]
  activeMessageId?: string | null
  selectedId?: string | null
  onSelect: (message: ChatMessage) => void
  className?: string
}

function SpeakerFilterChips({
  speakers,
  filterSpeakerId,
  onChange,
}: {
  speakers: ReturnType<typeof listTranscriptSpeakers>
  filterSpeakerId: string | null
  onChange: (id: string | null) => void
}) {
  if (speakers.length <= 1) return null

  return (
    <div className="mt-3 flex flex-wrap gap-1.5">
      <button
        type="button"
        onClick={() => onChange(null)}
        className={cn(
          'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
          filterSpeakerId == null
            ? 'bg-brand-soft text-brand ring-1 ring-brand/25'
            : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
        )}
      >
        全部
      </button>
      {speakers.map((speaker) => (
        <button
          key={speaker.id}
          type="button"
          onClick={() => onChange(filterSpeakerId === speaker.id ? null : speaker.id)}
          className={cn(
            'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
            filterSpeakerId === speaker.id
              ? 'bg-brand-soft text-brand ring-1 ring-brand/25'
              : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
          )}
        >
          {speaker.label}
        </button>
      ))}
    </div>
  )
}

function HistoryBubbleItem({
  message,
  sequence,
  selected,
  active,
  onSelect,
}: {
  message: ChatMessage
  sequence: number | null
  selected: boolean
  active: boolean
  onSelect: () => void
}) {
  const isUser = message.role === 'user'
  const label = messageLabel(message)
  const avatar = messageAvatar(message)
  const { summary } = condenseMessage(message.content, 120)
  const timeLabel = formatChatTime(message.createdAt)
  const showSequence = assignsTurn(message.role) && sequence != null

  return (
    <button
      type="button"
      onClick={onSelect}
      className={cn(
        'flex w-full flex-col gap-1 px-1 py-1 text-left transition-opacity',
        isUser ? 'items-end' : 'items-start',
        selected && 'opacity-100',
        !selected && 'opacity-90 hover:opacity-100',
      )}
    >
      <p
        className={cn(
          'px-0.5 text-[10px] font-medium',
          isUser && 'text-right',
          selected && 'font-semibold text-brand',
          active && !selected && 'text-ai',
          !selected && !active && 'text-text-tertiary',
        )}
      >
        {showSequence ? `#${sequence} · ${label}` : label}
        {timeLabel && (
          <span className="ml-1.5 font-normal tabular-nums opacity-80">{timeLabel}</span>
        )}
      </p>

      <div className={cn('flex max-w-[92%] items-start gap-2', isUser && 'flex-row-reverse')}>
        <ProfileAvatar id={avatar.id} name={avatar.name} size="sm" className="shrink-0" />
        <div
          className={cn(
            bubbleShellClass(message, isUser),
            'min-w-0 px-3 py-2 text-[12px] leading-snug',
          )}
        >
          <p className={cn('line-clamp-4 whitespace-pre-wrap', isUser ? 'text-white' : 'text-text-primary')}>
            {summary || '（空）'}
          </p>
        </div>
      </div>
    </button>
  )
}

/** Left gutter transcript list: header filters on top, IM-style bubbles below. */
export function TranscriptHistoryList({
  messages,
  activeMessageId,
  selectedId,
  onSelect,
  className,
}: TranscriptHistoryListProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [pinnedBottom, setPinnedBottom] = useState(true)
  const [newBelow, setNewBelow] = useState(0)
  const [filterSpeakerId, setFilterSpeakerId] = useState<string | null>(null)

  const speakers = useMemo(() => listTranscriptSpeakers(messages), [messages])
  const sequenceMap = useMemo(() => buildMessageSequenceMap(messages), [messages])
  const visibleMessages = useMemo(
    () => filterTranscriptBySpeaker(messages, filterSpeakerId),
    [messages, filterSpeakerId],
  )

  useEffect(() => {
    if (filterSpeakerId && !speakers.some((s) => s.id === filterSpeakerId)) {
      setFilterSpeakerId(null)
    }
  }, [speakers, filterSpeakerId])

  useEffect(() => {
    const el = scrollRef.current
    if (!el || !pinnedBottom) return
    el.scrollTop = el.scrollHeight
    setNewBelow(0)
  }, [visibleMessages, pinnedBottom])

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
  }, [visibleMessages.length, pinnedBottom])

  const scrollToBottom = () => {
    const el = scrollRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
    setPinnedBottom(true)
    setNewBelow(0)
  }

  const visibleCount = filterSpeakerId ? visibleMessages.length : messages.length
  const countLabel = filterSpeakerId
    ? `共 ${visibleCount}/${messages.length} 条`
    : `共 ${messages.length} 条`

  return (
    <div className={cn('relative flex min-h-0 flex-1 flex-col overflow-hidden', className)}>
      <div className="shrink-0 border-b border-black/[0.06] px-4 py-4 sm:px-5">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">发言记录</h2>
          <span className="shrink-0 text-[12px] tabular-nums text-text-tertiary">{countLabel}</span>
        </div>
        <SpeakerFilterChips
          speakers={speakers}
          filterSpeakerId={filterSpeakerId}
          onChange={setFilterSpeakerId}
        />
      </div>

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className={cn(
          'min-h-0 flex-1 space-y-2 overflow-y-auto px-3 py-3 sm:px-4',
          heScrollbar,
        )}
      >
        {visibleMessages.length === 0 && (
          <p className="py-6 text-center text-[12px] text-text-tertiary">
            {filterSpeakerId ? '该发言人暂无消息' : '暂无消息'}
          </p>
        )}
        {visibleMessages.map((message) => (
          <HistoryBubbleItem
            key={message.id}
            message={message}
            sequence={messageSequenceNumber(message, sequenceMap)}
            selected={selectedId === message.id}
            active={activeMessageId === message.id}
            onSelect={() => onSelect(message)}
          />
        ))}
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
