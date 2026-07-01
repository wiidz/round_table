import type { ReactNode } from 'react'
import {
  Bot,
  Layers,
  MessageCircle,
  MessageCircleOff,
  Users,
} from 'lucide-react'

import {
  MeetingModeInline,
  meetingModeFreeDialogueClass,
} from '@/components/meeting/meeting-mode-badge'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { useI18n } from '@/hooks/use-i18n'
import { hePageDesc, hePageTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

interface MeetingDetailHeaderProps {
  detail: MeetingDetail
  canReplay?: boolean
}

function MetaDot() {
  return <span className="hidden text-text-tertiary/35 sm:inline" aria-hidden>·</span>
}

function formatTokenCount(value: number, intlTag: string): string {
  if (value >= 1_000_000) {
    const compact = value / 1_000_000
    return `${compact >= 10 ? Math.round(compact) : compact.toFixed(1)}M`
  }
  if (value >= 10_000) {
    const compact = value / 1_000
    return `${compact >= 100 ? Math.round(compact) : compact.toFixed(1)}k`
  }
  return value.toLocaleString(intlTag)
}

function MetaItem({
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
        'inline-flex items-center gap-1 tabular-nums text-text-tertiary',
        className,
      )}
    >
      <Icon className="size-3 shrink-0 opacity-55" aria-hidden />
      {children}
    </span>
  )
}

export function MeetingDetailHeader({ detail, canReplay }: MeetingDetailHeaderProps) {
  const { t, intlTag } = useI18n()
  const topic = detail.topic?.trim() || t('meeting.topicEmpty')
  const startedAt = detail.started_at?.trim()
  const participants = detail.participant_count ?? 0
  const rounds = detail.max_rounds ?? 0
  const totalTokens = detail.total_tokens ?? 0

  return (
    <header className="space-y-4">
      <p className="text-[10px] font-medium uppercase tracking-[0.18em] text-text-tertiary">
        {t('meeting.reviewEyebrow')}
      </p>

      <h1 className={cn(hePageTitle, 'text-balance')}>{topic}</h1>

      <p className={hePageDesc}>
        {canReplay
          ? t('meetingUi.header.descriptionWithReplay')
          : t('meetingUi.header.descriptionNoReplay')}
      </p>

      <div className="flex flex-wrap items-center gap-x-2.5 gap-y-2 text-[13px]">
        <MeetingModeInline mode={detail.mode} modeKind={detail.mode_kind} />
        {detail.status && (
          <>
            <MetaDot />
            <MeetingStatusBadge status={detail.status} />
          </>
        )}
        <MetaDot />
        <MetaItem icon={Users}>
          {participants > 0
            ? t('meeting.meta.participants', { n: participants })
            : t('meeting.meta.participantsEmpty')}
        </MetaItem>
        <MetaDot />
        <MetaItem icon={Layers}>
          {rounds > 0 ? t('meeting.meta.rounds', { n: rounds }) : t('meeting.meta.roundsEmpty')}
        </MetaItem>
        <MetaDot />
        <MetaItem
          icon={detail.free_dialogue ? MessageCircle : MessageCircleOff}
          className={cn(
            detail.free_dialogue &&
              meetingModeFreeDialogueClass(detail.mode_kind, detail.mode),
          )}
        >
          {detail.free_dialogue
            ? t('meeting.meta.freeDialogueOn')
            : t('meeting.meta.freeDialogueOff')}
        </MetaItem>
        <MetaDot />
        <span className="inline-flex items-center gap-1 tabular-nums text-text-tertiary">
          <Bot className="size-3 shrink-0 text-ai/55" aria-hidden />
          <span className="font-mono text-[12px]">
            <span className="text-ai/75">
              {totalTokens > 0 ? formatTokenCount(totalTokens, intlTag) : '—'}
            </span>
            <span className="text-text-tertiary/45"> tok</span>
          </span>
        </span>
        <MetaDot />
        <span className="font-mono text-[12px] text-text-tertiary">{detail.id}</span>
        {startedAt && (
          <>
            <MetaDot />
            <span className="tabular-nums text-text-tertiary">{startedAt}</span>
          </>
        )}
      </div>
    </header>
  )
}
