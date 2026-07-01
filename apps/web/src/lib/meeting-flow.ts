import { meetingModeKind, type MeetingModeKind } from '@/lib/meeting-labels'

import type { MeetingDetail } from '@/types/meeting'

export type MeetingFlowStepKind =
  | 'pre-meeting'
  | 'debate-round'
  | 'free-dialogue'
  | 'synthesis'
  | 'confirmation'
  | 'closing'

export type MeetingFlowStepStatus = 'pending' | 'active' | 'completed' | 'skipped'

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

export function isMeetingFinishedStatus(status?: string): boolean {
  const s = status?.trim() ?? ''
  return (
    s === '已结束' ||
    s === 'Completed' ||
    s === '已归档' ||
    s === 'Archived' ||
    s === '共识达成' ||
    s === 'Consensus' ||
    s === '已中断' ||
    s === 'aborted' ||
    s === 'Aborted'
  )
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

function closingArtifactPath(
  files: Record<string, string>,
  modeKind?: MeetingModeKind,
): string {
  if (modeKind === 'decision' && hasFile(files, 'artifacts/minutes.md')) {
    return 'artifacts/minutes.md'
  }
  if (hasFile(files, 'artifacts/design-draft.md')) {
    return 'artifacts/design-draft.md'
  }
  if (hasFile(files, 'artifacts/minutes.md')) {
    return 'artifacts/minutes.md'
  }
  return modeKind === 'decision' ? 'artifacts/minutes.md' : 'artifacts/design-draft.md'
}

function stepCompletedByArtifact(
  files: Record<string, string>,
  path?: string,
): boolean {
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

/** Derive canonical meeting pipeline from workspace files + index metadata. */
export function buildMeetingFlow(detail: MeetingDetail): MeetingFlow {
  const files = detail.files ?? {}
  const meetingMd = files['MEETING.md'] ?? ''
  const modeKind = meetingModeKind(detail.mode_kind, detail.mode)
  const roundLabel = modeKind === 'deliberation' ? '研讨' : '辩论'
  const maxRounds = Math.max(1, detail.max_rounds ?? 1)
  const freeDialogue = detail.free_dialogue ?? false
  const confirmationRequired =
    parseConfirmationRequired(meetingMd) || hasFile(files, 'confirmation/brief.md')
  const finished = isMeetingFinishedStatus(detail.status)
  const running = isMeetingRunningStatus(detail.status)

  const steps: MeetingFlowStep[] = []

  const preMeetingPath = 'pre-meeting/perspectives.md'
  steps.push({
    id: 'pre-meeting',
    kind: 'pre-meeting',
    title: '会前准备',
    subtitle: 'Round 0：各专家独立提交初始观点',
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
      title: `第 ${round} 轮${roundLabel}`,
      subtitle: hasSummary
        ? '发言与立场记录 · 含 Moderator 轮次摘要'
        : '按顺序发言、回应与立场记录',
      status: stepCompletedByArtifact(files, roundPath) ? 'completed' : 'pending',
      filePath: hasFile(files, roundPath) ? roundPath : undefined,
    })

    if (round === 1 && freeDialogue) {
      const freeDialoguePath = 'free-dialogue/after-round-001.md'
      steps.push({
        id: 'free-dialogue',
        kind: 'free-dialogue',
        title: '自由问答',
        subtitle: 'Round 1 后参与者互相提问与回答',
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
      title: '方案合成',
      subtitle: 'Moderator 综合全场合议形成方案草案',
      status: stepCompletedByArtifact(files, draftPath) ? 'completed' : 'pending',
      filePath: hasFile(files, draftPath) ? draftPath : undefined,
    })
  }

  if (confirmationRequired) {
    const confirmationPath = 'confirmation/brief.md'
    steps.push({
      id: 'confirmation',
      kind: 'confirmation',
      title: 'Principal 确认',
      subtitle: 'Principal 审阅呈报清单并确认共识',
      status: stepCompletedByArtifact(files, confirmationPath) ? 'completed' : 'pending',
      filePath: hasFile(files, confirmationPath) ? confirmationPath : undefined,
    })
  }

  const closingPath = closingArtifactPath(files, modeKind)
  steps.push({
    id: 'closing',
    kind: 'closing',
    title: '结案',
    subtitle:
      modeKind === 'decision'
        ? '输出结论纪要并结束会议'
        : '归档交付物并结束会议',
    status:
      finished || stepCompletedByArtifact(files, closingPath) ? 'completed' : 'pending',
    filePath: hasFile(files, closingPath) ? closingPath : undefined,
  })

  applyFinishedMeetingSteps(steps, files, finished)
  if (running) {
    applyRunningActiveStep(steps)
  }

  return { steps, modeKind, roundLabel }
}

export function meetingFlowStepStatusLabel(status: MeetingFlowStepStatus): string {
  switch (status) {
    case 'completed':
      return '已完成'
    case 'active':
      return '进行中'
    case 'skipped':
      return '已跳过'
    default:
      return '待进行'
  }
}
