import { useEffect, useMemo, useState } from 'react'
import { ArrowLeft, FileText, Users } from 'lucide-react'
import { Link } from 'react-router-dom'

import { MeetingDetailHeader } from '@/components/meeting/meeting-detail-header'
import { MeetingFileNav } from '@/components/meeting/meeting-file-nav'
import { MeetingReplayViewer } from '@/components/meeting/meeting-replay-viewer'
import { MarkdownReader } from '@/components/markdown/markdown-reader'
import {
  MarkdownViewToggle,
  type MarkdownViewMode,
} from '@/components/markdown/markdown-view-toggle'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import {
  heFieldSurface,
  heFieldHint,
  hePanelShell,
  heSpring,
  heTextarea,
} from '@/lib/highend-styles'
import {
  hasWorkspaceTranscript,
  workspaceTranscriptMessages,
} from '@/lib/minutes-to-messages'
import {
  defaultMeetingFileSelection,
  groupMeetingFileNames,
  meetingFileCaption,
  meetingFileCategory,
  meetingFileDescription,
  meetingModeKind,
  MEETING_FILE_CATEGORY_LABELS,
} from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

type MeetingDetailView = 'files' | 'replay'

interface MeetingFilesViewerProps {
  backTo: string
  backLabel: string
  load: () => Promise<MeetingDetail>
}

export function MeetingFilesViewer({
  backTo,
  backLabel,
  load,
}: MeetingFilesViewerProps) {
  const [detail, setDetail] = useState<MeetingDetail | null>(null)
  const [activeFile, setActiveFile] = useState('')
  const [viewMode, setViewMode] = useState<MarkdownViewMode>('preview')
  const [detailView, setDetailView] = useState<MeetingDetailView>('files')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const modeKind = useMemo(
    () => meetingModeKind(detail?.mode_kind, detail?.mode),
    [detail?.mode_kind, detail?.mode],
  )

  const fileNames = useMemo(() => {
    if (!detail?.files) return []
    const grouped = groupMeetingFileNames(Object.keys(detail.files))
    return [...grouped.overview, ...grouped.deliverable, ...grouped.process]
  }, [detail])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    load()
      .then((data) => {
        if (cancelled) return
        setDetail(data)
        const names = Object.keys(data.files ?? {})
        const kind = meetingModeKind(data.mode_kind, data.mode)
        setActiveFile(defaultMeetingFileSelection(names, kind))
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载会议详情')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [load])

  const content = activeFile && detail?.files ? detail.files[activeFile] ?? '' : ''

  const canReplay = useMemo(
    () => (detail?.files ? hasWorkspaceTranscript(detail.files) : false),
    [detail?.files],
  )

  const replayMessages = useMemo(() => {
    if (!detail?.files || !canReplay) return []
    return workspaceTranscriptMessages(detail.files, detail.id, detail.started_at)
  }, [detail, canReplay])

  const meetingMd = detail?.files?.['MEETING.md'] ?? ''
  const topic = detail?.topic?.trim() || '（无主题）'

  return (
    <div className="space-y-8">
      <Link
        to={backTo}
        className={cn(
          'inline-flex items-center gap-2 text-sm text-text-secondary transition-colors hover:text-brand',
          heSpring,
        )}
      >
        <ArrowLeft className="size-4" />
        {backLabel}
      </Link>

      {loading && (
        <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
          加载会议 workspace…
        </div>
      )}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && detail && (
        <>
          <MeetingDetailHeader detail={detail} canReplay={canReplay} />

          {canReplay && (
            <div
              className="flex rounded-lg bg-black/[0.04] p-0.5 ring-1 ring-inset ring-black/[0.06]"
              role="group"
              aria-label="详情视图"
            >
              <button
                type="button"
                onClick={() => setDetailView('files')}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-[12px] font-medium transition-colors',
                  detailView === 'files'
                    ? 'bg-surface text-brand shadow-sm'
                    : 'text-text-tertiary hover:text-text-secondary',
                )}
              >
                <FileText className="size-3.5" aria-hidden />
                文档
              </button>
              <button
                type="button"
                onClick={() => setDetailView('replay')}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-[12px] font-medium transition-colors',
                  detailView === 'replay'
                    ? 'bg-surface text-brand shadow-sm'
                    : 'text-text-tertiary hover:text-text-secondary',
                )}
              >
                <Users className="size-3.5" aria-hidden />
                回放
              </button>
            </div>
          )}

          {detailView === 'replay' && canReplay ? (
            <div className="relative overflow-visible">
              <MeetingReplayViewer
                topic={topic}
                meetingMd={meetingMd}
                messages={replayMessages}
                className="h-[calc(100vh-16rem)] min-h-[28rem]"
              />
            </div>
          ) : fileNames.length === 0 ? (
            <ProfileStatePanel
              title="暂无 Markdown 文件"
              description={
                <>
                  在{' '}
                  <code className="font-mono text-xs">data/workspaces/{detail.id}/</code>{' '}
                  下尚未生成可读文档。
                </>
              }
            />
          ) : (
            <div className="grid gap-6 lg:grid-cols-[minmax(0,13rem)_minmax(0,1fr)]">
              <aside>
                <MeetingFileNav
                  names={fileNames}
                  activeFile={activeFile}
                  files={detail.files}
                  modeKind={modeKind}
                  onSelect={setActiveFile}
                />
              </aside>

              <section className="min-w-0 space-y-4 overflow-visible">
                {activeFile && (
                  <div className="space-y-2">
                    <p className="font-mono text-[12px] text-text-secondary">
                      {meetingFileCaption(activeFile, modeKind)}
                      <span className="ml-2 text-text-tertiary">
                        · {MEETING_FILE_CATEGORY_LABELS[meetingFileCategory(activeFile)]}
                      </span>
                    </p>
                    <p className={heFieldHint}>
                      {meetingFileDescription(activeFile, modeKind)}
                    </p>
                  </div>
                )}
                <MarkdownViewToggle mode={viewMode} onChange={setViewMode} />
                <div
                  className={cn(
                    heFieldSurface,
                    'relative min-h-[420px] overflow-visible p-5 sm:p-6',
                  )}
                >
                  {viewMode === 'preview' ? (
                    <MarkdownReader
                      key={activeFile}
                      documentKey={activeFile}
                      content={content}
                      constrained={false}
                    />
                  ) : (
                    <textarea
                      readOnly
                      value={content}
                      className={cn(heTextarea, 'min-h-[480px] cursor-default p-0')}
                      spellCheck={false}
                    />
                  )}
                </div>
              </section>
            </div>
          )}
        </>
      )}
    </div>
  )
}
