export interface MeetingIndex {
  id: string
  topic: string
  status: string
  mode?: string
  mode_kind?: 'decision' | 'deliberation'
  started_at?: string
  participant_count?: number
  max_rounds?: number
  free_dialogue?: boolean
  llm_call_count?: number
  total_tokens?: number
  updated_at: string
}

export interface MeetingDetail extends MeetingIndex {
  files: Record<string, string>
}

export interface MeetingsResponse {
  meetings: MeetingIndex[]
  total: number
  page: number
  page_size: number
}
