import type { LucideIcon } from 'lucide-react'
import {
  Flag,
  GitBranch,
  Layers,
  MessageCircle,
  Scan,
  Sparkles,
  UserCheck,
} from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import type {
  MeetingFlow,
  MeetingFlowStep,
  MeetingFlowStepKind,
  MeetingFlowStepStatus,
} from '@/lib/meeting-flow'
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

function stepStatusDotClass(status: MeetingFlowStepStatus): string {
  switch (status) {
    case 'completed':
      return 'bg-success'
    case 'active':
      return 'bg-brand ring-2 ring-brand/25'
    case 'skipped':
      return 'bg-text-tertiary/30'
    default:
      return 'bg-black/[0.08]'
  }
}

function CompactFlowStep({
  step,
  isLast,
  onOpenFile,
  statusLabel,
  clickToView,
}: {
  step: MeetingFlowStep
  isLast: boolean
  onOpenFile?: (path: string) => void
  statusLabel: (status: MeetingFlowStepStatus) => string
  clickToView: string
}) {
  const clickable = Boolean(step.filePath && onOpenFile)

  return (
    <li className="relative flex gap-2.5 pb-3 last:pb-0">
      {!isLast && (
        <span
          className="absolute left-[11px] top-6 bottom-0 w-px bg-black/[0.07]"
          aria-hidden
        />
      )}

      <span
        className={cn(
          'relative z-[1] mt-1.5 size-2 shrink-0 rounded-full',
          stepStatusDotClass(step.status),
        )}
        aria-hidden
      />

      <div className="min-w-0 flex-1">
        {clickable ? (
          <button
            type="button"
            className={cn(
              'group/step w-full rounded-xs px-1.5 py-1 text-left -mx-1.5',
              hePressable,
              heSpring,
              'hover:bg-black/[0.03]',
            )}
            onClick={() => step.filePath && onOpenFile?.(step.filePath)}
          >
            <StepContent
              step={step}
              clickable
              statusLabel={statusLabel}
              clickToView={clickToView}
            />
          </button>
        ) : (
          <div className="px-1.5 py-1 -mx-1.5">
            <StepContent step={step} statusLabel={statusLabel} clickToView={clickToView} />
          </div>
        )}
      </div>
    </li>
  )
}

function StepContent({
  step,
  clickable,
  statusLabel,
  clickToView,
}: {
  step: MeetingFlowStep
  clickable?: boolean
  statusLabel: (status: MeetingFlowStepStatus) => string
  clickToView: string
}) {
  const Icon = STEP_ICONS[step.kind]

  return (
    <>
      <div className="flex items-center gap-1.5">
        <Icon
          className={cn(
            'size-3 shrink-0',
            step.status === 'completed' && 'text-success',
            step.status === 'active' && 'text-brand',
            step.status === 'pending' && 'text-text-tertiary/60',
            step.status === 'skipped' && 'text-text-tertiary/50',
          )}
          aria-hidden
        />
        <p
          className={cn(
            'truncate text-[12px] font-medium',
            step.status === 'active' ? 'text-text-primary' : 'text-text-secondary',
            clickable && 'group-hover/step:text-brand',
          )}
        >
          {step.title}
        </p>
      </div>
      <p className="mt-0.5 text-[10px] text-text-tertiary">
        {statusLabel(step.status)}
        {clickable && step.filePath ? clickToView : ''}
      </p>
    </>
  )
}

interface MeetingFlowDockProps {
  detail: MeetingDetail
  className?: string
  sticky?: boolean
  onOpenFile?: (path: string) => void
}

export function MeetingFlowDock({
  detail,
  className,
  sticky = true,
  onOpenFile,
}: MeetingFlowDockProps) {
  const { t, buildMeetingFlow, meetingFlowStepStatusLabel, meetingModeShort } = useI18n()
  const flow = buildMeetingFlow(detail)
  const completedCount = flow.steps.filter((s) => s.status === 'completed').length
  const activeStep = flow.steps.find((s) => s.status === 'active')
  const modeLabel =
    flow.modeKind === 'deliberation'
      ? meetingModeShort('deliberation')
      : meetingModeShort('decision')

  return (
    <aside
      className={cn(
        hePanelShell,
        'overflow-visible p-4',
        sticky && 'lg:sticky lg:top-20 lg:self-start',
        className,
      )}
    >
      <div className="mb-3 flex items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <GitBranch className="size-3.5 shrink-0 text-info" aria-hidden />
            <p className="text-[12px] font-semibold text-text-primary">
              {t('meetingUi.flow.title')}
            </p>
          </div>
          <p className="mt-0.5 text-[10px] text-text-tertiary">{modeLabel}</p>
        </div>
        <span className="shrink-0 rounded-full bg-black/[0.03] px-2 py-0.5 text-[10px] font-medium tabular-nums text-text-tertiary ring-1 ring-inset ring-black/[0.05]">
          {completedCount}/{flow.steps.length}
        </span>
      </div>

      <div
        className="mb-3 h-1 overflow-hidden rounded-full bg-black/[0.05]"
        role="progressbar"
        aria-valuenow={completedCount}
        aria-valuemin={0}
        aria-valuemax={flow.steps.length}
        aria-label={t('meetingUi.flow.progressAriaLabel')}
      >
        <div
          className="h-full rounded-full bg-brand transition-[width] duration-500"
          style={{ width: `${(completedCount / Math.max(flow.steps.length, 1)) * 100}%` }}
        />
      </div>

      {activeStep && (
        <p className="mb-3 rounded-xs bg-brand-soft/50 px-2.5 py-2 text-[11px] leading-relaxed text-text-secondary ring-1 ring-inset ring-primary/10">
          {t('meetingUi.flow.currentStep', { title: activeStep.title })}
        </p>
      )}

      <ol className="list-none">
        {flow.steps.map((step, index) => (
          <CompactFlowStep
            key={step.id}
            step={step}
            isLast={index === flow.steps.length - 1}
            onOpenFile={onOpenFile}
            statusLabel={meetingFlowStepStatusLabel}
            clickToView={t('meetingUi.flow.clickToView')}
          />
        ))}
      </ol>
    </aside>
  )
}

export type { MeetingFlow }
