import { getTranslator } from '@/lib/i18n'

import type { AppLocale } from '@/lib/locale'

export function formatMeetingRoundsHint(
  locale: AppLocale,
  maxRounds: number,
  freeDialogueQuestions: number,
): string {
  const t = getTranslator(locale)
  if (maxRounds <= 0) return t('meetingUi.stats.roundsNotConfigured')
  if (freeDialogueQuestions > 0) {
    return t('meetingUi.stats.roundsHintWithFreeDialogue', {
      n: maxRounds,
      q: freeDialogueQuestions,
    })
  }
  return t('meetingUi.stats.roundsHintDebate', { n: maxRounds })
}
