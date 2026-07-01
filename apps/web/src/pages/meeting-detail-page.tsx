import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchMeeting } from '@/api/meetings'
import { MeetingFilesViewer } from '@/components/meeting/meeting-files-viewer'
import { useI18n } from '@/hooks/use-i18n'

export function MeetingDetailPage() {
  const { t } = useI18n()
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const load = useCallback(async () => fetchMeeting(id), [id])

  if (!id) {
    return null
  }

  return (
    <MeetingFilesViewer
      backTo="/meetings"
      backLabel={t('pages.meetingDetail.back')}
      load={load}
    />
  )
}
