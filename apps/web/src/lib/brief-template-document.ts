import type { BriefTemplateDocument } from '@/types/brief-template'

export function emptyBriefDocument(title = ''): BriefTemplateDocument {
  return {
    meta: { title },
    topic: '',
    brief: {
      goal: '',
      agenda: [''],
      in_scope: '',
      out_of_scope: '',
      done_criteria: '',
    },
    meeting: {
      mode: 'decision',
      max_rounds: 3,
      min_rounds_before_synthesis: 2,
      confirmation_mode: 'required',
      free_dialogue_max_questions: 1,
      participant_ids: [],
    },
  }
}

export function normalizeBriefDocument(doc: BriefTemplateDocument): BriefTemplateDocument {
  const agenda = (doc.brief.agenda ?? []).map((item) => item.trim()).filter(Boolean)
  return {
    ...doc,
    meta: {
      title: doc.meta.title.trim(),
      description: doc.meta.description?.trim() || undefined,
      owner: doc.meta.owner?.trim() || undefined,
    },
    topic: doc.topic?.trim() ?? '',
    brief: {
      goal: doc.brief.goal?.trim() ?? '',
      agenda: agenda.length > 0 ? agenda : [''],
      in_scope: doc.brief.in_scope?.trim() ?? '',
      out_of_scope: doc.brief.out_of_scope?.trim() ?? '',
      done_criteria: doc.brief.done_criteria?.trim() ?? '',
    },
    meeting: {
      mode: doc.meeting?.mode?.trim() || 'decision',
      max_rounds: doc.meeting?.max_rounds ?? 3,
      min_rounds_before_synthesis: doc.meeting?.min_rounds_before_synthesis ?? 2,
      confirmation_mode: doc.meeting?.confirmation_mode?.trim() || 'required',
      free_dialogue_max_questions: doc.meeting?.free_dialogue_max_questions ?? 1,
      participant_ids: doc.meeting?.participant_ids ?? [],
    },
  }
}

export function documentsEqual(a: BriefTemplateDocument, b: BriefTemplateDocument): boolean {
  return JSON.stringify(normalizeBriefDocument(a)) === JSON.stringify(normalizeBriefDocument(b))
}
