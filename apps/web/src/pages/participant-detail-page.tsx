import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchParticipant, saveParticipantFile } from '@/api/participants'
import { ProfileFilesEditor } from '@/components/profile-files-editor'
import { PARTICIPANT_FILE_HINTS, PARTICIPANT_STANDARD_FILES } from '@/lib/profile-labels'
import { domainPageEyebrow, domainNavLabel } from '@/lib/ui-labels'

export function ParticipantDetailPage() {
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const load = useCallback(() => fetchParticipant(id), [id])
  const save = useCallback(
    async (filename: string, content: string) => {
      await saveParticipantFile(id, filename, content)
    },
    [id],
  )

  if (!id) {
    return null
  }

  return (
    <ProfileFilesEditor
      role="participant"
      eyebrow={domainPageEyebrow('participant')}
      pageTitle={id}
      pageDescription="编辑 SOUL.md / AGENTS.md / TOOLS.md 档案，定义专家（Participant）人格与会议内行为。"
      backTo="/participants"
      backLabel={`返回${domainNavLabel('participant')}列表`}
      standardFiles={PARTICIPANT_STANDARD_FILES}
      fileHints={PARTICIPANT_FILE_HINTS}
      emptyHint="标准档案为 SOUL.md、AGENTS.md、TOOLS.md；打开页面时会从模板自动创建缺失文件。"
      load={load}
      save={save}
      resolveTitle={(data) => data.display_name?.trim() || data.id}
      resolveSubtitle={(data) =>
        [data.id, data.expertise].filter(Boolean).join(' · ')
      }
      resolveAvatar={(data) => ({
        id: data.id,
        name: data.display_name?.trim() || data.id,
      })}
    />
  )
}
