import { ChevronRight, FileText } from 'lucide-react'
import { Link } from 'react-router-dom'

import { useI18n } from '@/hooks/use-i18n'
import {
  heColumnTitleAI,
  heColumnTitleBrand,
  heFileBadge,
  hePanelShell,
  hePanelShellHover,
  heSpring,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { ProfileRole } from './profile-page-header'

interface ProfileFileBadge {
  name: string
  present?: boolean
}

interface ProfileListCardProps {
  role: ProfileRole
  href: string
  title: string
  subtitle?: string
  files: string[] | ProfileFileBadge[]
  meta?: string
}

function normalizeFileBadges(files: string[] | ProfileFileBadge[]): ProfileFileBadge[] {
  return files.map((file) =>
    typeof file === 'string' ? { name: file, present: true } : file,
  )
}

export function ProfileListCard({
  role,
  href,
  title,
  subtitle,
  files,
  meta,
}: ProfileListCardProps) {
  const { t, profileFileCaption } = useI18n()
  const fileBadges = normalizeFileBadges(files)

  return (
    <Link to={href} className="group block">
      <article
        className={cn(
          hePanelShell,
          hePanelShellHover,
          'px-6 py-5',
          heSpring,
        )}
      >
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1 space-y-3">
            <div>
              <p
                className={
                  role === 'principal' ? heColumnTitleBrand : heColumnTitleAI
                }
              >
                {role === 'principal'
                  ? t('profile.list.principalRole')
                  : t('profile.list.participantRole')}
              </p>
              <h2 className="mt-2 truncate text-[17px] font-semibold tracking-[-0.02em] text-text-primary">
                {title}
              </h2>
              {subtitle && (
                <p className="mt-1 truncate font-mono text-xs text-text-tertiary">
                  {subtitle}
                </p>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              {fileBadges.map((file) => (
                <span
                  key={file.name}
                  className={cn(
                    heFileBadge,
                    file.present === false && 'opacity-45 line-through decoration-text-tertiary/50',
                  )}
                >
                  {profileFileCaption(file.name)}
                </span>
              ))}
            </div>
          </div>
          <div className="flex shrink-0 flex-col items-end gap-2 pt-1">
            {meta && (
              <span className="text-[11px] tabular-nums text-text-tertiary">
                {meta}
              </span>
            )}
            <span
              className={cn(
                'inline-flex size-9 items-center justify-center rounded-full',
                'bg-black/[0.02] text-text-tertiary ring-1 ring-inset ring-black/[0.05]',
                'group-hover:bg-brand-soft group-hover:text-brand group-hover:ring-primary/25',
                heSpring,
              )}
            >
              <ChevronRight className="size-4" />
            </span>
          </div>
        </div>
      </article>
    </Link>
  )
}

export function ProfileListSkeleton() {
  const { t } = useI18n()

  return (
    <div className={cn(hePanelShell, 'px-8 py-10')}>
      <div className="flex items-center gap-3 text-sm text-text-secondary">
        <FileText className="size-4 animate-pulse opacity-50" />
        {t('profile.list.loadingIndex')}
      </div>
    </div>
  )
}
