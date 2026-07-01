import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'

import { fetchMeeting } from '@/api/meetings'
import { PageLayout } from '@/components/layout/page-main-layout'
import { MeetingReplayViewer } from '@/components/meeting/meeting-replay-viewer'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import { hePanelShell, heSpring } from '@/lib/highend-styles'
import {
  hasWorkspaceTranscript,
  workspaceTranscriptMessages,
} from '@/lib/minutes-to-messages'
import { cn } from '@/lib/utils'

export function MeetingReplayPage() {
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
          setError('本场暂无可回放 transcript')
          return
        }
        setTopic(detail.topic?.trim() || '（无主题）')
        setMeetingMd(detail.files?.['MEETING.md'] ?? '')
        setMessages(workspaceTranscriptMessages(detail.files, detail.id, detail.started_at))
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载会议回放')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id, load])

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
      返回会议详情
    </Link>
  )

  if (!id) {
    return null
  }

  if (loading) {
    return (
      <PageLayout header={backLink}>
        <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
          加载回放…
        </div>
      </PageLayout>
    )
  }

  if (error) {
    return (
      <PageLayout header={backLink}>
        <ProfileStatePanel variant="danger" title="无法回放" description={error} />
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
