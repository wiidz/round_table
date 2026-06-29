import { useEffect, useRef, useState } from 'react'
import { List, X } from 'lucide-react'

import { TocBackToTop } from '@/components/markdown/back-to-top-button'
import { heSpring } from '@/lib/highend-styles'
import {
  MARKDOWN_HEADING_SCROLL_OFFSET,
  scrollToHeading,
  type MarkdownHeading,
} from '@/lib/markdown-headings'
import { cn } from '@/lib/utils'

const TOC_MIN_HEADINGS = 2

function useActiveHeadingId(headings: MarkdownHeading[]) {
  const [activeId, setActiveId] = useState(headings[0]?.id ?? '')
  const pendingScrollId = useRef<string | null>(null)

  useEffect(() => {
    if (headings.length === 0) {
      setActiveId('')
      return
    }

    const sync = () => {
      if (pendingScrollId.current) {
        setActiveId(pendingScrollId.current)
        return
      }

      let current = headings[0].id
      for (const heading of headings) {
        const el = document.getElementById(heading.id)
        if (
          el &&
          el.getBoundingClientRect().top <= MARKDOWN_HEADING_SCROLL_OFFSET
        ) {
          current = heading.id
        }
      }
      setActiveId(current)
    }

    sync()
    window.addEventListener('scroll', sync, { passive: true })
    window.addEventListener('resize', sync)
    return () => {
      window.removeEventListener('scroll', sync)
      window.removeEventListener('resize', sync)
    }
  }, [headings])

  const navigateTo = (id: string, onNavigate?: () => void) => {
    pendingScrollId.current = id
    setActiveId(id)
    scrollToHeading(id, {
      onDone: () => {
        pendingScrollId.current = null
      },
    })
    onNavigate?.()
  }

  return { activeId, navigateTo }
}

function TocList({
  headings,
  activeId,
  onNavigate,
  navigateTo,
}: {
  headings: MarkdownHeading[]
  activeId: string
  onNavigate?: () => void
  navigateTo: (id: string, onNavigate?: () => void) => void
}) {
  return (
    <ul className="space-y-0.5">
      {headings.map((heading) => (
        <li key={heading.id}>
          <button
            type="button"
            onClick={() => navigateTo(heading.id, onNavigate)}
            className={cn(
              'w-full rounded-md py-1 text-left text-[13px] leading-snug',
              heSpring,
              heading.level === 1 && 'pl-0 font-medium',
              heading.level === 2 && 'pl-2',
              heading.level === 3 && 'pl-4 text-[11px]',
              activeId === heading.id
                ? 'font-medium text-brand'
                : 'text-text-tertiary hover:text-text-secondary',
            )}
            title={heading.text}
          >
            <span className="block break-words">{heading.text}</span>
          </button>
        </li>
      ))}
    </ul>
  )
}

interface MarkdownTocProps {
  headings: MarkdownHeading[]
}

export function MarkdownTocFloating({ headings }: MarkdownTocProps) {
  const { activeId, navigateTo } = useActiveHeadingId(headings)

  if (headings.length < TOC_MIN_HEADINGS) return null

  return (
    <aside className="absolute inset-y-0 left-full z-10 ml-5 hidden w-60 xl:block">
      <nav
        aria-label="文档目录"
        className={cn(
          'sticky top-20 flex max-h-[calc(100vh-5.5rem)] w-full flex-col rounded-xl p-3.5',
          'bg-surface/92 backdrop-blur-sm',
          'shadow-[var(--panel-shell-shadow)] ring-1 ring-inset ring-black/[0.06]',
        )}
      >
        <p className="mb-2 shrink-0 text-[10px] font-medium uppercase tracking-[0.14em] text-text-tertiary">
          目录
        </p>
        <div className="min-h-0 flex-1 overflow-y-auto">
          <TocList headings={headings} activeId={activeId} navigateTo={navigateTo} />
        </div>
        <TocBackToTop className="mt-2" />
      </nav>
    </aside>
  )
}

export function MarkdownTocMobile({ headings }: MarkdownTocProps) {
  const [open, setOpen] = useState(false)
  const { activeId, navigateTo } = useActiveHeadingId(headings)

  if (headings.length < TOC_MIN_HEADINGS) return null

  return (
    <div className="fixed bottom-6 right-[4.25rem] z-30 xl:hidden">
      {open && (
        <div
          className={cn(
            'mb-2 flex w-[min(16rem,calc(100vw-3rem))] max-h-[min(40vh,24rem)] flex-col rounded-xl bg-surface p-3',
            'shadow-[var(--panel-shell-shadow)] ring-1 ring-inset ring-black/[0.06]',
          )}
        >
          <div className="mb-2 flex shrink-0 items-center justify-between gap-2">
            <p className="text-[10px] font-medium uppercase tracking-[0.14em] text-text-tertiary">
              目录
            </p>
            <button
              type="button"
              onClick={() => setOpen(false)}
              className="rounded-md p-1 text-text-tertiary hover:bg-black/[0.04] hover:text-text-secondary"
              aria-label="关闭目录"
            >
              <X className="size-3.5" />
            </button>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto">
            <TocList
              headings={headings}
              activeId={activeId}
              navigateTo={navigateTo}
              onNavigate={() => setOpen(false)}
            />
          </div>
          <TocBackToTop
            className="mt-2"
            onClick={() => setOpen(false)}
          />
        </div>
      )}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'inline-flex size-10 items-center justify-center rounded-full bg-surface',
          'text-text-secondary shadow-[var(--panel-shell-shadow)] ring-1 ring-inset ring-black/[0.08]',
          heSpring,
          'hover:text-brand hover:ring-primary/25',
          open && 'text-brand ring-primary/30',
        )}
        aria-label={open ? '关闭目录' : '打开目录'}
        aria-expanded={open}
      >
        <List className="size-4" />
      </button>
    </div>
  )
}
