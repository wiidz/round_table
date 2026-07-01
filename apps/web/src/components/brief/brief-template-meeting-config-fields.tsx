import type { ReactNode } from 'react'

import { ParticipantMultiSelect } from '@/components/brief/participant-multi-select'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  BRIEF_MEETING_CONFIG_LABELS,
  BRIEF_TEMPLATE_SECTIONS,
  briefConfigPanelShell,
  briefMeetingConfigLabelClass,
  briefMeetingConfigRowGrid,
} from '@/components/brief/brief-template-sections'
import { FieldHintPopover } from '@/components/settings/field-hint-popover'
import { Input } from '@/components/ui/input'
import { heFieldSurface } from '@/lib/highend-styles'
import { emptyBriefDocument } from '@/lib/brief-template-document'
import type { BriefTemplateDocument } from '@/types/brief-template'
import { cn } from '@/lib/utils'

const SELECT_CHEVRON =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E\")"

const selectClassName = cn(
  heFieldSurface,
  'h-10 w-full cursor-pointer appearance-none bg-surface px-3 pr-9 text-sm text-text-primary',
)

const selectStyle = {
  backgroundImage: SELECT_CHEVRON,
  backgroundRepeat: 'no-repeat',
  backgroundPosition: 'right 0.65rem center',
  backgroundSize: '1rem',
} as const

function BriefMeetingConfigFieldRow({
  label,
  htmlFor,
  hint,
  children,
}: {
  label: string
  htmlFor?: string
  hint?: string
  children: ReactNode
}) {
  return (
    <div className={briefMeetingConfigRowGrid}>
      <label htmlFor={htmlFor} className={briefMeetingConfigLabelClass}>
        {label}
      </label>
      <div className="flex min-w-0 items-center gap-1.5">
        <div className="min-w-0 flex-1">{children}</div>
        {hint && <FieldHintPopover content={hint} ariaLabel={`${label} 说明`} />}
      </div>
    </div>
  )
}

interface BriefTemplateMeetingConfigFieldsProps {
  document: BriefTemplateDocument
  readonly?: boolean
  onChange: (next: BriefTemplateDocument) => void
  className?: string
}

export function BriefTemplateMeetingConfigFields({
  document,
  readonly,
  onChange,
  className,
}: BriefTemplateMeetingConfigFieldsProps) {
  function patchMeeting(partial: Partial<NonNullable<BriefTemplateDocument['meeting']>>) {
    onChange({
      ...document,
      meeting: { ...emptyBriefDocument().meeting, ...document.meeting, ...partial },
    })
  }

  return (
    <aside className={cn('space-y-5 lg:sticky lg:top-20 lg:self-start', className)}>
      <BriefSectionHeading
        title={BRIEF_TEMPLATE_SECTIONS.meeting.title}
        description={BRIEF_TEMPLATE_SECTIONS.meeting.description}
      />
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigFieldRow label={BRIEF_MEETING_CONFIG_LABELS.mode} htmlFor="brief-mode">
          <select
            id="brief-mode"
            value={document.meeting?.mode ?? 'decision'}
            disabled={readonly}
            className={selectClassName}
            style={selectStyle}
            onChange={(e) => patchMeeting({ mode: e.target.value })}
          >
            <option value="decision">裁决型（decision）</option>
            <option value="deliberation">研讨型（deliberation）</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={BRIEF_MEETING_CONFIG_LABELS.confirmation}
          htmlFor="brief-confirmation"
        >
          <select
            id="brief-confirmation"
            value={document.meeting?.confirmation_mode ?? 'required'}
            disabled={readonly}
            className={selectClassName}
            style={selectStyle}
            onChange={(e) => patchMeeting({ confirmation_mode: e.target.value })}
          >
            <option value="required">需要 Principal 确认</option>
            <option value="skip">跳过确认关</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={BRIEF_MEETING_CONFIG_LABELS.maxRounds}
          htmlFor="brief-max-rounds"
        >
          <Input
            id="brief-max-rounds"
            type="number"
            min={1}
            value={document.meeting?.max_rounds ?? 3}
            readOnly={readonly}
            onChange={(e) =>
              patchMeeting({ max_rounds: Number.parseInt(e.target.value, 10) || 1 })
            }
          />
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={BRIEF_MEETING_CONFIG_LABELS.minSynthesis}
          htmlFor="brief-min-synthesis"
          hint="仅研讨型生效"
        >
          <Input
            id="brief-min-synthesis"
            type="number"
            min={1}
            value={document.meeting?.min_rounds_before_synthesis ?? 2}
            readOnly={readonly}
            onChange={(e) =>
              patchMeeting({
                min_rounds_before_synthesis: Number.parseInt(e.target.value, 10) || 1,
              })
            }
          />
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={BRIEF_MEETING_CONFIG_LABELS.freeDialogue}
          htmlFor="brief-free-dialogue"
        >
          <Input
            id="brief-free-dialogue"
            type="number"
            min={0}
            value={document.meeting?.free_dialogue_max_questions ?? 1}
            readOnly={readonly}
            onChange={(e) =>
              patchMeeting({
                free_dialogue_max_questions: Number.parseInt(e.target.value, 10) || 0,
              })
            }
          />
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={BRIEF_MEETING_CONFIG_LABELS.experts}
          htmlFor="brief-participants"
          hint="可多选；留空表示默认邀请全部专家"
        >
          <ParticipantMultiSelect
            id="brief-participants"
            value={document.meeting?.participant_ids ?? []}
            disabled={readonly}
            onChange={(participant_ids) => patchMeeting({ participant_ids })}
          />
        </BriefMeetingConfigFieldRow>
      </div>
    </aside>
  )
}
