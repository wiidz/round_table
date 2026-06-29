import { useCallback, useState } from 'react'

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
}

export function MarkdownReader({
  content,
  className,
  constrained = true,
  documentKey,
}: MarkdownReaderProps) {
  const [headings, setHeadings] = useState<MarkdownHeading[]>([])
  const handleHeadingsChange = useCallback((next: MarkdownHeading[]) => {
    setHeadings((prev) => (headingsEqual(prev, next) ? prev : next))
  }, [])

  const showToc = headings.length >= 2

  return (
    <div
      className={cn(
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
      {showToc && <MarkdownTocFloating headings={headings} />}
      {showToc && <MarkdownTocMobile headings={headings} />}
    </div>
  )
}
