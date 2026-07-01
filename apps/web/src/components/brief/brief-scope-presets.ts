/** 讨论边界字段的通用填写引导（领域无关） */
export type { BriefScopePreset } from '@/lib/i18n/brief-scope-presets'
export { getBriefScopePresets } from '@/lib/i18n/brief-scope-presets'

export function applyScopePreset(current: string, snippet: string): string {
  const next = snippet.trim()
  if (!next) return current
  const trimmed = current.trim()
  if (!trimmed) return next
  if (trimmed.includes(next)) return trimmed
  return `${trimmed}\n${next}`
}

export function scopePresetApplied(current: string, snippet: string): boolean {
  const next = snippet.trim()
  if (!next) return false
  return current.trim().includes(next)
}
