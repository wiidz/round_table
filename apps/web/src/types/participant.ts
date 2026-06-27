export interface ParticipantIMBinding {
  platform: string
  application_id: string
  /** @deprecated 读取兼容 */
  bot_id?: string
}

export interface ParticipantRosterInput {
  id: string
  display_name: string
  expertise?: string
  im_bindings?: ParticipantIMBinding[]
}

export interface ParticipantIndex {
  id: string
  display_name?: string
  expertise?: string
  in_roster?: boolean
  im_bindings?: ParticipantIMBinding[]
  files: string[]
  updated_at: string
}

export interface ParticipantsResponse {
  participants: ParticipantIndex[]
}

export interface ParticipantDetail {
  id: string
  display_name?: string
  expertise?: string
  im_bindings?: ParticipantIMBinding[]
  files: Record<string, string>
}
