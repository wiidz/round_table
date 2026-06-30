import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchPrincipal, savePrincipalFile } from '@/api/principals'
import { ProfileFilesEditor } from '@/components/profile-files-editor'
import { PRINCIPAL_FILE_HINTS, PRINCIPAL_STANDARD_FILES } from '@/lib/profile-labels'
import { domainNavLabel, domainPageEyebrow } from '@/lib/ui-labels'

export function PrincipalDetailPage() {
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const load = useCallback(async () => {
    const data = await fetchPrincipal(id)
    return {
      id: data.id,
      files: data.files,
      display_name: data.display_name,
    }
  }, [id])

  const save = useCallback(
    async (filename: string, content: string) => {
      await savePrincipalFile(id, filename, content)
    },
    [id],
  )

  if (!id) {
    return null
  }

  return (
    <ProfileFilesEditor
      role="principal"
      eyebrow={domainPageEyebrow('principal')}
      pageTitle={`${domainNavLabel('principal')} · ${id}`}
      pageDescription="编辑 USER.md：定义你的语言偏好、Confirmation 审阅习惯与背景约束。会议议题请使用「简报模板」。"
      backTo="/principals"
      backLabel={`返回${domainNavLabel('principal')}列表`}
      standardFiles={PRINCIPAL_STANDARD_FILES}
      fileHints={PRINCIPAL_FILE_HINTS}
      emptyHint="保存 USER.md 后将写入 data/profiles/principals/ 目录。"
      load={load}
      save={save}
    />
  )
}
