/**
 * Web UI copy keyed by ROUND_TABLE_LOCALE (zh | en).
 * Code/API types stay English; user-facing text uses these helpers.
 */
import type { AppLocale } from '@/lib/locale'

export const UI_DOMAIN = {
  zh: {
    participant: '专家',
    principal: '委托人',
    moderator: '主持人',
    meeting: '会议',
  },
  en: {
    participant: 'Participant',
    principal: 'Principal',
    moderator: 'Moderator',
    meeting: 'Meeting',
  },
} as const

export type DomainKey = keyof (typeof UI_DOMAIN)['zh']

const DOMAIN_TERM: Record<DomainKey, string> = {
  participant: 'Participant',
  principal: 'Principal',
  moderator: 'Moderator',
  meeting: 'Meeting',
}

export type NavKey =
  | 'overview'
  | 'chat'
  | 'meetings'
  | 'briefTemplates'
  | 'settings'
  | 'workbench'

const NAV_LABELS: Record<AppLocale, Record<NavKey, string>> = {
  zh: {
    overview: '概览',
    chat: '聊天',
    meetings: '会议',
    briefTemplates: '简报模板',
    settings: '设置',
    workbench: '委托人工作台',
  },
  en: {
    overview: 'Overview',
    chat: 'Chat',
    meetings: 'Meetings',
    briefTemplates: 'Brief Templates',
    settings: 'Settings',
    workbench: 'Principal Workbench',
  },
}

/** Nav / short contexts */
export function domainNavLabel(locale: AppLocale, key: DomainKey): string {
  return UI_DOMAIN[locale][key]
}

/** Page title — single language per locale */
export function domainPageTitle(locale: AppLocale, key: DomainKey): string {
  return UI_DOMAIN[locale][key]
}

/** Caption pill beside page title */
export function domainPageEyebrow(locale: AppLocale, key: DomainKey): string {
  if (locale === 'en') {
    return key === 'principal'
      ? 'Decision Owner'
      : key === 'participant'
        ? 'Expert Profile'
        : `RoundTable ${DOMAIN_TERM[key]}`
  }
  return key === 'principal'
    ? 'Decision Owner'
    : key === 'participant'
      ? 'Expert Profile'
      : `RoundTable ${DOMAIN_TERM[key]}`
}

export function navLabel(locale: AppLocale, key: NavKey): string {
  return NAV_LABELS[locale][key]
}

export function briefTemplatePageTitle(locale: AppLocale): string {
  return locale === 'en' ? 'Brief Templates' : '简报模板'
}

export function briefTemplatePageEyebrow(locale: AppLocale): string {
  return locale === 'en' ? 'Meeting Brief' : '简报模板'
}

/** File sidebar caption: title only, or path when no title */
export function fileCaption(title: string, path: string): string {
  if (!title || title === path) return path
  return `${title} · ${path}`
}
