import { useEffect, useState } from 'react'

import { fetchMeetings, MEETINGS_PAGE_SIZE } from '@/api/meetings'
import { ApiError } from '@/api/client'
import { Pagination } from '@/components/ui/pagination'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import type { MeetingIndex } from '@/types/meeting'
import { formatDateYMD } from '@/lib/format-date'

function formatUpdatedAt(value: string) {
  if (!value) return '—'
  const formatted = formatDateYMD(value)
  return formatted || '—'
}

export function MeetingsPage() {
  const [meetings, setMeetings] = useState<MeetingIndex[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    fetchMeetings(page, MEETINGS_PAGE_SIZE)
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
  }, [page])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-[22px] font-semibold tracking-tight">会议</h1>
        <p className="mt-1 text-sm text-text-secondary">
          来自 <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">data/workspaces/</code>{' '}
          的 workspace 索引，每页 {MEETINGS_PAGE_SIZE} 条。
        </p>
      </div>

      {loading && (
        <Card>
          <CardContent className="py-8 text-sm text-text-secondary">加载中…</CardContent>
        </Card>
      )}

      {!loading && error && (
        <Card className="border-danger/30">
          <CardHeader>
            <CardTitle>加载失败</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-text-tertiary">
            请确认已运行 <code className="font-mono text-xs">make run</code>（:7777），且 Vite 代理{' '}
            <code className="font-mono text-xs">/api → :7777</code> 生效。
          </CardContent>
        </Card>
      )}

      {!loading && !error && total === 0 && (
        <Card>
          <CardHeader>
            <CardTitle>暂无会议</CardTitle>
            <CardDescription>
              还没有 workspace 目录。可通过 Discord{' '}
              <code className="font-mono text-xs">!rt meet</code> 或{' '}
              <code className="font-mono text-xs">make meet</code> 创建一场会议。
            </CardDescription>
          </CardHeader>
        </Card>
      )}

      {!loading && !error && total > 0 && (
        <>
          <ul className="space-y-3">
            {meetings.map((m) => (
              <li key={m.id}>
                <Card>
                  <CardHeader className="pb-3">
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div className="min-w-0 space-y-1">
                        <CardTitle className="truncate text-base">
                          {m.topic || '（无主题）'}
                        </CardTitle>
                        <CardDescription className="font-mono text-xs">
                          {m.id}
                        </CardDescription>
                      </div>
                      {m.status && (
                        <span className="rounded-md bg-brand-soft px-2 py-0.5 text-xs font-medium text-brand">
                          {m.status}
                        </span>
                      )}
                    </div>
                  </CardHeader>
                  <CardContent className="text-xs text-text-tertiary">
                    更新于 {formatUpdatedAt(m.updated_at)}
                  </CardContent>
                </Card>
              </li>
            ))}
          </ul>
          <Pagination
            page={page}
            pageSize={MEETINGS_PAGE_SIZE}
            total={total}
            onPageChange={setPage}
          />
        </>
      )}
    </div>
  )
}
