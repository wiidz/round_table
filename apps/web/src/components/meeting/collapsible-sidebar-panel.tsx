import { useState, type ReactNode } from 'react'
import { ChevronDown, type LucideIcon } from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import { hePanelShell, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface CollapsibleSidebarPanelProps {
  panelTitle: string
  subtitle?: ReactNode
  icon?: LucideIcon
  iconClassName?: string
  trailing?: ReactNode
  defaultExpanded?: boolean
  children: ReactNode
  className?: string
  bodyClassName?: string
}

export function CollapsibleSidebarPanel({
  panelTitle,
  subtitle,
  icon: Icon,
  iconClassName,
  trailing,
  defaultExpanded = true,
  children,
  className,
  bodyClassName,
}: CollapsibleSidebarPanelProps) {
  const { t } = useI18n()
  const [expanded, setExpanded] = useState(defaultExpanded)

  return (
    <aside className={cn(hePanelShell, 'overflow-visible p-4', className)}>
      <button
        type="button"
        aria-expanded={expanded}
        aria-label={
          expanded
            ? t('meetingUi.sidebar.collapseAria', { title: panelTitle })
            : t('meetingUi.sidebar.expandAria', { title: panelTitle })
        }
        onClick={() => setExpanded((v) => !v)}
        className={cn(
          'flex w-full items-start gap-2 rounded-md text-left',
          heSpring,
          'hover:bg-black/[0.03]',
        )}
      >
        {Icon && (
          <Icon
            className={cn('mt-0.5 size-3.5 shrink-0', iconClassName)}
            aria-hidden
          />
        )}
        <span className="min-w-0 flex-1">
          <span className="block text-[12px] font-semibold text-text-primary">{panelTitle}</span>
          {subtitle && (
            <span className="mt-0.5 block text-[10px] leading-relaxed text-text-tertiary">
              {subtitle}
            </span>
          )}
        </span>
        <span className="flex shrink-0 items-center gap-1.5 pt-0.5">
          {trailing}
          <ChevronDown
            className={cn(
              'size-3.5 text-text-tertiary transition-transform',
              !expanded && '-rotate-90',
            )}
            aria-hidden
          />
        </span>
      </button>
      {expanded && <div className={cn('mt-3', bodyClassName)}>{children}</div>}
    </aside>
  )
}
