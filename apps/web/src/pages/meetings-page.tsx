import { useEffect, useState } from 'react'

import {
  fetchMeetings,
  MEETINGS_PAGE_SIZE,
  MEETINGS_PAGE_SIZE_DESKTOP,
} from '@/api/meetings'
import { ApiError } from '@/api/client'
import {
  MeetingGridCard,
  MeetingGridSkeleton,
} from '@/components/meeting/meeting-list-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { Pagination } from '@/components/ui/pagination'
import type { MeetingIndex } from '@/types/meeting'

const DESKTOP_MEDIA = '(min-width: 1280px)'

function useMeetingsPageSize() {
  const [pageSize, setPageSize] = useState(() =>
    typeof window !== 'undefined' && window.matchMedia(DESKTOP_MEDIA).matches
      ? MEETINGS_PAGE_SIZE_DESKTOP
      : MEETINGS_PAGE_SIZE,
  )

  useEffect(() => {
    const mq = window.matchMedia(DESKTOP_MEDIA)
    const sync = () => {
      setPageSize(mq.matches ? MEETINGS_PAGE_SIZE_DESKTOP : MEETINGS_PAGE_SIZE)
    }
    sync()
    mq.addEventListener('change', sync)
    return () => mq.removeEventListener('change', sync)
  }, [])

  return pageSize
}

export function MeetingsPage() {
  const pageSize = useMeetingsPageSize()
  const [meetings, setMeetings] = useState<MeetingIndex[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setPage(1)
  }, [pageSize])

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    fetchMeetings(page, pageSize)
      .then((data) => {
        if (!cancelled) {
          setMeetings(data.meetings ?? [])
          setTotal(data.total ?? 0)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载会议列表')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [page, pageSize])

  return (
    <div className="space-y-8">
      <ProfilePageHeader
        role="principal"
        eyebrow="Meeting Browse"
        title="会议"
        description={
          <>
            浏览{' '}
            <code className="rounded-md bg-black/[0.04] px-1.5 py-0.5 font-mono text-[12px] ring-1 ring-inset ring-black/[0.05]">
              data/workspaces/
            </code>{' '}
            下的历史 Meeting。**天平**（明亮橙底）为裁决型、**云朵**为研讨型；卡片含人数、轮次与自由对话配置。
          </>
        }
      />

      {loading && <MeetingGridSkeleton count={pageSize} />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && total === 0 && (
        <ProfileStatePanel
          title="暂无会议"
          description={
            <>
              还没有 workspace 目录。可通过 Discord{' '}
              <code className="font-mono text-xs">!rt meet</code> 或{' '}
              <code className="font-mono text-xs">make meet</code> 创建一场会议。
            </>
          }
        />
      )}

      {!loading && !error && total > 0 && (
        <>
          <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {meetings.map((m) => (
              <li key={m.id} className="min-w-0">
                <MeetingGridCard meeting={m} />
              </li>
            ))}
          </ul>
          <Pagination
            page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={setPage}
          />
        </>
      )}
    </div>
  )
}
