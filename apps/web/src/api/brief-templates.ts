import { apiFetch } from '@/api/client'
import type {
  BriefTemplateDetail,
  BriefTemplateDocument,
  BriefTemplatesResponse,
  CloneBriefResponse,
} from '@/types/brief-template'

export function fetchBriefTemplates() {
  return apiFetch<BriefTemplatesResponse>('/brief-templates')
}

export function fetchBriefTemplate(id: string) {
  return apiFetch<BriefTemplateDetail>(`/brief-templates/${encodeURIComponent(id)}`)
}

export function createBriefTemplate(payload: { document: BriefTemplateDocument }) {
  return apiFetch<{ id: string; status: string }>('/brief-templates', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
}

export function saveBriefTemplate(id: string, payload: { document: BriefTemplateDocument }) {
  return apiFetch<{ status: string }>(`/brief-templates/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
}

export function cloneBriefFromMeeting(meetingId: string) {
  return apiFetch<CloneBriefResponse>('/meetings/clone-brief', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ meeting_id: meetingId }),
  })
}
