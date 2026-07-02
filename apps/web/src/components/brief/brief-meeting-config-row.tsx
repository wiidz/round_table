import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

import { briefMeetingConfigLabelClass, briefMeetingConfigRowGrid } from './brief-template-sections'

export function BriefMeetingConfigRow({
  label,
  value,
  valueAlign = 'center',
}: {
  label: string
  value: ReactNode
  valueAlign?: 'center' | 'start'
}) {
  const text = typeof value === 'string' ? value.trim() : ''

  return (
    <div
      className={cn(
        briefMeetingConfigRowGrid,
        valueAlign === 'start' && 'sm:items-start',
      )}
    >
      <p className={cn(briefMeetingConfigLabelClass, 'break-words pt-2 sm:max-w-[6rem] sm:pt-2.5')}>{label}</p>
      {typeof value === 'string' ? (
        <p
          className={cn(
            'text-[14px] leading-relaxed',
            text ? 'font-medium text-text-primary' : 'text-text-tertiary',
          )}
        >
          {text || '—'}
        </p>
      ) : (
        <div className="min-w-0">{value}</div>
      )}
    </div>
  )
}
