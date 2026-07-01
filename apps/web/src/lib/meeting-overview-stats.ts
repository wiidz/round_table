import { markdownCharCount, markdownReadingMinutes } from '@/lib/markdown-reading-stats'
import {
  meetingFileLabel,
  primaryDeliverablePath,
  type MeetingModeKind,
} from '@/lib/meeting-labels'
import { parseParticipantsFromMeetingMd } from '@/lib/meeting-participants'

import type { MeetingDetail } from '@/types/meeting'

export function formatTokenCount(value: number): string {
  if (value >= 1_000_000) {
    const compact = value / 1_000_000
    return `${compact >= 10 ? Math.round(compact) : compact.toFixed(1)}M`
  }
  if (value >= 10_000) {
    const compact = value / 1_000
    return `${compact >= 100 ? Math.round(compact) : compact.toFixed(1)}k`
  }
  return value.toLocaleString('zh-CN')
}

export interface MeetingOverviewStats {
  deliverable: {
    available: boolean
    title: string
    charCount: number
    readingMinutes: number
  }
  usage: {
    totalTokens: number
    llmCallCount: number
  }
  experts: {
    count: number
  }
  rounds: {
    maxRounds: number
    freeDialogueQuestions: number
  }
}

function extractMeetingTableCell(doc: string, rowLabel: string): string {
  const row = doc.match(new RegExp(`${rowLabel}\\s*\\|\\s*([^\\n|]+)`))
  return row?.[1]?.trim().replace(/^`|`$/g, '') ?? ''
}

function parseMaxRoundsFromMeetingMd(meetingMd: string): number {
  const cell = extractMeetingTableCell(meetingMd, '辩论轮次上限')
  if (!cell) return 0
  const match = cell.match(/(\d+)/)
  return match?.[1] ? Number.parseInt(match[1], 10) : 0
}

function parseFreeDialogueMaxQuestions(
  meetingMd: string,
  freeDialogueEnabled: boolean,
): number {
  if (!freeDialogueEnabled) return 0
  const cell = extractMeetingTableCell(meetingMd, 'Round 1 后自由对话')
  const match = cell.match(/每人最多\s*(\d+)\s*[轮问]/)
  if (match?.[1]) return Number.parseInt(match[1], 10)
  if (cell && !/^0(\s|$)/.test(cell)) return 1
  return 0
}

/** 辩论轮 + 自由对话问数，如 `3+1`；无自由对话时为 `3`。 */
export function formatMeetingRoundsValue(
  maxRounds: number,
  freeDialogueQuestions: number,
): string {
  if (maxRounds <= 0) return '—'
  if (freeDialogueQuestions > 0) {
    return `${maxRounds}+${freeDialogueQuestions}`
  }
  return String(maxRounds)
}

export function formatMeetingRoundsHint(
  maxRounds: number,
  freeDialogueQuestions: number,
): string {
  if (maxRounds <= 0) return '尚未配置轮次'
  if (freeDialogueQuestions > 0) {
    return `${maxRounds} 轮辩论 + ${freeDialogueQuestions} 问自由对话`
  }
  return `${maxRounds} 轮辩论`
}

export function buildMeetingOverviewStats(
  detail: MeetingDetail,
  modeKind?: MeetingModeKind,
): MeetingOverviewStats {
  const meetingMd = detail.files?.['MEETING.md'] ?? ''
  const deliverablePath = primaryDeliverablePath(modeKind)
  const deliverableContent = detail.files?.[deliverablePath]?.trim() ?? ''
  const parsedExperts = parseParticipantsFromMeetingMd(meetingMd)
  const expertCount =
    parsedExperts.length > 0 ? parsedExperts.length : (detail.participant_count ?? 0)
  const maxRounds = detail.max_rounds ?? parseMaxRoundsFromMeetingMd(meetingMd)
  const freeDialogueQuestions = parseFreeDialogueMaxQuestions(
    meetingMd,
    detail.free_dialogue ?? false,
  )

  return {
    deliverable: {
      available: deliverableContent.length > 0,
      title: meetingFileLabel(deliverablePath, modeKind),
      charCount: markdownCharCount(deliverableContent),
      readingMinutes: markdownReadingMinutes(deliverableContent),
    },
    usage: {
      totalTokens: detail.total_tokens ?? 0,
      llmCallCount: detail.llm_call_count ?? 0,
    },
    experts: {
      count: expertCount,
    },
    rounds: {
      maxRounds,
      freeDialogueQuestions,
    },
  }
}
