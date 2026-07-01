import { Cloud, Scale } from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import { meetingModeKind as meetingModeKindI18n } from '@/lib/i18n/meeting-labels'
import { cn } from '@/lib/utils'

export type MeetingModeKind = 'decision' | 'deliberation'

type ModeTagSize = 'sm' | 'md'

interface ModeTagProps {
  mode?: string
  modeKind?: string
  size?: ModeTagSize
  className?: string
}

const tagSizeClass: Record<ModeTagSize, { tag: string; icon: string; text: string }> = {
  sm: {
    tag: 'gap-1.5 px-2 py-0.5 text-[12px]',
    icon: 'size-3.5',
    text: 'font-semibold',
  },
  md: {
    tag: 'gap-2 px-2.5 py-1.5 text-[12px]',
    icon: 'size-4',
    text: 'font-bold tracking-[-0.01em]',
  },
}

const modeVisual: Record<
  MeetingModeKind,
  {
    icon: typeof Scale
    tag: string
    iconStroke: number
    meetingKey: 'meeting.mode.decisionMeeting' | 'meeting.mode.deliberationMeeting'
  }
> = {
  decision: {
    icon: Scale,
    tag: 'rounded-[10px] bg-[#F89540] text-white shadow-[0_2px_6px_-2px_rgba(248,149,64,0.35)] dark:bg-[#F97316]',
    iconStroke: 1.75,
    meetingKey: 'meeting.mode.decisionMeeting',
  },
  deliberation: {
    icon: Cloud,
    tag: 'rounded-full bg-[#F0F9FF] text-[#0369A1] ring-1 ring-inset ring-[#BAE6FD]/80 dark:bg-[#0C2340]/60 dark:text-[#BAE6FD] dark:ring-[#0369A1]/30',
    iconStroke: 1.65,
    meetingKey: 'meeting.mode.deliberationMeeting',
  },
}

function ModeTag({ mode, modeKind, size = 'sm', className }: ModeTagProps) {
  const { t, meetingModeShort, meetingModeKind } = useI18n()
  const kind = meetingModeKind(modeKind, mode)
  const label =
    meetingModeShort(mode) ?? (kind ? t(`meeting.mode.${kind}`) : t('meeting.mode.unknown'))
  const sizing = tagSizeClass[size]

  if (!kind) {
    return (
      <span
        className={cn(
          'inline-flex shrink-0 items-center rounded-lg bg-black/[0.04] text-text-secondary ring-1 ring-inset ring-black/[0.06]',
          sizing.tag,
          className,
        )}
      >
        <Scale className={cn(sizing.icon, 'shrink-0')} strokeWidth={1.65} aria-hidden />
        <span className={sizing.text}>{label}</span>
      </span>
    )
  }

  const visual = modeVisual[kind]
  const Icon = visual.icon

  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center',
        visual.tag,
        sizing.tag,
        className,
      )}
      aria-label={t(visual.meetingKey)}
    >
      <Icon className={cn(sizing.icon, 'shrink-0')} strokeWidth={visual.iconStroke} aria-hidden />
      <span className={sizing.text}>{label}</span>
    </span>
  )
}

export function MeetingModeMark(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="md" />
}

export function MeetingModeInline(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="sm" />
}

export function MeetingModeBadge(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="sm" />
}

export function meetingModeHoverClass(modeKind?: string, mode?: string): string {
  const kind = meetingModeKindI18n(modeKind, mode)
  if (kind === 'decision') return 'group-hover:text-[#EA580C] dark:group-hover:text-[#FDBA74]'
  if (kind === 'deliberation') return 'group-hover:text-[#0284C7] dark:group-hover:text-[#7DD3FC]'
  return 'group-hover:text-text-primary'
}

export function meetingModeFreeDialogueClass(modeKind?: string, mode?: string): string {
  const kind = meetingModeKindI18n(modeKind, mode)
  if (kind === 'decision') return 'font-medium text-[#EA580C] dark:text-[#FDBA74]'
  if (kind === 'deliberation') return 'font-medium text-[#0284C7] dark:text-[#7DD3FC]'
  return 'font-medium text-text-secondary'
}
