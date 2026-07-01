import { Hash } from 'lucide-react'

import { briefFieldCaptionClass } from '@/components/brief/brief-template-sections'
import { heFieldHint, heSectionTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface BriefSectionHeadingProps {
  title: string
  description?: string
  className?: string
  as?: 'h2' | 'h3'
}

export function BriefSectionHeading({
  title,
  description,
  className,
  as: Tag = 'h2',
}: BriefSectionHeadingProps) {
  return (
    <div className={cn('space-y-1.5', className)}>
      <div className="flex min-w-0 items-center gap-2">
        <Hash className="size-4 shrink-0 text-info" strokeWidth={2} aria-hidden />
        <Tag className={heSectionTitle}>{title}</Tag>
      </div>
      {description && <p className={cn(heFieldHint, 'text-text-tertiary/90')}>{description}</p>}
    </div>
  )
}

interface BriefPreviewFieldProps {
  label: string
  children?: string
  empty?: string
  valueClassName?: string
  emphasize?: boolean
}

export function BriefPreviewField({
  label,
  children,
  empty = '—',
  valueClassName,
  emphasize = true,
}: BriefPreviewFieldProps) {
  const text = children?.trim()
  const hasValue = Boolean(text)

  return (
    <div className="space-y-1.5">
      <p className={briefFieldCaptionClass}>{label}</p>
      <p
        className={cn(
          'text-[14px] leading-relaxed',
          hasValue
            ? emphasize
              ? 'font-medium text-text-primary'
              : 'text-text-secondary'
            : 'text-text-tertiary',
          valueClassName,
        )}
      >
        {text || empty}
      </p>
    </div>
  )
}
