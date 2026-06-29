import { apiFetch } from '@/api/client'
import type { RuntimeResponse } from '@/types/runtime'

export function fetchRuntime() {
  return apiFetch<RuntimeResponse>('/system/runtime')
}
