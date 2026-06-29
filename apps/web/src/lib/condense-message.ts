/** Strip common Markdown decoration for summary text only. */
export function stripMarkdownForSummary(content: string): string {
  return content
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/\*\*|__/g, '')
    .replace(/`/g, '')
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/^\s*\d+\.\s+/gm, '')
    .replace(/\s+/g, ' ')
    .trim()
}

/** First sentence by CJK/Western punctuation or first line. */
export function firstSentence(text: string): string {
  const trimmed = text.trim()
  if (!trimmed) return ''

  const lineBreak = trimmed.indexOf('\n')
  if (lineBreak >= 0 && lineBreak < 80) {
    return trimmed.slice(0, lineBreak).trim()
  }

  const match = trimmed.match(/^[\s\S]*?[。！？.!?]/)
  if (match) {
    return match[0].trim()
  }
  return trimmed
}

function charCount(text: string): number {
  return [...text].length
}

function truncateChars(text: string, max: number): string {
  const chars = [...text]
  if (chars.length <= max) return text
  return `${chars.slice(0, max).join('')}…`
}

export interface CondensedMessage {
  summary: string
  /** True when summary is shorter than plain source (Drawer warranted). */
  truncated: boolean
}

/** ADR-0013 §4: first sentence or ~60 chars; skip truncation when ≤80 chars. */
export function condenseMessage(content: string, maxChars = 60): CondensedMessage {
  const plain = stripMarkdownForSummary(content)
  if (!plain) {
    return { summary: '', truncated: false }
  }

  const total = charCount(plain)
  if (total <= 80) {
    return { summary: plain, truncated: false }
  }

  const sentence = firstSentence(plain)
  const sentenceLen = charCount(sentence)

  if (sentenceLen >= 8 && sentenceLen <= maxChars) {
    return { summary: sentence, truncated: sentenceLen < total }
  }

  if (total <= maxChars) {
    return { summary: plain, truncated: false }
  }

  return { summary: truncateChars(plain, maxChars), truncated: true }
}
