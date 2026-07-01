import { primaryDeliverablePath, type MeetingModeKind } from '@/lib/meeting-labels'

import type { MeetingDetail } from '@/types/meeting'

export interface MeetingBriefPreview {
  topic: string
  goal: string
  agenda: string[]
  inScope: string
  outOfScope: string
  doneCriteria: string
  conclusion?: string
  conclusionSource?: string
}

export function parseNumberedList(body: string): string[] {
  const text = body.trim()
  if (!text) return []

  const items: string[] = []
  for (const line of text.split('\n')) {
    let item = line.trim()
    if (!item) continue

    const ordered = item.match(/^(\d{1,3})\.\s+(.+)$/)
    if (ordered) {
      item = ordered[2]!.trim()
    }

    item = item.replace(/^\*\*(.+)\*\*$/, '$1').replace(/\*\*/g, '')
    if (item) items.push(item)
  }
  return items
}

export function extractMarkdownSection(doc: string, heading: string): string {
  const marker = `## ${heading}`
  const idx = doc.indexOf(marker)
  if (idx < 0) return ''

  let rest = doc.slice(idx + marker.length).replace(/^[\s\r\n]+/, '')
  const nextHeading = rest.search(/\n## /)
  if (nextHeading >= 0) {
    rest = rest.slice(0, nextHeading)
  }
  const hr = rest.indexOf('\n---')
  if (hr >= 0) {
    rest = rest.slice(0, hr)
  }
  return rest.trim()
}

const CONCLUSION_HEADINGS = [
  '结论',
  '核心结论',
  '已决要点',
  '核心方案',
  'Consensus',
  'Executive Summary',
] as const

function firstMeaningfulParagraph(content: string): string {
  for (const line of content.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed) continue
    if (trimmed.startsWith('#')) continue
    if (trimmed.startsWith('|')) continue
    if (trimmed.startsWith('**Topic')) continue
    if (trimmed.startsWith('Total tokens')) continue
    if (trimmed.startsWith('Strategy:')) continue
    return trimmed.replace(/^[-*]\s+/, '')
  }
  return ''
}

export function extractDeliverableSummary(content: string): string {
  const doc = content.trim()
  if (!doc) return ''

  for (const heading of CONCLUSION_HEADINGS) {
    const section = extractMarkdownSection(doc, heading)
    if (section) {
      const paragraph = firstMeaningfulParagraph(section) || section.split('\n')[0]?.trim()
      if (paragraph) {
        if (heading === 'Consensus' && /^Strategy:/i.test(paragraph)) {
          const strategy = paragraph.replace(/^Strategy:\s*/i, '').replace(/\s*\(.*\)$/, '').trim()
          return strategy ? `已达成共识（${strategy}）` : paragraph
        }
        return paragraph
      }
    }
  }

  return firstMeaningfulParagraph(doc)
}

export function parseMeetingBriefPreview(
  detail: MeetingDetail,
  modeKind?: MeetingModeKind,
): MeetingBriefPreview {
  const meetingMd = detail.files?.['MEETING.md'] ?? ''
  const topicFromDoc = extractMarkdownSection(meetingMd, '会议主题')
  const topic = topicFromDoc || detail.topic?.trim() || '（无主题）'
  const goal = extractMarkdownSection(meetingMd, '会议目标')

  const deliverablePath = primaryDeliverablePath(modeKind)
  let deliverable = detail.files?.[deliverablePath]?.trim() ?? ''
  let conclusion = extractDeliverableSummary(deliverable)
  let conclusionSource = conclusion ? deliverablePath : undefined

  if (!conclusion) {
    const recap = detail.files?.['moderator/executive-recap.md']?.trim() ?? ''
    const recapSummary = extractDeliverableSummary(recap)
    if (recapSummary) {
      conclusion = recapSummary
      conclusionSource = 'moderator/executive-recap.md'
    }
  }

  return {
    topic,
    goal,
    agenda: parseNumberedList(extractMarkdownSection(meetingMd, '议程')),
    inScope: extractMarkdownSection(meetingMd, '讨论范围'),
    outOfScope: extractMarkdownSection(meetingMd, '不在范围'),
    doneCriteria: extractMarkdownSection(meetingMd, '完成标准'),
    conclusion: conclusion || undefined,
    conclusionSource,
  }
}
