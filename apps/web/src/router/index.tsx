import { createBrowserRouter } from 'react-router-dom'

import { AppShell } from '@/components/layout/app-shell'
import { PageMainLayout } from '@/components/layout/page-main-layout'
import { BriefTemplateDetailPage } from '@/pages/brief-template-detail-page'
import { BriefTemplatesPage } from '@/pages/brief-templates-page'
import { ChatPage } from '@/pages/chat-page'
import { HomePage } from '@/pages/home-page'
import { MeetingDetailPage } from '@/pages/meeting-detail-page'
import { MeetingReplayPage } from '@/pages/meeting-replay-page'
import { MeetingsPage } from '@/pages/meetings-page'
import { NotFoundPage } from '@/pages/not-found-page'
import { ParticipantDetailPage } from '@/pages/participant-detail-page'
import { ParticipantsPage } from '@/pages/participants-page'
import { PrincipalDetailPage } from '@/pages/principal-detail-page'
import { PrincipalsPage } from '@/pages/principals-page'
import { SettingsPage } from '@/pages/settings-page'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      {
        element: <PageMainLayout />,
        children: [
          { index: true, element: <HomePage /> },
          { path: 'chat', element: <ChatPage /> },
          { path: 'meetings', element: <MeetingsPage /> },
          { path: 'principals', element: <PrincipalsPage /> },
          { path: 'principals/:id', element: <PrincipalDetailPage /> },
          { path: 'brief-templates', element: <BriefTemplatesPage /> },
          { path: 'brief-templates/:id', element: <BriefTemplateDetailPage /> },
          { path: 'participants', element: <ParticipantsPage /> },
          { path: 'participants/:id', element: <ParticipantDetailPage /> },
          { path: 'settings', element: <SettingsPage /> },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
      { path: 'meetings/:id', element: <MeetingDetailPage /> },
      { path: 'meetings/:id/replay', element: <MeetingReplayPage /> },
    ],
  },
])
