export interface SettingsSubsectionMeta {
  id: string
  label: string
  available: boolean
}

export interface DiscordBotState {
  /** moderator 为 "moderator"；参与 Bot 为 Discord Application ID */
  id: string
  label?: string
  display_name?: string
  primary: boolean
  deletable: boolean
  env_key: string
  configured: boolean
  restart_required?: boolean
  discord_application_id?: string
  discord_username?: string
  avatar_url?: string
  profile_fetched_at?: string
  token_masked?: string
  token?: string
  bound_participant_id?: string
}

export interface DiscordBotInput {
  application_id?: string
  token?: string
  bound_participant_id?: string
}

export interface DiscordBotsUpdate {
  moderator_token?: string
  moderator_role_token?: string
  moderator_role_id?: string
  participants: DiscordBotInput[]
}

export interface SettingsFieldState {
  key: string
  value?: string
  configured: boolean
  secret: boolean
  editable: boolean
  restart_required?: boolean
  label: string
  group: string
  subsection?: string
  section?: string
  placeholder?: string
  description?: string
  input_type?: 'number' | 'select' | 'switch' | 'radio'
  options?: { value: string; label: string }[]
  min?: number
  max?: number
}

export type DiscordTransportPhase = 'stopped' | 'starting' | 'ready'

export interface DiscordTransportStatus {
  running: boolean
  phase?: DiscordTransportPhase
  pid?: number
  started_at?: string
  ready_at?: string
  last_exit?: string
  log_path?: string
}

export interface DiscordTransportLogs {
  path: string
  lines: string
}

export interface MeetPresetConfig {
  id: string
  group: 'deliberation' | 'decision' | string
  icon: string
  name_zh: string
  name_en: string
  mode: string
  max_rounds: number
  confirmation: string
  free_dialogue_questions: number
  command?: string
}

export interface SettingsResponse {
  source: string
  secrets_path: string
  groups: string[]
  subsections: Record<string, SettingsSubsectionMeta[]>
  fields: SettingsFieldState[]
  discord_bots?: DiscordBotState[]
  meet_presets?: MeetPresetConfig[]
  meet_presets_defaults?: MeetPresetConfig[]
}

export type SettingsValues = Record<string, string>
