import { Plus, Trash2 } from 'lucide-react'

import { BriefTemplateMeetingConfigFields } from '@/components/brief/brief-template-meeting-config-fields'
import { BriefTemplateMetaFields } from '@/components/brief/brief-template-meta-fields'
import { briefTemplateBodyGridClass } from '@/components/brief/brief-template-preview'
import { BriefTemplateScopeFields } from '@/components/brief/brief-template-scope-fields'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  BRIEF_TEMPLATE_SECTIONS,
  BRIEF_TOPIC_EMPTY_COPY,
  briefFieldCaptionClass,
  briefTemplateLeftColumnClass,
  briefTemplateRightColumnClass,
} from '@/components/brief/brief-template-sections'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { hePressable, heSpring } from '@/lib/highend-styles'
import { emptyBriefDocument } from '@/lib/brief-template-document'
import type { BriefTemplateDocument } from '@/types/brief-template'
import { cn } from '@/lib/utils'

interface BriefTemplateFormFieldsProps {
  document: BriefTemplateDocument
  readonly?: boolean
  onChange: (next: BriefTemplateDocument) => void
}

const agendaInputClass =
  'h-9 border-0 bg-transparent shadow-none ring-0 focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-primary/45'

export function BriefTemplateFormFields({
  document,
  readonly,
  onChange,
}: BriefTemplateFormFieldsProps) {
  function patch(partial: Partial<BriefTemplateDocument>) {
    onChange({ ...document, ...partial })
  }

  function currentBrief() {
    return { ...emptyBriefDocument().brief, ...document.brief }
  }

  function patchBrief(partial: Partial<BriefTemplateDocument['brief']>) {
    onChange({ ...document, brief: { ...currentBrief(), ...partial } })
  }

  const agenda = currentBrief().agenda?.length ? currentBrief().agenda! : ['']

  return (
    <div className="space-y-8">
      <BriefTemplateMetaFields document={document} readonly={readonly} onChange={onChange} />

      <div className={briefTemplateBodyGridClass}>
        <div className={briefTemplateLeftColumnClass}>
          <section className="space-y-4">
            <BriefSectionHeading
              title={BRIEF_TEMPLATE_SECTIONS.topicGoal.title}
              description={BRIEF_TEMPLATE_SECTIONS.topicGoal.description}
              as="h3"
            />
            <div className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="brief-topic" className={briefFieldCaptionClass}>
                  主题
                </label>
                <Input
                  id="brief-topic"
                  value={document.topic ?? ''}
                  readOnly={readonly}
                  placeholder={BRIEF_TOPIC_EMPTY_COPY.placeholder}
                  onChange={(e) => patch({ topic: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <label htmlFor="brief-goal" className={briefFieldCaptionClass}>
                  会议目标 <span className="text-destructive">*</span>
                </label>
                <Textarea
                  id="brief-goal"
                  value={document.brief.goal ?? ''}
                  readOnly={readonly}
                  rows={3}
                  className="min-h-[6rem] font-sans text-sm leading-relaxed"
                  onChange={(e) => patchBrief({ goal: e.target.value })}
                />
              </div>
            </div>
          </section>

          <section className="space-y-3">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <BriefSectionHeading
                title={BRIEF_TEMPLATE_SECTIONS.agenda.title}
                description={BRIEF_TEMPLATE_SECTIONS.agenda.description}
                as="h3"
              />
              {!readonly && (
                <button
                  type="button"
                  className={cn(
                    'inline-flex items-center gap-1 text-[12px] text-brand',
                    hePressable,
                    heSpring,
                  )}
                  onClick={() => patchBrief({ agenda: [...agenda, ''] })}
                >
                  <Plus className="size-3.5" />
                  添加议程
                </button>
              )}
            </div>

            <ul className="space-y-2">
                {agenda.map((item, index) => (
                  <li
                    key={`agenda-row-${index}`}
                    className="grid grid-cols-[2.5rem_minmax(0,1fr)_2.5rem] items-center gap-3 rounded-xs bg-black/[0.025] px-3 py-2"
                  >
                    <span className="flex size-7 items-center justify-center rounded-full bg-brand-soft/80 text-[12px] font-semibold tabular-nums text-brand">
                      {index + 1}
                    </span>
                    <Input
                      value={item}
                      readOnly={readonly}
                      placeholder={`议题 ${index + 1}`}
                      className={agendaInputClass}
                      onChange={(e) => {
                        const next = [...agenda]
                        next[index] = e.target.value
                        patchBrief({ agenda: next })
                      }}
                    />
                    {!readonly ? (
                      <button
                        type="button"
                        aria-label={`删除议题 ${index + 1}`}
                        className={cn(
                          'inline-flex size-9 items-center justify-center rounded-lg text-text-tertiary hover:bg-black/[0.04] hover:text-destructive',
                          hePressable,
                          agenda.length <= 1 && 'pointer-events-none opacity-30',
                        )}
                        disabled={agenda.length <= 1}
                        onClick={() => {
                          if (agenda.length <= 1) return
                          const next = agenda.filter((_, i) => i !== index)
                          patchBrief({ agenda: next.length ? next : [''] })
                        }}
                      >
                        <Trash2 className="size-4" />
                      </button>
                    ) : (
                      <span />
                    )}
                  </li>
                ))}
            </ul>
          </section>

          <div>
            <BriefTemplateScopeFields
              inScope={document.brief.in_scope ?? ''}
              outOfScope={document.brief.out_of_scope ?? ''}
              doneCriteria={document.brief.done_criteria ?? ''}
              readonly={readonly}
              onChange={(patch) => patchBrief(patch)}
            />
          </div>
        </div>

        <BriefTemplateMeetingConfigFields
          document={document}
          readonly={readonly}
          onChange={onChange}
          className={briefTemplateRightColumnClass}
        />
      </div>
    </div>
  )
}
