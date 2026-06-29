import { Users } from 'lucide-react'

import { heSubsectionTitleNeutral } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface RoundTableStagePlaceholderProps {
  turnCount: number
  className?: string
}

/** M2 will replace this with RoundTableStage. */
export function RoundTableStagePlaceholder({
  turnCount,
  className,
}: RoundTableStagePlaceholderProps) {
  return (
    <div
      className={cn(
        'flex min-h-0 flex-1 flex-col items-center justify-center px-6 py-8 text-center',
        className,
      )}
    >
      <div className="flex size-16 items-center justify-center rounded-2xl bg-ai-soft ring-1 ring-ai/15">
        <Users className="size-8 text-ai/70" aria-hidden />
      </div>
      <h3 className={cn(heSubsectionTitleNeutral, 'mt-4')}>圆桌 Live 视图</h3>
      <p className="mt-2 max-w-sm text-[13px] leading-relaxed text-text-tertiary">
        围坐发言与席位气泡将在下一阶段（M2）呈现。当前可在下方发言记录查看摘要，点击打开详情 Drawer。
      </p>
      {turnCount > 0 && (
        <p className="mt-3 font-mono text-[12px] text-text-tertiary">
          已记录 {turnCount} 轮专家/司仪发言
        </p>
      )}
    </div>
  )
}
