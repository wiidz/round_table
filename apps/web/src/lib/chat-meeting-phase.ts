import type { ChatMessage } from '@/types/chat'

export type ChatMeetingPhase = 'idle' | 'setup' | 'running' | 'post'

/** Parse moderator 会议状态 / input phase replies (generic transport copy). */
export function parsePhaseFromStatusReply(content: string): ChatMeetingPhase | null {
  const text = content.trim()
  if (!text) return null

  if (
    text.includes('会议进行中') ||
    text.includes('Meeting running') ||
    text.includes('自由问答') ||
    text.includes('Free dialogue') ||
    text.includes('会议已暂停') ||
    text.includes('Meeting paused') ||
    text.includes('确认关') ||
    text.includes('Confirmation')
  ) {
    return 'running'
  }

  if (text.includes('会议已结束') || text.includes('Meeting finished')) {
    return 'post'
  }

  if (
    text.includes('配置 ·') ||
    text.includes('Setup ·') ||
    text.includes('接待 ·') ||
    text.includes('Reception ·') ||
    text.includes('专家 ·') ||
    text.includes('Expert ·')
  ) {
    return 'setup'
  }

  if (text.includes('当前输入态：空闲') || text.includes('Input phase: Idle')) {
    return 'idle'
  }

  return null
}

const MEETING_ID_PATTERN = /`(mtg-[a-zA-Z0-9-]+)`/

/** Extract meeting id from transport status / launch replies. */
export function parseMeetingIdFromStatusReply(content: string): string | null {
  const text = content.trim()
  if (!text) return null

  const idLine = text.match(/🆔\s*`(mtg-[^`\s]+)`/)
  if (idLine?.[1]) return idLine[1]

  const backtick = text.match(MEETING_ID_PATTERN)
  if (backtick?.[1]) return backtick[1]

  return null
}

export function parseMeetingIdFromMessages(messages: ChatMessage[]): string | null {
  for (let i = messages.length - 1; i >= 0; i--) {
    const message = messages[i]!
    if (message.role !== 'moderator' && message.role !== 'system') continue
    const id = parseMeetingIdFromStatusReply(message.content)
    if (id) return id
  }
  return null
}

export function inferChatMeetingPhase(messages: ChatMessage[]): ChatMeetingPhase {
  for (let i = messages.length - 1; i >= 0; i--) {
    const message = messages[i]!
    if (message.role !== 'moderator' && message.role !== 'system') continue
    const parsed = parsePhaseFromStatusReply(message.content)
    if (parsed) return parsed
  }

  const hasParticipantTurn = messages.some(
    (m) => m.role === 'participant' && m.turn != null,
  )
  if (hasParticipantTurn) return 'running'

  const hasTurn = messages.some((m) => m.turn != null)
  if (hasTurn) return 'setup'

  return 'idle'
}

export function phaseLabel(phase: ChatMeetingPhase): string {
  switch (phase) {
    case 'setup':
      return '配置中'
    case 'running':
      return '会议进行中'
    case 'post':
      return '已结束'
    default:
      return '空闲'
  }
}

export function suggestedViewMode(phase: ChatMeetingPhase): 'list' | 'roundtable' {
  return phase === 'running' || phase === 'post' ? 'roundtable' : 'list'
}
