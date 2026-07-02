import { useEffect, useState } from 'react'
import {
  ChevronDown,
  FolderOpen,
  LayoutList,
  Package,
  type LucideIcon,
} from 'lucide-react'

import { useI18n } from '@/hooks/use-i18n'
import {
  heFieldLabel,
  heFileNavItem,
  heFileNavItemSelected,
  hePrimaryDeliverableBadge,
  hePrimaryDeliverableBadgeOnBrand,
  heSpring,
} from '@/lib/highend-styles'
import {
  groupMeetingFileNames,
  isPrimaryDeliverable,
  meetingFileCategory,
  meetingFileHasTitle,
  type MeetingFileCategory,
  type MeetingModeKind,
} from '@/lib/meeting-labels'
import { cn } from '@/lib/utils'

const FILE_CATEGORY_ORDER: MeetingFileCategory[] = ['overview', 'deliverable', 'process']

const FILE_CATEGORY_ICONS: Record<MeetingFileCategory, LucideIcon> = {
  overview: LayoutList,
  deliverable: Package,
  process: FolderOpen,
}

interface MeetingFileNavProps {
  names: string[]
  activeFile: string
  files?: Record<string, string>
  modeKind?: MeetingModeKind
  onSelect: (path: string) => void
}

interface FileNavSectionProps {
  category: MeetingFileCategory
  items: string[]
  activeFile: string
  files?: Record<string, string>
  modeKind?: MeetingModeKind
  collapsed?: boolean
  onToggle?: () => void
  onSelect: (path: string) => void
}

function FileNavSectionHeader({
  category,
  count,
  collapsed,
  collapsible,
  onToggle,
}: {
  category: MeetingFileCategory
  count: number
  collapsed?: boolean
  collapsible?: boolean
  onToggle?: () => void
}) {
  const { t } = useI18n()
  const isDeliverable = category === 'deliverable'
  const isProcess = category === 'process'
  const Icon = FILE_CATEGORY_ICONS[category]

  const title = (
    <span className="flex min-w-0 items-center gap-1.5">
      <Icon
        className={cn(
          'size-3.5 shrink-0',
          isDeliverable && 'text-brand',
          isProcess && 'text-text-secondary',
          !isDeliverable && !isProcess && 'text-text-tertiary',
        )}
        aria-hidden
      />
      <span
        className={cn(
          heFieldLabel,
          'min-w-0 truncate',
          isDeliverable && 'text-brand',
          isProcess && 'text-text-secondary',
        )}
      >
        {t(`meeting.fileCategory.${category}`)}
      </span>
    </span>
  )

  const countBadge = (
    <span
      className={cn(
        'shrink-0 tabular-nums text-[11px]',
        isDeliverable && 'text-brand/65',
        isProcess && 'text-text-tertiary',
        !isDeliverable && !isProcess && 'text-text-tertiary/70',
      )}
    >
      {t('meetingUi.fileNav.sectionCount', { count })}
    </span>
  )

  if (collapsible) {
    return (
      <button
        type="button"
        onClick={onToggle}
        className={cn(
          'flex w-full items-center gap-2 rounded-md px-1 py-0.5 text-left',
          heSpring,
          'hover:bg-black/[0.03]',
        )}
      >
        {title}
        <span className="ml-auto flex shrink-0 items-center gap-1.5">
          {countBadge}
          <ChevronDown
            className={cn(
              'size-3.5 text-text-tertiary transition-transform',
              collapsed && '-rotate-90',
            )}
            aria-hidden
          />
        </span>
      </button>
    )
  }

  return (
    <div className="flex items-center gap-2 px-1">
      {title}
      <span className="ml-auto">{countBadge}</span>
    </div>
  )
}
function FileNavActiveStats({ content }: { content: string }) {
  const { formatMarkdownReadingStats } = useI18n()

  return (
    <>
      <span
        className="mt-1.5 block border-t border-black/[0.08] pt-1.5"
        aria-hidden
      />
      <span className="block text-[10px] tabular-nums text-text-tertiary">
        {formatMarkdownReadingStats(content)}
      </span>
    </>
  )
}

