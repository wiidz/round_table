import type { ReactNode } from 'react'
import { CalendarDays, ChevronRight, Layers, MessageCircle, MessageCircleOff, Users } from 'lucide-react'
import { Link } from 'react-router-dom'

import {
  MeetingModeMark,
  meetingModeFreeDialogueClass,
  meetingModeHoverClass,
} from '@/components/meeting/meeting-mode-badge'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { formatDateTimeYMDHMS, formatDateYMD } from '@/lib/format-date'
import {
  hePanelShell,
  hePanelShellHover,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
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

function StatItem({
  icon: Icon,
  children,
  className,
}: {
  icon: typeof Users
  children: ReactNode
  className?: string
}) {
  return (
    <span className={cn('inline-flex items-center gap-1 tabular-nums', className)}>
      <Icon className="size-3 shrink-0 opacity-70" />
      {children}
    </span>
  )
}

export function MeetingGridCard({ meeting }: MeetingGridCardProps) {
  const topic = meeting.topic?.trim() || '（无主题）'
  const timeLabel = formatMeetingTime(meeting)
  const participants = meeting.participant_count ?? 0
  const rounds = meeting.max_rounds ?? 0

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
          'flex h-full min-h-[168px] flex-col gap-3 p-4',
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <MeetingModeMark mode={meeting.mode} modeKind={meeting.mode_kind} />
          <div className="flex shrink-0 items-center gap-2 pt-1">
            {meeting.status && <MeetingStatusBadge status={meeting.status} />}
            <span
              className={cn(
                'inline-flex size-8 items-center justify-center rounded-full',
                'bg-black/[0.02] text-text-tertiary ring-1 ring-inset ring-black/[0.05]',
                'group-hover:bg-black/[0.04] group-hover:text-text-secondary',
                heSpring,
              )}
              aria-hidden
            >
              <ChevronRight className="size-4 shrink-0" strokeWidth={2} />
            </span>
          </div>
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

        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-text-secondary">
          <StatItem icon={Users}>
            {participants > 0 ? `${participants} 人` : '— 人'}
          </StatItem>
          <span className="text-text-tertiary/50">·</span>
          <StatItem icon={Layers}>
            {rounds > 0 ? `${rounds} 轮` : '— 轮'}
          </StatItem>
          <span className="text-text-tertiary/50">·</span>
          <StatItem
            icon={meeting.free_dialogue ? MessageCircle : MessageCircleOff}
            className={cn(
              meeting.free_dialogue
                ? meetingModeFreeDialogueClass(meeting.mode_kind, meeting.mode)
                : 'text-text-tertiary',
            )}
          >
            {meeting.free_dialogue ? '自由对话' : '无自由'}
          </StatItem>
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
          className={cn(hePanelShell, 'h-[168px] animate-pulse bg-black/[0.02]')}
        />
      ))}
    </div>
  )
}
