import { useMemo, useState } from 'react'

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
import {
  chatSideRailLeftClass,
  chatSideRailRightClass,
  hePanelShell,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { maxTurnNumber } from '@/lib/meeting-transcript-projection'
import { buildMessageSequenceMap, messageSequenceNumber } from '@/lib/message-sequence'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface MeetingReplayViewerProps {
  topic: string
  meetingMd: string
  messages: ChatMessage[]
  className?: string
}

export function MeetingReplayViewer({
  topic,
  meetingMd,
  messages,
  className,
}: MeetingReplayViewerProps) {
  const [drawerMessage, setDrawerMessage] = useState<ChatMessage | null>(null)
  const [scrubTurn, setScrubTurn] = useState<number | null>(null)
  const narrow = useNarrowScreen()
  const wideSidePanel = useMediaQuery('(min-width: 96rem)')
  const roundtableSidePanel = !narrow && wideSidePanel

  const { turns, activeSpeakerId, latestBySeat, activeMessage, isScrubbing } =
    useMeetingTranscript(messages, scrubTurn)
  const { seats, participants, loading, rosterFromApi, rosterTotal } =
    useMeetingReplaySeats(meetingMd, messages)

  const maxTurn = maxTurnNumber(turns)
  const activeMessageId = activeMessage?.id ?? null
  const sequenceMap = useMemo(() => buildMessageSequenceMap(messages), [messages])
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

  return (
    <>
      <div className={cn('relative min-h-0', className)}>
        <div
          className={cn(
            hePanelShell,
            'relative flex h-full min-h-0 flex-col overflow-hidden',
          )}
        >
          <div className="flex shrink-0 items-center justify-between gap-3 border-b border-black/[0.05] px-5 py-4">
            <div>
              <h2 className={heSubsectionTitleNeutral}>会议回放</h2>
              <p className="mt-1 text-[12px] text-text-tertiary">
                圆桌 Live · 发言进度 · 侧栏详情
                {narrow && ' · 窄屏记录列表'}
              </p>
            </div>
          </div>

          <div className="relative flex min-h-0 flex-1 flex-col overflow-hidden">
            {!narrow ? (
              <RoundTableView
                seats={seats}
                messages={messages}
                latestBySeat={latestBySeat}
                activeSpeakerId={activeSpeakerId}
                focusedSeatId={focusedSeatId}
                turnCount={scrubTurn ?? turns.length}
                maxTurn={maxTurn}
                scrubTurn={scrubTurn}
                onScrubTurnChange={setScrubTurn}
                activeMessageId={activeMessageId}
                selectedMessageId={drawerMessage?.id ?? null}
                rosterLoading={loading}
                rosterFromApi={rosterFromApi}
                rosterTotal={rosterTotal}
                seatedExpertCount={participants.length}
                centerTitle={topic}
                centerSubtitle={centerSubtitle}
                onSelectMessage={setDrawerMessage}
                showTranscriptStrip={!roundtableSidePanel}
              />
            ) : (
              <StripOnlyView
                messages={messages}
                maxTurn={maxTurn}
                scrubTurn={scrubTurn}
                onScrubTurnChange={setScrubTurn}
                activeMessageId={activeMessageId}
                selectedMessageId={drawerMessage?.id ?? null}
                onSelectMessage={setDrawerMessage}
              />
            )}
          </div>
        </div>

        {roundtableSidePanel && (
          <>
            <TranscriptHistoryPanel
              messages={messages}
              activeMessageId={activeMessageId}
              selectedId={drawerMessage?.id ?? null}
              onSelect={setDrawerMessage}
              className={chatSideRailLeftClass}
            />
            <TranscriptDetailPanel
              message={drawerMessage}
              sequence={selectedSequence}
              onClear={() => setDrawerMessage(null)}
              className={chatSideRailRightClass}
            />
          </>
        )}
      </div>

      {!roundtableSidePanel && (
        <TranscriptDrawer
          message={drawerMessage}
          sequence={selectedSequence}
          onClose={() => setDrawerMessage(null)}
        />
      )}
    </>
  )
}
