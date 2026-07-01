import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { downloadMeetingArchive, deleteMeeting } from '@/api/meetings'

import { PageLayout } from '@/components/layout/page-main-layout'
import { MarkdownTocAside } from '@/components/markdown/markdown-toc'
import { MeetingDeleteDialog } from '@/components/meeting/meeting-delete-dialog'
import { MeetingDetailOverview } from '@/components/meeting/meeting-detail-overview'
import { MeetingDetailSidebar } from '@/components/meeting/meeting-detail-sidebar'
import {
  MeetingDetailViewTabs,
  type MeetingDetailView,
} from '@/components/meeting/meeting-detail-view-tabs'
import { MeetingDocumentsPanel } from '@/components/meeting/meeting-documents-panel'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { ApiError } from '@/api/client'
import { useI18n } from '@/hooks/use-i18n'
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
  const { t } = useI18n()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [detail, setDetail] = useState<MeetingDetail | null>(null)
  const [activeFile, setActiveFile] = useState('')
  const [viewMode, setViewMode] = useState<MarkdownViewMode>('preview')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [articleHeadings, setArticleHeadings] = useState<MarkdownHeading[]>([])
  const [downloading, setDownloading] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

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
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('meetingUi.files.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [load, t])

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

  const handleDownload = useCallback(async () => {
    if (!detail) return
    setActionError(null)
    setDownloading(true)
    try {
      await downloadMeetingArchive(detail.id)
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setActionError(
          t('meetingUi.files.downloadFailedWithStatus', {
            status: err.status,
            message: err.message,
          }),
        )
      } else if (err instanceof Error) {
        setActionError(err.message)
      } else {
        setActionError(t('meetingUi.files.downloadFailed'))
      }
    } finally {
      setDownloading(false)
    }
  }, [detail, t])

  const handleConfirmDelete = useCallback(async () => {
    if (!detail) return
    setActionError(null)
    setDeleting(true)
    try {
      await deleteMeeting(detail.id)
      setDeleteDialogOpen(false)
      navigate(backTo)
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setActionError(
          t('meetingUi.files.deleteFailedWithStatus', { status: err.status, message: err.message }),
        )
      } else if (err instanceof Error) {
        setActionError(err.message)
      } else {
        setActionError(t('meetingUi.files.deleteFailed'))
      }
    } finally {
      setDeleting(false)
    }
  }, [backTo, detail, navigate, t])

  const deleteTopic = brief?.topic?.trim() || detail?.topic?.trim() || detail?.id || ''

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
      <PageLayout
        header={
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
        }
      >
        <div className={cn(hePanelShell, 'px-8 py-10 text-sm text-text-secondary')}>
          {t('meetingUi.files.loading')}
        </div>
      </PageLayout>
    )
  }

  if (error) {
    return (
      <PageLayout
        header={
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
        }
      >
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.loadFailed')}
          description={error}
        />
      </PageLayout>
    )
  }

  if (!detail || !brief || !sidebarProps) {
    return null
  }

  return (
    <>
    <PageLayout
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

      {actionError && (
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.actionFailed')}
          description={actionError}
          className="mb-5"
        />
      )}

      {view === 'overview' ? (
        <MeetingDetailOverview
          detail={detail}
          brief={brief}
          modeKind={modeKind}
          canReplay={canReplay}
          onOpenDocuments={() => openDocuments()}
          onOpenConclusion={brief.conclusion ? openConclusion : undefined}
          onDownload={handleDownload}
          onDelete={() => setDeleteDialogOpen(true)}
          downloading={downloading}
          deleting={deleting}
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
    </PageLayout>

    <MeetingDeleteDialog
      open={deleteDialogOpen}
      onOpenChange={(open) => {
        if (deleting) return
        setDeleteDialogOpen(open)
      }}
      topic={deleteTopic}
      meetingId={detail.id}
      deleting={deleting}
      onConfirm={() => void handleConfirmDelete()}
    />
  </>
  )
}
