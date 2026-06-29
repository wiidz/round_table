export type ChatFrameType = 'connected' | 'message' | 'error' | 'typing'

export type ChatRole = 'user' | 'moderator' | 'participant' | 'system'

export interface ChatFrame {
  type: ChatFrameType
  id?: string
  session_id?: string
  role?: ChatRole
  author_id?: string
  author_name?: string
  at?: string
  content?: string
  error?: string
}

export interface ChatMessage {
  id: string
  role: ChatRole
  content: string
  authorId?: string
  authorName?: string
  createdAt: number
  /** Global meeting speech index; moderator/participant only. */
  turn?: number
  pending?: boolean
  error?: boolean
}

export type ChatConnectionState = 'connecting' | 'open' | 'closed' | 'error'
