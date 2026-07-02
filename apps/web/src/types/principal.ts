export interface PrincipalIndex {
  id: string
  display_name?: string
  files: string[]
  updated_at: string
}

export interface PrincipalsResponse {
  principals: PrincipalIndex[]
}

export interface PrincipalUserProfile {
  language: string
  confirmation?: string
  context?: string
}

export interface PrincipalPersonaMeta {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export interface PrincipalDetail {
  id: string
  display_name?: string
  active_persona_id: string
  personas: PrincipalPersonaMeta[]
  user_profile: PrincipalUserProfile
}

export const EMPTY_PRINCIPAL_USER_PROFILE: PrincipalUserProfile = {
  language: 'zh-CN',
  confirmation: '',
  context: '',
}
