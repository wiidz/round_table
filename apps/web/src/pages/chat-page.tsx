import { ChatWindow } from '@/components/chat/chat-window'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useChatSocket } from '@/hooks/use-chat-socket'

export function ChatPage() {
  const { connectionState, sessionId, messages, lastError, sendMessage, reconnect } =
    useChatSocket()

  return (
    <div className="flex h-[calc(100vh-7.5rem)] min-h-[28rem] flex-col gap-4">
      <div className="shrink-0">
        <ProfilePageHeader
          role="participant"
          eyebrow="Chat"
          title="聊天"
          description="浏览器 Transport：Setup 默认列表，会议进行中自动切圆桌；768px 以下窄屏会议期降级为发言记录列表。"
        />
      </div>

      {connectionState === 'error' && (
        <ProfileStatePanel
          variant="danger"
          title="连接失败"
          description={lastError ?? '无法连接聊天服务，请确认 API 已启动。'}
        />
      )}

      <div className="min-h-0 flex-1">
        <ChatWindow
          className="h-full"
          connectionState={connectionState}
          messages={messages}
          sessionId={sessionId}
          lastError={lastError}
          onSend={sendMessage}
          onReconnect={reconnect}
        />
      </div>
    </div>
  )
}
