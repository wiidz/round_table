import { useRef } from 'react'
import { SendHorizonal } from 'lucide-react'

import {
  chatComposerInnerClass,
  chatComposerOuterClass,
  chatComposerSendClass,
  chatComposerTextareaClass,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface ChatComposerProps {
  draft: string
  onDraftChange: (value: string) => void
  onSend: (text: string) => void
  disabled?: boolean
  className?: string
}

function isImeComposing(event: React.KeyboardEvent<HTMLTextAreaElement>): boolean {
  return event.nativeEvent.isComposing || event.keyCode === 229
}

export function ChatComposer({
  draft,
  onDraftChange,
  onSend,
  disabled = false,
  className,
}: ChatComposerProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const composingRef = useRef(false)
  const pendingEnterSendRef = useRef(false)

  const readText = () => textareaRef.current?.value ?? draft

  const trySend = (textOverride?: string) => {
    const text = (textOverride ?? readText()).trim()
    if (!text || disabled) return
    onSend(text)
  }

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    trySend()
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key !== 'Enter' || event.shiftKey) return

    if (isImeComposing(event) || composingRef.current) {
      pendingEnterSendRef.current = true
      return
    }

    event.preventDefault()
    trySend()
  }

  const handleCompositionStart = () => {
    composingRef.current = true
  }

  const handleCompositionEnd = (event: React.CompositionEvent<HTMLTextAreaElement>) => {
    composingRef.current = false
    if (!pendingEnterSendRef.current) return

    pendingEnterSendRef.current = false
    const text = event.currentTarget.value.trim()
    if (text && !disabled) {
      onSend(text)
    }
  }

  return (
    <div
      className={cn(
        'pointer-events-none absolute inset-x-0 bottom-0 z-30',
        'bg-gradient-to-t from-surface from-55% via-surface/95 to-transparent px-4 pb-4 pt-8 sm:px-5',
        className,
      )}
    >
      <form
        onSubmit={handleSubmit}
        className={cn(chatComposerOuterClass, 'pointer-events-auto mx-auto max-w-full')}
      >
        <div className={chatComposerInnerClass}>
          <textarea
            ref={textareaRef}
            value={draft}
            onChange={(e) => onDraftChange(e.target.value)}
            onKeyDown={handleKeyDown}
            onCompositionStart={handleCompositionStart}
            onCompositionEnd={handleCompositionEnd}
            rows={2}
            placeholder={disabled ? '连接中…' : '输入消息，Enter 发送，Shift+Enter 换行'}
            disabled={disabled}
            aria-label="聊天输入"
            className={chatComposerTextareaClass}
          />
          <button
            type="submit"
            disabled={disabled || !draft.trim()}
            className={chatComposerSendClass}
          >
            <SendHorizonal className="size-4" aria-hidden />
            发送
          </button>
        </div>
      </form>
    </div>
  )
}
