import { useLocale } from '@/contexts/locale-context'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useI18n } from '@/hooks/use-i18n'
import { hePanelShell } from '@/lib/highend-styles'
import { localeToUserLanguage } from '@/lib/locale'
import { cn } from '@/lib/utils'

import type { PrincipalUserProfile } from '@/types/principal'

export function PrincipalUserPreview({
  profile,
  embedded,
}: {
  profile: PrincipalUserProfile
  embedded?: boolean
}) {
  const { t } = useI18n()
  const { locale: appLocale } = useLocale()
  const languageCode = localeToUserLanguage(appLocale)
  const languageDisplay =
    appLocale === 'en'
      ? `${t('pages.settings.localeEn')} (${languageCode})`
      : `${t('pages.settings.localeZh')} (${languageCode})`

  const fields = (
    <div className="space-y-8">
      <SettingsFieldRow
        label={t('profile.principal.languageLabel')}
        hint={t('profile.principal.languageHint')}
      >
        <Input value={languageDisplay} readOnly tabIndex={-1} />
      </SettingsFieldRow>
      <SettingsFieldRow
        label={t('profile.principal.confirmationLabel')}
        hint={t('profile.principal.confirmationHint')}
      >
        <Input
          value={profile.confirmation?.trim() || t('profile.principal.persona.emptyValue')}
          readOnly
          tabIndex={-1}
        />
      </SettingsFieldRow>
      <SettingsFieldRow
        label={t('profile.principal.contextLabel')}
        hint={t('profile.principal.contextHint')}
      >
        <Textarea
          value={profile.context?.trim() || t('profile.principal.persona.emptyValue')}
          readOnly
          tabIndex={-1}
          rows={5}
          className="min-h-[8rem] font-sans"
        />
      </SettingsFieldRow>
    </div>
  )

  if (embedded) {
    return fields
  }

  return (
    <div className={cn(hePanelShell, 'space-y-6 p-6 sm:p-8')}>
      <p className="text-sm leading-relaxed text-text-secondary">
        {t('profile.principal.preview.description')}
      </p>
      {fields}
    </div>
  )
}
