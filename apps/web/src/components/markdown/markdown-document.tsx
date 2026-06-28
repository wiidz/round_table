import type { Components } from 'react-markdown'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

import { cn } from '@/lib/utils'

const markdownComponents: Components = {
  h1: ({ children }) => (
    <h1 className="mt-8 mb-4 text-[22px] font-semibold tracking-[-0.02em] text-text-primary first:mt-0">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="mt-8 mb-3 text-lg font-semibold tracking-[-0.02em] text-text-primary first:mt-0">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="mt-6 mb-2 text-base font-semibold text-text-primary first:mt-0">
      {children}
    </h3>
  ),
  p: ({ children }) => (
    <p className="mb-4 text-[15px] leading-[1.85] text-text-primary last:mb-0">
      {children}
    </p>
  ),
  ul: ({ children }) => (
    <ul className="mb-4 list-disc space-y-1.5 pl-5 text-[15px] leading-[1.85] text-text-primary last:mb-0">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="mb-4 list-decimal space-y-1.5 pl-5 text-[15px] leading-[1.85] text-text-primary last:mb-0">
      {children}
    </ol>
  ),
  li: ({ children }) => <li className="text-text-primary">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="mb-4 border-l-[3px] border-ai/35 bg-ai-soft/40 py-2 pl-4 text-[15px] leading-[1.85] text-text-secondary italic last:mb-0">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-8 border-border-subtle" />,
  a: ({ href, children }) => (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="text-info underline decoration-info/30 underline-offset-2 hover:decoration-info/60"
    >
      {children}
    </a>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold text-text-primary">{children}</strong>
  ),
  em: ({ children }) => <em className="italic text-text-secondary">{children}</em>,
  code: ({ className, children }) => {
    const isBlock = /language-/.test(className ?? '')
    if (isBlock) {
      return (
        <code className={cn('font-mono text-[13px] leading-[1.7] text-text-primary', className)}>
          {children}
        </code>
      )
    }
    return (
      <code className="rounded-md bg-black/[0.05] px-1.5 py-0.5 font-mono text-[13px] text-text-primary ring-1 ring-inset ring-black/[0.05]">
        {children}
      </code>
    )
  },
  pre: ({ children }) => (
    <pre className="mb-4 overflow-x-auto rounded-lg bg-black/[0.04] px-4 py-3 ring-1 ring-inset ring-black/[0.06] last:mb-0">
      {children}
    </pre>
  ),
  table: ({ children }) => (
    <div className="mb-4 overflow-x-auto last:mb-0">
      <table className="w-full min-w-[320px] border-collapse text-left text-[14px] leading-relaxed">
        {children}
      </table>
    </div>
  ),
  thead: ({ children }) => (
    <thead className="border-b border-border-subtle bg-black/[0.02]">{children}</thead>
  ),
  tbody: ({ children }) => <tbody className="divide-y divide-border-subtle/80">{children}</tbody>,
  tr: ({ children }) => <tr>{children}</tr>,
  th: ({ children }) => (
    <th className="px-3 py-2.5 font-medium text-text-secondary">{children}</th>
  ),
  td: ({ children }) => (
    <td className="px-3 py-2.5 align-top text-text-primary">{children}</td>
  ),
}

interface MarkdownDocumentProps {
  content: string
  className?: string
  /** DESIGN.md Document Layout — default max 760px */
  constrained?: boolean
}

export function MarkdownDocument({
  content,
  className,
  constrained = true,
}: MarkdownDocumentProps) {
  if (!content.trim()) {
    return (
      <p className="text-sm text-text-tertiary italic">（空文档）</p>
    )
  }

  return (
    <article
      className={cn(
        'min-w-0',
        constrained && 'max-w-[760px]',
        className,
      )}
    >
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
        {content}
      </ReactMarkdown>
    </article>
  )
}
