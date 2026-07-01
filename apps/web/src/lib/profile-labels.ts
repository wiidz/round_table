import type { AppLocale } from '@/lib/locale'
import * as profile from '@/lib/i18n/profile-labels'

export {
  PARTICIPANT_STANDARD_FILES,
  PRINCIPAL_STANDARD_FILES,
  profileFileHasTitle,
} from '@/lib/i18n/profile-labels'

const fallbackLocale: AppLocale = 'zh'

export const PROFILE_FILE_LABELS: Record<string, string> = {
  'USER.md': '偏好画像',
  'SOUL.md': '人格',
  'AGENTS.md': '行为规则',
  'TOOLS.md': '工具约定',
}

export function profileFileCaption(filename: string): string {
  return profile.profileFileCaption(fallbackLocale, filename)
}

export const PARTICIPANT_FILE_HINTS: Record<string, string> = {
  'SOUL.md': '人格、语气与边界（ADR-0010 标准档案）',
  'AGENTS.md': 'Meeting 内行为规则与发言方式',
  'TOOLS.md': '工具与环境约定',
}

export const PRINCIPAL_FILE_HINTS: Record<string, string> = {
  'USER.md':
    'Principal 偏好与背景（语言、Confirmation 审阅习惯、行业约束）。Moderator 服务你的长期设定，不是单次会议议题。',
}

export function participantFileHints(locale: AppLocale): Record<string, string> {
  const hints: Record<string, string> = {}
  for (const file of profile.PARTICIPANT_STANDARD_FILES) {
    hints[file] = profile.participantFileHint(locale, file)
  }
  return hints
}

export function principalFileHints(locale: AppLocale): Record<string, string> {
  const hints: Record<string, string> = {}
  for (const file of profile.PRINCIPAL_STANDARD_FILES) {
    hints[file] = profile.principalFileHint(locale, file)
  }
  return hints
}
