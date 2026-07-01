/** 讨论边界字段的通用填写引导（领域无关） */
export interface BriefScopePreset {
  label: string
  value: string
}

export const BRIEF_SCOPE_PRESETS = {
  inScope: [
    { label: 'Topic 方案', value: '围绕本次 Topic 的方案、选项与取舍' },
    { label: '议程相关', value: '与议程项直接相关的结论与依据' },
    { label: '风险假设', value: '主要风险、关键假设与待验证点' },
  ],
  outOfScope: [
    { label: '排期人力', value: '实施排期与人力分配' },
    { label: '延伸话题', value: '未列入议程的延伸话题' },
    { label: '长期规划', value: '需另行立项的长期规划' },
  ],
  doneCriteria: [
    { label: '可执行结论', value: '每议题至少 1 条可执行结论' },
    { label: 'Principal 确认', value: 'Principal 确认共识摘要' },
    { label: '行动项', value: '明确后续行动项与负责人（或待指派）' },
  ],
} satisfies Record<string, BriefScopePreset[]>

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
