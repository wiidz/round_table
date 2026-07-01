import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'

export type MeetingStatusTone = 'neutral' | 'running' | 'warning' | 'success' | 'danger'

const STATUS_TONES: Record<string, MeetingStatusTone> = {
  准备中: 'neutral',
  Preparing: 'neutral',
  进行中: 'running',
  Running: 'running',
  已暂停: 'warning',
  Paused: 'warning',
  共识达成: 'success',
  Consensus: 'success',
  'Principal 确认中': 'warning',
  Confirmation: 'warning',
  已结束: 'success',
  Completed: 'success',
  已中断: 'danger',
  aborted: 'danger',
  Aborted: 'danger',
  已归档: 'neutral',
  Archived: 'neutral',
}

export function meetingStatusLabel(locale: AppLocale, status: string): string {
  if (!status) return getTranslator(locale)('meeting.status.unknown')
  const key = `meeting.status.${status}`
  const translated = getTranslator(locale)(key)
  if (translated !== key) return translated
  if (status === '已中断') return getTranslator(locale)('meeting.status.aborted')
  return status
}

export function meetingStatusTone(status: string): MeetingStatusTone {
  if (!status) return 'neutral'
  return STATUS_TONES[status] ?? 'neutral'
}

export function meetingModeShort(locale: AppLocale, mode?: string): string | undefined {
  if (!mode) return undefined
  const t = getTranslator(locale)
  if (mode.includes('研讨') || mode.includes('deliberation')) return t('meeting.mode.deliberation')
  if (mode.includes('裁决') || mode.includes('decision')) return t('meeting.mode.decision')
  return mode.split('（')[0]?.trim() || mode
}

export type MeetingModeKind = 'decision' | 'deliberation'

export function meetingModeKind(
  modeKind?: string,
  mode?: string,
): MeetingModeKind | undefined {
  if (modeKind === 'decision' || modeKind === 'deliberation') return modeKind
  if (!mode) return undefined
  if (mode.includes('研讨') || mode.toLowerCase().includes('deliberation')) return 'deliberation'
  if (mode.includes('裁决') || mode.toLowerCase().includes('decision')) return 'decision'
  return undefined
}

export type MeetingFileCategory = 'overview' | 'deliverable' | 'process'

const MEETING_OVERVIEW_FILES = new Set(['MEETING.md', 'usage/summary.md'])

const MEETING_FILE_CATEGORY_ORDER: Record<MeetingFileCategory, number> = {
  overview: 0,
  deliverable: 1,
  process: 2,
}

export function meetingFileCategory(path: string): MeetingFileCategory {
  if (MEETING_OVERVIEW_FILES.has(path)) return 'overview'
  if (
    path.startsWith('artifacts/') ||
    path.startsWith('confirmation/') ||
    path === 'action-items.md'
  ) {
    return 'deliverable'
  }
  return 'process'
}

export function meetingFileCategoryLabel(locale: AppLocale, category: MeetingFileCategory): string {
  return getTranslator(locale)(`meeting.fileCategory.${category}`)
}

const STATIC_FILE_KEYS: Record<string, string> = {
  'MEETING.md': 'meeting.files.MEETING.md',
  'MINUTES.md': 'meeting.files.MINUTES.md',
  'action-items.md': 'meeting.files.action-items.md',
  'pre-meeting/perspectives.md': 'meeting.files.pre-meeting/perspectives.md',
  'artifacts/design-draft.md': 'meeting.files.artifacts/design-draft.md',
  'artifacts/open-questions.md': 'meeting.files.artifacts/open-questions.md',
  'usage/summary.md': 'meeting.files.usage/summary.md',
  'confirmation/brief.md': 'meeting.files.confirmation/brief.md',
  'moderator/executive-recap.md': 'meeting.files.moderator/executive-recap.md',
}

