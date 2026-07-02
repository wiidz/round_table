import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'
import {
  isMeetingAbortedStatus,
  isMeetingRunningStatus,
  isMeetingSuccessfullyCompletedStatus,
} from '@/lib/meeting-flow'
import type { MeetingDetail } from '@/types/meeting'

export type MeetingFlowStepKind =
  | 'pre-meeting'
  | 'debate-round'
  | 'free-dialogue'
  | 'synthesis'
  | 'confirmation'
  | 'closing'

export type MeetingFlowStepStatus =
  | 'pending'
  | 'active'
  | 'completed'
  | 'skipped'
  | 'interrupted'

export type MeetingFlowOutcome = 'running' | 'completed' | 'aborted'

export interface MeetingFlowStep {
  id: string
  kind: MeetingFlowStepKind
  title: string
  subtitle?: string
  status: MeetingFlowStepStatus
  filePath?: string
  round?: number
}

export interface MeetingFlow {
  steps: MeetingFlowStep[]
  modeKind?: MeetingModeKind
  roundLabel: string
  outcome: MeetingFlowOutcome
  /** 中断时停在哪一步（若有） */
  interruptedStepId?: string
}

function roundFilePath(n: number): string {
  return `rounds/round-${String(n).padStart(3, '0')}.md`
}

function hasFile(files: Record<string, string>, path: string): boolean {
  return Boolean(files[path]?.trim())
}

export function parseConfirmationRequired(meetingMd: string): boolean {
  const text = meetingMd.trim()
  if (!text) return false
  const row = text.match(/确认模式\s*\|\s*([^\n|]+)/)
  if (row?.[1]) {
    const value = row[1].trim()
    if (/跳过|skip/i.test(value)) return false
    if (/需要|required/i.test(value)) return true
  }
  if (/confirmation_mode:\s*skip/i.test(text)) return false
  if (/confirmation_mode:\s*required/i.test(text)) return true
  return false
}

import { meetingModeKind, type MeetingModeKind } from '@/lib/i18n/meeting-labels'

function closingArtifactPath(
  files: Record<string, string>,
  modeKind?: MeetingModeKind,
): string {
  if (modeKind === 'decision' && hasFile(files, 'artifacts/minutes.md')) {
    return 'artifacts/minutes.md'
  }
  if (hasFile(files, 'artifacts/design-draft.md')) return 'artifacts/design-draft.md'
  if (hasFile(files, 'artifacts/minutes.md')) return 'artifacts/minutes.md'
  return modeKind === 'decision' ? 'artifacts/minutes.md' : 'artifacts/design-draft.md'
}

function stepCompletedByArtifact(files: Record<string, string>, path?: string): boolean {
  return path ? hasFile(files, path) : false
}

function applyRunningActiveStep(steps: MeetingFlowStep[]): void {
  let activeAssigned = false
  for (const step of steps) {
    if (step.status === 'skipped' || step.status === 'completed') continue
    if (!activeAssigned) {
      step.status = 'active'
      activeAssigned = true
    } else {
      step.status = 'pending'
    }
  }
}

function applyAbortedMeetingSteps(
  steps: MeetingFlowStep[],
  files: Record<string, string>,
): string | undefined {
  for (const step of steps) {
    if (step.status === 'completed' || step.status === 'skipped') continue
    if (step.filePath && stepCompletedByArtifact(files, step.filePath)) {
      step.status = 'completed'
    }
  }

  const interruptIndex = steps.findIndex(
    (s) => s.status !== 'completed' && s.status !== 'skipped',
  )
  if (interruptIndex < 0) return undefined

  steps[interruptIndex].status = 'interrupted'
  return steps[interruptIndex].id
}

function resolveMeetingFlowOutcome(
  aborted: boolean,
  successfullyCompleted: boolean,
  running: boolean,
): MeetingFlowOutcome {
  if (aborted) return 'aborted'
  if (successfullyCompleted) return 'completed'
  if (running) return 'running'
  return 'running'
}

function applyFinishedMeetingSteps(
  steps: MeetingFlowStep[],
  files: Record<string, string>,
  finished: boolean,
): void {
  if (!finished) return
  for (const step of steps) {
    if (step.status === 'completed' || step.status === 'skipped') continue
    if (step.kind === 'closing') {
      step.status = 'completed'
      continue
    }
    if (step.filePath && stepCompletedByArtifact(files, step.filePath)) {
      step.status = 'completed'
    }
  }
}

