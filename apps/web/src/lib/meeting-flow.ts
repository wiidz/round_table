import type { AppLocale } from '@/lib/locale'
import * as flow from '@/lib/i18n/meeting-flow'
import { meetingModeKind, type MeetingModeKind } from '@/lib/i18n/meeting-labels'
import type { MeetingDetail } from '@/types/meeting'

export type {
  MeetingFlow,
  MeetingFlowOutcome,
  MeetingFlowStep,
  MeetingFlowStepKind,
  MeetingFlowStepStatus,
} from '@/lib/i18n/meeting-flow'

export {
  parseConfirmationRequired,
  listDebateRoundNumbers,
  parseConfirmationCycle,
  parseConfirmationRejectionCount,
} from '@/lib/i18n/meeting-flow'

const fallbackLocale: AppLocale = 'zh'

export function isMeetingAbortedStatus(status?: string): boolean {
  const s = status?.trim() ?? ''
  return s === '已中断' || s === 'aborted' || s === 'Aborted'
}

export function isMeetingSuccessfullyCompletedStatus(status?: string): boolean {
  const s = status?.trim() ?? ''
  return (
    s === '已结束' ||
    s === 'Completed' ||
    s === '已归档' ||
    s === 'Archived' ||
    s === '共识达成' ||
    s === 'Consensus'
  )
}

export function isMeetingFinishedStatus(status?: string): boolean {
  return isMeetingSuccessfullyCompletedStatus(status) || isMeetingAbortedStatus(status)
}

export function isMeetingRunningStatus(status?: string): boolean {
  const s = status?.trim() ?? ''
  return (
    s === '进行中' ||
    s === 'Running' ||
    s === 'Principal 确认中' ||
    s === 'Confirmation' ||
    s === '已暂停' ||
    s === 'Paused' ||
    s === '准备中' ||
    s === 'Preparing'
  )
}

export function buildMeetingFlow(detail: MeetingDetail): flow.MeetingFlow {
  return flow.buildMeetingFlow(fallbackLocale, detail)
}

export function meetingFlowStepStatusLabel(status: flow.MeetingFlowStepStatus): string {
  return flow.meetingFlowStepStatusLabel(fallbackLocale, status)
}

export { meetingModeKind, type MeetingModeKind }
