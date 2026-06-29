import { SendHorizonal } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { heFormEmbed, hePressable, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface ChatComposerProps {
  draft: string
  onDraftChange: (value: string) => void
  onSend: () => void
  disabled?: boolean
  className?: string
}

export function ChatComposer({
  draft,
  onDraftChange,
  onSend,
  disabled = false,
  className,
}: ChatComposerProps) {
  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (disabled) return
    onSend()
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (disabled) return
      onSend()
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className={cn('shrink-0 border-t border-black/[0.05] bg-black/[0.015] px-5 py-4', className)}
    >
      <div className={cn(heFormEmbed, 'flex items-end gap-3 p-3')}>
        <textarea
          value={draft}
          onChange={(e) => onDraftChange(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={2}
          placeholder={disabled ? '连接中…' : '输入消息，Enter 发送，Shift+Enter 换行'}
          disabled={disabled}
          className={cn(
            'min-h-[3rem] flex-1 resize-none border-0 bg-transparent px-1 py-1 text-[14px]',
            'text-text-primary placeholder:text-text-tertiary focus:outline-none',
            'disabled:cursor-not-allowed disabled:opacity-60',
          )}
        />
        <Button
          type="submit"
          disabled={disabled || !draft.trim()}
          className={cn(hePressable, heSpring, 'shrink-0 gap-1.5 rounded-xs px-4')}
        >
          <SendHorizonal className="size-4" aria-hidden />
          发送
        </Button>
      </div>
    </form>
  )
}
