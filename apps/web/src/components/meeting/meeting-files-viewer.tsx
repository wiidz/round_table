import { useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'

import { MeetingDetailHeader } from '@/components/meeting/meeting-detail-header'
import { MarkdownDocument } from '@/components/markdown/markdown-document'
import {
  MarkdownViewToggle,
  type MarkdownViewMode,
} from '@/components/markdown/markdown-view-toggle'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import {
  heFieldLabel,
  heFieldSurface,
  heFilePill,
  heFilePillSelected,
  hePanelShell,
  heSpring,
  heTextarea,
} from '@/lib/highend-styles'
import {
  meetingFileLabel,
  sortMeetingFileNames,
} from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

import type { MeetingDetail } from '@/types/meeting'

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
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fileNames = useMemo(() => {
    if (!detail?.files) return []
    return sortMeetingFileNames(Object.keys(detail.files))
  }, [detail])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    load()
      .then((data) => {
        if (cancelled) return
        setDetail(data)
        const names = sortMeetingFileNames(Object.keys(data.files ?? {}))
        setActiveFile(names[0] ?? '')
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
          <MeetingDetailHeader detail={detail} />

          {fileNames.length === 0 ? (
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
            <div className="grid gap-6 lg:grid-cols-[minmax(0,15rem)_minmax(0,1fr)]">
              <aside className="space-y-2">
                <p className={heFieldLabel}>Workspace 文件</p>
                <nav className="flex flex-col gap-1.5">
                  {fileNames.map((name) => (
                    <button
                      key={name}
                      type="button"
                      onClick={() => setActiveFile(name)}
                      className={cn(
                        activeFile === name ? heFilePillSelected : heFilePill,
                        'w-full truncate text-left',
                      )}
                      title={name}
                    >
                      {meetingFileLabel(name)}
                    </button>
                  ))}
                </nav>
                <p className="pt-1 font-mono text-[10px] text-text-tertiary/80">
                  {activeFile}
                </p>
              </aside>

              <section className="min-w-0 space-y-4">
                <MarkdownViewToggle mode={viewMode} onChange={setViewMode} />
                <div className={cn(heFieldSurface, 'min-h-[420px] p-5 sm:p-6')}>
                  {viewMode === 'preview' ? (
                    <MarkdownDocument content={content} constrained={false} />
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
