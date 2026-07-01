import { useI18n } from '@/hooks/use-i18n'

/** @deprecated Use useI18n() instead */
export function useUILabels() {
  const i18n = useI18n()
  return {
    locale: i18n.locale,
    domainNavLabel: i18n.domainNavLabel,
    domainPageTitle: i18n.domainPageTitle,
    domainPageEyebrow: i18n.domainPageEyebrow,
    navLabel: i18n.navLabel,
    briefTemplatePageTitle: i18n.briefTemplatePageTitle,
    briefTemplatePageEyebrow: i18n.briefTemplatePageEyebrow,
    t: i18n.t,
  }
}
