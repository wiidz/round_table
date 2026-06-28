import { MeetingModeInline } from '@/components/meeting/meeting-mode-badge'
import { MeetingStatusBadge } from '@/components/meeting/meeting-status-badge'
import { hePageDesc, hePageTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

interface MeetingDetailHeaderProps {
  detail: MeetingDetail
}

function MetaDot() {
  return <span className="hidden text-text-tertiary/35 sm:inline" aria-hidden>·</span>
}

export function MeetingDetailHeader({ detail }: MeetingDetailHeaderProps) {
  const topic = detail.topic?.trim() || '（无主题）'
  const startedAt = detail.started_at?.trim()

  return (
    <header className="space-y-4">
      <p className="text-[10px] font-medium uppercase tracking-[0.18em] text-text-tertiary">
        Meeting Review
      </p>

      <h1 className={cn(hePageTitle, 'text-balance')}>{topic}</h1>

      <p className={hePageDesc}>
        阅读 MEETING.md、MINUTES 与 Artifacts。默认渲染 Markdown，可切换查看源码。
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
