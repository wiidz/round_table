import { useCallback, useEffect, useRef, useState } from 'react'

import { assignTurnForRole } from '@/lib/assign-turn'
import type { ChatConnectionState, ChatFrame, ChatMessage, ChatRole } from '@/types/chat'

function chatWsUrl(): string {
  const wsDev = import.meta.env.VITE_CHAT_WS_DEV as string | undefined
  if (import.meta.env.DEV && wsDev?.trim()) {
    return wsDev.trim()
  }
  const apiBase = import.meta.env.VITE_API_BASE ?? '/api'
  if (apiBase.startsWith('http://') || apiBase.startsWith('https://')) {
    const url = new URL(apiBase)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.pathname = `${url.pathname.replace(/\/$/, '')}/chat/ws`
    url.search = ''
    url.hash = ''
    return url.toString()
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}${apiBase.replace(/\/$/, '')}/chat/ws`
}

function nextMessageId(): string {
  return crypto.randomUUID()
}

function parseFrameRole(role: string | undefined): ChatRole {
  switch (role) {
    case 'user':
      return 'user'
    case 'system':
      return 'system'
    case 'participant':
      return 'participant'
    default:
      return 'moderator'
  }
}

function parseFrameTime(at: string | undefined): number {
  if (at?.trim()) {
    const parsed = Date.parse(at)
    if (!Number.isNaN(parsed)) return parsed
  }
  return Date.now()
}

export function useChatSocket() {
  const [connectionState, setConnectionState] = useState<ChatConnectionState>('connecting')
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [lastError, setLastError] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<number | null>(null)
  const connectGenRef = useRef(0)
  const disposedRef = useRef(false)
  const nextTurnRef = useRef(1)

  const pushMessage = useCallback(
    (
      msg: Omit<ChatMessage, 'turn'> & { turn?: number },
      options?: { skipTurnAssign?: boolean },
    ) => {
      let withTurn: ChatMessage
      if (msg.turn != null) {
        withTurn = msg as ChatMessage
        nextTurnRef.current = Math.max(nextTurnRef.current, msg.turn + 1)
      } else if (options?.skipTurnAssign) {
        withTurn = { ...msg }
      } else {
        const assigned = assignTurnForRole(msg.role, nextTurnRef.current)
        nextTurnRef.current = assigned.nextTurn
        withTurn = assigned.turn != null ? { ...msg, turn: assigned.turn } : { ...msg }
      }

      setMessages((prev) => {
        const index = prev.findIndex((item) => item.id === withTurn.id)
        if (index >= 0) {
          const next = [...prev]
          next[index] = { ...next[index], ...withTurn }
          return next
        }
        return [...prev, withTurn]
      })
    },
    [],
  )

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return
    if (wsRef.current?.readyState === WebSocket.CONNECTING) return

    if (reconnectTimer.current != null) {
      window.clearTimeout(reconnectTimer.current)
      reconnectTimer.current = null
    }

    disposedRef.current = false
    nextTurnRef.current = 1
    const gen = ++connectGenRef.current

    setConnectionState('connecting')
    setLastError(null)

    const ws = new WebSocket(chatWsUrl())
    wsRef.current = ws

    ws.onopen = () => {
      if (gen !== connectGenRef.current) return
      setConnectionState('open')
    }

    ws.onmessage = (event) => {
      if (gen !== connectGenRef.current) return

      let frame: ChatFrame
      try {
        frame = JSON.parse(String(event.data)) as ChatFrame
      } catch {
        setLastError('无法解析服务端消息')
        return
      }

      if (frame.type === 'connected') {
        if (frame.session_id) {
          setSessionId(frame.session_id)
        }
        return
      }

      if (frame.type === 'error') {
        const errText = frame.error?.trim() || '请求失败'
        setLastError(errText)
        pushMessage({
          id: nextMessageId(),
          role: 'system',
          content: errText,
          createdAt: Date.now(),
          error: true,
        })
        return
      }

      if (frame.type === 'message' && frame.content?.trim()) {
        const serverTurn =
          typeof frame.turn === 'number' && frame.turn > 0 ? frame.turn : undefined
        pushMessage({
          id: frame.id ?? nextMessageId(),
          role: parseFrameRole(frame.role),
          content: frame.content,
          authorId: frame.author_id,
          authorName: frame.author_name,
          createdAt: parseFrameTime(frame.at),
          turn: serverTurn,
        })
      }
    }

    ws.onerror = () => {
      if (gen !== connectGenRef.current) return
      setConnectionState('error')
      setLastError('WebSocket 连接异常')
    }

    ws.onclose = () => {
      if (gen !== connectGenRef.current) return
      wsRef.current = null
      setSessionId(null)
      setConnectionState('closed')
      if (disposedRef.current || reconnectTimer.current != null) return
      reconnectTimer.current = window.setTimeout(() => {
        reconnectTimer.current = null
        connect()
      }, 3000)
    }
  }, [pushMessage])

  useEffect(() => {
    connect()
    return () => {
      disposedRef.current = true
      connectGenRef.current += 1
      if (reconnectTimer.current != null) {
        window.clearTimeout(reconnectTimer.current)
        reconnectTimer.current = null
      }
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [connect])

  const sendMessage = useCallback(
    (content: string) => {
      const text = content.trim()
      if (!text) return false
      const ws = wsRef.current
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        setLastError('尚未连接，请稍候重试')
        return false
      }

      const id = nextMessageId()
      pushMessage({ id, role: 'user', content: text, createdAt: Date.now() }, { skipTurnAssign: true })
      ws.send(JSON.stringify({ type: 'message', id, content: text }))
      return true
    },
    [pushMessage],
  )

  return {
    connectionState,
    sessionId,
    messages,
    lastError,
    sendMessage,
    reconnect: connect,
  }
}
