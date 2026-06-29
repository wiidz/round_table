export const PROFILE_FILE_LABELS: Record<string, string> = {
  'USER.md': '偏好画像',
  'SOUL.md': '人格',
  'AGENTS.md': '行为规则',
  'TOOLS.md': '工具约定',
}

/** Sidebar / header: 人格 · SOUL.md */
export function profileFileCaption(filename: string): string {
  const title = PROFILE_FILE_LABELS[filename]
  if (!title) return filename
  return `${title} · ${filename}`
}

export function profileFileHasTitle(filename: string): boolean {
  return filename in PROFILE_FILE_LABELS
}

/** ADR-0010 规定的 Participant 标准档案三件套 */
export const PARTICIPANT_STANDARD_FILES = ['SOUL.md', 'AGENTS.md', 'TOOLS.md'] as const

export const PARTICIPANT_FILE_HINTS: Record<string, string> = {
  'SOUL.md': '人格、语气与边界（ADR-0010 标准档案）',
  'AGENTS.md': 'Meeting 内行为规则与发言方式',
  'TOOLS.md': '工具与环境约定',
}

export const PRINCIPAL_FILE_HINTS: Record<string, string> = {
  'USER.md': 'Principal 偏好与背景（ADR-0010 标准档案）',
  'SOUL.md': '可选：人格与语气',
  'AGENTS.md': '可选：Meeting 内行为规则',
  'TOOLS.md': '可选：工具与环境约定',
}
