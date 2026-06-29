import { Link } from 'react-router-dom'

import { cn } from '@/lib/utils'

interface RoundTableEmptyHintProps {
  loading: boolean
  rosterFromApi: boolean
  participantCount: number
  className?: string
}

export function RoundTableEmptyHint({
  loading,
  rosterFromApi,
  participantCount,
  className,
}: RoundTableEmptyHintProps) {
  if (participantCount > 0) return null

  let message: string
  if (loading) {
    message = '正在加载专家名录…'
  } else if (!rosterFromApi) {
    message = '暂无 roster 专家；会议中的发言者将临时入座。'
  } else {
    message = '专家名录为空，请先在设置中添加专家。'
  }

  return (
    <div
      className={cn(
        'pointer-events-none absolute inset-x-4 bottom-4 z-20 rounded-xl bg-surface/95 px-4 py-3 text-center shadow-sm ring-1 ring-black/[0.06] backdrop-blur-sm',
        className,
      )}
    >
      <p className="text-[12px] leading-relaxed text-text-secondary">{message}</p>
      {!loading && rosterFromApi && (
        <p className="pointer-events-auto mt-2 text-[11px]">
          <Link to="/participants" className="text-brand hover:underline">
            前往专家管理
          </Link>
          <span className="text-text-tertiary"> · 或发送「!rt 专家 列表」查看</span>
        </p>
      )}
    </div>
  )
}
