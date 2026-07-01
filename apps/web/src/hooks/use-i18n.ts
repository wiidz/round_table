import { useMemo } from 'react'

import { useLocale } from '@/contexts/locale-context'
import { getTranslator } from '@/lib/i18n'
import { localeIntlTag } from '@/lib/i18n/translate'
import type { AppLocale } from '@/lib/locale'
import {
  meetingFileCaption as meetingFileCaptionI18n,
  meetingFileDescription as meetingFileDescriptionI18n,
  meetingFileLabel as meetingFileLabelI18n,
  meetingModeKind,
  meetingModeShort as meetingModeShortI18n,
  meetingStatusLabel as meetingStatusLabelI18n,
  meetingStatusTone,
  type MeetingModeKind,
} from '@/lib/i18n/meeting-labels'
import { profileFileCaption as profileFileCaptionI18n } from '@/lib/i18n/profile-labels'
import {
  buildMeetingFlow as buildMeetingFlowI18n,
  meetingFlowStepStatusLabel as meetingFlowStepStatusLabelI18n,
  type MeetingFlowStepStatus,
} from '@/lib/i18n/meeting-flow'
import type { MeetingDetail } from '@/types/meeting'
import {
  messageAvatar as messageAvatarI18n,
  messageLabel as messageLabelI18n,
} from '@/lib/i18n/chat-display'
import { formatMarkdownReadingStats as formatMarkdownReadingStatsI18n } from '@/lib/i18n/markdown-reading-stats'
import { formatMeetingRoundsHint as formatMeetingRoundsHintI18n } from '@/lib/i18n/meeting-overview-stats'

export function useI18n() {
  const { locale, loading, applyLocaleFromFields, reloadLocale } = useLocale()

  return useMemo(() => {
    const t = getTranslator(locale)
    const intlTag = localeIntlTag(locale)

    return {
      locale,
      loading,
      applyLocaleFromFields,
      reloadLocale,
      t,
      intlTag,
      meetingStatusLabel: (status: string) => meetingStatusLabelI18n(locale, status),
      meetingStatusTone,
      meetingModeShort: (mode?: string) => meetingModeShortI18n(locale, mode),
      meetingModeKind,
      meetingFileLabel: (path: string, modeKind?: MeetingModeKind) =>
        meetingFileLabelI18n(locale, path, modeKind),
      meetingFileCaption: (path: string, modeKind?: MeetingModeKind) =>
        meetingFileCaptionI18n(locale, path, modeKind),
      meetingFileDescription: (path: string, modeKind?: MeetingModeKind) =>
        meetingFileDescriptionI18n(locale, path, modeKind),
      profileFileCaption: (filename: string) => profileFileCaptionI18n(locale, filename),
      buildMeetingFlow: (detail: MeetingDetail) => buildMeetingFlowI18n(locale, detail),
      meetingFlowStepStatusLabel: (status: MeetingFlowStepStatus) =>
        meetingFlowStepStatusLabelI18n(locale, status),
      messageLabel: (message: Parameters<typeof messageLabelI18n>[1]) =>
        messageLabelI18n(locale, message),
      messageAvatar: (message: Parameters<typeof messageAvatarI18n>[1]) =>
        messageAvatarI18n(locale, message),
      formatMarkdownReadingStats: (content: string) =>
        formatMarkdownReadingStatsI18n(locale, content),
      formatMeetingRoundsHint: (maxRounds: number, freeDialogueQuestions: number) =>
        formatMeetingRoundsHintI18n(locale, maxRounds, freeDialogueQuestions),
      formatNumber: (value: number) => value.toLocaleString(intlTag),
      domainNavLabel: (key: 'participant' | 'principal' | 'moderator' | 'meeting') =>
        t(`domain.${key}`),
      domainPageTitle: (key: 'participant' | 'principal' | 'moderator' | 'meeting') =>
        t(`domain.${key}`),
      domainPageEyebrow: (key: 'participant' | 'principal') => t(`eyebrow.${key}`),
      navLabel: (
        key: 'overview' | 'chat' | 'meetings' | 'briefTemplates' | 'settings' | 'workbench',
      ) => t(`nav.${key}`),
      briefTemplatePageTitle: () => t('brief.pageTitle'),
      briefTemplatePageEyebrow: () => t('brief.pageEyebrow'),
      settingsTabKey: (serverGroup: string) => {
        const mapped = t(`settings.groups.${serverGroup}`)
        return mapped.startsWith('settings.groups.') ? serverGroup : mapped
      },
      settingsTabLabel: (tabKey: string) => {
        const key = tabKey as 'service' | 'storage' | 'llm' | 'meeting' | 'im'
        return t(`pages.settings.tabs.${key}`)
      },
    }
  }, [locale, loading, applyLocaleFromFields, reloadLocale])
}

/** Non-React helper when only locale + t are needed */
export function i18nFor(locale: AppLocale) {
  const t = getTranslator(locale)
  return { locale, t, intlTag: localeIntlTag(locale) }
}
