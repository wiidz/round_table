import type { ReactNode } from 'react'
import { useCallback, useMemo, useState } from 'react'

import { PageLayout } from '@/components/layout/page-main-layout'
import { RoundTableView } from '@/components/round-table/round-table-view'
import { StripOnlyView } from '@/components/round-table/strip-only-view'
import { TranscriptDetailPanel } from '@/components/round-table/transcript-detail-panel'
import { TranscriptDrawer } from '@/components/round-table/transcript-drawer'
import { TranscriptHistoryPanel } from '@/components/round-table/transcript-history-panel'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { useMeetingReplaySeats } from '@/hooks/use-meeting-replay-seats'
import { useMeetingTranscript } from '@/hooks/use-meeting-transcript'
import { useMediaQuery, useNarrowScreen } from '@/hooks/use-media-query'
import { speakerId } from '@/lib/chat-display'
import { hePanelShell, heSubsectionTitleNeutral } from '@/lib/highend-styles'
import { maxTurnNumber, scrubTurnForMessage, activeMessageAtScrubTurn } from '@/lib/meeting-transcript-projection'
import { buildMessageSequenceMap, messageSequenceNumber } from '@/lib/message-sequence'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface MeetingReplayViewerProps {
  topic: string
  meetingMd: string
  messages: ChatMessage[]
  header?: ReactNode
  pageShell?: (slots: {
    main: ReactNode
    left?: ReactNode
    right?: ReactNode
    drawer?: ReactNode
  }) => ReactNode
}

const replayMainPanelClass = cn(
  hePanelShell,
  'flex h-full min-h-[36rem] flex-col overflow-hidden lg:min-h-[calc(100vh-14rem)]',
)

export function MeetingReplayViewer({
  topic,
  meetingMd,
  messages,
  header,
  pageShell,
}: MeetingReplayViewerProps) {
  const [drawerMessage, setDrawerMessage] = useState<ChatMessage | null>(null)
  const [scrubTurn, setScrubTurn] = useState<number | null>(null)
  const narrow = useNarrowScreen()
  const wideSidePanel = useMediaQuery('(min-width: 96rem)')
  const roundtableSidePanel = !narrow && wideSidePanel

  const { turns, latestBySeat, activeMessage, isScrubbing } =
    useMeetingTranscript(messages, scrubTurn)
  const { seats, participants, loading, rosterFromApi, rosterTotal } =
    useMeetingReplaySeats(meetingMd, messages)

  const maxTurn = maxTurnNumber(turns)
  const activeMessageId = activeMessage?.id ?? null
  const sequenceMap = useMemo(() => buildMessageSequenceMap(messages), [messages])

  const selectMessage = useCallback((msg: ChatMessage) => {
    setDrawerMessage(msg)
  }, [])

  const handleScrubTurnChange = useCallback(
    (turn: number | null) => {
      setScrubTurn(turn)
      setDrawerMessage(activeMessageAtScrubTurn(turns, turn))
    },
    [turns],
  )

  const selectHistoryMessage = useCallback(
    (msg: ChatMessage) => {
      setScrubTurn(scrubTurnForMessage(msg, maxTurn))
      setDrawerMessage(msg)
    },
    [maxTurn],
  )

  const highlightMessageId = drawerMessage?.id ?? activeMessageId
  const referenceTurn = drawerMessage?.turn ?? activeMessage?.turn ?? null
  const selectedSequence = drawerMessage
    ? messageSequenceNumber(drawerMessage, sequenceMap)
    : null

  const centerSubtitle = useMemo(() => {
    if (isScrubbing && scrubTurn != null) {
      return `回放 · 第 ${scrubTurn} 轮发言`
    }
    if (turns.length > 0) return `共 ${turns.length} 条发言 · 拖动进度条回放`
    return undefined
  }, [isScrubbing, scrubTurn, turns.length])

  const focusedSeatId = useMemo(
    () => (drawerMessage ? speakerId(drawerMessage) : null),
    [drawerMessage],
  )

  if (messages.length === 0) {
    return (
      <ProfileStatePanel
        title="暂无可回放发言"
        description="本场 Workspace 尚未生成 MINUTES.md 或各轮研讨记录。"
      />
    )
  }

  const mainPanel = (
    <div className={replayMainPanelClass}>
      <div className="flex shrink-0 items-center justify-between gap-3 border-b border-black/[0.05] px-5 py-4">
        <div>
          <h2 className={heSubsectionTitleNeutral}>会议回放</h2>
          <p className="mt-1 text-[12px] text-text-tertiary">
            圆桌 Live · 发言进度 · 侧栏详情
            {narrow && ' · 窄屏记录列表'}
          </p>
        </div>
      </div>

      <div className="relative flex min-h-[28rem] flex-1 flex-col overflow-hidden">
        {!narrow ? (
          <RoundTableView
            seats={seats}
            messages={messages}
            latestBySeat={latestBySeat}
            focusedSeatId={focusedSeatId}
            turnCount={scrubTurn ?? turns.length}
            maxTurn={maxTurn}
            scrubTurn={scrubTurn}
            onScrubTurnChange={handleScrubTurnChange}
            activeMessageId={activeMessageId}
            selectedMessageId={drawerMessage?.id ?? null}
            highlightMessageId={highlightMessageId}
            referenceTurn={referenceTurn}
            rosterLoading={loading}
            rosterFromApi={rosterFromApi}
            rosterTotal={rosterTotal}
            seatedExpertCount={participants.length}
            centerTitle={topic}
            centerSubtitle={centerSubtitle}
            onSelectMessage={selectMessage}
            showTranscriptStrip={!roundtableSidePanel}
          />
        ) : (
          <StripOnlyView
            messages={messages}
            maxTurn={maxTurn}
            scrubTurn={scrubTurn}
            onScrubTurnChange={handleScrubTurnChange}
            activeMessageId={activeMessageId}
            selectedMessageId={drawerMessage?.id ?? null}
            onSelectMessage={selectMessage}
          />
        )}
      </div>
    </div>
  )

  const leftPanel = roundtableSidePanel ? (
    <TranscriptHistoryPanel
      messages={messages}
      activeMessageId={activeMessageId}
      selectedId={highlightMessageId}
      onSelect={selectHistoryMessage}
    />
  ) : undefined

  const rightPanel = roundtableSidePanel ? (
    <TranscriptDetailPanel
      message={drawerMessage}
      sequence={selectedSequence}
      onClear={() => setDrawerMessage(null)}
    />
  ) : undefined

  const drawer = !roundtableSidePanel ? (
    <TranscriptDrawer
      message={drawerMessage}
      sequence={selectedSequence}
      onClose={() => setDrawerMessage(null)}
    />
  ) : undefined

  if (pageShell) {
    return pageShell({ main: mainPanel, left: leftPanel, right: rightPanel, drawer })
  }

  return (
    <>
      <PageLayout
        header={header}
        sidebarFrom="96rem"
        sideColumnWidth="gutter"
        left={leftPanel}
        right={rightPanel}
        bodyClassName="min-[96rem]:h-full"
      >
        <div className="h-full min-h-0">{mainPanel}</div>
      </PageLayout>
      {drawer}
    </>
  )
}
