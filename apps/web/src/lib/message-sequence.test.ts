import { describe, expect, it } from 'vitest'

import {
  buildMessageSequenceMap,
  formatMessageSequence,
  messageSequenceNumber,
} from '@/lib/message-sequence'
import type { ChatMessage } from '@/types/chat'

function msg(partial: Partial<ChatMessage> & Pick<ChatMessage, 'id' | 'role' | 'content'>): ChatMessage {
  return { createdAt: 0, ...partial }
}

describe('buildMessageSequenceMap', () => {
  it('includes user speech in global order', () => {
    const messages = [
      msg({ id: 'a', role: 'moderator', content: 'hi', turn: 1, createdAt: 1 }),
      msg({ id: 'b', role: 'user', content: 'question', turn: 2, createdAt: 2 }),
      msg({ id: 'c', role: 'participant', content: 'answer', turn: 3, createdAt: 3 }),
    ]

    const map = buildMessageSequenceMap(messages)
    expect(map.get('a')).toBe(1)
    expect(map.get('b')).toBe(2)
    expect(map.get('c')).toBe(3)
  })

  it('backfills missing turns in chronological order', () => {
    const messages = [
      msg({ id: 'a', role: 'user', content: 'q', createdAt: 2 }),
      msg({ id: 'b', role: 'moderator', content: 'a', createdAt: 3 }),
    ]

    const map = buildMessageSequenceMap(messages)
    expect(map.get('a')).toBe(1)
    expect(map.get('b')).toBe(2)
    expect(formatMessageSequence(messages[0], map)).toBe('#1')
  })

  it('skips system messages', () => {
    const messages = [
      msg({ id: 's', role: 'system', content: 'err', createdAt: 1 }),
      msg({ id: 'u', role: 'user', content: 'ok', createdAt: 2 }),
    ]

    const map = buildMessageSequenceMap(messages)
    expect(map.has('s')).toBe(false)
    expect(messageSequenceNumber(messages[0], map)).toBeNull()
    expect(map.get('u')).toBe(1)
  })
})
