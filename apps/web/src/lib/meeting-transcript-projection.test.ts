import { describe, expect, it } from 'vitest'

import type { ChatMessage } from '@/types/chat'
import { activeMessageAtScrubTurn, maxTurnNumber, projectTranscriptAtTurn, scrubTurnForMessage } from '@/lib/meeting-transcript-projection'

function turnMsg(turn: number, role: ChatMessage['role'], id: string): ChatMessage {
  return {
    id,
    role,
    content: `msg ${turn}`,
    turn,
    createdAt: turn,
    authorId: role === 'participant' ? id : undefined,
  }
}

describe('projectTranscriptAtTurn', () => {
  const turns = [
    turnMsg(1, 'moderator', 'm1'),
    turnMsg(2, 'participant', 'dev'),
    turnMsg(3, 'participant', 'design'),
  ]

  it('uses full tail when scrubTurn is null', () => {
    const { activeSpeakerId, activeMessage } = projectTranscriptAtTurn(turns, null)
    expect(activeSpeakerId).toBe('design')
    expect(activeMessage?.turn).toBe(3)
  })

  it('projects historical state at scrub turn', () => {
    const { latestBySeat, activeSpeakerId } = projectTranscriptAtTurn(turns, 2)
    expect(activeSpeakerId).toBe('dev')
    expect(latestBySeat.has('design')).toBe(false)
    expect(latestBySeat.get('dev')?.turn).toBe(2)
  })
})

describe('maxTurnNumber', () => {
  it('returns highest turn', () => {
    expect(maxTurnNumber([turnMsg(1, 'moderator', 'a'), turnMsg(5, 'participant', 'b')])).toBe(5)
  })
})

describe('activeMessageAtScrubTurn', () => {
  it('returns the last message visible at scrub turn', () => {
    const turns = [
      turnMsg(1, 'moderator', 'm1'),
      turnMsg(2, 'participant', 'dev'),
      turnMsg(3, 'participant', 'design'),
    ]
    expect(activeMessageAtScrubTurn(turns, 2)?.id).toBe('dev')
    expect(activeMessageAtScrubTurn(turns, null)?.id).toBe('design')
  })
})

describe('scrubTurnForMessage', () => {
  it('returns null at live tail', () => {
    const msg = turnMsg(5, 'participant', 'dev')
    expect(scrubTurnForMessage(msg, 5)).toBeNull()
  })

  it('returns turn when before max', () => {
    const msg = turnMsg(3, 'participant', 'dev')
    expect(scrubTurnForMessage(msg, 5)).toBe(3)
  })
})
