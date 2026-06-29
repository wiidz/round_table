/** MEETING.md table / domain status → display label */
export const MEETING_STATUS_LABELS: Record<string, string> = {
  Preparing: '准备中',
  Running: '进行中',
  Paused: '已暂停',
  Consensus: '共识达成',
  Confirmation: 'Principal 确认中',
  Completed: '已结束',
  Archived: '已归档',
}

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
  已归档: 'neutral',
  Archived: 'neutral',
}

export function meetingStatusLabel(status: string): string {
  if (!status) return '未知'
  return MEETING_STATUS_LABELS[status] ?? status
}

export function meetingStatusTone(status: string): MeetingStatusTone {
  if (!status) return 'neutral'
  return STATUS_TONES[status] ?? STATUS_TONES[meetingStatusLabel(status)] ?? 'neutral'
}

/** Short mode label from MEETING.md 会议模式 row */
export function meetingModeShort(mode?: string): string | undefined {
  if (!mode) return undefined
  if (mode.includes('研讨') || mode.includes('deliberation')) return '研讨型'
  if (mode.includes('裁决') || mode.includes('decision')) return '裁决型'
  return mode.split('（')[0]?.trim() || mode
}

export type MeetingModeKind = 'decision' | 'deliberation'

export function meetingModeKind(
  modeKind?: string,
  mode?: string,
): MeetingModeKind | undefined {
  if (modeKind === 'decision' || modeKind === 'deliberation') {
    return modeKind
  }
  if (!mode) return undefined
  if (mode.includes('研讨') || mode.toLowerCase().includes('deliberation')) {
    return 'deliberation'
  }
  if (mode.includes('裁决') || mode.toLowerCase().includes('decision')) {
    return 'decision'
  }
  return undefined
}

/** Workspace markdown file → readable label (Chinese) */
export const MEETING_FILE_LABELS: Record<string, string> = {
  'MEETING.md': '会议简报',
  'MINUTES.md': '会议纪要',
  'action-items.md': '行动项',
  'pre-meeting/perspectives.md': '会前观点',
  'artifacts/design-draft.md': '方案草案',
  'artifacts/open-questions.md': '待决问题',
  'usage/summary.md': 'Token 用量',
  'confirmation/brief.md': '确认呈报清单',
  'moderator/executive-recap.md': '会议回顾',
}

export type MeetingFileCategory = 'overview' | 'deliverable' | 'process'

const MEETING_OVERVIEW_FILES = new Set(['MEETING.md', 'usage/summary.md'])

const MEETING_FILE_CATEGORY_ORDER: Record<MeetingFileCategory, number> = {
  overview: 0,
  deliverable: 1,
  process: 2,
}

/** 概览（简报 / 用量） vs 交付物 vs 过程文档 */
export function meetingFileCategory(path: string): MeetingFileCategory {
  if (MEETING_OVERVIEW_FILES.has(path)) {
    return 'overview'
  }
  if (
    path.startsWith('artifacts/') ||
    path.startsWith('confirmation/') ||
    path === 'action-items.md'
  ) {
    return 'deliverable'
  }
  return 'process'
}

export const MEETING_FILE_CATEGORY_LABELS: Record<MeetingFileCategory, string> = {
  overview: '概览',
  deliverable: '交付物',
  process: '过程文档',
}

/** Dynamic workspace paths (round number from filename) */
const MEETING_FILE_LABEL_PATTERNS: Array<{
  re: RegExp
  label: (round: number) => string
}> = [
  { re: /^rounds\/round-(\d+)\.md$/, label: (n) => `第 ${n} 轮研讨记录` },
  { re: /^moderator\/round-(\d+)-summary\.md$/, label: (n) => `第 ${n} 轮摘要` },
  { re: /^moderator\/round-(\d+)-readiness\.md$/, label: (n) => `第 ${n} 轮研讨就绪` },
  { re: /^free-dialogue\/after-round-(\d+)\.md$/, label: (n) => `第 ${n} 轮后自由问答` },
]

function parseRoundFromPath(path: string): number | undefined {
  const m = path.match(/round-(\d+)/)
  if (!m) return undefined
  return parseInt(m[1], 10)
}

function meetingFileResolvedTitle(
  path: string,
  modeKind?: MeetingModeKind,
): string | undefined {
  if (path === 'artifacts/minutes.md') {
    return modeKind === 'decision' ? '结论纪要' : '归档纪要'
  }
  if (path in MEETING_FILE_LABELS) {
    return MEETING_FILE_LABELS[path]
  }
  for (const { re, label } of MEETING_FILE_LABEL_PATTERNS) {
    const m = path.match(re)
    if (m) {
      return label(parseInt(m[1], 10))
    }
  }
  return undefined
}

function overviewSortRank(path: string): number {
  if (path === 'MEETING.md') return 0
  if (path === 'usage/summary.md') return 1
  return 2
}

/** Sort bucket for paths not in MEETING_FILE_ORDER (lower = earlier in sidebar) */
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

export function meetingFileLabel(path: string, modeKind?: MeetingModeKind): string {
  return meetingFileResolvedTitle(path, modeKind) ?? path.split('/').pop() ?? path
}

