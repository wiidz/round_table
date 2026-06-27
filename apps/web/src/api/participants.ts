import { apiFetch } from '@/api/client'
import type {
  ParticipantDetail,
  ParticipantRosterInput,
  ParticipantsResponse,
} from '@/types/participant'

export function fetchParticipants() {
  return apiFetch<ParticipantsResponse>('/participants')
}

export function fetchParticipant(id: string) {
  return apiFetch<ParticipantDetail>(`/participants/${encodeURIComponent(id)}`)
}

export function createParticipant(input: ParticipantRosterInput) {
  return apiFetch<ParticipantsResponse>('/participants', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
}

export function updateParticipant(id: string, input: ParticipantRosterInput) {
  return apiFetch<ParticipantsResponse>(`/participants/${encodeURIComponent(id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
}

export function deleteParticipant(id: string) {
  return apiFetch<ParticipantsResponse>(`/participants/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export function saveParticipantFile(id: string, filename: string, content: string) {
  return apiFetch<{ status: string }>(
    `/participants/${encodeURIComponent(id)}/files/${encodeURIComponent(filename)}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    },
  )
}
