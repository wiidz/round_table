import { describe, expect, it } from 'vitest'

import type { ChatMessage } from '@/types/chat'
import { filterTranscriptBySpeaker, listTranscriptSpeakers } from '@/lib/transcript-speakers'

function msg(partial: Partial<ChatMessage> & Pick<ChatMessage, 'id' | 'role' | 'content'>): ChatMessage {
  return { createdAt: 0, ...partial }
}

describe('listTranscriptSpeakers', () => {
  it('lists unique speakers in order, skips system', () => {
    const messages = [
      msg({ id: '1', role: 'moderator', content: 'a' }),
      msg({ id: '2', role: 'participant', content: 'b', authorId: 'dev', authorName: '开发' }),
      msg({ id: '3', role: 'system', content: 'c' }),
      msg({ id: '4', role: 'participant', content: 'd', authorId: 'dev', authorName: '开发' }),
    ]
    expect(listTranscriptSpeakers(messages)).toEqual([
      { id: 'moderator', label: '主持人' },
      { id: 'dev', label: '开发' },
    ])
  })
})

describe('filterTranscriptBySpeaker', () => {
  it('returns all messages when filter is null', () => {
    const messages = [msg({ id: '1', role: 'user', content: 'hi' })]
    expect(filterTranscriptBySpeaker(messages, null)).toHaveLength(1)
  })

  it('filters by speaker id', () => {
    const messages = [
      msg({ id: '1', role: 'moderator', content: 'a' }),
      msg({ id: '2', role: 'user', content: 'b' }),
    ]
    expect(filterTranscriptBySpeaker(messages, 'user')).toHaveLength(1)
    expect(filterTranscriptBySpeaker(messages, 'user')[0]?.id).toBe('2')
  })
})
