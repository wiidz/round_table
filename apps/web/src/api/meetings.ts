import { apiFetch } from '@/api/client'
import type { MeetingsResponse } from '@/types/meeting'

const PAGE_SIZE = 10

export function fetchMeetings(page = 1, pageSize = PAGE_SIZE) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })
  return apiFetch<MeetingsResponse>(`/meetings?${params}`)
}

export { PAGE_SIZE as MEETINGS_PAGE_SIZE }
