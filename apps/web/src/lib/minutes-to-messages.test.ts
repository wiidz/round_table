import { describe, expect, it } from 'vitest'

import {
  parseMinutesMarkdown,
  workspaceTranscriptMessages,
} from '@/lib/minutes-to-messages'

const SAMPLE_MINUTES = `# Minutes

**Topic:** 测试会议

## Pre-meeting (Round 0)

Pre-meeting perspectives

- **design** (design): 初始观点 A
- **player** (player): 初始观点 B

## Round 1

Round 1

- **design** (design): 辩论观点 A _[abstain]_
- **player** (player): 辩论观点 B _[agree]_

### Moderator summary

## Round 1 研讨摘要

本轮达成共识。

## Token usage

Total tokens: **100**
`

describe('parseMinutesMarkdown', () => {
  it('parses participant bullets with stance stripped and sequential turns', () => {
    const messages = parseMinutesMarkdown(SAMPLE_MINUTES, { meetingId: 'mtg-1' })
    expect(messages).toHaveLength(5)
    expect(messages[0]).toMatchObject({
      role: 'participant',
      authorId: 'design',
      content: '初始观点 A',
      turn: 1,
    })
    expect(messages[3]).toMatchObject({
      role: 'participant',
      authorId: 'player',
      content: '辩论观点 B',
      turn: 4,
    })
    expect(messages[4]).toMatchObject({
      role: 'moderator',
      turn: 5,
    })
    expect(messages[4]?.content).toContain('本轮达成共识')
  })

  it('falls back to round files when MINUTES.md is missing', () => {
    const messages = workspaceTranscriptMessages(
      {
        'rounds/round-001.md': '# Round 1\n\n- **dev** (dev): hello _[agree]_',
      },
      'mtg-2',
    )
    expect(messages).toHaveLength(1)
    expect(messages[0]).toMatchObject({ authorId: 'dev', content: 'hello', turn: 1 })
  })
})
