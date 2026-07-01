import { useEffect, useMemo, useRef, useState } from 'react'

import { bubbleShellClass } from '@/components/chat/chat-bubble'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import {
  sequenceBadgeToneFromHistoryItem,
  TranscriptSequenceBadge,
} from '@/components/round-table/transcript-sequence-badge'
import { useI18n } from '@/hooks/use-i18n'
import { condenseMessage } from '@/lib/condense-message'
import { assignsTurn } from '@/lib/chat-display'
import { formatChatTime } from '@/lib/format-date'
import {
  buildMessageSequenceMap,
  messageSequenceNumber,
} from '@/lib/message-sequence'
import {
  filterTranscriptBySpeaker,
  listTranscriptSpeakers,
} from '@/lib/transcript-speakers'
import { transcriptBubbleBadgePaddingTop, transcriptSpeakerLabelClass } from '@/lib/transcript-speaker-label'
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
  const { t } = useI18n()

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
        {t('transcript.list.all')}
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
  const { locale, t, messageLabel, messageAvatar } = useI18n()
  const isUser = message.role === 'user'
  const label = messageLabel(message)
  const avatar = messageAvatar(message)
  const { summary } = condenseMessage(message.content, 120)
  const timeLabel = formatChatTime(message.createdAt, locale)
  const showSequence = assignsTurn(message.role) && sequence != null

  return (
    <button
      type="button"
      data-message-id={message.id}
      onClick={onSelect}
      className={cn(
        'group flex w-full rounded-lg px-1 py-1 text-left transition-[opacity,background-color] duration-200',
        isUser ? 'justify-end' : 'justify-start',
        selected && 'bg-ai-soft/40 opacity-100',
        !selected && 'opacity-90 hover:bg-black/[0.02] hover:opacity-100',
      )}
    >
      <div className={cn('flex max-w-[92%] items-start gap-2', isUser && 'flex-row-reverse')}>
        <div className="flex shrink-0 flex-col items-center gap-1.5">
          <ProfileAvatar id={avatar.id} name={avatar.name} size="sm" className="shrink-0" />
          {label && (
            <p
              className={transcriptSpeakerLabelClass({
                highlighted: selected,
                focused: active && !selected,
              })}
            >
              {label}
            </p>
          )}
        </div>
        <div className="relative min-w-0 flex-1">
          {showSequence && sequence != null && (
            <TranscriptSequenceBadge
              sequence={sequence}
              tone={sequenceBadgeToneFromHistoryItem(selected, active)}
            />
          )}
          <div
            className={cn(
              bubbleShellClass(message, isUser),
              'chat-bubble--interactive relative min-w-0 !px-3 !pb-5 text-[12px] leading-snug transition-shadow duration-200',
              showSequence ? transcriptBubbleBadgePaddingTop : '!pt-2',
              selected && 'chat-bubble--live-highlight',
              active && !selected && 'ring-1 ring-ai/20',
            )}
          >
            <p className={cn('line-clamp-3 whitespace-pre-wrap', isUser ? 'text-white' : 'text-text-primary')}>
              {summary || t('transcript.list.emptyContent')}
            </p>
            {timeLabel && (
              <span
                className={cn(
                  'pointer-events-none absolute bottom-1 left-3 text-[9px] font-normal tabular-nums',
                  isUser ? 'text-white/45' : 'text-text-tertiary/70',
                )}
              >
                {timeLabel}
              </span>
            )}
          </div>
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
  const { t } = useI18n()
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
    if (!selectedId) return
    const el = scrollRef.current?.querySelector<HTMLElement>(`[data-message-id="${selectedId}"]`)
    el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  }, [selectedId, visibleMessages])

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
    ? t('transcript.list.count', { visible: visibleCount, total: messages.length })
    : t('transcript.list.countAll', { total: messages.length })

  return (
    <div className={cn('relative flex min-h-0 flex-1 flex-col overflow-hidden', className)}>
      <div className="shrink-0 border-b border-black/[0.06] px-4 py-4 sm:px-5">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">
            {t('transcript.list.title')}
          </h2>
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
            {filterSpeakerId ? t('transcript.list.noSpeakerMessages') : t('transcript.list.noMessages')}
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
          {t('transcript.list.newMessages')}
        </button>
      )}
    </div>
  )
}
