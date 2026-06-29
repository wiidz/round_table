import { useState } from 'react'
import { Loader2, SendHorizonal, Wifi, WifiOff } from 'lucide-react'

import { RoundTableStage } from '@/components/round-table/round-table-stage'
import { TranscriptDrawer } from '@/components/round-table/transcript-drawer'
import { TranscriptStrip } from '@/components/round-table/transcript-strip'
import { Button } from '@/components/ui/button'
import { useMeetingTranscript } from '@/hooks/use-meeting-transcript'
import { useRosterSeats } from '@/hooks/use-roster-seats'
import {
  heFormEmbed,
  hePanelShell,
  hePressable,
  heSpring,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { ChatConnectionState, ChatMessage } from '@/types/chat'

function ConnectionBadge({ state }: { state: ChatConnectionState }) {
  const label =
    state === 'open'
      ? '已连接'
      : state === 'connecting'
        ? '连接中'
        : state === 'error'
          ? '连接异常'
          : '已断开'

  const tone =
    state === 'open' ? 'success' : state === 'connecting' ? 'warning' : 'neutral'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset',
        tone === 'success' && 'bg-success-soft text-success ring-success/20',
        tone === 'warning' && 'bg-warning-soft text-warning ring-warning/25',
        tone === 'neutral' && 'bg-black/[0.04] text-text-tertiary ring-black/[0.06]',
      )}
    >
      {state === 'open' ? (
        <Wifi className="size-3" aria-hidden />
      ) : state === 'connecting' ? (
        <Loader2 className="size-3 animate-spin" aria-hidden />
      ) : (
        <WifiOff className="size-3" aria-hidden />
      )}
      {label}
    </span>
  )
}

interface ChatWindowProps {
  className?: string
  connectionState: ChatConnectionState
  messages: ChatMessage[]
  sessionId: string | null
  lastError: string | null
  onSend: (content: string) => boolean
  onReconnect: () => void
}

export function ChatWindow({
  className,
  connectionState,
  messages,
  sessionId,
  lastError,
  onSend,
  onReconnect,
}: ChatWindowProps) {
  const [draft, setDraft] = useState('')
  const [drawerMessage, setDrawerMessage] = useState<ChatMessage | null>(null)
  const canSend = connectionState === 'open'

  const { turns, activeSpeakerId, latestBySeat } = useMeetingTranscript(messages)
  const { seats } = useRosterSeats(messages)
  const activeMessageId =
    activeSpeakerId != null ? latestBySeat.get(activeSpeakerId)?.id ?? null : null

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!canSend) return
    if (onSend(draft)) {
      setDraft('')
    }
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (!canSend) return
      if (onSend(draft)) {
        setDraft('')
      }
    }
  }

  return (
    <>
      <div className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}>
        <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-black/[0.05] px-5 py-4">
          <div>
            <h2 className={heSubsectionTitleNeutral}>与司仪对话</h2>
            <p className="mt-1 text-[12px] text-text-tertiary">
              浏览器 Transport · 无需 Principal · 可发起会议
            </p>
          </div>
          <div className="flex items-center gap-2">
            <ConnectionBadge state={connectionState} />
            {connectionState !== 'open' && (
              <Button type="button" variant="outline" size="sm" onClick={onReconnect}>
                重连
              </Button>
            )}
          </div>
        </div>

        {sessionId && connectionState === 'open' && (
          <p className="shrink-0 border-b border-black/[0.04] px-5 py-2 font-mono text-[11px] text-text-tertiary">
            会话 {sessionId.slice(0, 8)}…
          </p>
        )}

        <RoundTableStage
          seats={seats}
          latestBySeat={latestBySeat}
          activeSpeakerId={activeSpeakerId}
          turnCount={turns.length}
          onLiveMessageClick={setDrawerMessage}
        />

        <TranscriptStrip
          messages={messages}
          activeMessageId={activeMessageId}
          selectedId={drawerMessage?.id ?? null}
          onSelect={setDrawerMessage}
        />

        {lastError && connectionState === 'error' && (
          <p className="shrink-0 px-5 pb-2 text-[12px] text-danger">{lastError}</p>
        )}

        <form
          onSubmit={handleSubmit}
          className="shrink-0 border-t border-black/[0.05] bg-black/[0.015] px-5 py-4"
        >
          <div className={cn(heFormEmbed, 'flex items-end gap-3 p-3')}>
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={2}
              placeholder={canSend ? '输入消息，Enter 发送，Shift+Enter 换行' : '连接中…'}
              disabled={!canSend}
              className={cn(
                'min-h-[3rem] flex-1 resize-none border-0 bg-transparent px-1 py-1 text-[14px]',
                'text-text-primary placeholder:text-text-tertiary focus:outline-none',
                'disabled:cursor-not-allowed disabled:opacity-60',
              )}
            />
            <Button
              type="submit"
              disabled={!canSend || !draft.trim()}
              className={cn(hePressable, heSpring, 'shrink-0 gap-1.5 rounded-xs px-4')}
            >
              <SendHorizonal className="size-4" aria-hidden />
              发送
            </Button>
          </div>
        </form>
      </div>

      <TranscriptDrawer message={drawerMessage} onClose={() => setDrawerMessage(null)} />
    </>
  )
}
