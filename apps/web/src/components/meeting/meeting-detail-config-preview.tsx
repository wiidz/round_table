import { BriefMeetingExpertsList } from '@/components/brief/brief-meeting-experts-list'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefMeetingConfigRow } from '@/components/brief/brief-meeting-config-row'
import { briefConfigPanelShell, meetingDetailConfigPanelClass } from '@/components/brief/brief-template-sections'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefMeetingConfigLabels, getBriefSections } from '@/lib/i18n/brief-sections'
import { meetingModeKind, type MeetingModeKind } from '@/lib/meeting-labels'
import { parseParticipantsFromMeetingMd } from '@/lib/meeting-participants'
import { parseConfirmationRequired } from '@/lib/meeting-flow'
import { hePanelShell } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

function extractMeetingTableCell(doc: string, rowLabel: string): string {
  const row = doc.match(new RegExp(`${rowLabel}\\s*\\|\\s*([^\\n|]+)`))
  return row?.[1]?.trim().replace(/^`|`$/g, '') ?? ''
}

interface MeetingDetailConfigPreviewProps {
  detail: MeetingDetail
  modeKind?: MeetingModeKind
  className?: string
}

export function MeetingDetailConfigPreview({
  detail,
  modeKind,
  className,
}: MeetingDetailConfigPreviewProps) {
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const configLabels = getBriefMeetingConfigLabels(locale)
  const meetingMd = detail.files?.['MEETING.md'] ?? ''
  const kind = modeKind ?? meetingModeKind(detail.mode_kind, detail.mode)
  const isDeliberation = kind === 'deliberation'

  function confirmationLabel(): string {
    const cell = extractMeetingTableCell(meetingMd, '确认模式')
    if (/跳过|skip/i.test(cell)) return t('brief.config.confirmationSkipPrincipal')
    if (/需要|required/i.test(cell)) return t('brief.config.confirmationRequired')
    return parseConfirmationRequired(meetingMd)
      ? t('brief.config.confirmationRequired')
      : t('brief.config.confirmationSkipPrincipal')
  }

  function freeDialogueLabel(): string {
    if (!detail.free_dialogue) return t('common.disabled')
    const cell = extractMeetingTableCell(meetingMd, 'Round 1 后自由对话')
    const match = cell.match(/每人最多\s*(\d+)\s*轮/)
    if (match?.[1]) {
      return t('brief.config.freeDialogueQuestions', { n: Number.parseInt(match[1], 10) })
    }
    return t('common.enabled')
  }

  function minSynthesisLabel(): string | null {
    const cell = extractMeetingTableCell(meetingMd, '合成轮次')
    const match = cell.match(/(\d+)/)
    if (match?.[1]) {
      return t('brief.config.minSynthesisFrom', { n: Number.parseInt(match[1], 10) })
    }
    const yaml = meetingMd.match(/min_rounds_before_synthesis:\s*(\d+)/i)
    if (yaml?.[1]) {
      return t('brief.config.minSynthesisFrom', { n: Number.parseInt(yaml[1], 10) })
    }
    return null
  }

  function expertEntries() {
    const participants = parseParticipantsFromMeetingMd(meetingMd)
    if (participants.length > 0) {
      return participants.map((p) => ({ id: p.id, name: p.label }))
    }
    return []
  }

  function expertEmptyLabel(experts: ReturnType<typeof expertEntries>): string {
    if (experts.length > 0) return '—'
    if (detail.participant_count && detail.participant_count > 0) {
      return t('meeting.meta.participants', { n: detail.participant_count })
    }
    return '—'
  }

  const synthesisLabel = isDeliberation ? minSynthesisLabel() : null
  const experts = expertEntries()
  const expertsEmptyLabel = expertEmptyLabel(experts)

  return (
    <aside className={cn(hePanelShell, meetingDetailConfigPanelClass, 'overflow-visible p-4', className)}>
      <div className="mb-3">
        <BriefSectionHeading
          title={sections.meeting.title}
          description={sections.meeting.description}
        />
      </div>
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigRow
          layout="detail"
          label={configLabels.mode}
          value={
            isDeliberation ? t('brief.config.modeDeliberation') : t('brief.config.modeDecision')
          }
        />
        <BriefMeetingConfigRow
          layout="detail"
          label={configLabels.confirmation}
          value={confirmationLabel()}
        />
        <BriefMeetingConfigRow
          layout="detail"
          label={configLabels.maxRounds}
          value={
            detail.max_rounds && detail.max_rounds > 0
              ? t('brief.config.maxRoundsValue', { n: detail.max_rounds })
              : '—'
          }
        />
        {synthesisLabel && (
          <BriefMeetingConfigRow
            layout="detail"
            label={configLabels.minSynthesis}
            value={synthesisLabel}
          />
        )}
        <BriefMeetingConfigRow
          layout="detail"
          label={configLabels.freeDialogue}
          value={freeDialogueLabel()}
        />
        <BriefMeetingConfigRow
          layout="detail"
          label={configLabels.experts}
          valueAlign="start"
          value={
            <BriefMeetingExpertsList
              experts={experts}
              emptyLabel={expertsEmptyLabel}
            />
          }
        />
      </div>
    </aside>
  )
}
