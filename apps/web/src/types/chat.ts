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
  turn?: number
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
  /** Global speech index; moderator, participant, and user (委托人). */
  turn?: number
  pending?: boolean
  error?: boolean
}

export type ChatConnectionState = 'connecting' | 'open' | 'closed' | 'error'

/** Per-seat typing indicator state. Key = seat id (speakerId). */
export interface TypingState {
  role: ChatRole
  authorId?: string
  authorName?: string
}

export type TypingStates = Map<string, TypingState>
