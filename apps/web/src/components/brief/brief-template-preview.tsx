import { BriefTemplateMeetingConfigPreview } from '@/components/brief/brief-template-meeting-config-preview'
import { BriefTemplateScopePreview } from '@/components/brief/brief-template-scope-fields'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  briefAgendaItemShell,
  briefFieldCaptionClass,
  briefTemplateBodyGridClass,
  briefTemplateLeftColumnClass,
  briefTemplateRightColumnClass,
} from '@/components/brief/brief-template-sections'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefSections, getBriefTopicEmptyCopy } from '@/lib/i18n/brief-sections'
import { heFieldHint } from '@/lib/highend-styles'
import { normalizeBriefDocument } from '@/lib/brief-template-document'
import { cn } from '@/lib/utils'

import type { BriefTemplateDocument } from '@/types/brief-template'

export function BriefTemplatePreview({ document }: { document: BriefTemplateDocument }) {
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const topicEmpty = getBriefTopicEmptyCopy(locale)
  const doc = normalizeBriefDocument(document)
  const agenda = doc.brief.agenda?.filter(Boolean) ?? []
  const topic = doc.topic?.trim()
  const goal = doc.brief.goal?.trim()

  return (
    <div className={briefTemplateBodyGridClass}>
      <div className={briefTemplateLeftColumnClass}>
        <section className="space-y-4">
          <BriefSectionHeading
            title={sections.topicGoal.title}
            description={sections.topicGoal.description}
          />
          <div className="space-y-5">
            <div className="space-y-1.5">
              <p className={briefFieldCaptionClass}>{t('brief.fields.topic')}</p>
              <p
                className={cn(
                  'text-[18px] font-semibold leading-snug tracking-[-0.02em]',
                  topic ? 'text-text-primary' : 'font-normal text-text-tertiary',
                )}
              >
                {topic || topicEmpty.preview}
              </p>
            </div>
            <div className="space-y-1.5">
              <p className={briefFieldCaptionClass}>{t('brief.fields.goal')}</p>
              <p
                className={cn(
                  'text-[15px] leading-relaxed',
                  goal ? 'font-medium text-text-primary' : 'text-text-tertiary',
                )}
              >
                {goal || t('brief.fields.goalEmpty')}
              </p>
            </div>
          </div>
        </section>

        <section className="space-y-4">
          <BriefSectionHeading
            title={sections.agenda.title}
            description={sections.agenda.description}
          />
          {agenda.length > 0 ? (
            <ol className="space-y-2.5">
              {agenda.map((item, index) => (
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
          ) : (
            <p className={heFieldHint}>{t('brief.fields.agendaEmpty')}</p>
          )}
        </section>

        <div>
          <BriefTemplateScopePreview
            inScope={doc.brief.in_scope}
            outOfScope={doc.brief.out_of_scope}
            doneCriteria={doc.brief.done_criteria}
          />
        </div>
      </div>

      <BriefTemplateMeetingConfigPreview
        document={doc}
        className={briefTemplateRightColumnClass}
      />
    </div>
  )
}
