import type { ReactNode } from 'react'

import {
  heEyebrowAI,
  heEyebrowBrand,
  hePageDesc,
  hePageTitle,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export type ProfileRole = 'principal' | 'participant'

interface ProfilePageHeaderProps {
  role: ProfileRole
  eyebrow: string
  title: string
  description: ReactNode
}

export function ProfilePageHeader({
  role,
  eyebrow,
  title,
  description,
}: ProfilePageHeaderProps) {
  return (
    <header className="space-y-3">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
        <h1 className={hePageTitle}>{title}</h1>
        <span className={role === 'principal' ? heEyebrowBrand : heEyebrowAI}>
          {eyebrow}
        </span>
      </div>
      <p className={hePageDesc}>{description}</p>
    </header>
  )
}

interface ProfileStatePanelProps {
  variant?: 'default' | 'danger'
  title: string
  description: ReactNode
  className?: string
}

export function ProfileStatePanel({
  variant = 'default',
  title,
  description,
  className,
}: ProfileStatePanelProps) {
  return (
    <div
      className={cn(
        'overflow-hidden rounded-[1.75rem] border-0 bg-surface px-8 py-10',
        'ring-1 ring-[var(--panel-shell-ring)] shadow-[var(--panel-shell-shadow)]',
        variant === 'danger' && 'ring-destructive/25',
        className,
      )}
    >
      <h2 className="text-base font-semibold text-text-primary">{title}</h2>
      <div className="mt-2 text-sm leading-relaxed text-text-secondary">
        {description}
      </div>
    </div>
  )
}
