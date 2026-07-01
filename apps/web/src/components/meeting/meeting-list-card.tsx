import type { ReactNode } from 'react'
import {
  Bot,
  CalendarDays,
  Layers,
  MessageCircle,
  MessageCircleOff,
  Users,
} from 'lucide-react'
import { Link } from 'react-router-dom'

import {
  MeetingModeMark,
  meetingModeFreeDialogueClass,
  meetingModeHoverClass,
} from '@/components/meeting/meeting-mode-badge'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { useI18n } from '@/hooks/use-i18n'
import { formatDateTimeYMDHMS, formatDateYMD } from '@/lib/format-date'
import {
  heFileBadge,
  hePanelShell,
  hePanelShellHover,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { formatTokenCount } from '@/lib/meeting-overview-stats'
import { cn } from '@/lib/utils'

import type { MeetingIndex } from '@/types/meeting'

interface MeetingGridCardProps {
  meeting: MeetingIndex
}

function formatMeetingTime(meeting: MeetingIndex): string {
  const raw = meeting.started_at?.trim()
  if (raw) {
    const withoutTz = raw.replace(/\s*\([A-Z]+\)\s*$/, '').trim()
    if (/^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}/.test(withoutTz)) {
      return withoutTz.slice(0, 16)
    }
    return withoutTz
  }
  const fromUpdated = formatDateTimeYMDHMS(meeting.updated_at)
  if (fromUpdated) return fromUpdated.slice(0, 16)
  return formatDateYMD(meeting.updated_at) || '—'
}

function formatMeetingId(id: string): string {
  if (id.length <= 22) return id
  return `${id.slice(0, 10)}…${id.slice(-8)}`
}

function MetaChip({
  icon: Icon,
  children,
  className,
}: {
  icon: typeof Users
  children: ReactNode
  className?: string
}) {
  return (
    <span
      className={cn(
        heFileBadge,
        'inline-flex items-center gap-1 tabular-nums',
        className,
      )}
    >
      <Icon className="size-3 shrink-0 opacity-55" />
      {children}
    </span>
  )
}

export function MeetingGridCard({ meeting }: MeetingGridCardProps) {
  const { t, intlTag } = useI18n()
  const topic = meeting.topic?.trim() || t('meeting.topicEmpty')
  const timeLabel = formatMeetingTime(meeting)
  const participants = meeting.participant_count ?? 0
  const rounds = meeting.max_rounds ?? 0
  const llmCalls = meeting.llm_call_count ?? 0
  const totalTokens = meeting.total_tokens ?? 0
  const hasUsage = llmCalls > 0 || totalTokens > 0

  return (
    <Link
      to={`/meetings/${encodeURIComponent(meeting.id)}`}
      className={cn('group block h-full', hePressable)}
    >
      <article
        className={cn(
          hePanelShell,
          hePanelShellHover,
          heSpring,
          'flex h-full min-h-[172px] flex-col gap-3 p-4',
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <MeetingModeMark mode={meeting.mode} modeKind={meeting.mode_kind} />
          {meeting.status && (
            <MeetingStatusBadge status={meeting.status} className="shrink-0" />
          )}
        </div>

        <h2
          className={cn(
            'line-clamp-2 min-h-[2.75rem] text-[15px] font-semibold leading-snug tracking-[-0.02em] text-text-primary',
            meetingModeHoverClass(meeting.mode_kind, meeting.mode),
            heSpring,
          )}
          title={topic}
        >
          {topic}
        </h2>

        <div className="space-y-2">
          <div className="flex flex-wrap gap-1.5">
            <MetaChip icon={Users}>
              {participants > 0
                ? t('meeting.meta.participants', { n: participants })
                : t('meeting.meta.participantsEmpty')}
            </MetaChip>
            <MetaChip icon={Layers}>
              {rounds > 0 ? t('meeting.meta.rounds', { n: rounds }) : t('meeting.meta.roundsEmpty')}
            </MetaChip>
            <MetaChip
              icon={meeting.free_dialogue ? MessageCircle : MessageCircleOff}
              className={cn(
                meeting.free_dialogue &&
                  meetingModeFreeDialogueClass(meeting.mode_kind, meeting.mode),
              )}
            >
              {meeting.free_dialogue
                ? t('meeting.meta.freeDialogueOn')
                : t('meeting.meta.freeDialogueOff')}
            </MetaChip>
          </div>
          {hasUsage && (
            <div className="flex items-center gap-2 border-t border-dashed border-black/[0.08] pt-2 mt-4 dark:border-white/[0.1]">
              <Bot className="size-3 shrink-0 text-ai/55" aria-hidden />
              <p className="min-w-0 truncate font-mono text-[10px] leading-none tabular-nums">
                <span className="text-text-secondary">{llmCalls > 0 ? llmCalls : '—'}</span>
                <span className="text-text-tertiary/45"> calls</span>
                <span className="mx-1.5 text-text-tertiary/25">/</span>
                <span className="text-ai/75">
                  {totalTokens > 0 ? formatTokenCount(totalTokens, intlTag) : '—'}
                </span>
                <span className="text-text-tertiary/45"> tok</span>
              </p>
            </div>
          )}
        </div>

        <div className="mt-auto flex items-end justify-between gap-2 border-t border-black/[0.05] pt-3 dark:border-white/[0.06]">
          <p
            className="min-w-0 truncate font-mono text-[10px] text-text-tertiary"
            title={meeting.id}
          >
            {formatMeetingId(meeting.id)}
          </p>
          <p className="flex shrink-0 items-center gap-1 text-[11px] tabular-nums text-text-tertiary">
            <CalendarDays className="size-3 opacity-60" />
            {timeLabel}
          </p>
        </div>
      </article>
    </Link>
  )
}

interface MeetingGridSkeletonProps {
  count?: number
}

export function MeetingGridSkeleton({ count = 12 }: MeetingGridSkeletonProps) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: count }, (_, i) => (
        <div
          key={i}
          className={cn(hePanelShell, 'h-[172px] animate-pulse bg-black/[0.02]')}
        />
      ))}
    </div>
  )
}
