export interface MeetingIndex {
  id: string
  topic: string
  status: string
  updated_at: string
}

export interface MeetingsResponse {
  meetings: MeetingIndex[]
  total: number
  page: number
  page_size: number
}
