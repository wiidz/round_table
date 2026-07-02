import type { BriefTemplateDocument, MeetingDefaults } from '@/types/brief-template'

export function emptyBriefDocument(title = ''): BriefTemplateDocument {
  return {
    meta: { title },
    topic: '',
    brief: {
      goal: '',
      agenda: [],
      in_scope: '',
      out_of_scope: '',
      done_criteria: '',
    },
  }
}

function trimOptionalString(value: string | undefined): string | undefined {
  const trimmed = value?.trim()
  return trimmed ? trimmed : undefined
}

function normalizeMeetingDefaults(
  meeting: BriefTemplateDocument['meeting'],
): MeetingDefaults | undefined {
  if (!meeting) return undefined

  const out: MeetingDefaults = {}
  const mode = trimOptionalString(meeting.mode)
  if (mode) out.mode = mode

  const confirmation = trimOptionalString(meeting.confirmation_mode)
  if (confirmation) out.confirmation_mode = confirmation

  if (meeting.max_rounds != null && meeting.max_rounds > 0) {
    out.max_rounds = meeting.max_rounds
  }
  if (meeting.min_rounds_before_synthesis != null && meeting.min_rounds_before_synthesis > 0) {
    out.min_rounds_before_synthesis = meeting.min_rounds_before_synthesis
  }
  if (meeting.free_dialogue_max_questions != null && meeting.free_dialogue_max_questions >= 0) {
    out.free_dialogue_max_questions = meeting.free_dialogue_max_questions
  }

  const participantIds = (meeting.participant_ids ?? []).map((id) => id.trim()).filter(Boolean)
  if (participantIds.length > 0) out.participant_ids = participantIds

  return Object.keys(out).length > 0 ? out : undefined
}

/** 保存用：去掉空白，不填充开会时默认项 */
export function normalizeBriefDocument(doc: BriefTemplateDocument): BriefTemplateDocument {
  const agenda = (doc.brief.agenda ?? []).map((item) => item.trim()).filter(Boolean)

  return {
    ...doc,
    meta: {
      title: doc.meta.title.trim(),
      description: trimOptionalString(doc.meta.description),
      owner: trimOptionalString(doc.meta.owner),
    },
    topic: doc.topic?.trim() ?? '',
    brief: {
      goal: doc.brief.goal?.trim() ?? '',
      agenda,
      in_scope: doc.brief.in_scope?.trim() ?? '',
      out_of_scope: doc.brief.out_of_scope?.trim() ?? '',
      done_criteria: doc.brief.done_criteria?.trim() ?? '',
    },
    meeting: normalizeMeetingDefaults(doc.meeting),
  }
}

/** 除模板名称外，是否至少有一项预填内容（保存校验） */
export function briefTemplateHasSubstantiveContent(doc: BriefTemplateDocument): boolean {
  const normalized = normalizeBriefDocument(doc)
  if (normalized.meta.description) return true
  if (normalized.topic) return true
  if (normalized.brief.goal) return true
  if (normalized.brief.agenda?.length) return true
  if (normalized.brief.in_scope) return true
  if (normalized.brief.out_of_scope) return true
  if (normalized.brief.done_criteria) return true
  if (normalized.meeting && Object.keys(normalized.meeting).length > 0) return true
  return false
}

export function documentsEqual(a: BriefTemplateDocument, b: BriefTemplateDocument): boolean {
  return JSON.stringify(normalizeBriefDocument(a)) === JSON.stringify(normalizeBriefDocument(b))
}

export function meetingFieldIsSet(
  meeting: MeetingDefaults | undefined,
  field: keyof MeetingDefaults,
): boolean {
  if (!meeting) return false
  const value = meeting[field]
  if (field === 'participant_ids') {
    return Array.isArray(value) && value.length > 0
  }
  if (typeof value === 'number') {
    return value !== undefined && !Number.isNaN(value)
  }
  return typeof value === 'string' && value.trim() !== ''
}
