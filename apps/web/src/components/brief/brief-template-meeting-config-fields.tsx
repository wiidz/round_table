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
import { heInputEditable } from '@/lib/highend-styles'
import type { BriefTemplateDocument, MeetingDefaults } from '@/types/brief-template'
import { cn } from '@/lib/utils'

const SELECT_CHEVRON =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E\")"

const selectClassName = cn(
  heInputEditable,
  'h-10 w-full cursor-pointer appearance-none px-3 pr-9 text-sm text-text-primary',
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
      <label htmlFor={htmlFor} className={cn(briefMeetingConfigLabelClass, 'break-words pt-2 sm:max-w-[6rem] sm:pt-2.5')}>
        {label}
      </label>
      <div className="flex min-w-0 items-start gap-1.5">
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

function patchMeetingDocument(
  document: BriefTemplateDocument,
  partial: Partial<MeetingDefaults>,
): BriefTemplateDocument {
  const nextMeeting: MeetingDefaults = { ...document.meeting }
  for (const [key, value] of Object.entries(partial) as [keyof MeetingDefaults, MeetingDefaults[keyof MeetingDefaults]][]) {
    if (value === undefined || value === null || (Array.isArray(value) && value.length === 0)) {
      delete nextMeeting[key]
    } else {
      nextMeeting[key] = value as never
    }
  }
  return {
    ...document,
    meeting: Object.keys(nextMeeting).length > 0 ? nextMeeting : undefined,
  }
}

function parseOptionalInt(raw: string): number | undefined {
  const trimmed = raw.trim()
  if (!trimmed) return undefined
  const n = Number.parseInt(trimmed, 10)
  return Number.isNaN(n) ? undefined : n
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
  const meeting = document.meeting
  const mode = meeting?.mode ?? ''
  const showMinSynthesis = mode === 'deliberation'

  function patchMeeting(partial: Partial<MeetingDefaults>) {
    onChange(patchMeetingDocument(document, partial))
  }

  return (
    <aside className={cn('space-y-5 lg:sticky lg:top-20 lg:self-start', className)}>
      <BriefSectionHeading
        title={sections.meeting.title}
        description={sections.meeting.description}
      />
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigFieldRow
          label={configLabels.mode}
          htmlFor="brief-mode"
          hint={t('brief.config.deferHint')}
        >
          <select
            id="brief-mode"
            value={mode}
            disabled={readonly}
            className={selectClassName}
            style={selectStyle}
            onChange={(e) =>
              patchMeeting({
                mode: e.target.value || undefined,
                ...(e.target.value !== 'deliberation'
                  ? { min_rounds_before_synthesis: undefined }
                  : {}),
              })
            }
          >
            <option value="">{t('brief.config.deferToMeeting')}</option>
            <option value="decision">{t('brief.config.modeDecision')}</option>
            <option value="deliberation">{t('brief.config.modeDeliberation')}</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={configLabels.confirmation}
          htmlFor="brief-confirmation"
          hint={t('brief.config.deferHint')}
        >
          <select
            id="brief-confirmation"
            value={meeting?.confirmation_mode ?? ''}
            disabled={readonly}
            className={selectClassName}
            style={selectStyle}
            onChange={(e) => patchMeeting({ confirmation_mode: e.target.value || undefined })}
          >
            <option value="">{t('brief.config.deferToMeeting')}</option>
            <option value="required">{t('brief.config.confirmationRequired')}</option>
            <option value="skip">{t('brief.config.confirmationSkip')}</option>
          </select>
        </BriefMeetingConfigFieldRow>
        <BriefMeetingConfigFieldRow
          label={configLabels.maxRounds}
          htmlFor="brief-max-rounds"
          hint={t('brief.config.deferHint')}
        >
          <Input
            id="brief-max-rounds"
            type="number"
            min={1}
            value={meeting?.max_rounds ?? ''}
            readOnly={readonly}
            placeholder={t('brief.config.deferToMeeting')}
            onChange={(e) => patchMeeting({ max_rounds: parseOptionalInt(e.target.value) })}
          />
        </BriefMeetingConfigFieldRow>
        {showMinSynthesis && (
          <BriefMeetingConfigFieldRow
            label={configLabels.minSynthesis}
            htmlFor="brief-min-synthesis"
            hint={t('brief.config.minSynthesisHint')}
          >
            <Input
              id="brief-min-synthesis"
              type="number"
              min={1}
              value={meeting?.min_rounds_before_synthesis ?? ''}
              readOnly={readonly}
              placeholder={t('brief.config.deferToMeeting')}
              onChange={(e) =>
                patchMeeting({ min_rounds_before_synthesis: parseOptionalInt(e.target.value) })
              }
            />
          </BriefMeetingConfigFieldRow>
        )}
        <BriefMeetingConfigFieldRow
          label={configLabels.freeDialogue}
          htmlFor="brief-free-dialogue"
          hint={t('brief.config.deferHint')}
        >
          <Input
            id="brief-free-dialogue"
            type="number"
            min={0}
            value={meeting?.free_dialogue_max_questions ?? ''}
            readOnly={readonly}
            placeholder={t('brief.config.deferToMeeting')}
            onChange={(e) =>
              patchMeeting({ free_dialogue_max_questions: parseOptionalInt(e.target.value) })
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
            value={meeting?.participant_ids ?? []}
            disabled={readonly}
            onChange={(participant_ids) => patchMeeting({ participant_ids })}
          />
        </BriefMeetingConfigFieldRow>
      </div>
    </aside>
  )
}
