import { describe, expect, it } from 'vitest'

import { meetingFileLabel } from '@/lib/i18n/meeting-labels'

describe('meetingFileLabel', () => {
  it('resolves paths with dots and slashes without returning i18n keys', () => {
    expect(meetingFileLabel('zh', 'MEETING.md')).toBe('会议简报')
    expect(meetingFileLabel('zh', 'artifacts/design-draft.md')).toBe('方案草案')
    expect(meetingFileLabel('zh', 'confirmation/brief.md')).toBe('确认呈报清单')
    expect(meetingFileLabel('zh', 'usage/summary.md')).toBe('Token 用量')
  })

  it('resolves round pattern labels', () => {
    expect(meetingFileLabel('zh', 'rounds/round-003.md')).toBe('第 3 轮研讨记录')
  })
})
