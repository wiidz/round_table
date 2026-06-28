import { cn } from '@/lib/utils'
import {
  meetingStatusLabel,
  meetingStatusTone,
  type MeetingStatusTone,
} from '@/lib/meeting-labels'

const toneClass: Record<MeetingStatusTone, string> = {
  neutral:
    'bg-black/[0.04] text-text-secondary ring-black/[0.06]',
  running:
    'bg-info/10 text-info ring-info/20',
  warning:
    'bg-warning/10 text-warning ring-warning/20',
  success:
    'bg-success/10 text-success ring-success/20',
  danger:
    'bg-danger/10 text-danger ring-danger/20',
}

interface MeetingStatusBadgeProps {
  status: string
  className?: string
}

export function MeetingStatusBadge({ status, className }: MeetingStatusBadgeProps) {
  const tone = meetingStatusTone(status)
  return (
    <span
      className={cn(
        'inline-flex shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium ring-1 ring-inset',
        toneClass[tone],
        className,
      )}
    >
      {meetingStatusLabel(status)}
    </span>
  )
}
