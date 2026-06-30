import { describe, expect, it } from 'vitest'

import type { ChatMessage } from '@/types/chat'
import { resolveLiveBubbleVariant } from '@/lib/live-bubble-variant'

function msg(turn: number, id: string): ChatMessage {
  return {
    id,
    role: 'participant',
    content: 'hello',
    turn,
    createdAt: turn,
    authorId: id,
  }
}

describe('resolveLiveBubbleVariant', () => {
  it('marks highlighted message as active', () => {
    const message = msg(3, 'a')
    expect(resolveLiveBubbleVariant(message, 'a', 5)).toBe('active')
  })

  it('classifies turns before reference as before', () => {
    expect(resolveLiveBubbleVariant(msg(2, 'a'), 'b', 5)).toBe('before')
  })

  it('classifies turns after reference as after', () => {
    expect(resolveLiveBubbleVariant(msg(7, 'c'), 'a', 5)).toBe('after')
  })
})
