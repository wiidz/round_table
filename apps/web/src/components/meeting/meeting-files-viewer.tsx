import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link, useSearchParams } from 'react-router-dom'

import { PageThreeColumnLayout } from '@/components/layout/page-three-column-layout'
import { MarkdownTocAside } from '@/components/markdown/markdown-toc'
import { MeetingDetailOverview } from '@/components/meeting/meeting-detail-overview'
import { MeetingDetailSidebar } from '@/components/meeting/meeting-detail-sidebar'
import {
  MeetingDetailViewTabs,
  type MeetingDetailView,
} from '@/components/meeting/meeting-detail-view-tabs'
import { MeetingDocumentsPanel } from '@/components/meeting/meeting-documents-panel'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import { useMediaQuery } from '@/hooks/use-media-query'
import { hePanelShell, heSpring } from '@/lib/highend-styles'
import { headingsEqual, type MarkdownHeading } from '@/lib/markdown-headings'
import { hasWorkspaceTranscript } from '@/lib/minutes-to-messages'
import { parseMeetingBriefPreview } from '@/lib/meeting-brief-preview'
import {
  defaultMeetingFileSelection,
  groupMeetingFileNames,
  meetingModeKind,
  primaryDeliverablePath,
  type MeetingFileCategory,
} from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

import type { MarkdownViewMode } from '@/components/markdown/markdown-view-toggle'
import type { MeetingDetail } from '@/types/meeting'

interface MeetingFilesViewerProps {
  backTo: string
  backLabel: string
  load: () => Promise<MeetingDetail>
}

function firstFileInCategory(
  fileNames: string[],
  category: MeetingFileCategory,
  modeKind?: ReturnType<typeof meetingModeKind>,
): string {
  const grouped = groupMeetingFileNames(fileNames)
  return grouped[category][0] ?? defaultMeetingFileSelection(fileNames, modeKind)
}

function parseMeetingDetailView(value: string | null): MeetingDetailView {
  return value === 'documents' ? 'documents' : 'overview'
}

export function MeetingFilesViewer({
  backTo,
  backLabel,
  load,
}: MeetingFilesViewerProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [detail, setDetail] = useState<MeetingDetail | null>(null)
  const [activeFile, setActiveFile] = useState('')
  const [viewMode, setViewMode] = useState<MarkdownViewMode>('preview')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [articleHeadings, setArticleHeadings] = useState<MarkdownHeading[]>([])

  const view = parseMeetingDetailView(searchParams.get('view'))
  const documentsWideLayout = useMediaQuery('(min-width: 96rem)')

  const modeKind = useMemo(
    () => meetingModeKind(detail?.mode_kind, detail?.mode),
    [detail?.mode_kind, detail?.mode],
  )

  const fileNames = useMemo(() => {
    if (!detail?.files) return []
    const grouped = groupMeetingFileNames(Object.keys(detail.files))
    return [...grouped.overview, ...grouped.deliverable, ...grouped.process]
  }, [detail])

  const brief = useMemo(
    () => (detail ? parseMeetingBriefPreview(detail, modeKind) : null),
    [detail, modeKind],
  )

  const setView = useCallback(
    (next: MeetingDetailView) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev)
          if (next === 'overview') {
            params.delete('view')
          } else {
            params.set('view', 'documents')
          }
          return params
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )

  const handleHeadingsCollected = useCallback((next: MarkdownHeading[]) => {
    setArticleHeadings((prev) => (headingsEqual(prev, next) ? prev : next))
  }, [])

  useEffect(() => {
    setArticleHeadings([])
  }, [activeFile])

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

  const canReplay = useMemo(
    () => (detail?.files ? hasWorkspaceTranscript(detail.files) : false),
    [detail?.files],
  )

  function openDocuments(category?: MeetingFileCategory) {
    if (!detail) return
    const names = Object.keys(detail.files ?? {})
    const kind = meetingModeKind(detail.mode_kind, detail.mode)
    const nextFile = category
      ? firstFileInCategory(names, category, kind)
      : defaultMeetingFileSelection(names, kind)
    if (nextFile) setActiveFile(nextFile)
    setView('documents')
  }

  function openFile(path: string) {
    setActiveFile(path)
    setView('documents')
  }

  function openConclusion() {
    if (!detail || !modeKind) return
    setActiveFile(primaryDeliverablePath(modeKind))
    setView('documents')
  }

  const sidebarProps = detail
    ? {
        detail,
        modeKind,
        onOpenFile: openFile,
      }
    : null

  const mobileSidebarClass =
    view === 'documents' ? 'min-[96rem]:hidden' : 'xl:hidden'

  if (loading) {
    return (
      <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
        加载会议 workspace…
      </div>
    )
  }

  if (error) {
    return <ProfileStatePanel variant="danger" title="加载失败" description={error} />
  }

  if (!detail || !brief || !sidebarProps) {
    return null
  }

  return (
    <PageThreeColumnLayout
      header={
        <div className="flex flex-col items-start gap-3">
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
          <MeetingDetailViewTabs
            value={view}
            documentCount={fileNames.length}
            onChange={setView}
          />
        </div>
      }
      left={
        view === 'documents' ? <MeetingDetailSidebar {...sidebarProps} /> : undefined
      }
      right={
        view === 'overview' ? (
          <MeetingDetailSidebar {...sidebarProps} />
        ) : (
          <MarkdownTocAside headings={articleHeadings} />
        )
      }
      sidebarFrom={view === 'documents' ? '96rem' : 'xl'}
    >
      {sidebarProps && (
        <div className={cn('mb-5', mobileSidebarClass)}>
          <MeetingDetailSidebar {...sidebarProps} />
        </div>
      )}

      {view === 'overview' ? (
        <MeetingDetailOverview
          detail={detail}
          brief={brief}
          modeKind={modeKind}
          canReplay={canReplay}
          onOpenDocuments={() => openDocuments()}
          onOpenConclusion={brief.conclusion ? openConclusion : undefined}
        />
      ) : (
        <MeetingDocumentsPanel
          detail={detail}
          fileNames={fileNames}
          activeFile={activeFile}
          viewMode={viewMode}
          modeKind={modeKind}
          externalToc={documentsWideLayout}
          onHeadingsCollected={
            documentsWideLayout ? handleHeadingsCollected : undefined
          }
          onSelectFile={setActiveFile}
          onViewModeChange={setViewMode}
        />
      )}
    </PageThreeColumnLayout>
  )
}