/** Sidebar / header: 会议纪要 · MINUTES.md */
export function meetingFileCaption(path: string, modeKind?: MeetingModeKind): string {
  const title = meetingFileLabel(path, modeKind)
  const resolved = meetingFileResolvedTitle(path, modeKind)
  if (!resolved) return path
  return `${title} · ${path}`
}

/** Known file has a Chinese title distinct from path */
export function meetingFileHasTitle(path: string, modeKind?: MeetingModeKind): boolean {
  return meetingFileResolvedTitle(path, modeKind) !== undefined
}

export function sortMeetingFileNames(names: string[]): string[] {
  const order = new Map<string, number>(
    MEETING_FILE_ORDER.map((name, i) => [name, i]),
  )
  return [...names].sort((a, b) => {
    const catA = meetingFileCategory(a)
    const catB = meetingFileCategory(b)
    if (catA !== catB) {
      return MEETING_FILE_CATEGORY_ORDER[catA] - MEETING_FILE_CATEGORY_ORDER[catB]
    }
    if (catA === 'overview') {
      return overviewSortRank(a) - overviewSortRank(b)
    }
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

export function defaultMeetingFileSelection(
  names: string[],
  modeKind?: MeetingModeKind,
): string {
  const groups = groupMeetingFileNames(names)
  const primary = primaryDeliverablePath(modeKind)
  if (groups.deliverable.includes(primary)) {
    return primary
  }
  return (
    groups.deliverable[0] ??
    groups.overview[0] ??
    groups.process[0] ??
    ''
  )
}

/** Principal-facing headline artifact per meeting mode */
export function primaryDeliverablePath(modeKind?: MeetingModeKind): string {
  if (modeKind === 'decision') {
    return 'artifacts/minutes.md'
  }
  return 'artifacts/design-draft.md'
}

export function isPrimaryDeliverable(
  path: string,
  modeKind?: MeetingModeKind,
): boolean {
  return path === primaryDeliverablePath(modeKind)
}

type MeetingFileDescription =
  | string
  | ((modeKind?: MeetingModeKind) => string)

const MEETING_FILE_DESCRIPTIONS: Record<string, MeetingFileDescription> = {
  'MEETING.md':
    '议题、议程、参会者与会议状态快照；Participant 讨论的 bootstrap 上下文。',
  'MINUTES.md':
    '会议进行中由 Engine 持续更新的结构化纪要；与结束时写入的 artifacts/minutes.md 不同。',
  'usage/summary.md': '本次会议 LLM Token 用量汇总。',
  'artifacts/design-draft.md':
    '研讨型主交付物：Moderator 综合全场合议记录形成的方案草案。',
  'artifacts/open-questions.md': '研讨过程中尚未收敛的待决问题清单。',
  'artifacts/minutes.md': (modeKind) =>
    modeKind === 'decision'
      ? '裁决型主交付物：会议结束时的结论纪要快照。'
      : '会议结束时的纪要归档副本；过程稿见 MINUTES.md。',
  'confirmation/brief.md': 'Principal 确认阶段使用的呈报清单。',
  'action-items.md': '会后需跟进的行动项清单。',
  'pre-meeting/perspectives.md':
    'Round 0 各 Participant 独立撰写的会前观点。',
  'moderator/executive-recap.md': 'Moderator 对整场会议的 Executive Recap 回顾。',
}

const MEETING_FILE_DESCRIPTION_PATTERNS: Array<{
  re: RegExp
  describe: (round: number) => string
}> = [
  {
    re: /^rounds\/round-(\d+)\.md$/,
    describe: (n) => `第 ${n} 轮辩论发言与立场记录。`,
  },
  {
    re: /^moderator\/round-(\d+)-summary\.md$/,
    describe: (n) => `第 ${n} 轮结束后 Moderator 提炼的轮次摘要。`,
  },
  {
    re: /^moderator\/round-(\d+)-readiness\.md$/,
    describe: (n) => `第 ${n} 轮合成就绪评估。`,
  },
  {
    re: /^free-dialogue\/after-round-(\d+)\.md$/,
    describe: (n) => `第 ${n} 轮后的自由问答记录。`,
  },
]

const MEETING_FILE_CATEGORY_DESCRIPTIONS: Record<MeetingFileCategory, string> = {
  overview: '会议概览信息。',
  deliverable: '本场会议交付产出。',
  process: '会议过程记录。',
}

/** One-line purpose for the active workspace file (reader panel). */
export function meetingFileDescription(
  path: string,
  modeKind?: MeetingModeKind,
): string {
  const entry = MEETING_FILE_DESCRIPTIONS[path]
  if (typeof entry === 'function') {
    return entry(modeKind)
  }
  if (typeof entry === 'string') {
    return entry
  }
  for (const { re, describe } of MEETING_FILE_DESCRIPTION_PATTERNS) {
    const match = path.match(re)
    if (match) {
      return describe(parseInt(match[1], 10))
    }
  }
  return MEETING_FILE_CATEGORY_DESCRIPTIONS[meetingFileCategory(path)]
}
