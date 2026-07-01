import { FileText, LayoutDashboard } from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import { hePressable, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export type MeetingDetailView = 'overview' | 'documents'

interface MeetingDetailViewTabsProps {
  value: MeetingDetailView
  documentCount?: number
  onChange: (view: MeetingDetailView) => void
  className?: string
}

export function MeetingDetailViewTabs({
  value,
  documentCount,
  onChange,
  className,
}: MeetingDetailViewTabsProps) {
  const { t } = useI18n()

  const tabs: {
    id: MeetingDetailView
    label: string
    icon: typeof LayoutDashboard
  }[] = [
    { id: 'overview', label: t('meeting.tabs.overview'), icon: LayoutDashboard },
    { id: 'documents', label: t('meeting.tabs.documents'), icon: FileText },
  ]

  return (
    <div
      className={cn(
        'flex w-fit rounded-xl bg-black/[0.03] p-1 ring-1 ring-inset ring-black/[0.06]',
        className,
      )}
      role="tablist"
      aria-label={t('meeting.tabs.ariaLabel')}
    >
      {tabs.map((tab) => {
        const selected = value === tab.id
        const Icon = tab.icon
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={selected}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-lg px-4 py-2 text-[13px] font-medium',
              hePressable,
              heSpring,
              selected
                ? 'bg-brand text-white shadow-[0_8px_20px_-8px_rgba(232,93,4,0.45)] ring-1 ring-inset ring-primary/25'
                : 'text-text-secondary hover:bg-black/[0.03] hover:text-text-primary',
            )}
            onClick={() => onChange(tab.id)}
          >
            <Icon
              className={cn('size-3.5 shrink-0', selected ? 'text-white' : 'text-text-tertiary')}
              aria-hidden
            />
            {tab.label}
            {tab.id === 'documents' && documentCount !== undefined && (
              <span
                className={cn(
                  'rounded-full px-1.5 py-px text-[10px] font-semibold tabular-nums',
                  selected
                    ? 'bg-white/20 text-white ring-1 ring-inset ring-white/25'
                    : 'bg-black/[0.05] text-text-tertiary',
                )}
              >
                {documentCount}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
