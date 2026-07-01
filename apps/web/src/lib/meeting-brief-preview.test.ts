import { describe, expect, it } from 'vitest'

import {
  extractDeliverableSummary,
  extractMarkdownSection,
  parseMeetingBriefPreview,
} from '@/lib/meeting-brief-preview'

import type { MeetingDetail } from '@/types/meeting'

const SAMPLE_MEETING_MD = `# 会议简报

## 会议主题

Auth 拆分

## 会议目标

形成可执行共识。

## 讨论范围

方案取舍

## 不在范围

排期

## 完成标准

每议题 1 条结论
`

describe('extractMarkdownSection', () => {
  it('extracts section body until next heading', () => {
    expect(extractMarkdownSection(SAMPLE_MEETING_MD, '会议目标')).toBe('形成可执行共识。')
  })
})

describe('extractDeliverableSummary', () => {
  it('prefers conclusion heading', () => {
    const md = `# Minutes\n\n## 结论\n\n批准上线，补充 QA 角色。`
    expect(extractDeliverableSummary(md)).toBe('批准上线，补充 QA 角色。')
  })
})

describe('parseMeetingBriefPreview', () => {
  it('merges meeting doc and deliverable summary', () => {
    const detail: MeetingDetail = {
      id: 'mtg-a',
      topic: 'fallback',
      status: '已结束',
      mode_kind: 'decision',
      updated_at: '',
      files: {
        'MEETING.md': SAMPLE_MEETING_MD,
        'artifacts/minutes.md': '## 结论\n\n批准拆分方案。',
      },
    }

    const brief = parseMeetingBriefPreview(detail, 'decision')
    expect(brief.topic).toBe('Auth 拆分')
    expect(brief.goal).toBe('形成可执行共识。')
    expect(brief.inScope).toBe('方案取舍')
    expect(brief.conclusion).toBe('批准拆分方案。')
    expect(brief.conclusionSource).toBe('artifacts/minutes.md')
  })
})
