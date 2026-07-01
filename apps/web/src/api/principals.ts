import { apiFetch } from '@/api/client'
import type { PrincipalDetail, PrincipalUserProfile, PrincipalsResponse } from '@/types/principal'

export function fetchPrincipals() {
  return apiFetch<PrincipalsResponse>('/principals')
}

export function fetchPrincipal(id: string) {
  return apiFetch<PrincipalDetail>(`/principals/${encodeURIComponent(id)}`)
}

export function savePrincipalUserProfile(id: string, profile: PrincipalUserProfile) {
  return apiFetch<{ status: string; user_profile: PrincipalUserProfile }>(
    `/principals/${encodeURIComponent(id)}/user-profile`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(profile),
    },
  )
}
