import type { ReactNode } from 'react'

import { ParticipantMultiSelect } from '@/components/brief/participant-multi-select'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  briefConfigPanelShell,
  briefMeetingConfigLabelClass,
  briefMeetingConfigRowGrid,
} from '@/components/brief/brief-template-sections'
import { FieldHintPopover } from '@/components/settings/field-hint-popover'
import { Input } from '@/components/ui/input'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefMeetingConfigLabels, getBriefSections } from '@/lib/i18n/brief-sections'
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
  const { t } = useI18n()

  return (
    <div className={briefMeetingConfigRowGrid}>
      <label htmlFor={htmlFor} className={briefMeetingConfigLabelClass}>
        {label}
      </label>
      <div className="flex min-w-0 items-center gap-1.5">
        <div className="min-w-0 flex-1">{children}</div>
        {hint && (
          <FieldHintPopover
            content={hint}
            ariaLabel={t('brief.fieldHintAria', { label })}
          />
        )}
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
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const configLabels = getBriefMeetingConfigLabels(locale)

  function patchMeeting(partial: Partial<NonNullable<BriefTemplateDocument['meeting']>>) {
    onChange({
      ...document,
      meeting: { ...emptyBriefDocument().meeting, ...document.meeting, ...partial },
    })
  }

  return (
    <aside className={cn('space-y-5 lg:sticky lg:top-20 lg:self-start', className)}>
      <BriefSectionHeading
        title={sections.meeting.title}
        description={sections.meeting.description}
      />
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigFieldRow label={configLabels.mode} htmlFor="brief-mode">
          <select
            id="brief-mode"
            value={document.meeting?.mode ?? 'decision'}
            disabled={readonly}
            className={selectClassName}
            style={selectStyle}
            onChange={(e) => patchMeeting({ mode: e.target.value })}
          >
            <option value="decision">{t('brief.config.modeDecision')}</option>
            <option value="deliberation">{t('brief.config.modeDeliberation')}</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={configLabels.confirmation}
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
            <option value="required">{t('brief.config.confirmationRequired')}</option>
            <option value="skip">{t('brief.config.confirmationSkip')}</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={configLabels.maxRounds}
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
          label={configLabels.minSynthesis}
          htmlFor="brief-min-synthesis"
          hint={t('brief.config.minSynthesisHint')}
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
          label={configLabels.freeDialogue}
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
          label={configLabels.experts}
          htmlFor="brief-participants"
          hint={t('brief.config.expertsHint')}
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
