/**
 * Web UI copy: Chinese label + English domain term (see docs/NAMING.md).
 * Code/API types stay English; user-facing text uses these helpers.
 */
export const UI_DOMAIN = {
  participant: { label: '专家', term: 'Participant' },
  principal: { label: '委托人', term: 'Principal' },
  moderator: { label: '司仪', term: 'Moderator' },
  meeting: { label: '会议', term: 'Meeting' },
} as const

export type DomainKey = keyof typeof UI_DOMAIN

/** Nav / short contexts — Chinese only */
export function domainNavLabel(key: DomainKey): string {
  return UI_DOMAIN[key].label
}

/** Page title — 专家 · Participant */
export function domainPageTitle(key: DomainKey): string {
  const { label, term } = UI_DOMAIN[key]
  return `${label} · ${term}`
}

/** Caption under page title */
export function domainPageEyebrow(key: DomainKey): string {
  return `RoundTable ${UI_DOMAIN[key].term}`
}

/** Bilingual file caption: 会议纪要 · MINUTES.md */
export function fileCaption(title: string, path: string): string {
  if (!title || title === path) return path
  return `${title} · ${path}`
}
