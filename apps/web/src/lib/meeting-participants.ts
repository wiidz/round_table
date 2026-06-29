import type { RosterSeatInput } from '@/lib/round-table-layout'
import type { ChatMessage } from '@/types/chat'

/** Experts who have sent at least one participant message in this session. */
export function participantsFromMessages(messages: ChatMessage[]): RosterSeatInput[] {
  const seen = new Map<string, string>()
  for (const message of messages) {
    if (message.role !== 'participant') continue
    const id = message.authorId?.trim()
    if (!id || seen.has(id)) continue
    seen.set(id, message.authorName?.trim() || id)
  }
  return [...seen.entries()].map(([id, label]) => ({ id, label }))
}

const PARTICIPANT_PAIR = /([a-zA-Z0-9_-]+)\u00b7([^,\n]+)/g

/** Parse `id·display` pairs from Discord setup ParticipantsSummary. */
export function parseParticipantSummaryText(text: string): RosterSeatInput[] {
  const seen = new Map<string, string>()
  for (const match of text.matchAll(PARTICIPANT_PAIR)) {
    const id = match[1]?.trim()
    const label = match[2]?.trim()
    if (!id || !label || seen.has(id)) continue
    seen.set(id, label)
  }
  return [...seen.entries()].map(([id, label]) => ({ id, label }))
}

/** Parse participant table under ## 参会人员 in MEETING.md. */
export function parseParticipantsFromMeetingMd(content: string): RosterSeatInput[] {
  const section = content.match(/##\s*参会人员[\s\S]*?(?=\n##|\n---|\n\*\*Token|$)/)
  if (!section) return []

  const seen = new Map<string, string>()
  for (const line of section[0].split('\n')) {
    if (!line.trim().startsWith('|')) continue
    if (line.includes('参会者') || line.includes('---')) continue
    const cells = line
      .split('|')
      .map((c) => c.trim())
      .filter(Boolean)
    if (cells.length < 2) continue
    const id = cells[0]!
    const role = cells[1]!
    if (id === '参会者' || id === 'Participant') continue
    if (!seen.has(id)) {
      seen.set(id, role !== id ? role : id)
    }
  }
  return [...seen.entries()].map(([id, label]) => ({ id, label }))
}

/** Parse display-name list from launch ack (`👥 参会：A、B`). */
export function parseParticipantDisplayList(text: string): string[] {
  const line = text.match(/[-\s]*👥\s*(?:参会|Participants?)[:：]\s*([^\n]+)/i)
  if (!line?.[1]) return []
  const raw = line[1].trim()
  if (!raw || raw === '全员' || raw.toLowerCase() === 'full roster') return []
  return raw
    .split(/[,、]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

function resolveDisplayNames(names: string[], roster: RosterSeatInput[]): RosterSeatInput[] {
  const byLabel = new Map<string, string>()
  for (const p of roster) {
    byLabel.set(p.label.toLowerCase(), p.id)
    byLabel.set(p.id.toLowerCase(), p.id)
  }
  const out: RosterSeatInput[] = []
  const seen = new Set<string>()
  for (const name of names) {
    const id = byLabel.get(name.toLowerCase())
    if (!id || seen.has(id)) continue
    seen.add(id)
    const rosterEntry = roster.find((p) => p.id === id)
    out.push({ id, label: rosterEntry?.label ?? name })
  }
  return out
}

/** Latest meeting lineup from moderator/system messages (setup confirm / launch). */
export function parseMeetingParticipantsFromMessages(
  messages: ChatMessage[],
  roster: RosterSeatInput[],
): RosterSeatInput[] {
  for (let i = messages.length - 1; i >= 0; i--) {
    const message = messages[i]!
    if (message.role !== 'moderator' && message.role !== 'system') continue
    const content = message.content

    const fromSummary = parseParticipantSummaryText(content)
    if (fromSummary.length > 0) return fromSummary

    const displayNames = parseParticipantDisplayList(content)
    if (displayNames.length > 0) {
      const resolved = resolveDisplayNames(displayNames, roster)
      if (resolved.length > 0) return resolved
    }
  }
  return []
}

export function resolveMeetingLineup(
  phase: 'idle' | 'setup' | 'running' | 'post',
  options: {
    roster: RosterSeatInput[]
    meetingMdParticipants: RosterSeatInput[]
    messageParticipants: RosterSeatInput[]
    spokenParticipants: RosterSeatInput[]
  },
): RosterSeatInput[] {
  const { roster, meetingMdParticipants, messageParticipants, spokenParticipants } = options

  if (phase === 'running' || phase === 'post') {
    if (meetingMdParticipants.length > 0) return meetingMdParticipants
    if (messageParticipants.length > 0) return messageParticipants
    if (spokenParticipants.length > 0) return spokenParticipants
    return roster
  }

  return []
}
