import { AlertTriangle, CheckCircle2 } from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import type { MeetingFlow, MeetingFlowOutcome } from '@/lib/meeting-flow'
import { cn } from '@/lib/utils'

export function meetingFlowProgressBarClass(outcome: MeetingFlowOutcome): string {
  switch (outcome) {
    case 'completed':
      return 'bg-success'
    case 'aborted':
      return 'bg-danger'
    default:
      return 'bg-brand'
  }
}

export function meetingFlowProgressLabel(
  t: ReturnType<typeof useI18n>['t'],
  flow: MeetingFlow,
  completedCount: number,
): string {
  if (flow.outcome === 'completed') {
    return t('meetingUi.flow.progressCompleted')
  }
  if (flow.outcome === 'aborted') {
    return t('meetingUi.flow.progressAborted', {
      done: completedCount,
      total: flow.steps.length,
    })
  }
  return `${completedCount}/${flow.steps.length}`
}

interface MeetingFlowOutcomeBannerProps {
  flow: MeetingFlow
  className?: string
}

export function MeetingFlowOutcomeBanner({ flow, className }: MeetingFlowOutcomeBannerProps) {
  const { t } = useI18n()

  if (flow.outcome === 'completed') {
    return (
      <div
        className={cn(
          'mb-3 flex gap-2.5 rounded-xs bg-success-soft/55 px-2.5 py-2 ring-1 ring-inset ring-success/20',
          className,
        )}
      >
        <CheckCircle2 className="mt-0.5 size-3.5 shrink-0 text-success" aria-hidden />
        <div className="min-w-0 space-y-0.5">
          <p className="text-[11px] font-medium text-success">
            {t('meetingUi.flow.outcomeCompletedTitle')}
          </p>
          <p className="text-[10px] leading-relaxed text-text-secondary">
            {flow.confirmationRejections && flow.confirmationRejections > 0
              ? t('meetingUi.flow.outcomeCompletedWithRejections', {
                  count: flow.confirmationRejections,
                })
              : t('meetingUi.flow.outcomeCompletedHint')}
          </p>
        </div>
      </div>
    )
  }

  if (flow.outcome === 'aborted') {
    const interruptedStep = flow.steps.find((s) => s.id === flow.interruptedStepId)
    return (
      <div
        className={cn(
          'mb-3 flex gap-2.5 rounded-xs bg-danger-soft/55 px-2.5 py-2 ring-1 ring-inset ring-danger/20',
          className,
        )}
      >
        <AlertTriangle className="mt-0.5 size-3.5 shrink-0 text-danger" aria-hidden />
        <div className="min-w-0 space-y-0.5">
          <p className="text-[11px] font-medium text-danger">
            {t('meetingUi.flow.outcomeAbortedTitle')}
          </p>
          <p className="text-[10px] leading-relaxed text-text-secondary">
            {interruptedStep
              ? t('meetingUi.flow.outcomeAbortedAt', { title: interruptedStep.title })
              : t('meetingUi.flow.outcomeAbortedHint')}
          </p>
        </div>
      </div>
    )
  }

  return null
}
