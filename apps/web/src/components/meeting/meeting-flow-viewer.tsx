import type { LucideIcon } from 'lucide-react'
import {
  CheckCircle2,
  Circle,
  CircleDashed,
  Flag,
  Layers,
  MessageCircle,
  Scan,
  Sparkles,
  UserCheck,
} from 'lucide-react'

import {
  buildMeetingFlow,
  meetingFlowStepStatusLabel,
  type MeetingFlowStep,
  type MeetingFlowStepKind,
  type MeetingFlowStepStatus,
} from '@/lib/meeting-flow'
import { meetingFileLabel } from '@/lib/meeting-labels'
import { hePanelShell, hePressable, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

const STEP_ICONS: Record<MeetingFlowStepKind, LucideIcon> = {
  'pre-meeting': Scan,
  'debate-round': Layers,
  'free-dialogue': MessageCircle,
  synthesis: Sparkles,
  confirmation: UserCheck,
  closing: Flag,
}

function statusTone(status: MeetingFlowStepStatus): string {
  switch (status) {
    case 'completed':
      return 'text-success'
    case 'active':
      return 'text-brand'
    case 'skipped':
      return 'text-text-tertiary'
    default:
      return 'text-text-tertiary'
  }
}

function statusBadgeClass(status: MeetingFlowStepStatus): string {
  switch (status) {
    case 'completed':
      return 'bg-success-soft text-success ring-success/20'
    case 'active':
      return 'bg-brand-soft text-brand ring-primary/20'
    case 'skipped':
      return 'bg-black/[0.03] text-text-tertiary ring-black/[0.06]'
    default:
      return 'bg-black/[0.03] text-text-tertiary ring-black/[0.06]'
  }
}

function StatusIcon({ status }: { status: MeetingFlowStepStatus }) {
  if (status === 'completed') {
    return <CheckCircle2 className="size-4 text-success" aria-hidden />
  }
  if (status === 'active') {
    return <Circle className="size-4 fill-brand/15 text-brand" aria-hidden />
  }
  return <CircleDashed className="size-4 text-text-tertiary/70" aria-hidden />
}

function FlowStepRow({
  step,
  modeKind,
  isLast,
  onOpenFile,
}: {
  step: MeetingFlowStep
  modeKind?: MeetingDetail['mode_kind']
  isLast: boolean
  onOpenFile?: (path: string) => void
}) {
  const Icon = STEP_ICONS[step.kind]
  const canOpen = Boolean(step.filePath && onOpenFile)

  return (
    <li className="relative flex gap-4 pb-8 last:pb-0">
      {!isLast && (
        <span
          className="absolute left-[15px] top-9 bottom-0 w-px bg-black/[0.08]"
          aria-hidden
        />
      )}

      <div
        className={cn(
          'relative z-[1] flex size-8 shrink-0 items-center justify-center rounded-full bg-surface ring-1 ring-inset',
          step.status === 'active' && 'ring-primary/25',
          step.status === 'completed' && 'ring-success/25',
          step.status === 'pending' && 'ring-black/[0.06]',
        )}
      >
        <Icon className={cn('size-3.5', statusTone(step.status))} aria-hidden />
      </div>

      <div className="min-w-0 flex-1 space-y-2 pt-0.5">
        <div className="flex flex-wrap items-start justify-between gap-x-3 gap-y-1">
          <div className="min-w-0 space-y-1">
            <p className="text-[15px] font-semibold tracking-[-0.01em] text-text-primary">
              {step.title}
            </p>
            {step.subtitle && (
              <p className="text-[13px] leading-relaxed text-text-secondary">{step.subtitle}</p>
            )}
          </div>
          <span
            className={cn(
              'inline-flex shrink-0 items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-medium ring-1 ring-inset',
              statusBadgeClass(step.status),
            )}
          >
            <StatusIcon status={step.status} />
            {meetingFlowStepStatusLabel(step.status)}
          </span>
        </div>

        {step.filePath && (
          <button
            type="button"
            disabled={!canOpen}
            className={cn(
              'inline-flex max-w-full items-center gap-1.5 rounded-xs bg-black/[0.025] px-2.5 py-1.5 text-left text-[12px] text-text-secondary ring-1 ring-inset ring-black/[0.05]',
              canOpen && cn(hePressable, heSpring, 'hover:bg-black/[0.04] hover:text-brand'),
              !canOpen && 'cursor-default opacity-70',
            )}
            onClick={() => {
              if (step.filePath && onOpenFile) onOpenFile(step.filePath)
            }}
          >
            <span className="truncate font-mono">{step.filePath}</span>
            {canOpen && <span className="shrink-0 text-brand">查看</span>}
          </button>
        )}

        {step.filePath && (
          <p className="text-[12px] text-text-tertiary">
            {meetingFileLabel(step.filePath, modeKind)}
          </p>
        )}
      </div>
    </li>
  )
}

interface MeetingFlowViewerProps {
  detail: MeetingDetail
  className?: string
  onOpenFile?: (path: string) => void
}

export function MeetingFlowViewer({ detail, className, onOpenFile }: MeetingFlowViewerProps) {
  const flow = buildMeetingFlow(detail)
  const completedCount = flow.steps.filter((s) => s.status === 'completed').length

  return (
    <section className={cn(hePanelShell, 'overflow-visible p-6 sm:p-8', className)}>
      <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
        <div className="space-y-1">
          <p className="text-[10px] font-medium uppercase tracking-[0.16em] text-text-tertiary">
            会议流程
          </p>
          <p className="text-sm text-text-secondary">
            从会前准备到结案的 Engine 标准路径
            {flow.modeKind === 'deliberation' ? '（研讨型）' : '（裁决型）'}
          </p>
        </div>
        <p className="text-[12px] tabular-nums text-text-tertiary">
          {completedCount}/{flow.steps.length} 已完成
        </p>
      </div>

      <ol className="list-none">
        {flow.steps.map((step, index) => (
          <FlowStepRow
            key={step.id}
            step={step}
            modeKind={flow.modeKind}
            isLast={index === flow.steps.length - 1}
            onOpenFile={onOpenFile}
          />
        ))}
      </ol>
    </section>
  )
}
