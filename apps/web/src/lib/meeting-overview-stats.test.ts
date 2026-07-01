import { describe, expect, it } from 'vitest'

import {
  buildMeetingOverviewStats,
  formatMeetingRoundsValue,
  formatTokenCount,
} from '@/lib/meeting-overview-stats'

import type { MeetingDetail } from '@/types/meeting'

describe('formatTokenCount', () => {
  it('formats large values compactly', () => {
    expect(formatTokenCount(11389)).toBe('11.4k')
    expect(formatTokenCount(1_500_000)).toBe('1.5M')
  })
})

describe('formatMeetingRoundsValue', () => {
  it('uses plus format for debate and free dialogue', () => {
    expect(formatMeetingRoundsValue(3, 1)).toBe('3+1')
    expect(formatMeetingRoundsValue(3, 0)).toBe('3')
    expect(formatMeetingRoundsValue(0, 1)).toBe('—')
  })
})

describe('buildMeetingOverviewStats', () => {
  it('derives deliverable, usage, and expert stats', () => {
    const detail: MeetingDetail = {
      id: 'mtg-test',
      topic: '测试',
      status: '已结束',
      mode_kind: 'decision',
      participant_count: 3,
      max_rounds: 3,
      free_dialogue: true,
      total_tokens: 11389,
      llm_call_count: 10,
      updated_at: '2026-01-01',
      files: {
        'MEETING.md': `## 配置\n| 辩论轮次上限 | 3（不含 Pre-meeting Round 0） |\n| Round 1 后自由对话 | 每人最多 1 问 |\n## 参会人员\n| 参会者 | 角色 |\n| --- | --- |\n| design | 设计 |\n| dev | 开发 |`,
        'artifacts/minutes.md': '# 纪要\n\n'.padEnd(401, '字'),
      },
    }

    const stats = buildMeetingOverviewStats(detail, 'decision')

    expect(stats.deliverable.available).toBe(true)
    expect(stats.deliverable.charCount).toBeGreaterThan(0)
    expect(stats.deliverable.readingMinutes).toBeGreaterThan(0)
    expect(stats.usage.totalTokens).toBe(11389)
    expect(stats.usage.llmCallCount).toBe(10)
    expect(stats.experts.count).toBe(2)
    expect(stats.rounds.maxRounds).toBe(3)
    expect(stats.rounds.freeDialogueQuestions).toBe(1)
    expect(formatMeetingRoundsValue(stats.rounds.maxRounds, stats.rounds.freeDialogueQuestions)).toBe(
      '3+1',
    )
  })
})
