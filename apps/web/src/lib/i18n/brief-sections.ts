import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'

export function getBriefSections(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    topicGoal: {
      title: t('brief.sections.topicGoal.title'),
      description: t('brief.sections.topicGoal.description'),
    },
    agenda: {
      title: t('brief.sections.agenda.title'),
      description: t('brief.sections.agenda.description'),
    },
    scope: {
      title: t('brief.sections.scope.title'),
      description: t('brief.sections.scope.description'),
    },
    meeting: {
      title: t('brief.sections.meeting.title'),
      description: t('brief.sections.meeting.description'),
    },
  } as const
}

export function getBriefTopicEmptyCopy(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    preview: t('brief.topic.emptyPreview'),
    placeholder: t('brief.topic.emptyPlaceholder'),
  } as const
}

export function getBriefMeetingConfigLabels(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    mode: t('brief.config.mode'),
    confirmation: t('brief.config.confirmation'),
    maxRounds: t('brief.config.maxRounds'),
    minSynthesis: t('brief.config.minSynthesis'),
    freeDialogue: t('brief.config.freeDialogue'),
    experts: t('brief.config.experts'),
  } as const
}

export function getBriefScopeEmptyCopy(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    inScope: t('brief.scope.emptyInScope'),
    outOfScope: t('brief.scope.emptyOutOfScope'),
    doneCriteria: t('brief.scope.emptyDone'),
  } as const
}