function FileNavItem({
  name,
  activeFile,
  category,
  files,
  modeKind,
  onSelect,
}: {
  name: string
  activeFile: string
  category: MeetingFileCategory
  files?: Record<string, string>
  modeKind?: MeetingModeKind
  onSelect: (path: string) => void
}) {
  const { t, meetingFileCaption, meetingFileLabel } = useI18n()
  const isActive = activeFile === name
  const isPrimary =
    category === 'deliverable' && isPrimaryDeliverable(name, modeKind)
  const isProcess = category === 'process'
  const content = isActive ? (files?.[name] ?? '') : ''

  return (
    <button
      type="button"
      onClick={() => onSelect(name)}
      className={cn(
        'w-full text-left',
        heSpring,
        isProcess
          ? cn(
              'rounded-md px-2 py-1.5',
              isActive
                ? 'bg-black/[0.05] font-medium text-text-secondary ring-1 ring-inset ring-black/[0.08]'
                : 'text-text-secondary/85 hover:bg-black/[0.03] hover:text-text-secondary',
            )
          : isActive
            ? heFileNavItemSelected
            : heFileNavItem,
      )}
      title={meetingFileCaption(name, modeKind)}
    >
      {meetingFileHasTitle(name, modeKind) ? (
        <span className="flex min-w-0 flex-col gap-0.5">
          <span className="flex min-w-0 items-center gap-2">
            <span className="min-w-0 truncate text-[13px]">
              {meetingFileLabel(name, modeKind)}
            </span>
            {isPrimary && (
              <span
                className={cn(
                  isActive
                    ? hePrimaryDeliverableBadgeOnBrand
                    : hePrimaryDeliverableBadge,
                  'ml-auto shrink-0',
                )}
              >
                {t('meetingUi.fileNav.primaryDeliverable')}
              </span>
            )}
          </span>
          <span
            className={cn(
              'truncate font-mono text-[10px] text-text-tertiary/90',
            )}
          >
            {name}
          </span>
          {isActive && <FileNavActiveStats content={content} />}
        </span>
      ) : (
        <span className="flex min-w-0 flex-col gap-0.5">
          <span
            className={cn(
              'block truncate font-mono',
              isProcess ? 'text-[12px] text-text-tertiary' : 'text-[12px]',
            )}
          >
            {name}
          </span>
          {isActive && <FileNavActiveStats content={content} />}
        </span>
      )}
    </button>
  )
}

function FileNavSection({
  category,
  items,
  activeFile,
  files,
  modeKind,
  collapsed,
  onToggle,
  onSelect,
}: FileNavSectionProps) {
  if (items.length === 0) return null

  const collapsible = category === 'process'
  const isCollapsed = collapsible && collapsed

  return (
    <section>
      <FileNavSectionHeader
        category={category}
        count={items.length}
        collapsed={isCollapsed}
        collapsible={collapsible}
        onToggle={onToggle}
      />
      {!isCollapsed && (
        <nav
          className={cn(
            'mt-2 flex flex-col',
            category === 'process' ? 'gap-0.5' : 'gap-1.5',
          )}
        >
          {items.map((name) => (
            <FileNavItem
              key={name}
              name={name}
              activeFile={activeFile}
              category={category}
              files={files}
              modeKind={modeKind}
              onSelect={onSelect}
            />
          ))}
        </nav>
      )}
    </section>
  )
}

export function MeetingFileNav({
  names,
  activeFile,
  files,
  modeKind,
  onSelect,
}: MeetingFileNavProps) {
  const groups = groupMeetingFileNames(names)
  const [processExpanded, setProcessExpanded] = useState(false)

  useEffect(() => {
    if (activeFile && meetingFileCategory(activeFile) === 'process') {
      setProcessExpanded(true)
    }
  }, [activeFile])

  return (
    <div className="space-y-4">
      {FILE_CATEGORY_ORDER.map((category) => (
        <FileNavSection
          key={category}
          category={category}
          items={groups[category]}
          activeFile={activeFile}
          files={files}
          modeKind={modeKind}
          collapsed={category === 'process' ? !processExpanded : undefined}
          onToggle={
            category === 'process'
              ? () => setProcessExpanded((v) => !v)
              : undefined
          }
          onSelect={onSelect}
        />
      ))}
    </div>
  )
}
