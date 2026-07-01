import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'

export interface BriefScopePreset {
  label: string
  value: string
}

export function getBriefScopePresets(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    inScope: [
      {
        label: t('brief.scopePresets.inScope.topicPlan.label'),
        value: t('brief.scopePresets.inScope.topicPlan.value'),
      },
      {
        label: t('brief.scopePresets.inScope.agendaRelated.label'),
        value: t('brief.scopePresets.inScope.agendaRelated.value'),
      },
      {
        label: t('brief.scopePresets.inScope.riskAssumptions.label'),
        value: t('brief.scopePresets.inScope.riskAssumptions.value'),
      },
    ],
    outOfScope: [
      {
        label: t('brief.scopePresets.outOfScope.scheduleStaff.label'),
        value: t('brief.scopePresets.outOfScope.scheduleStaff.value'),
      },
      {
        label: t('brief.scopePresets.outOfScope.extendedTopics.label'),
        value: t('brief.scopePresets.outOfScope.extendedTopics.value'),
      },
      {
        label: t('brief.scopePresets.outOfScope.longTermPlanning.label'),
        value: t('brief.scopePresets.outOfScope.longTermPlanning.value'),
      },
    ],
    doneCriteria: [
      {
        label: t('brief.scopePresets.doneCriteria.actionableConclusions.label'),
        value: t('brief.scopePresets.doneCriteria.actionableConclusions.value'),
      },
      {
        label: t('brief.scopePresets.doneCriteria.principalConfirmation.label'),
        value: t('brief.scopePresets.doneCriteria.principalConfirmation.value'),
      },
      {
        label: t('brief.scopePresets.doneCriteria.actionItems.label'),
        value: t('brief.scopePresets.doneCriteria.actionItems.value'),
      },
    ],
  } satisfies Record<string, BriefScopePreset[]>
}
