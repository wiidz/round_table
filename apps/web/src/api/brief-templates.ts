import { apiFetch } from '@/api/client'
import type {
  BriefTemplateDetail,
  BriefTemplatesResponse,
  CloneBriefResponse,
} from '@/types/brief-template'

export function fetchBriefTemplates() {
  return apiFetch<BriefTemplatesResponse>('/brief-templates')
}

export function fetchBriefTemplate(id: string) {
  return apiFetch<BriefTemplateDetail>(`/brief-templates/${encodeURIComponent(id)}`)
}

export function saveBriefTemplate(id: string, content: string) {
  return apiFetch<{ status: string }>(`/brief-templates/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content }),
  })
}

export function cloneBriefFromMeeting(meetingId: string) {
  return apiFetch<CloneBriefResponse>('/meetings/clone-brief', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ meeting_id: meetingId }),
  })
}
