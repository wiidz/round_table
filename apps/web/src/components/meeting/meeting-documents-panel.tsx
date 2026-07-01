import { useCallback, useEffect, useState } from 'react'

import { MeetingFileNav } from '@/components/meeting/meeting-file-nav'
import { MarkdownReader } from '@/components/markdown/markdown-reader'
import {
  MarkdownViewToggle,
  type MarkdownViewMode,
} from '@/components/markdown/markdown-view-toggle'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import {
  heFieldHint,
  heFieldSurface,
  heTextarea,
} from '@/lib/highend-styles'
import {
  meetingFileCaption,
  meetingFileCategory,
  meetingFileDescription,
  MEETING_FILE_CATEGORY_LABELS,
  type MeetingModeKind,
} from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

import type { MarkdownHeading } from '@/lib/markdown-headings'
import type { MeetingDetail } from '@/types/meeting'

interface MeetingDocumentsPanelProps {
  detail: MeetingDetail
  fileNames: string[]
  activeFile: string
  viewMode: MarkdownViewMode
  modeKind?: MeetingModeKind
  /** 宽屏三栏时由右侧 aside 渲染文章目录 */
  externalToc?: boolean
  onHeadingsCollected?: (headings: MarkdownHeading[]) => void
  onSelectFile: (path: string) => void
  onViewModeChange: (mode: MarkdownViewMode) => void
  className?: string
}

export function MeetingDocumentsPanel({
  detail,
  fileNames,
  activeFile,
  viewMode,
  modeKind,
  externalToc = false,
  onHeadingsCollected,
  onSelectFile,
  onViewModeChange,
  className,
}: MeetingDocumentsPanelProps) {
  const content = activeFile && detail.files ? detail.files[activeFile] ?? '' : ''

  return (
    <section className={cn('space-y-4', className)}>
      <div className="space-y-1">
        <h2 className="text-[15px] font-semibold text-text-primary">Workspace 文档</h2>
        <p className={heFieldHint}>浏览本场 Workspace 产出的 Markdown 文件</p>
      </div>

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
        <div className="grid gap-6 lg:grid-cols-[minmax(0,13rem)_minmax(0,1fr)]">
          <aside>
            <MeetingFileNav
              names={fileNames}
              activeFile={activeFile}
              files={detail.files}
              modeKind={modeKind}
              onSelect={onSelectFile}
            />
          </aside>

          <div className="min-w-0 space-y-4">
            {activeFile && (
              <div className="space-y-2">
                <p className="font-mono text-[12px] text-text-secondary">
                  {meetingFileCaption(activeFile, modeKind)}
                  <span className="ml-2 text-text-tertiary">
                    · {MEETING_FILE_CATEGORY_LABELS[meetingFileCategory(activeFile)]}
                  </span>
                </p>
                <p className={heFieldHint}>{meetingFileDescription(activeFile, modeKind)}</p>
              </div>
            )}
            <MarkdownViewToggle mode={viewMode} onChange={onViewModeChange} />
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
                  tocInGutter={externalToc}
                  onHeadingsCollected={onHeadingsCollected}
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
          </div>
        </div>
      )}
    </section>
  )
}
