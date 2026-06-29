import { apiFetch } from '@/api/client'
import type { MeetingDetail, MeetingsResponse } from '@/types/meeting'

const PAGE_SIZE = 12
const PAGE_SIZE_DESKTOP = 12

export function fetchMeetings(page = 1, pageSize = PAGE_SIZE) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })
  return apiFetch<MeetingsResponse>(`/meetings?${params}`)
}

export function fetchMeeting(id: string) {
  return apiFetch<MeetingDetail>(`/meetings/${encodeURIComponent(id)}`)
}

export { PAGE_SIZE as MEETINGS_PAGE_SIZE, PAGE_SIZE_DESKTOP as MEETINGS_PAGE_SIZE_DESKTOP }
