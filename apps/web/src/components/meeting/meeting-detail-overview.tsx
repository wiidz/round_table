import { Link } from 'react-router-dom'
import { ChevronRight, ExternalLink, FileText, Users } from 'lucide-react'

import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefTemplateScopePreview } from '@/components/brief/brief-template-scope-fields'
import {
  BRIEF_TEMPLATE_SECTIONS,
  briefAgendaItemShell,
  briefFieldLabelClass,
} from '@/components/brief/brief-template-sections'
import {
  MeetingModeInline,
} from '@/components/meeting/meeting-mode-badge'
import { MeetingOverviewStatCards } from '@/components/meeting/meeting-overview-stat-cards'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { hePageTitle, hePanelShell, hePressable, heSpring } from '@/lib/highend-styles'
import { meetingFileLabel, primaryDeliverablePath, type MeetingModeKind } from '@/lib/meeting-labels'
import type { MeetingBriefPreview } from '@/lib/meeting-brief-preview'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

interface MeetingDetailOverviewProps {
  detail: MeetingDetail
  brief: MeetingBriefPreview
  modeKind?: MeetingModeKind
  canReplay: boolean
  onOpenDocuments: () => void
  onOpenConclusion?: () => void
}

export function MeetingDetailOverview({
  detail,
  brief,
  modeKind,
  canReplay,
  onOpenDocuments,
  onOpenConclusion,
}: MeetingDetailOverviewProps) {
  const primaryPath = primaryDeliverablePath(modeKind)
  const primaryTitle = meetingFileLabel(primaryPath, modeKind)
  const topic = brief.topic.trim()
  const goal = brief.goal.trim()
  const conclusion = brief.conclusion?.trim()

  const sessionMeta = detail.started_at?.trim()

  const conclusionEmpty =
    detail.status === '进行中' || detail.status === 'Running'
      ? '会议进行中，结论尚未产出'
      : detail.status === '已中断' || detail.status === 'aborted' || detail.status === 'Aborted'
        ? '会议已中断，未产出结论'
        : '暂无总结结论'

  return (
    <section className={cn(hePanelShell, 'overflow-visible p-5 sm:p-6')}>
      <div className="space-y-8">
        <div className="space-y-3">
          <p className="text-[10px] font-medium uppercase tracking-[0.18em] text-text-tertiary">
            会议复盘
          </p>
          <h1
            className={cn(
              hePageTitle,
              'text-balance',
              !topic && 'font-normal text-text-tertiary',
            )}
          >
            {topic || '（无主题）'}
          </h1>
          <div className="flex flex-wrap items-center gap-2">
            <MeetingModeInline mode={detail.mode} modeKind={detail.mode_kind} />
            {detail.status && <MeetingStatusBadge status={detail.status} />}
          </div>
          {sessionMeta && (
            <p className="text-[13px] text-text-secondary">{sessionMeta}</p>
          )}
        </div>

        <MeetingOverviewStatCards detail={detail} modeKind={modeKind} />

        <div className="space-y-5">
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="size-1.5 shrink-0 rounded-full bg-info" aria-hidden />
              <p className={briefFieldLabelClass}>会议目标</p>
            </div>
            <p
              className={cn(
                'pl-3.5 text-[15px] leading-relaxed',
                goal ? 'font-medium text-text-primary' : 'text-text-tertiary',
              )}
            >
              {goal || '未填写会议目标'}
            </p>
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="size-1.5 shrink-0 rounded-full bg-brand" aria-hidden />
              <p className={briefFieldLabelClass}>总结结论</p>
            </div>
            <p
              className={cn(
                'pl-3.5 text-[15px] leading-relaxed',
                conclusion ? 'font-medium text-text-primary' : 'text-text-tertiary',
              )}
            >
              {conclusion || conclusionEmpty}
            </p>
            {conclusion && brief.conclusionSource && onOpenConclusion && (
              <button
                type="button"
                className={cn(
                  'ml-3.5 inline-flex items-center gap-1 text-[12px] font-medium text-brand',
                  hePressable,
                  heSpring,
                )}
                onClick={onOpenConclusion}
              >
                查看完整{primaryTitle}
                <ChevronRight className="size-3.5" aria-hidden />
              </button>
            )}
          </div>
        </div>

        {brief.agenda.length > 0 && (
          <section className="space-y-4">
            <BriefSectionHeading
              title={BRIEF_TEMPLATE_SECTIONS.agenda.title}
              description={BRIEF_TEMPLATE_SECTIONS.agenda.description}
            />
            <ol className="space-y-2.5">
              {brief.agenda.map((item, index) => (
                <li key={`${index}-${item}`} className={briefAgendaItemShell}>
                  <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-brand text-[12px] font-bold tabular-nums text-white">
                    {index + 1}
                  </span>
                  <p className="min-w-0 pt-0.5 text-[14px] font-medium leading-relaxed text-text-primary">
                    {item}
                  </p>
                </li>
              ))}
            </ol>
          </section>
        )}

        <BriefTemplateScopePreview
          inScope={brief.inScope}
          outOfScope={brief.outOfScope}
          doneCriteria={brief.doneCriteria}
        />

        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            className={cn(
              'inline-flex items-center gap-2 rounded-xl bg-brand px-4 py-2.5 text-[13px] font-medium text-white shadow-sm',
              hePressable,
              heSpring,
            )}
            onClick={onOpenDocuments}
          >
            <FileText className="size-4" aria-hidden />
            浏览文档
          </button>
          {canReplay && (
            <Link
              to={`/meetings/${encodeURIComponent(detail.id)}/replay`}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(
                'inline-flex items-center gap-2 rounded-xl bg-surface px-4 py-2.5 text-[13px] font-medium text-text-primary ring-1 ring-inset ring-black/[0.08]',
                hePressable,
                heSpring,
                'hover:bg-black/[0.02]',
              )}
            >
              <Users className="size-4 text-brand" aria-hidden />
              圆桌回放
              <ExternalLink className="size-3.5 text-text-tertiary" aria-hidden />
            </Link>
          )}
          <span
            className="ml-auto max-w-[min(100%,14rem)] truncate font-mono text-[11px] text-text-tertiary"
            title={detail.id}
          >
            {detail.id}
          </span>
        </div>
      </div>
    </section>
  )
}
