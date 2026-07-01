import { BriefMeetingExpertsList } from '@/components/brief/brief-meeting-experts-list'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefMeetingConfigRow } from '@/components/brief/brief-meeting-config-row'
import {
  BRIEF_MEETING_CONFIG_LABELS,
  BRIEF_TEMPLATE_SECTIONS,
  briefConfigPanelShell,
} from '@/components/brief/brief-template-sections'
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

function confirmationLabel(meetingMd: string): string {
  const cell = extractMeetingTableCell(meetingMd, '确认模式')
  if (/跳过|skip/i.test(cell)) return '跳过 Principal 确认'
  if (/需要|required/i.test(cell)) return '需要 Principal 确认'
  return parseConfirmationRequired(meetingMd)
    ? '需要 Principal 确认'
    : '跳过 Principal 确认'
}

function freeDialogueLabel(detail: MeetingDetail, meetingMd: string): string {
  if (!detail.free_dialogue) return '未启用'
  const cell = extractMeetingTableCell(meetingMd, 'Round 1 后自由对话')
  const match = cell.match(/每人最多\s*(\d+)\s*轮/)
  if (match?.[1]) return `${match[1]} 问`
  return '已启用'
}

function minSynthesisLabel(meetingMd: string): string | null {
  const cell = extractMeetingTableCell(meetingMd, '合成轮次')
  const match = cell.match(/(\d+)/)
  if (match?.[1]) return `第 ${match[1]} 轮起`
  const yaml = meetingMd.match(/min_rounds_before_synthesis:\s*(\d+)/i)
  if (yaml?.[1]) return `第 ${yaml[1]} 轮起`
  return null
}

function expertEntries(detail: MeetingDetail, meetingMd: string) {
  const participants = parseParticipantsFromMeetingMd(meetingMd)
  if (participants.length > 0) {
    return participants.map((p) => ({ id: p.id, name: p.label }))
  }
  return []
}

function expertEmptyLabel(
  detail: MeetingDetail,
  experts: ReturnType<typeof expertEntries>,
): string {
  if (experts.length > 0) return '—'
  if (detail.participant_count && detail.participant_count > 0) {
    return `${detail.participant_count} 人`
  }
  return '—'
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
  const meetingMd = detail.files?.['MEETING.md'] ?? ''
  const kind = modeKind ?? meetingModeKind(detail.mode_kind, detail.mode)
  const isDeliberation = kind === 'deliberation'
  const synthesisLabel = isDeliberation ? minSynthesisLabel(meetingMd) : null
  const experts = expertEntries(detail, meetingMd)
  const expertsEmptyLabel = expertEmptyLabel(detail, experts)

  return (
    <aside className={cn(hePanelShell, 'overflow-visible p-4', className)}>
      <div className="mb-3">
        <BriefSectionHeading
          title={BRIEF_TEMPLATE_SECTIONS.meeting.title}
          description={BRIEF_TEMPLATE_SECTIONS.meeting.description}
        />
      </div>
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.mode}
          value={isDeliberation ? '研讨型（deliberation）' : '裁决型（decision）'}
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.confirmation}
          value={confirmationLabel(meetingMd)}
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.maxRounds}
          value={detail.max_rounds && detail.max_rounds > 0 ? `${detail.max_rounds} 轮` : '—'}
        />
        {synthesisLabel && (
          <BriefMeetingConfigRow
            label={BRIEF_MEETING_CONFIG_LABELS.minSynthesis}
            value={synthesisLabel}
          />
        )}
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.freeDialogue}
          value={freeDialogueLabel(detail, meetingMd)}
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.experts}
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
