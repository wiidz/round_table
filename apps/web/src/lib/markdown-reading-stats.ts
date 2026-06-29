/** 中文 Markdown 阅读速度估算（字/分钟） */
const CHARS_PER_MINUTE = 400

export function markdownCharCount(content: string): number {
  return [...content].length
}

export function markdownReadingMinutes(content: string): number {
  const chars = markdownCharCount(content)
  if (chars === 0) return 0
  return Math.max(1, Math.ceil(chars / CHARS_PER_MINUTE))
}

/** 侧栏 active 项：共 3,842 字 · 约 8 分钟 */
export function formatMarkdownReadingStats(content: string): string {
  const chars = markdownCharCount(content)
  if (chars === 0) {
    return '共 0 字'
  }
  const minutes = markdownReadingMinutes(content)
  return `共 ${chars.toLocaleString('zh-CN')} 字 · 约 ${minutes} 分钟`
}
