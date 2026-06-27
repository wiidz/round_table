import { apiFetch } from '@/api/client'
import type { PrincipalDetail, PrincipalsResponse } from '@/types/principal'

export function fetchPrincipals() {
  return apiFetch<PrincipalsResponse>('/principals')
}

export function fetchPrincipal(id: string) {
  return apiFetch<PrincipalDetail>(`/principals/${encodeURIComponent(id)}`)
}

export function savePrincipalFile(id: string, filename: string, content: string) {
  return apiFetch<{ status: string }>(
    `/principals/${encodeURIComponent(id)}/files/${encodeURIComponent(filename)}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    },
  )
}
