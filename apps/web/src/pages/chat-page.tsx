import { ChatWindow } from '@/components/chat/chat-window'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useChatSocket } from '@/hooks/use-chat-socket'

export function ChatPage() {
  const { connectionState, sessionId, messages, lastError, typingStates, sendMessage, reconnect } =
    useChatSocket()

  return (
    <ChatWindow
      className="h-full"
      connectionState={connectionState}
      messages={messages}
      sessionId={sessionId}
      lastError={lastError}
      typingStates={typingStates}
      onSend={sendMessage}
      onReconnect={reconnect}
      pageShell={({ main, left, right, drawer }) => (
        <PageLayout
          header={
            <ProfilePageHeader
              role="participant"
              eyebrow="Chat"
              title="聊天"
              description="浏览器 Transport：Setup 默认列表，会议进行中自动切圆桌；768px 以下窄屏会议期降级为发言记录列表。"
            />
          }
          left={left}
          right={right}
          sidebarFrom="96rem"
          sideColumnWidth="gutter"
          className="h-[calc(100vh-7.5rem)] min-h-[28rem]"
          bodyClassName="min-h-0 flex-1 min-[96rem]:h-full"
        >
          {connectionState === 'error' && (
            <ProfileStatePanel
              variant="danger"
              title="连接失败"
              description={lastError ?? '无法连接聊天服务，请确认 API 已启动。'}
            />
          )}
          <div className="relative h-full min-h-0">{main}</div>
          {drawer}
        </PageLayout>
      )}
    />
  )
}
