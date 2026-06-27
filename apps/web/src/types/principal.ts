export interface PrincipalIndex {
  id: string
  display_name?: string
  files: string[]
  updated_at: string
}

export interface PrincipalsResponse {
  principals: PrincipalIndex[]
}

export interface PrincipalDetail {
  id: string
  display_name?: string
  files: Record<string, string>
}
