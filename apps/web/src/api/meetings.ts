import { apiFetch, ApiError } from '@/api/client'
import type { MeetingDetail, MeetingIndex, MeetingsResponse } from '@/types/meeting'

const API_BASE = import.meta.env.VITE_API_BASE ?? '/api'

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

function parseDownloadFilename(disposition: string | null, fallback: string): string {
  if (!disposition) return fallback
  const match = disposition.match(/filename="([^"]+)"/)
  return match?.[1]?.trim() || fallback
}

/** 下载会议 workspace 压缩包（application/zip） */
export async function downloadMeetingArchive(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/meetings/${encodeURIComponent(id)}/archive`)
  if (!response.ok) {
    let message = response.statusText || 'Download failed'
    try {
      const body = (await response.json()) as { error?: string }
      if (body.error) message = body.error
    } catch {
      // ignore non-json error bodies
    }
    throw new ApiError(message, response.status)
  }

  const blob = await response.blob()
  const filename = parseDownloadFilename(
    response.headers.get('Content-Disposition'),
    `${id}.zip`,
  )
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.click()
  URL.revokeObjectURL(url)
}

export function deleteMeeting(id: string) {
  return apiFetch<{ status: string; meeting_id: string }>(
    `/meetings/${encodeURIComponent(id)}`,
    { method: 'DELETE' },
  )
}

export { PAGE_SIZE as MEETINGS_PAGE_SIZE, PAGE_SIZE_DESKTOP as MEETINGS_PAGE_SIZE_DESKTOP }
