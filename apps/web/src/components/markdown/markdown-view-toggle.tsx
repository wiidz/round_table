import { cn } from '@/lib/utils'
import { heFilePill, heFilePillSelected, hePressable } from '@/lib/highend-styles'

export type MarkdownViewMode = 'preview' | 'source'

interface MarkdownViewToggleProps {
  mode: MarkdownViewMode
  onChange: (mode: MarkdownViewMode) => void
  className?: string
}

export function MarkdownViewToggle({ mode, onChange, className }: MarkdownViewToggleProps) {
  return (
    <div className={cn('inline-flex flex-wrap gap-2', className)}>
      <button
        type="button"
        onClick={() => onChange('preview')}
        className={cn(
          hePressable,
          mode === 'preview' ? heFilePillSelected : heFilePill,
        )}
      >
        预览
      </button>
      <button
        type="button"
        onClick={() => onChange('source')}
        className={cn(
          hePressable,
          mode === 'source' ? heFilePillSelected : heFilePill,
        )}
      >
        源码
      </button>
    </div>
  )
}
