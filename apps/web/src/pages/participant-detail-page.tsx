import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchParticipant, saveParticipantFile } from '@/api/participants'
import { ProfileFilesEditor } from '@/components/profile-files-editor'
import { useI18n } from '@/hooks/use-i18n'
import { PARTICIPANT_FILE_HINTS, PARTICIPANT_STANDARD_FILES } from '@/lib/profile-labels'

export function ParticipantDetailPage() {
  const i18n = useI18n()
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
      eyebrow={i18n.domainPageEyebrow('participant')}
      pageTitle={id}
      pageDescription={i18n.t('profile.filesEditor.participantDescription')}
      backTo="/participants"
      backLabel={i18n.t('profile.filesEditor.participantBack', {
        participant: i18n.domainNavLabel('participant'),
      })}
      standardFiles={PARTICIPANT_STANDARD_FILES}
      fileHints={PARTICIPANT_FILE_HINTS}
      emptyHint={i18n.t('profile.filesEditor.participantEmptyHint')}
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