const MEETING_FILE_LABEL_PATTERNS: Array<{
  re: RegExp
  key: 'roundRecord' | 'roundSummary' | 'roundReadiness' | 'freeDialogue'
}> = [
  { re: /^rounds\/round-(\d+)\.md$/, key: 'roundRecord' },
  { re: /^moderator\/round-(\d+)-summary\.md$/, key: 'roundSummary' },
  { re: /^moderator\/round-(\d+)-readiness\.md$/, key: 'roundReadiness' },
  { re: /^free-dialogue\/after-round-(\d+)\.md$/, key: 'freeDialogue' },
]

function parseRoundFromPath(path: string): number | undefined {
  const m = path.match(/round-(\d+)/)
  if (!m) return undefined
  return parseInt(m[1], 10)
}

function meetingFileResolvedTitle(
  locale: AppLocale,
  path: string,
  modeKind?: MeetingModeKind,
): string | undefined {
  const t = getTranslator(locale)
  if (path === 'artifacts/minutes.md') {
    return modeKind === 'decision'
      ? t('meeting.files.minutesDecision')
      : t('meeting.files.minutesArchive')
  }
  const staticKey = STATIC_FILE_KEYS[path]
  if (staticKey) return t(staticKey)
  for (const { re, key } of MEETING_FILE_LABEL_PATTERNS) {
    const m = path.match(re)
    if (m) return t(`meeting.files.${key}`, { n: parseInt(m[1], 10) })
  }
  return undefined
}

function overviewSortRank(path: string): number {
  if (path === 'MEETING.md') return 0
  if (path === 'usage/summary.md') return 1
  return 2
}

function meetingFileSortBucket(path: string): number {
  if (path.startsWith('rounds/')) return 100 + (parseRoundFromPath(path) ?? 0)
  if (path.startsWith('free-dialogue/')) return 200 + (parseRoundFromPath(path) ?? 0)
  if (path.includes('-summary.md')) return 300 + (parseRoundFromPath(path) ?? 0)
  if (path.includes('-readiness.md')) return 400 + (parseRoundFromPath(path) ?? 0)
  if (path === 'moderator/executive-recap.md') return 450
  if (path.startsWith('confirmation/')) return 500
  return 900
}

export const MEETING_FILE_ORDER = [
  'MEETING.md',
  'MINUTES.md',
  'artifacts/design-draft.md',
  'artifacts/open-questions.md',
  'artifacts/minutes.md',
  'confirmation/brief.md',
  'pre-meeting/perspectives.md',
  'action-items.md',
  'usage/summary.md',
] as const

export function meetingFileLabel(
  locale: AppLocale,
  path: string,
  modeKind?: MeetingModeKind,
): string {
  return meetingFileResolvedTitle(locale, path, modeKind) ?? path.split('/').pop() ?? path
}

export function meetingFileCaption(
  locale: AppLocale,
  path: string,
  modeKind?: MeetingModeKind,
): string {
  const title = meetingFileLabel(locale, path, modeKind)
  const resolved = meetingFileResolvedTitle(locale, path, modeKind)
  if (!resolved) return path
  return `${title} · ${path}`
}

export function meetingFileHasTitle(
  locale: AppLocale,
  path: string,
  modeKind?: MeetingModeKind,
): boolean {
  return meetingFileResolvedTitle(locale, path, modeKind) !== undefined
}

export function sortMeetingFileNames(names: string[]): string[] {
  const order = new Map<string, number>(MEETING_FILE_ORDER.map((name, i) => [name, i]))
  return [...names].sort((a, b) => {
    const catA = meetingFileCategory(a)
    const catB = meetingFileCategory(b)
    if (catA !== catB) {
      return MEETING_FILE_CATEGORY_ORDER[catA] - MEETING_FILE_CATEGORY_ORDER[catB]
    }
    if (catA === 'overview') return overviewSortRank(a) - overviewSortRank(b)
    const ai = order.get(a) ?? meetingFileSortBucket(a)
    const bi = order.get(b) ?? meetingFileSortBucket(b)
    if (ai !== bi) return ai - bi
    return a.localeCompare(b)
  })
}

