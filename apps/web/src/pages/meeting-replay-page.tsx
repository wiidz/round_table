import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'

import { fetchMeeting } from '@/api/meetings'
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

  if (!id) {
    return null
  }

  return (
    <div className="flex h-[calc(100vh-4rem)] min-h-[32rem] flex-col gap-4">
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

      {loading && (
        <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
          加载回放…
        </div>
      )}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="无法回放" description={error} />
      )}

      {!loading && !error && (
        <MeetingReplayViewer
          topic={topic}
          meetingMd={meetingMd}
          messages={messages}
          className="min-h-0 flex-1"
        />
      )}
    </div>
  )
}