export function buildMeetingFlow(locale: AppLocale, detail: MeetingDetail): MeetingFlow {
  const t = getTranslator(locale)
  const files = detail.files ?? {}
  const meetingMd = files['MEETING.md'] ?? ''
  const modeKind = meetingModeKind(detail.mode_kind, detail.mode)
  const roundLabel =
    modeKind === 'deliberation'
      ? t('meeting.roundLabel.deliberation')
      : t('meeting.roundLabel.debate')
  const maxRounds = Math.max(1, detail.max_rounds ?? 1)
  const freeDialogue = detail.free_dialogue ?? false
  const confirmationRequired =
    parseConfirmationRequired(meetingMd) || hasFile(files, 'confirmation/brief.md')
  const aborted = isMeetingAbortedStatus(detail.status)
  const successfullyCompleted = isMeetingSuccessfullyCompletedStatus(detail.status)
  const running = isMeetingRunningStatus(detail.status)
  const outcome = resolveMeetingFlowOutcome(aborted, successfullyCompleted, running)

  const steps: MeetingFlowStep[] = []
  const preMeetingPath = 'pre-meeting/perspectives.md'
  steps.push({
    id: 'pre-meeting',
    kind: 'pre-meeting',
    title: t('meeting.flow.preMeeting'),
    subtitle: t('meeting.flow.preMeetingSub'),
    status: stepCompletedByArtifact(files, preMeetingPath) ? 'completed' : 'pending',
    filePath: hasFile(files, preMeetingPath) ? preMeetingPath : undefined,
  })

  for (let round = 1; round <= maxRounds; round++) {
    const roundPath = roundFilePath(round)
    const summaryPath = `moderator/round-${String(round).padStart(3, '0')}-summary.md`
    const hasSummary = hasFile(files, summaryPath)
    steps.push({
      id: `round-${round}`,
      kind: 'debate-round',
      round,
      title: t('meeting.flow.roundTitle', { round, label: roundLabel }),
      subtitle: hasSummary ? t('meeting.flow.roundSubWithSummary') : t('meeting.flow.roundSub'),
      status: stepCompletedByArtifact(files, roundPath) ? 'completed' : 'pending',
      filePath: hasFile(files, roundPath) ? roundPath : undefined,
    })
    if (round === 1 && freeDialogue) {
      const freeDialoguePath = 'free-dialogue/after-round-001.md'
      steps.push({
        id: 'free-dialogue',
        kind: 'free-dialogue',
        title: t('meeting.flow.freeDialogue'),
        subtitle: t('meeting.flow.freeDialogueSub'),
        status: stepCompletedByArtifact(files, freeDialoguePath) ? 'completed' : 'pending',
        filePath: hasFile(files, freeDialoguePath) ? freeDialoguePath : undefined,
      })
    }
  }

  if (modeKind === 'deliberation') {
    const draftPath = 'artifacts/design-draft.md'
    steps.push({
      id: 'synthesis',
      kind: 'synthesis',
      title: t('meeting.flow.synthesis'),
      subtitle: t('meeting.flow.synthesisSub'),
      status: stepCompletedByArtifact(files, draftPath) ? 'completed' : 'pending',
      filePath: hasFile(files, draftPath) ? draftPath : undefined,
    })
  }

  if (confirmationRequired) {
    const confirmationPath = 'confirmation/brief.md'
    steps.push({
      id: 'confirmation',
      kind: 'confirmation',
      title: t('meeting.flow.confirmation'),
      subtitle: t('meeting.flow.confirmationSub'),
      status: stepCompletedByArtifact(files, confirmationPath) ? 'completed' : 'pending',
      filePath: hasFile(files, confirmationPath) ? confirmationPath : undefined,
    })
  }

  const closingPath = closingArtifactPath(files, modeKind)
  steps.push({
    id: 'closing',
    kind: 'closing',
    title: t('meeting.flow.closing'),
    subtitle:
      modeKind === 'decision'
        ? t('meeting.flow.closingDecision')
        : t('meeting.flow.closingDeliberation'),
    status:
      (successfullyCompleted && !aborted) || stepCompletedByArtifact(files, closingPath)
        ? 'completed'
        : 'pending',
    filePath: hasFile(files, closingPath) ? closingPath : undefined,
  })

  let interruptedStepId: string | undefined
  if (aborted) {
    interruptedStepId = applyAbortedMeetingSteps(steps, files)
  } else if (successfullyCompleted) {
    applyFinishedMeetingSteps(steps, files, true)
  } else if (running) {
    applyRunningActiveStep(steps)
  }

  return { steps, modeKind, roundLabel, outcome, interruptedStepId }
}

export function meetingFlowStepStatusLabel(
  locale: AppLocale,
  status: MeetingFlowStepStatus,
): string {
  const t = getTranslator(locale)
  switch (status) {
    case 'completed':
      return t('meeting.flow.stepCompleted')
    case 'active':
      return t('meeting.flow.stepActive')
    case 'skipped':
      return t('meeting.flow.stepSkipped')
    case 'interrupted':
      return t('meeting.flow.stepInterrupted')
    default:
      return t('meeting.flow.stepPending')
  }
}
