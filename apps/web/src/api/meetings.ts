import { apiFetch } from '@/api/client'
import type { MeetingDetail, MeetingIndex, MeetingsResponse } from '@/types/meeting'

const PAGE_SIZE = 12
const PAGE_SIZE_DESKTOP = 12

export function fetchMeetings(page = 1, pageSize = PAGE_SIZE) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })
  return apiFetch<MeetingsResponse>(`/meetings?${params}`)
}

/** 拉取全部会议索引（用于仪表盘等全量统计） */
export async function fetchAllMeetings(): Promise<MeetingIndex[]> {
  const peek = await fetchMeetings(1, 1)
  const total = peek.total ?? peek.meetings?.length ?? 0
  if (total === 0) return []
  if (total <= (peek.meetings?.length ?? 0)) {
    return peek.meetings ?? []
  }
  const full = await fetchMeetings(1, total)
  return full.meetings ?? []
}

export function fetchMeeting(id: string) {
  return apiFetch<MeetingDetail>(`/meetings/${encodeURIComponent(id)}`)
}

export { PAGE_SIZE as MEETINGS_PAGE_SIZE, PAGE_SIZE_DESKTOP as MEETINGS_PAGE_SIZE_DESKTOP }
