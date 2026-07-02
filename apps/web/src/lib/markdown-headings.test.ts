import { describe, expect, it } from 'vitest'

import { extractMarkdownHeadings } from '@/lib/markdown-headings'

describe('extractMarkdownHeadings', () => {
  it('extracts a single top-level heading (pre-meeting style)', () => {
    const headings = extractMarkdownHeadings('# Pre-meeting (Round 0)\n\nBody text')
    expect(headings).toEqual([
      { level: 1, text: 'Pre-meeting (Round 0)', id: 'md-h-0' },
    ])
  })

  it('skips headings inside fenced code blocks', () => {
    const headings = extractMarkdownHeadings(
      '# Real\n\n```md\n# Not a heading\n```\n\n## Section',
    )
    expect(headings.map((h) => h.text)).toEqual(['Real', 'Section'])
  })
})
