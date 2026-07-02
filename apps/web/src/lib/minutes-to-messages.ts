import { assignTurnForRole } from '@/lib/assign-turn'
import type { ChatMessage, ChatRole } from '@/types/chat'

const PARTICIPANT_LINE =
  /^-\s*\*\*([^*]+)\*\*\s*\(([^)]+)\):\s*(.*)$/
const STANCE_SUFFIX = /\s+_\[[^\]]+\]_\s*$/
const SECTION_H2 = /^##\s+(.+)$/
const SECTION_H3 = /^###\s+(.+)$/
const FREE_DIALOGUE_Q = /^Q:\s*(.*)$/
const FREE_DIALOGUE_A = /^A:\s*(.*)$/

interface ParseContext {
  meetingId: string
  baseTime: number
  messages: ChatMessage[]
  nextTurn: number
  msgIndex: number
}

function pushMessage(
  ctx: ParseContext,
  role: ChatRole,
  content: string,
  authorId?: string,
  authorName?: string,
): void {
  const trimmed = content.trim()
  if (!trimmed) return

  const { turn, nextTurn } = assignTurnForRole(role, ctx.nextTurn)
  ctx.nextTurn = nextTurn

  ctx.messages.push({
    id: `${ctx.meetingId}-t${ctx.msgIndex++}`,
    role,
    content: trimmed,
    authorId,
    authorName,
    turn,
    createdAt: ctx.baseTime + (turn ?? ctx.msgIndex) * 1000,
  })
}

function flushParticipantBullet(
  ctx: ParseContext,
  authorId: string,
  content: string,
): void {
  const body = content.replace(STANCE_SUFFIX, '').trim()
  pushMessage(ctx, 'participant', body, authorId.trim(), authorId.trim())
}

function parseParticipantBullets(ctx: ParseContext, body: string): void {
  let pendingId: string | null = null
  let pendingContent: string[] = []

  const flush = () => {
    if (!pendingId) return
    flushParticipantBullet(ctx, pendingId, pendingContent.join('\n'))
    pendingId = null
    pendingContent = []
  }

  for (const rawLine of body.split('\n')) {
    const line = rawLine.trimEnd()
    const match = line.match(PARTICIPANT_LINE)
    if (match) {
      flush()
      pendingId = match[1] ?? ''
      pendingContent = [match[3] ?? '']
      continue
    }
    if (pendingId && line.trim()) {
      pendingContent.push(line.trim())
    }
  }
  flush()
}

function parseFreeDialogueBlock(ctx: ParseContext, body: string): void {
  let pendingAsker: string | null = null
  let pendingAnswerer: string | null = null
  let principalQuestion = false
  let pendingQuestion = ''

  for (const rawLine of body.split('\n')) {
    const line = rawLine.trim()
    if (!line || line.startsWith('Free dialogue after Round')) continue

    if (line.includes('Principal') && line.includes('via Moderator')) {
      principalQuestion = true
      const answerer = line.match(/→ \*\*([^*]+)\*\*/)
      pendingAnswerer = answerer?.[1]?.trim() ?? null
      pendingAsker = null
      continue
    }

    const route = line.match(/^\*\*([^*]+)\*\* \([^)]+\) → \*\*([^*]+)\*\*/)
    if (route) {
      principalQuestion = false
      pendingAsker = route[1]?.trim() ?? null
      pendingAnswerer = route[2]?.trim() ?? null
      continue
    }

    const qMatch = line.match(FREE_DIALOGUE_Q)
    if (qMatch) {
      pendingQuestion = qMatch[1]?.trim() ?? ''
      continue
    }

    const aMatch = line.match(FREE_DIALOGUE_A)
    if (aMatch && pendingAnswerer) {
      if (principalQuestion) {
        pushMessage(ctx, 'user', pendingQuestion)
      } else if (pendingAsker) {
        pushMessage(ctx, 'participant', pendingQuestion, pendingAsker, pendingAsker)
      }
      pushMessage(ctx, 'participant', aMatch[1]?.trim() ?? '', pendingAnswerer, pendingAnswerer)
      pendingQuestion = ''
      pendingAsker = null
      pendingAnswerer = null
      principalQuestion = false
    }
  }
}

