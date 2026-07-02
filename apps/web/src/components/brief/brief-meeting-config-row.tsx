import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

import { briefMeetingConfigLabelClass, briefMeetingConfigRowGrid, meetingDetailConfigRowGrid } from './brief-template-sections'

export function BriefMeetingConfigRow({
  label,
  value,
  valueAlign = 'center',
  layout = 'template',
}: {
  label: string
  value: ReactNode
  valueAlign?: 'center' | 'start'
  /** template：简报编辑/预览；detail：会议详情侧栏 */
  layout?: 'template' | 'detail'
}) {
  const text = typeof value === 'string' ? value.trim() : ''
  const rowGrid = layout === 'detail' ? meetingDetailConfigRowGrid : briefMeetingConfigRowGrid
  const labelWidth =
    layout === 'detail' ? 'sm:max-w-[4rem]' : 'sm:max-w-[6rem]'

  return (
    <div
      className={cn(
        rowGrid,
        valueAlign === 'start' && 'sm:items-start',
      )}
    >
      <p className={cn(briefMeetingConfigLabelClass, 'break-words pt-2 sm:pt-2.5', labelWidth)}>{label}</p>
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
