/**
 * Backward-compatible re-exports defaulting to zh locale.
 * Prefer useI18n() in React components.
 */
import type { AppLocale } from '@/lib/locale'
import * as labels from '@/lib/i18n/meeting-labels'

export type {
  MeetingFileCategory,
  MeetingModeKind,
  MeetingStatusTone,
} from '@/lib/i18n/meeting-labels'

export {
  MEETING_FILE_ORDER,
  defaultMeetingFileSelection,
  groupMeetingFileNames,
  isPrimaryDeliverable,
  meetingFileCategory,
  meetingModeKind,
  meetingStatusTone,
  primaryDeliverablePath,
  sortMeetingFileNames,
} from '@/lib/i18n/meeting-labels'

const fallbackLocale: AppLocale = 'zh'

export const MEETING_STATUS_LABELS: Record<string, string> = {
  Preparing: '准备中',
  Running: '进行中',
  Paused: '已暂停',
  Consensus: '共识达成',
  Confirmation: '委托人确认中',
  Completed: '已结束',
  Archived: '已归档',
  aborted: '已中断',
  Aborted: '已中断',
  已中断: '已中断',
}

export const MEETING_FILE_LABELS: Record<string, string> = {
  'MEETING.md': '会议简报',
  'MINUTES.md': '会议纪要',
  'action-items.md': '行动项',
  'pre-meeting/perspectives.md': '会前观点',
  'artifacts/design-draft.md': '方案草案',
  'artifacts/open-questions.md': '待决问题',
  'usage/summary.md': 'Token 用量',
  'confirmation/brief.md': '确认呈报清单',
  'moderator/executive-recap.md': '会议回顾',
}

export const MEETING_FILE_CATEGORY_LABELS: Record<
  labels.MeetingFileCategory,
  string
> = {
  overview: '概览',
  deliverable: '交付物',
  process: '过程文档',
}

export function meetingStatusLabel(status: string): string {
  return labels.meetingStatusLabel(fallbackLocale, status)
}

export function meetingModeShort(mode?: string): string | undefined {
  return labels.meetingModeShort(fallbackLocale, mode)
}

export function meetingFileCategoryLabel(category: labels.MeetingFileCategory): string {
  return labels.meetingFileCategoryLabel(fallbackLocale, category)
}

export function meetingFileLabel(path: string, modeKind?: labels.MeetingModeKind): string {
  return labels.meetingFileLabel(fallbackLocale, path, modeKind)
}

export function meetingFileCaption(path: string, modeKind?: labels.MeetingModeKind): string {
  return labels.meetingFileCaption(fallbackLocale, path, modeKind)
}

export function meetingFileHasTitle(path: string, modeKind?: labels.MeetingModeKind): boolean {
  return labels.meetingFileHasTitle(fallbackLocale, path, modeKind)
}

export function meetingFileDescription(
  path: string,
  modeKind?: labels.MeetingModeKind,
): string {
  return labels.meetingFileDescription(fallbackLocale, path, modeKind)
}
