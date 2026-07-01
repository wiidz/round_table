import { MeetingDetailConfigPreview } from '@/components/meeting/meeting-detail-config-preview'
import { MeetingFlowDock } from '@/components/meeting/meeting-flow-dock'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'
import type { MeetingModeKind } from '@/lib/meeting-labels'

interface MeetingDetailSidebarProps {
  detail: MeetingDetail
  modeKind?: MeetingModeKind
  className?: string
  onOpenFile?: (path: string) => void
}

/** 配置 + 流程侧栏内容（由 PageThreeColumnLayout 或页面自行包 aside） */
export function MeetingDetailSidebar({
  detail,
  modeKind,
  className,
  onOpenFile,
}: MeetingDetailSidebarProps) {
  return (
    <div className={cn('space-y-5', className)}>
      <MeetingDetailConfigPreview detail={detail} modeKind={modeKind} />
      <MeetingFlowDock detail={detail} sticky={false} onOpenFile={onOpenFile} />
    </div>
  )
}
