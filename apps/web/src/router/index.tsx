import { createBrowserRouter } from 'react-router-dom'

import { AppShell } from '@/components/layout/app-shell'
import { HomePage } from '@/pages/home-page'
import { MeetingDetailPage } from '@/pages/meeting-detail-page'
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
      { index: true, element: <HomePage /> },
      { path: 'meetings', element: <MeetingsPage /> },
      { path: 'meetings/:id', element: <MeetingDetailPage /> },
      { path: 'principals', element: <PrincipalsPage /> },
      { path: 'principals/:id', element: <PrincipalDetailPage /> },
      { path: 'participants', element: <ParticipantsPage /> },
      { path: 'participants/:id', element: <ParticipantDetailPage /> },
      { path: 'settings', element: <SettingsPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