export function groupMeetingFileNames(
  names: string[],
): Record<MeetingFileCategory, string[]> {
  const sorted = sortMeetingFileNames(names)
  const groups: Record<MeetingFileCategory, string[]> = {
    overview: [],
    deliverable: [],
    process: [],
  }
  for (const name of sorted) {
    groups[meetingFileCategory(name)].push(name)
  }
  return groups
}

export function primaryDeliverablePath(modeKind?: MeetingModeKind): string {
  if (modeKind === 'decision') return 'artifacts/minutes.md'
  return 'artifacts/design-draft.md'
}

export function isPrimaryDeliverable(path: string, modeKind?: MeetingModeKind): boolean {
  return path === primaryDeliverablePath(modeKind)
}

export function defaultMeetingFileSelection(
  names: string[],
  modeKind?: MeetingModeKind,
): string {
  const groups = groupMeetingFileNames(names)
  const primary = primaryDeliverablePath(modeKind)
  if (groups.deliverable.includes(primary)) return primary
  return groups.deliverable[0] ?? groups.overview[0] ?? groups.process[0] ?? ''
}

const MEETING_FILE_DESCRIPTION_KEYS: Record<string, string> = {
  'MEETING.md': 'meeting.fileDesc.meetingMd',
  'MINUTES.md': 'meeting.fileDesc.minutesMd',
  'usage/summary.md': 'meeting.fileDesc.usageSummary',
  'artifacts/design-draft.md': 'meeting.fileDesc.designDraft',
  'artifacts/open-questions.md': 'meeting.fileDesc.openQuestions',
  'confirmation/brief.md': 'meeting.fileDesc.confirmationBrief',
  'action-items.md': 'meeting.fileDesc.actionItems',
  'pre-meeting/perspectives.md': 'meeting.fileDesc.perspectives',
  'moderator/executive-recap.md': 'meeting.fileDesc.executiveRecap',
}

const MEETING_FILE_DESCRIPTION_PATTERNS: Array<{
  re: RegExp
  key: 'roundRecord' | 'roundSummary' | 'roundReadiness' | 'freeDialogue'
}> = [
  { re: /^rounds\/round-(\d+)\.md$/, key: 'roundRecord' },
  { re: /^moderator\/round-(\d+)-summary\.md$/, key: 'roundSummary' },
  { re: /^moderator\/round-(\d+)-readiness\.md$/, key: 'roundReadiness' },
  { re: /^free-dialogue\/after-round-(\d+)\.md$/, key: 'freeDialogue' },
]

export function meetingFileDescription(
  locale: AppLocale,
  path: string,
  modeKind?: MeetingModeKind,
): string {
  const t = getTranslator(locale)
  if (path === 'artifacts/minutes.md') {
    return modeKind === 'decision'
      ? t('meeting.fileDesc.minutesDecision')
      : t('meeting.fileDesc.minutesArchive')
  }
  const descKey = MEETING_FILE_DESCRIPTION_KEYS[path]
  if (descKey) return t(descKey)
  for (const { re, key } of MEETING_FILE_DESCRIPTION_PATTERNS) {
    const match = path.match(re)
    if (match) return t(`meeting.fileDesc.${key}`, { n: parseInt(match[1], 10) })
  }
  const cat = meetingFileCategory(path)
  if (cat === 'overview') return t('meeting.fileDesc.categoryOverview')
  if (cat === 'deliverable') return t('meeting.fileDesc.categoryDeliverable')
  return t('meeting.fileDesc.categoryProcess')
}

/** @deprecated Use meetingFileCategoryLabel(locale, category) */
export const MEETING_FILE_CATEGORY_LABELS: Record<MeetingFileCategory, string> = {
  overview: '概览',
  deliverable: '交付物',
  process: '过程文档',
}
