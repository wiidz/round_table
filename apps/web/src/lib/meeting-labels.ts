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

/** Workspace markdown file → readable label */
export const MEETING_FILE_LABELS: Record<string, string> = {
  'MEETING.md': '会议简报',
  'MINUTES.md': '会议纪要',
  'action-items.md': '行动项',
  'pre-meeting/perspectives.md': '会前观点',
  'artifacts/design-draft.md': '方案草案',
  'artifacts/open-questions.md': '待决问题',
  'artifacts/minutes.md': '结论纪要',
  'usage/summary.md': 'Token 用量',
}

export const MEETING_FILE_ORDER = [
  'MEETING.md',
  'MINUTES.md',
  'artifacts/design-draft.md',
  'artifacts/open-questions.md',
  'artifacts/minutes.md',
  'pre-meeting/perspectives.md',
  'action-items.md',
  'usage/summary.md',
] as const

export function meetingFileLabel(path: string): string {
  return MEETING_FILE_LABELS[path] ?? path
}

export function sortMeetingFileNames(names: string[]): string[] {
  const order = new Map<string, number>(
    MEETING_FILE_ORDER.map((name, i) => [name, i]),
  )
  return [...names].sort((a, b) => {
    const ai = order.get(a) ?? 999
    const bi = order.get(b) ?? 999
    if (ai !== bi) return ai - bi
    return a.localeCompare(b)
  })
}
