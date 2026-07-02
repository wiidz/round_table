import { ChatWindow } from '@/components/chat/chat-window'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useChatSocket } from '@/hooks/use-chat-socket'
import { useI18n } from '@/hooks/use-i18n'

export function ChatPage() {
  const { t } = useI18n()
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
              title={t('chat.page.title')}
              description={t('chat.page.description')}
            />
          }
          left={left}
          right={right}
          sidebarFrom="96rem"
          sideColumnWidth="gutter"
          fillHeight
          className="h-[calc(100vh-7.5rem)] min-h-[28rem]"
          bodyClassName="min-h-0 flex-1"
        >
          {connectionState === 'error' && (
            <ProfileStatePanel
              variant="danger"
              title={t('chat.page.connectionFailed')}
              description={lastError ?? t('chat.page.cannotConnect')}
            />
          )}
          <div className="relative flex min-h-0 flex-1 flex-col">{main}</div>
          {drawer}
        </PageLayout>
      )}
    />
  )
}
