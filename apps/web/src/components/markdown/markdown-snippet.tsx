import type { Components } from 'react-markdown'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import { cn } from '@/lib/utils'

const snippetComponents: Components = {
  p: ({ children }) => (
    <p className="mb-0 text-[15px] leading-relaxed font-medium text-text-primary last:mb-0">
      {children}
    </p>
  ),
  strong: ({ children }) => <strong className="font-semibold text-text-primary">{children}</strong>,
  em: ({ children }) => <em className="italic text-text-secondary">{children}</em>,
  ul: ({ children }) => (
    <ul className="mb-0 list-disc space-y-1 pl-5 text-[15px] leading-relaxed last:mb-0">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="mb-0 list-decimal space-y-1 pl-5 text-[15px] leading-relaxed last:mb-0">
      {children}
    </ol>
  ),
  li: ({ children }) => <li className="text-text-primary">{children}</li>,
}

interface MarkdownSnippetProps {
  content: string
  className?: string
}

export function MarkdownSnippet({ content, className }: MarkdownSnippetProps) {
  const text = content.trim()
  if (!text) return null

  return (
    <div className={cn('min-w-0', className)}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={snippetComponents}>
        {text}
      </ReactMarkdown>
    </div>
  )
}

export function MarkdownSnippetOrEmpty({
  content,
  empty,
  className,
  emptyClassName,
}: MarkdownSnippetProps & { empty: string; emptyClassName?: string }) {
  const text = content.trim()
  if (!text) {
    return <p className={cn('text-[15px] leading-relaxed text-text-tertiary', emptyClassName)}>{empty}</p>
  }
  return <MarkdownSnippet content={text} className={className} />
}
