export type BriefTemplateSource = 'builtin' | 'custom'

export interface BriefBody {
  goal?: string
  agenda?: string[]
  in_scope?: string
  out_of_scope?: string
  done_criteria?: string
}

export interface MeetingDefaults {
  mode?: string
  max_rounds?: number
  min_rounds_before_synthesis?: number
  confirmation_mode?: string
  free_dialogue_max_questions?: number
  participant_ids?: string[]
}

export interface LaunchDraft {
  topic: string
  brief: BriefBody
  meeting: MeetingDefaults
}

export interface BriefTemplateIndex {
  id: string
  title: string
  description?: string
  source: BriefTemplateSource
  updated_at: string
}

export interface BriefTemplateDetail extends BriefTemplateIndex {
  content: string
  launch: LaunchDraft
}

export interface BriefTemplatesResponse {
  templates: BriefTemplateIndex[]
}

export interface CloneBriefResponse {
  launch: LaunchDraft
}
