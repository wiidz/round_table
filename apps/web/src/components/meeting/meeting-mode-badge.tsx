import { Cloud, Scale } from 'lucide-react'

import { cn } from '@/lib/utils'
import { meetingModeKind, meetingModeShort } from '@/lib/meeting-labels'

export type MeetingModeKind = 'decision' | 'deliberation'

type ModeTagSize = 'sm' | 'md'

interface ModeTagProps {
  mode?: string
  modeKind?: string
  size?: ModeTagSize
  className?: string
}

/**
 * 裁决 vs 研讨 — 图标 + 文案在同一 tag 内
 * - 裁决：方角 tag · 明亮橙底 · 白字
 * - 研讨：圆角 tag · 淡蓝底 · 蓝字
 */
const modeVisual: Record<
  MeetingModeKind,
  {
    icon: typeof Scale
    label: string
    tag: string
    iconStroke: number
  }
> = {
  decision: {
    icon: Scale,
    label: '裁决型',
    tag: 'rounded-[10px] bg-[#F89540] text-white shadow-[0_2px_6px_-2px_rgba(248,149,64,0.35)] dark:bg-[#F97316]',
    iconStroke: 1.75,
  },
  deliberation: {
    icon: Cloud,
    label: '研讨型',
    tag: 'rounded-full bg-[#F0F9FF] text-[#0369A1] ring-1 ring-inset ring-[#BAE6FD]/80 dark:bg-[#0C2340]/60 dark:text-[#BAE6FD] dark:ring-[#0369A1]/30',
    iconStroke: 1.65,
  },
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

function ModeTag({ mode, modeKind, size = 'sm', className }: ModeTagProps) {
  const kind = meetingModeKind(modeKind, mode)
  const label = meetingModeShort(mode) ?? (kind ? modeVisual[kind].label : '未知模式')
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
      aria-label={`${visual.label}会议`}
    >
      <Icon className={cn(sizing.icon, 'shrink-0')} strokeWidth={visual.iconStroke} aria-hidden />
      <span className={sizing.text}>{visual.label}</span>
    </span>
  )
}

/** 列表卡片顶栏 */
export function MeetingModeMark(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="md" />
}

/** 详情页元信息行 */
export function MeetingModeInline(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="sm" />
}

/** 紧凑 pill（与 Inline 同款，保留别名） */
export function MeetingModeBadge(props: Omit<ModeTagProps, 'size'>) {
  return <ModeTag {...props} size="sm" />
}

export function meetingModeHoverClass(modeKind?: string, mode?: string): string {
  const kind = meetingModeKind(modeKind, mode)
  if (kind === 'decision') return 'group-hover:text-[#EA580C] dark:group-hover:text-[#FDBA74]'
  if (kind === 'deliberation') return 'group-hover:text-[#0284C7] dark:group-hover:text-[#7DD3FC]'
  return 'group-hover:text-text-primary'
}

export function meetingModeFreeDialogueClass(modeKind?: string, mode?: string): string {
  const kind = meetingModeKind(modeKind, mode)
  if (kind === 'decision') return 'font-medium text-[#EA580C] dark:text-[#FDBA74]'
  if (kind === 'deliberation') return 'font-medium text-[#0284C7] dark:text-[#7DD3FC]'
  return 'font-medium text-text-secondary'
}
