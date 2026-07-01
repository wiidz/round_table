import { useCallback, useEffect, useState } from 'react'

import { MarkdownDocument } from '@/components/markdown/markdown-document'
import {
  MarkdownTocFloating,
  MarkdownTocMobile,
} from '@/components/markdown/markdown-toc'
import {
  headingsEqual,
  type MarkdownHeading,
} from '@/lib/markdown-headings'
import { cn } from '@/lib/utils'

interface MarkdownReaderProps {
  content: string
  className?: string
  constrained?: boolean
  /** Remount collector when switching documents */
  documentKey?: string
  /**
   * Wide meeting documents: parent renders gutter TOC; reader only collects headings.
   */
  tocInGutter?: boolean
  onHeadingsCollected?: (headings: MarkdownHeading[]) => void
}

export function MarkdownReader({
  content,
  className,
  constrained = true,
  documentKey,
  tocInGutter = false,
  onHeadingsCollected,
}: MarkdownReaderProps) {
  const [headings, setHeadings] = useState<MarkdownHeading[]>([])
  const handleHeadingsChange = useCallback((next: MarkdownHeading[]) => {
    setHeadings((prev) => (headingsEqual(prev, next) ? prev : next))
  }, [])

  useEffect(() => {
    onHeadingsCollected?.(headings)
  }, [headings, onHeadingsCollected])

  const showToc = headings.length >= 2

  return (
    <div
      className={cn(
        'relative',
        constrained && 'max-w-[760px]',
        className,
      )}
    >
      <MarkdownDocument
        key={documentKey ?? content}
        content={content}
        constrained={false}
        onHeadingsChange={handleHeadingsChange}
      />
      {showToc && !tocInGutter && <MarkdownTocFloating headings={headings} />}
      {showToc && !tocInGutter && <MarkdownTocMobile headings={headings} />}
      {showToc && tocInGutter && (
        <MarkdownTocMobile headings={headings} hideFrom="96rem" />
      )}
    </div>
  )
}
