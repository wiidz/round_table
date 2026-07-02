import { apiFetch } from '@/api/client'
import type {
  PrincipalDetail,
  PrincipalPersonaMeta,
  PrincipalUserProfile,
  PrincipalsResponse,
} from '@/types/principal'

export function fetchPrincipals() {
  return apiFetch<PrincipalsResponse>('/principals')
}

export function fetchPrincipal(id: string) {
  return apiFetch<PrincipalDetail>(`/principals/${encodeURIComponent(id)}`)
}

export function fetchPrincipalPersona(id: string, personaId: string) {
  return apiFetch<{ id: string; persona_id: string; user_profile: PrincipalUserProfile }>(
    `/principals/${encodeURIComponent(id)}/personas/${encodeURIComponent(personaId)}`,
  )
}

export function savePrincipalUserProfile(
  id: string,
  profile: PrincipalUserProfile,
  personaId?: string,
) {
  const path = personaId
    ? `/principals/${encodeURIComponent(id)}/personas/${encodeURIComponent(personaId)}/user-profile`
    : `/principals/${encodeURIComponent(id)}/user-profile`
  return apiFetch<{ status: string; user_profile: PrincipalUserProfile }>(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(profile),
  })
}

export function createPrincipalPersona(id: string, title: string) {
  return apiFetch<{ persona: PrincipalPersonaMeta }>(
    `/principals/${encodeURIComponent(id)}/personas`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title }),
    },
  )
}

export function setActivePrincipalPersona(id: string, personaId: string) {
  return apiFetch<{
    status: string
    active_persona_id: string
    personas: PrincipalPersonaMeta[]
  }>(`/principals/${encodeURIComponent(id)}/personas/active`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ persona_id: personaId }),
  })
}
