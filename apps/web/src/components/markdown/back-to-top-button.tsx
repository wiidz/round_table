import { ArrowUp } from 'lucide-react'

import { heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface TocBackToTopProps {
  onClick?: () => void
  className?: string
}

export function TocBackToTop({ onClick, className }: TocBackToTopProps) {
  return (
    <div className={cn('shrink-0 border-t border-border-subtle pt-2', className)}>
      <button
        type="button"
        onClick={() => {
          window.scrollTo({ top: 0, behavior: 'smooth' })
          onClick?.()
        }}
        className={cn(
          'flex w-full items-center gap-1.5 rounded-md py-1.5 text-left',
          'text-[12px] text-text-tertiary',
          heSpring,
          'hover:bg-black/[0.03] hover:text-text-secondary',
        )}
      >
        <ArrowUp className="size-3.5 shrink-0" aria-hidden />
        回到顶部
      </button>
    </div>
  )
}
