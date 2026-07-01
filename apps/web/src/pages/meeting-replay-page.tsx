import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'

import { fetchMeeting } from '@/api/meetings'
import { PageLayout } from '@/components/layout/page-main-layout'
import { MeetingReplayViewer } from '@/components/meeting/meeting-replay-viewer'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import { useI18n } from '@/hooks/use-i18n'
import { hePanelShell, heSpring } from '@/lib/highend-styles'
import {
  hasWorkspaceTranscript,
  workspaceTranscriptMessages,
} from '@/lib/minutes-to-messages'
import { cn } from '@/lib/utils'

export function MeetingReplayPage() {
  const { t } = useI18n()
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const load = useCallback(async () => fetchMeeting(id), [id])

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [topic, setTopic] = useState('')
  const [meetingMd, setMeetingMd] = useState('')
  const [messages, setMessages] = useState<ReturnType<typeof workspaceTranscriptMessages>>([])

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    load()
      .then((detail) => {
        if (cancelled) return
        if (!hasWorkspaceTranscript(detail.files ?? {})) {
          setError(t('pages.meetingReplay.noTranscript'))
          return
        }
        setTopic(detail.topic?.trim() || t('meeting.topicEmpty'))
        setMeetingMd(detail.files?.['MEETING.md'] ?? '')
        setMessages(workspaceTranscriptMessages(detail.files, detail.id, detail.started_at))
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('pages.meetingReplay.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id, load, t])

  const detailPath = useMemo(
    () => (id ? `/meetings/${encodeURIComponent(id)}` : '/meetings'),
    [id],
  )

  const backLink = (
    <Link
      to={detailPath}
      className={cn(
        'inline-flex shrink-0 items-center gap-2 text-sm text-text-secondary transition-colors hover:text-brand',
        heSpring,
      )}
    >
      <ArrowLeft className="size-4" />
      {t('pages.meetingReplay.back')}
    </Link>
  )

  if (!id) {
    return null
  }

  if (loading) {
    return (
      <PageLayout header={backLink}>
        <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
          {t('pages.meetingReplay.loading')}
        </div>
      </PageLayout>
    )
  }

  if (error) {
    return (
      <PageLayout header={backLink}>
        <ProfileStatePanel
          variant="danger"
          title={t('pages.meetingReplay.cannotReplay')}
          description={error}
        />
      </PageLayout>
    )
  }

  return (
    <MeetingReplayViewer
      topic={topic}
      meetingMd={meetingMd}
      messages={messages}
      pageShell={({ main, left, right, drawer }) => (
        <PageLayout
          header={backLink}
          left={left}
          right={right}
          sidebarFrom="96rem"
          sideColumnWidth="gutter"
          className="min-[96rem]:min-h-[calc(100vh-7.5rem)]"
          bodyClassName="min-[96rem]:h-full"
        >
          <div className="h-full min-h-0">{main}</div>
          {drawer}
        </PageLayout>
      )}
    />
  )
}
