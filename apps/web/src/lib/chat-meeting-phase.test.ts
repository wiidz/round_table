import { describe, expect, it } from 'vitest'

import {
  parseMeetingIdFromMessages,
  parseMeetingIdFromStatusReply,
  parsePhaseFromStatusReply,
} from '@/lib/chat-meeting-phase'
import type { ChatMessage } from '@/types/chat'

describe('parsePhaseFromStatusReply', () => {
  it('detects running phase from Chinese status', () => {
    expect(parsePhaseFromStatusReply('📍 **当前输入态：会议进行中**')).toBe('running')
  })
})

describe('parseMeetingIdFromStatusReply', () => {
  it('extracts id from 🆔 line', () => {
    const text = '📍 **当前输入态：会议进行中**\n🆔 `mtg-dc-1782559391`\n\nhint'
    expect(parseMeetingIdFromStatusReply(text)).toBe('mtg-dc-1782559391')
  })

  it('extracts id from launch ack backticks', () => {
    expect(parseMeetingIdFromStatusReply('会议 `mtg-web-abc123` 已启动')).toBe('mtg-web-abc123')
  })
})

describe('parseMeetingIdFromMessages', () => {
  it('uses latest moderator/system status with meeting id', () => {
    const messages: ChatMessage[] = [
      {
        id: '1',
        role: 'moderator',
        content: '📍 **当前输入态：空闲**',
        createdAt: 1,
      },
      {
        id: '2',
        role: 'moderator',
        content: '📍 **当前输入态：会议进行中**\n🆔 `mtg-dc-test`\n\n',
        createdAt: 2,
      },
    ]
    expect(parseMeetingIdFromMessages(messages)).toBe('mtg-dc-test')
  })
})
