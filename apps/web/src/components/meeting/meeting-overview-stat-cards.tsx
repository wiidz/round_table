import type { LucideIcon } from 'lucide-react'
import { Bot, FileText, Layers, Users } from 'lucide-react'

import {
  buildMeetingOverviewStats,
  formatMeetingRoundsHint,
  formatMeetingRoundsValue,
  formatTokenCount,
} from '@/lib/meeting-overview-stats'
import { type MeetingModeKind } from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

interface MeetingOverviewStatCardsProps {
  detail: MeetingDetail
  modeKind?: MeetingModeKind
  className?: string
}

function StatCard({
  label,
  value,
  hint,
  icon: Icon,
  accent = 'neutral',
}: {
  label: string
  value: string
  hint?: string
  icon: LucideIcon
  accent?: 'brand' | 'ai' | 'neutral'
}) {
  return (
    <article
      className={cn(
        'flex min-w-0 flex-col gap-2 rounded-lg bg-black/[0.025] px-4 py-3.5',
        'ring-1 ring-inset ring-black/[0.05]',
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <p className="text-[11px] font-medium uppercase tracking-[0.12em] text-text-tertiary">
          {label}
        </p>
        <span
          className={cn(
            'inline-flex size-7 shrink-0 items-center justify-center rounded-md',
            accent === 'brand' && 'bg-brand-soft text-brand',
            accent === 'ai' && 'bg-ai-soft text-ai',
            accent === 'neutral' && 'bg-black/[0.04] text-text-secondary',
          )}
        >
          <Icon className="size-3.5" strokeWidth={1.75} aria-hidden />
        </span>
      </div>
      <p className="text-[22px] font-semibold leading-none tracking-[-0.03em] tabular-nums text-text-primary">
        {value}
      </p>
      {hint && <p className="text-[12px] leading-relaxed text-text-tertiary">{hint}</p>}
    </article>
  )
}

export function MeetingOverviewStatCards({
  detail,
  modeKind,
  className,
}: MeetingOverviewStatCardsProps) {
  const stats = buildMeetingOverviewStats(detail, modeKind)
  const { deliverable, usage, experts, rounds } = stats

  const deliverableValue = deliverable.available
    ? `${deliverable.charCount.toLocaleString('zh-CN')} 字`
    : '—'
  const deliverableHint = deliverable.available
    ? `约 ${deliverable.readingMinutes} 分钟 · ${deliverable.title}`
    : '主交付物尚未产出'

  const usageValue =
    usage.totalTokens > 0 ? formatTokenCount(usage.totalTokens) : '—'
  const usageHint =
    usage.llmCallCount > 0
      ? `${usage.llmCallCount.toLocaleString('zh-CN')} 次 LLM 调用`
      : '暂无 Token 用量'

  const expertsValue = experts.count > 0 ? `${experts.count} 人` : '—'
  const expertsHint = experts.count > 0 ? '参会专家' : '尚未记录阵容'

  const roundsValue = formatMeetingRoundsValue(
    rounds.maxRounds,
    rounds.freeDialogueQuestions,
  )
  const roundsHint = formatMeetingRoundsHint(
    rounds.maxRounds,
    rounds.freeDialogueQuestions,
  )

  return (
    <div className={cn('grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4', className)}>
      <StatCard
        label="交付物"
        value={deliverableValue}
        hint={deliverableHint}
        icon={FileText}
        accent="brand"
      />
      <StatCard
        label="专家"
        value={expertsValue}
        hint={expertsHint}
        icon={Users}
      />
      <StatCard
        label="辩论轮次"
        value={roundsValue}
        hint={roundsHint}
        icon={Layers}
      />
      <StatCard
        label="Token 用量"
        value={usageValue}
        hint={usageHint}
        icon={Bot}
        accent="ai"
      />
    </div>
  )
}
