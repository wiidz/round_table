import { describe, expect, it } from 'vitest'

import { buildMeetingFlow, parseConfirmationRequired } from '@/lib/meeting-flow'
import {
  listDebateRoundNumbers,
  parseConfirmationCycle,
  parseConfirmationRejectionCount,
} from '@/lib/i18n/meeting-flow'

import type { MeetingDetail } from '@/types/meeting'

describe('parseConfirmationRequired', () => {
  it('detects skip from MEETING.md table', () => {
    const md = '| 确认模式 | skip |'
    expect(parseConfirmationRequired(md)).toBe(false)
  })

  it('detects required from MEETING.md table', () => {
    const md = '| 确认模式 | 需要 Principal 确认 |'
    expect(parseConfirmationRequired(md)).toBe(true)
  })
})

describe('buildMeetingFlow', () => {
  it('builds decision flow with pre-meeting, one round, and closing', () => {
    const detail: MeetingDetail = {
      id: 'mtg-a',
      topic: '测试',
      status: '已结束',
      mode: '裁决型（decision）',
      mode_kind: 'decision',
      max_rounds: 1,
      free_dialogue: false,
      updated_at: '',
      files: {
        'MEETING.md': '| 确认模式 | skip |',
        'pre-meeting/perspectives.md': '# views',
        'rounds/round-001.md': '# round 1',
        'artifacts/minutes.md': '# done',
      },
    }

    const flow = buildMeetingFlow(detail)
    expect(flow.steps.map((s) => s.id)).toEqual([
      'pre-meeting',
      'round-1',
      'closing',
    ])
    expect(flow.steps.every((s) => s.status === 'completed')).toBe(true)
  })

  it('inserts free dialogue after round 1 when enabled', () => {
    const detail: MeetingDetail = {
      id: 'mtg-b',
      topic: '测试',
      status: 'Running',
      mode_kind: 'deliberation',
      max_rounds: 2,
      free_dialogue: true,
      updated_at: '',
      files: {
        'pre-meeting/perspectives.md': '# views',
        'rounds/round-001.md': '# r1',
      },
    }

    const flow = buildMeetingFlow(detail)
    expect(flow.steps.map((s) => s.id)).toEqual([
      'pre-meeting',
      'round-1',
      'free-dialogue',
      'round-2',
      'synthesis',
      'closing',
    ])
    expect(flow.steps.find((s) => s.id === 'pre-meeting')?.status).toBe('completed')
    expect(flow.steps.find((s) => s.id === 'round-1')?.status).toBe('completed')
    expect(flow.steps.find((s) => s.id === 'free-dialogue')?.status).toBe('active')
  })

  it('marks interrupt point and outcome for aborted meetings', () => {
    const detail: MeetingDetail = {
      id: 'mtg-abort',
      topic: '测试',
      status: '已中断',
      mode_kind: 'decision',
      max_rounds: 1,
      free_dialogue: false,
      updated_at: '',
      files: {
        'MEETING.md': '| 确认模式 | skip |',
        'pre-meeting/perspectives.md': '# views',
        'rounds/round-001.md': '# round 1',
      },
    }

    const flow = buildMeetingFlow(detail)
    expect(flow.outcome).toBe('aborted')
    expect(flow.interruptedStepId).toBe('closing')
    expect(flow.steps.find((s) => s.id === 'pre-meeting')?.status).toBe('completed')
    expect(flow.steps.find((s) => s.id === 'round-1')?.status).toBe('completed')
    expect(flow.steps.find((s) => s.id === 'closing')?.status).toBe('interrupted')
  })

  it('reports completed outcome when meeting finished normally', () => {
    const detail: MeetingDetail = {
      id: 'mtg-done',
      topic: '测试',
      status: '已结束',
      mode_kind: 'decision',
      max_rounds: 1,
      free_dialogue: false,
      updated_at: '',
      files: {
        'MEETING.md': '| 确认模式 | skip |',
        'pre-meeting/perspectives.md': '# views',
        'rounds/round-001.md': '# round 1',
        'artifacts/minutes.md': '# done',
      },
    }

    const flow = buildMeetingFlow(detail)
    expect(flow.outcome).toBe('completed')
    expect(flow.steps.every((s) => s.status === 'completed')).toBe(true)
  })

  it('omits unplayed rounds when meeting completed with bumped max_rounds', () => {
    const detail: MeetingDetail = {
      id: 'mtg-extra-cap',
      topic: '测试',
      status: '已结束',
      mode_kind: 'deliberation',
      max_rounds: 5,
      free_dialogue: true,
      updated_at: '',
      files: {
        'MEETING.md': '| 确认模式 | 需要 Principal 确认 |',
        'pre-meeting/perspectives.md': '# views',
        'rounds/round-001.md': '# r1',
        'rounds/round-002.md': '# r2',
        'rounds/round-003.md': '# r3',
        'rounds/round-004.md': '# r4',
        'free-dialogue/after-round-001.md': '# fd',
        'artifacts/design-draft.md': '# draft',
        'confirmation/brief.md': '# Confirmation Brief (cycle 3)\n\napproved',
        'artifacts/minutes.md': '# archive',
      },
    }

    const flow = buildMeetingFlow(detail)
    expect(flow.steps.map((s) => s.id)).toEqual([
      'pre-meeting',
      'round-1',
      'free-dialogue',
      'round-2',
      'round-3',
      'round-4',
      'synthesis',
      'confirmation-rejected-1',
      'confirmation-rejected-2',
      'confirmation',
      'closing',
    ])
    expect(flow.confirmationRejections).toBe(2)
    expect(flow.steps.every((s) => s.status === 'completed')).toBe(true)
  })
})

describe('confirmation rejection helpers', () => {
  it('parses confirmation cycle from brief header', () => {
    expect(parseConfirmationCycle('# Confirmation Brief (cycle 3)\n\nbody')).toBe(3)
    expect(parseConfirmationCycle('# other')).toBeUndefined()
  })

  it('counts rejections as cycle minus one when completed', () => {
    const files = {
      'confirmation/brief.md': '# Confirmation Brief (cycle 3)\n',
    }
    expect(parseConfirmationRejectionCount(files, true)).toBe(2)
    expect(parseConfirmationRejectionCount(files, false)).toBe(0)
  })

  it('lists debate round numbers from workspace files', () => {
    expect(
      listDebateRoundNumbers({
        'rounds/round-001.md': 'a',
        'rounds/round-004.md': 'b',
        'MINUTES.md': 'c',
      }),
    ).toEqual([1, 4])
  })
})
