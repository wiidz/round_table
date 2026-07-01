import { X } from 'lucide-react'

import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { cn } from '@/lib/utils'

export type MeetingExpertEntry = {
  id: string
  name: string
}

interface BriefMeetingExpertsListProps {
  experts: MeetingExpertEntry[]
  emptyLabel?: string
  removable?: boolean
  disabled?: boolean
  onRemove?: (id: string) => void
  className?: string
}

export function BriefMeetingExpertsList({
  experts,
  emptyLabel = '—',
  removable = false,
  disabled = false,
  onRemove,
  className,
}: BriefMeetingExpertsListProps) {
  if (experts.length === 0) {
    return (
      <p className={cn('text-[14px] leading-relaxed text-text-tertiary', className)}>
        {emptyLabel}
      </p>
    )
  }

  return (
    <ul className={cn('flex flex-col gap-2.5', className)}>
      {experts.map((expert) => (
        <li key={expert.id} className="flex min-w-0 items-center gap-2.5">
          <ProfileAvatar id={expert.id} name={expert.name} size="xs" className="shrink-0" />
          <span className="min-w-0 flex-1 truncate text-[14px] font-medium leading-snug text-text-primary">
            {expert.name}
          </span>
          {removable && !disabled && onRemove && (
            <button
              type="button"
              className="shrink-0 rounded p-0.5 text-text-tertiary hover:bg-black/[0.06] hover:text-text-primary"
              aria-label={`移除 ${expert.name}`}
              onClick={() => onRemove(expert.id)}
            >
              <X className="size-3.5" />
            </button>
          )}
        </li>
      ))}
    </ul>
  )
}
