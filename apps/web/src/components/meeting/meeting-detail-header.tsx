import type { ReactNode } from 'react'
import {
  Bot,
  Layers,
  MessageCircle,
  MessageCircleOff,
  Users,
} from 'lucide-react'

import {
  MeetingModeInline,
  meetingModeFreeDialogueClass,
} from '@/components/meeting/meeting-mode-badge'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { hePageDesc, hePageTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

interface MeetingDetailHeaderProps {
  detail: MeetingDetail
  canReplay?: boolean
}

function MetaDot() {
  return <span className="hidden text-text-tertiary/35 sm:inline" aria-hidden>·</span>
}

function formatTokenCount(value: number): string {
  if (value >= 1_000_000) {
    const compact = value / 1_000_000
    return `${compact >= 10 ? Math.round(compact) : compact.toFixed(1)}M`
  }
  if (value >= 10_000) {
    const compact = value / 1_000
    return `${compact >= 100 ? Math.round(compact) : compact.toFixed(1)}k`
  }
  return value.toLocaleString('zh-CN')
}

function MetaItem({
  icon: Icon,
  children,
  className,
}: {
  icon: typeof Users
  children: ReactNode
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 tabular-nums text-text-tertiary',
        className,
      )}
    >
      <Icon className="size-3 shrink-0 opacity-55" aria-hidden />
      {children}
    </span>
  )
}

export function MeetingDetailHeader({ detail, canReplay }: MeetingDetailHeaderProps) {
  const topic = detail.topic?.trim() || '（无主题）'
  const startedAt = detail.started_at?.trim()
  const participants = detail.participant_count ?? 0
  const rounds = detail.max_rounds ?? 0
  const totalTokens = detail.total_tokens ?? 0

  return (
    <header className="space-y-4">
      <p className="text-[10px] font-medium uppercase tracking-[0.18em] text-text-tertiary">
        会议复盘
      </p>

      <h1 className={cn(hePageTitle, 'text-balance')}>{topic}</h1>

      <p className={hePageDesc}>
        {canReplay
          ? '在「文档」中浏览 Workspace 产出；在「回放」中以圆桌复盘讨论；在「流程」中查看会前准备到结案的 Engine 路径。'
          : '在「文档」中浏览 Workspace 产出；在「流程」中查看会前准备到结案的 Engine 路径。'}
      </p>

      <div className="flex flex-wrap items-center gap-x-2.5 gap-y-2 text-[13px]">
        <MeetingModeInline mode={detail.mode} modeKind={detail.mode_kind} />
        {detail.status && (
          <>
            <MetaDot />
            <MeetingStatusBadge status={detail.status} />
          </>
        )}
        <MetaDot />
        <MetaItem icon={Users}>
          {participants > 0 ? `${participants} 人` : '— 人'}
        </MetaItem>
        <MetaDot />
        <MetaItem icon={Layers}>
          {rounds > 0 ? `${rounds} 轮` : '— 轮'}
        </MetaItem>
        <MetaDot />
        <MetaItem
          icon={detail.free_dialogue ? MessageCircle : MessageCircleOff}
          className={cn(
            detail.free_dialogue &&
              meetingModeFreeDialogueClass(detail.mode_kind, detail.mode),
          )}
        >
          {detail.free_dialogue ? '自由对话' : '无自由'}
        </MetaItem>
        <MetaDot />
        <span className="inline-flex items-center gap-1 tabular-nums text-text-tertiary">
          <Bot className="size-3 shrink-0 text-ai/55" aria-hidden />
          <span className="font-mono text-[12px]">
            <span className="text-ai/75">
              {totalTokens > 0 ? formatTokenCount(totalTokens) : '—'}
            </span>
            <span className="text-text-tertiary/45"> tok</span>
          </span>
        </span>
        <MetaDot />
        <span className="font-mono text-[12px] text-text-tertiary">{detail.id}</span>
        {startedAt && (
          <>
            <MetaDot />
            <span className="tabular-nums text-text-tertiary">{startedAt}</span>
          </>
        )}
      </div>
    </header>
  )
}
