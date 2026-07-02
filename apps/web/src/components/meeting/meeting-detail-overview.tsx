import { Link } from 'react-router-dom'
import { ChevronRight, Download, ExternalLink, FileText, Trash2, Users } from 'lucide-react'

import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefTemplateScopePreview } from '@/components/brief/brief-template-scope-fields'
import {
  briefAgendaItemShell,
  briefFieldLabelClass,
} from '@/components/brief/brief-template-sections'
import { MarkdownSnippet } from '@/components/markdown/markdown-snippet'
import { MeetingModeInline } from '@/components/meeting/meeting-mode-badge'
import { MeetingOverviewStatCards } from '@/components/meeting/meeting-overview-stat-cards'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefSections } from '@/lib/i18n/brief-sections'
import { hePageTitle, hePanelShell, hePressable, heSpring } from '@/lib/highend-styles'
import { primaryDeliverablePath, type MeetingModeKind } from '@/lib/meeting-labels'
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
  onDownload?: () => void
  onDelete?: () => void
  downloading?: boolean
  deleting?: boolean
}

function isMeetingRunning(status: string): boolean {
  return status === 'Running' || status === '进行中'
}

function isMeetingAborted(status: string): boolean {
  return status === 'Aborted' || status === 'aborted' || status === '已中断'
}

export function MeetingDetailOverview({
  detail,
  brief,
  modeKind,
  canReplay,
  onOpenDocuments,
  onOpenConclusion,
  onDownload,
  onDelete,
  downloading = false,
  deleting = false,
}: MeetingDetailOverviewProps) {
  const { t, locale, meetingFileLabel } = useI18n()
  const sections = getBriefSections(locale)
  const primaryPath = primaryDeliverablePath(modeKind)
  const primaryTitle = meetingFileLabel(primaryPath, modeKind)
  const topic = brief.topic.trim()
  const goal = brief.goal.trim()
  const conclusion = brief.conclusion?.trim()

  const sessionMeta = detail.started_at?.trim()

  const conclusionEmpty = isMeetingRunning(detail.status)
    ? t('meetingUi.overview.conclusionRunning')
    : isMeetingAborted(detail.status)
      ? t('meetingUi.overview.conclusionAborted')
      : t('meetingUi.overview.conclusionEmpty')

  return (
    <section className={cn(hePanelShell, 'overflow-visible p-5 sm:p-6')}>
      <div className="space-y-8">
        <div className="space-y-3">
          <p className="text-[10px] font-medium uppercase tracking-[0.18em] text-text-tertiary">
            {t('meeting.reviewEyebrow')}
          </p>
          <h1
            className={cn(
              hePageTitle,
              'text-balance',
              !topic && 'font-normal text-text-tertiary',
            )}
          >
            {topic || t('meeting.topicEmpty')}
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
              <p className={briefFieldLabelClass}>{t('meetingUi.overview.goal')}</p>
            </div>
            <p
              className={cn(
                'pl-3.5 text-[15px] leading-relaxed',
                goal ? 'font-medium text-text-primary' : 'text-text-tertiary',
              )}
            >
              {goal || t('meetingUi.overview.goalEmpty')}
            </p>
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="size-1.5 shrink-0 rounded-full bg-brand" aria-hidden />
              <p className={briefFieldLabelClass}>{t('meetingUi.overview.conclusion')}</p>
            </div>
            {conclusion ? (
              <MarkdownSnippet content={conclusion} className="pl-3.5" />
            ) : (
              <p className="pl-3.5 text-[15px] leading-relaxed text-text-tertiary">{conclusionEmpty}</p>
            )}
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
                {t('meetingUi.overview.viewFullDeliverable', { title: primaryTitle })}
                <ChevronRight className="size-3.5" aria-hidden />
              </button>
            )}
          </div>
        </div>

        {brief.agenda.length > 0 && (
          <section className="space-y-4">
            <BriefSectionHeading
              title={sections.agenda.title}
              description={sections.agenda.description}
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
            {t('meetingUi.overview.browseDocuments')}
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
              {t('meetingUi.overview.replay')}
              <ExternalLink className="size-3.5 text-text-tertiary" aria-hidden />
            </Link>
          )}
          {onDownload && (
            <button
              type="button"
              disabled={downloading}
              className={cn(
                'inline-flex items-center gap-2 rounded-xl bg-surface px-4 py-2.5 text-[13px] font-medium text-text-primary ring-1 ring-inset ring-black/[0.08]',
                hePressable,
                heSpring,
                'hover:bg-black/[0.02] disabled:cursor-not-allowed disabled:opacity-60',
              )}
              onClick={onDownload}
            >
              <Download className="size-4 text-brand" aria-hidden />
              {downloading
                ? t('meetingUi.overview.downloading')
                : t('meetingUi.overview.download')}
            </button>
          )}
          {onDelete && (
            <button
              type="button"
              disabled={deleting}
              className={cn(
                'inline-flex items-center gap-2 rounded-xl bg-surface px-4 py-2.5 text-[13px] font-medium text-danger ring-1 ring-inset ring-danger/20',
                hePressable,
                heSpring,
                'hover:bg-black/[0.05] disabled:cursor-not-allowed disabled:opacity-60',
              )}
              onClick={onDelete}
            >
              <Trash2 className="size-4" aria-hidden />
              {deleting ? t('meetingUi.overview.deleting') : t('meetingUi.overview.delete')}
            </button>
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
