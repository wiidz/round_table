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
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { Pagination } from '@/components/ui/pagination'
import { useI18n } from '@/hooks/use-i18n'
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
  const { t } = useI18n()
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
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('pages.meetings.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [page, pageSize, t])

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={t('pages.meetings.eyebrow')}
          title={t('pages.meetings.title')}
          description={t('pages.meetings.description')}
        />
      }
    >
    <div className="space-y-8">
      {loading && <MeetingGridSkeleton count={pageSize} />}

      {!loading && error && (
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.loadFailed')}
          description={error}
        />
      )}

      {!loading && !error && total === 0 && (
        <ProfileStatePanel
          title={t('pages.meetings.emptyTitle')}
          description={t('pages.meetings.emptyDescription')}
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
    </PageLayout>
  )
}
