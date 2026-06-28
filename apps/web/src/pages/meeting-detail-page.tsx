import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchMeeting } from '@/api/meetings'
import { MeetingFilesViewer } from '@/components/meeting/meeting-files-viewer'

export function MeetingDetailPage() {
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const load = useCallback(async () => fetchMeeting(id), [id])

  if (!id) {
    return null
  }

  return (
    <MeetingFilesViewer
      backTo="/meetings"
      backLabel="返回会议列表"
      load={load}
    />
  )
}
