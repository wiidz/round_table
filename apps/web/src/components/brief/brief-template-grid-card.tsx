import { ChevronRight, FileStack } from 'lucide-react'
import { Link } from 'react-router-dom'

import {
  heEyebrowBrand,
  heFileBadge,
  hePanelShell,
  hePanelShellHover,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { BriefTemplateIndex } from '@/types/brief-template'

function sourceLabel(source: BriefTemplateIndex['source']): string {
  return source === 'builtin' ? '内置' : '自定义'
}

export function BriefTemplateGridCard({ template }: { template: BriefTemplateIndex }) {
  const description = template.description?.trim()

  return (
    <Link
      to={`/brief-templates/${encodeURIComponent(template.id)}`}
      className={cn('group block h-full', hePressable)}
    >
      <article
        className={cn(
          hePanelShell,
          hePanelShellHover,
          heSpring,
          'flex h-full min-h-[168px] flex-col gap-3 p-5',
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <span
            className={cn(
              heFileBadge,
              template.source === 'builtin'
                ? 'text-text-secondary'
                : 'bg-brand-soft/70 text-brand ring-primary/15',
            )}
          >
            {sourceLabel(template.source)}
          </span>
          <span
            className={cn(
              'inline-flex size-8 items-center justify-center rounded-full',
              'bg-black/[0.02] text-text-tertiary ring-1 ring-inset ring-black/[0.05]',
              'group-hover:bg-brand-soft group-hover:text-brand group-hover:ring-primary/25',
              heSpring,
            )}
          >
            <ChevronRight className="size-4" />
          </span>
        </div>

        <div className="min-w-0 flex-1 space-y-2">
          <div className="flex items-start gap-2">
            <FileStack className="mt-0.5 size-4 shrink-0 text-brand/70" aria-hidden />
            <h2 className="line-clamp-2 text-[16px] font-semibold leading-snug tracking-[-0.02em] text-text-primary group-hover:text-brand">
              {template.title}
            </h2>
          </div>
          <p className="line-clamp-3 min-h-[3.75rem] text-[13px] leading-relaxed text-text-secondary">
            {description || '暂无说明，点击进入编辑模板内容。'}
          </p>
        </div>

        <div className="flex items-center justify-between gap-2 border-t border-black/[0.04] pt-3">
          <span className={heEyebrowBrand}>Meeting Brief</span>
          {template.source === 'custom' && (
            <span className="truncate font-mono text-[10px] text-text-tertiary">
              {template.id}
            </span>
          )}
        </div>
      </article>
    </Link>
  )
}

export function BriefTemplateGridSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 3 }, (_, i) => (
        <div
          key={i}
          className={cn(hePanelShell, 'min-h-[168px] animate-pulse bg-black/[0.02]')}
        />
      ))}
    </div>
  )
}