function parseMinutesSections(ctx: ParseContext, minutesMd: string): void {
  const lines = minutesMd.split('\n')
  let roundOpen = false
  let sectionKind: 'round' | 'free-dialogue' | 'moderator-summary' | 'skip' = 'skip'
  let sectionBody: string[] = []

  const flushSection = () => {
    const body = sectionBody.join('\n').trim()
    sectionBody = []
    if (!body) return

    if (sectionKind === 'round') {
      parseParticipantBullets(ctx, body)
    } else if (sectionKind === 'free-dialogue') {
      parseFreeDialogueBlock(ctx, body)
    } else if (sectionKind === 'moderator-summary') {
      pushMessage(ctx, 'moderator', body, undefined, '主持人')
    }
  }

  for (const rawLine of lines) {
    const inSubSection =
      sectionKind === 'free-dialogue' || sectionKind === 'moderator-summary'

    const h3 = rawLine.match(SECTION_H3)
    if (h3 && roundOpen && !inSubSection) {
      flushSection()
      const title = h3[1]?.trim() ?? ''
      if (title === 'Free dialogue') {
        sectionKind = 'free-dialogue'
      } else if (title === 'Moderator summary') {
        sectionKind = 'moderator-summary'
      } else {
        sectionKind = 'skip'
      }
      continue
    }

    const h2 = rawLine.match(SECTION_H2)
    if (h2 && !inSubSection) {
      flushSection()
      roundOpen = false
      const title = h2[1]?.trim() ?? ''
      if (title.startsWith('Token usage') || title === 'Synthesis' || title.startsWith('Consensus')) {
        sectionKind = 'skip'
      } else if (title.startsWith('Pre-meeting') || /^Round \d+/.test(title)) {
        sectionKind = 'round'
        roundOpen = true
      } else {
        sectionKind = 'skip'
      }
      continue
    }

    if (sectionKind !== 'skip') {
      sectionBody.push(rawLine)
    }
  }
  flushSection()
}

function parseRoundFileContent(ctx: ParseContext, content: string): void {
  const body = content.replace(/^#[^\n]*\n+/, '').trim()
  parseParticipantBullets(ctx, body)
}

function sortedRoundPaths(files: Record<string, string>): string[] {
  return Object.keys(files)
    .filter((p) => /^rounds\/round-\d+\.md$/.test(p) || p === 'pre-meeting/perspectives.md')
    .sort((a, b) => {
      const rank = (path: string) => {
        if (path === 'pre-meeting/perspectives.md') return 0
        const n = path.match(/round-(\d+)/)
        return n ? parseInt(n[1], 10) : 999
      }
      return rank(a) - rank(b)
    })
}

/** Return the highest round number covered by a MINUTES.md string (0 = none). */
function maxRoundInMinutes(minutesMd: string): number {
  let max = 0
  for (const m of minutesMd.matchAll(/^## Round (\d+)/gm)) {
    const n = parseInt(m[1]!, 10)
    if (n > max) max = n
  }
  return max
}

function roundNumberFromPath(path: string): number {
  if (path === 'pre-meeting/perspectives.md') return 0
  const m = path.match(/round-(\d+)/)
  return m ? parseInt(m[1]!, 10) : -1
}

export function parseMinutesMarkdown(
  minutesMd: string,
  options: { meetingId: string; startedAt?: string },
): ChatMessage[] {
  const ctx: ParseContext = {
    meetingId: options.meetingId,
    baseTime: options.startedAt ? Date.parse(options.startedAt) || Date.now() : Date.now(),
    messages: [],
    nextTurn: 1,
    msgIndex: 0,
  }
  parseMinutesSections(ctx, minutesMd)
  return ctx.messages
}

/**
 * Build replay messages from workspace files.
 *
 * Priority:
 * 1. MINUTES.md (root) – preferred; contains moderator summaries + all rounds.
 * 2. artifacts/minutes.md – alternate path for the same document.
 * 3. round files fallback when neither MINUTES.md variant exists.
 *
 * When a MINUTES.md is found, we also check whether newer round files exist
 * (higher round number than the last ## Round N in the minutes) and append
 * them so that an in-progress meeting shows the most recent rounds.
 */
export function workspaceTranscriptMessages(
  files: Record<string, string>,
  meetingId: string,
  startedAt?: string,
): ChatMessage[] {
  const minutesContent =
    files['MINUTES.md']?.trim() || files['artifacts/minutes.md']?.trim()

  const baseTime = startedAt ? Date.parse(startedAt) || Date.now() : Date.now()

  if (minutesContent) {
    const messages = parseMinutesMarkdown(minutesContent, { meetingId, startedAt })

    // Supplement with round files that represent rounds newer than MINUTES.md
    const coveredMax = maxRoundInMinutes(minutesContent)
    const extraPaths = sortedRoundPaths(files).filter(
      (p) => roundNumberFromPath(p) > coveredMax,
    )
    if (extraPaths.length > 0) {
      const ctx: ParseContext = {
        meetingId,
        baseTime,
        messages,
        nextTurn: messages.length > 0
          ? Math.max(...messages.map((m) => (m.turn ?? 0))) + 1
          : 1,
        msgIndex: messages.length,
      }
      for (const path of extraPaths) {
        const content = files[path]
        if (content?.trim()) {
          parseRoundFileContent(ctx, content)
        }
      }
    }

    return messages
  }

  // Fallback: parse all available round files
  const ctx: ParseContext = {
    meetingId,
    baseTime,
    messages: [],
    nextTurn: 1,
    msgIndex: 0,
  }
  for (const path of sortedRoundPaths(files)) {
    const content = files[path]
    if (content?.trim()) {
      parseRoundFileContent(ctx, content)
    }
  }
  return ctx.messages
}

export function hasWorkspaceTranscript(files: Record<string, string>): boolean {
  if (files['MINUTES.md']?.trim()) return true
  if (files['artifacts/minutes.md']?.trim()) return true
  return sortedRoundPaths(files).some((p) => files[p]?.trim())
}
