import { describe, expect, it } from 'vitest'

import { condenseMessage, firstSentence, stripMarkdownForSummary } from '@/lib/condense-message'

describe('stripMarkdownForSummary', () => {
  it('removes heading and list markers', () => {
    expect(stripMarkdownForSummary('## 标题\n- **要点** one')).toBe('标题 要点 one')
  })
})

describe('firstSentence', () => {
  it('splits on CJK period', () => {
    expect(firstSentence('第一句。第二句')).toBe('第一句。')
  })

  it('splits on Western period', () => {
    expect(firstSentence('Hello world. More text')).toBe('Hello world.')
  })
})

describe('condenseMessage', () => {
  it('returns full text when ≤80 chars', () => {
    const short = '这是一段较短的说明。'
    const { summary, truncated } = condenseMessage(short)
    expect(summary).toBe(short)
    expect(truncated).toBe(false)
  })

  it('truncates long markdown to first sentence or char limit', () => {
    const long =
      '## 方案\n\n' +
      '这是第一句很长的讨论内容，需要被摘要。'.repeat(3) +
      '这是第二句不应出现在摘要里除非第一句太短。'.repeat(2)
    const { summary, truncated } = condenseMessage(long)
    expect(summary.length).toBeLessThan(stripMarkdownForSummary(long).length)
    expect(truncated).toBe(true)
  })
})
