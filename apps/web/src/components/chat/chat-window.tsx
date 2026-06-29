import { useMemo, useState } from 'react'
import { Loader2, LayoutList, Users, Wifi, WifiOff } from 'lucide-react'

import { ChatComposer } from '@/components/chat/chat-composer'
import { ImTranscriptView } from '@/components/chat/im-transcript-view'
import { RoundTableView } from '@/components/round-table/round-table-view'
import { StripOnlyView } from '@/components/round-table/strip-only-view'
import { TranscriptDrawer } from '@/components/round-table/transcript-drawer'
import { Button } from '@/components/ui/button'
import { useChatViewMode } from '@/hooks/use-chat-view-mode'
import { useMeetingTranscript } from '@/hooks/use-meeting-transcript'
import { useNarrowScreen } from '@/hooks/use-media-query'
import { useRosterSeats } from '@/hooks/use-roster-seats'
import { speakerId } from '@/lib/chat-display'
import { phaseLabel } from '@/lib/chat-meeting-phase'
import { hePanelShell, heSubsectionTitleNeutral } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { ChatConnectionState, ChatMessage } from '@/types/chat'

type ChatLayout = 'list' | 'roundtable' | 'strip-only'

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

function ViewModeToggle({
  mode,
  onChange,
  disabled,
}: {
  mode: 'list' | 'roundtable'
  onChange: (mode: 'list' | 'roundtable') => void
  disabled?: boolean
}) {
  return (
    <div
      className={cn(
        'flex rounded-lg bg-black/[0.04] p-0.5 ring-1 ring-inset ring-black/[0.06]',
        disabled && 'opacity-50',
      )}
      role="group"
      aria-label="视图模式"
    >
      <button
        type="button"
        disabled={disabled}
        onClick={() => onChange('roundtable')}
        className={cn(
          'inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-[12px] font-medium transition-colors',
          mode === 'roundtable'
            ? 'bg-surface text-brand shadow-sm'
            : 'text-text-tertiary hover:text-text-secondary',
          disabled && 'cursor-not-allowed',
        )}
      >
        <Users className="size-3.5" aria-hidden />
        圆桌
      </button>
      <button
        type="button"
        disabled={disabled}
        onClick={() => onChange('list')}
        className={cn(
          'inline-flex items-center gap-1 rounded-md px-2.5 py-1 text-[12px] font-medium transition-colors',
          mode === 'list'
            ? 'bg-surface text-brand shadow-sm'
            : 'text-text-tertiary hover:text-text-secondary',
          disabled && 'cursor-not-allowed',
        )}
      >
        <LayoutList className="size-3.5" aria-hidden />
        列表
      </button>
    </div>
  )
}

function resolveLayout(
  narrow: boolean,
  phase: ReturnType<typeof useChatViewMode>['phase'],
  mode: 'list' | 'roundtable',
): ChatLayout {
  if (narrow) {
    return phase === 'running' || phase === 'post' ? 'strip-only' : 'list'
  }
  return mode
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
  const narrow = useNarrowScreen()

  const { mode, phase, setMode } = useChatViewMode(messages)
  const layout = resolveLayout(narrow, phase, mode)
  const { turns, activeSpeakerId, latestBySeat } = useMeetingTranscript(messages)
  const { seats, participants, loading, rosterFromApi } = useRosterSeats(messages)
  const activeMessageId =
    activeSpeakerId != null ? latestBySeat.get(activeSpeakerId)?.id ?? null : null

  const focusedSeatId = useMemo(
    () => (drawerMessage ? speakerId(drawerMessage) : null),
    [drawerMessage],
  )

  const submitDraft = () => {
    if (onSend(draft)) setDraft('')
  }

  return (
    <>
      <div className={cn(hePanelShell, 'flex h-full min-h-0 flex-col overflow-hidden', className)}>
        <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-black/[0.05] px-5 py-4">
          <div>
            <h2 className={heSubsectionTitleNeutral}>与司仪对话</h2>
            <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[12px] text-text-tertiary">
              <span>浏览器 Transport · 无需 Principal</span>
              <span className="text-black/20">·</span>
              <span>{phaseLabel(phase)}</span>
              {narrow && (
                <>
                  <span className="text-black/20">·</span>
                  <span>{layout === 'strip-only' ? '窄屏记录' : '窄屏列表'}</span>
                </>
              )}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <div className="hidden md:block">
              <ViewModeToggle mode={mode} onChange={setMode} />
            </div>
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

        {layout === 'roundtable' && (
          <RoundTableView
            seats={seats}
            messages={messages}
            latestBySeat={latestBySeat}
            activeSpeakerId={activeSpeakerId}
            focusedSeatId={focusedSeatId}
            turnCount={turns.length}
            activeMessageId={activeMessageId}
            selectedMessageId={drawerMessage?.id ?? null}
            rosterLoading={loading}
            rosterFromApi={rosterFromApi}
            participantCount={participants.length}
            onSelectMessage={setDrawerMessage}
          />
        )}

        {layout === 'list' && <ImTranscriptView messages={messages} />}

        {layout === 'strip-only' && (
          <StripOnlyView
            messages={messages}
            activeMessageId={activeMessageId}
            selectedMessageId={drawerMessage?.id ?? null}
            onSelectMessage={setDrawerMessage}
          />
        )}

        {lastError && connectionState === 'error' && (
          <p className="shrink-0 px-5 pb-2 text-[12px] text-danger">{lastError}</p>
        )}

        <ChatComposer
          draft={draft}
          onDraftChange={setDraft}
          onSend={submitDraft}
          disabled={!canSend}
        />
      </div>

      <TranscriptDrawer message={drawerMessage} onClose={() => setDrawerMessage(null)} />
    </>
  )
}
