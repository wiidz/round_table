import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchPrincipal, savePrincipalFile } from '@/api/principals'
import { ProfileFilesEditor } from '@/components/profile-files-editor'
import { PRINCIPAL_FILE_HINTS } from '@/lib/profile-labels'
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
      pageTitle={`${domainNavLabel('principal')}档案 · ${id}`}
      pageDescription="编辑 USER.md 等档案，定义委托人（Principal）偏好与背景；保存后立即写入 data/profiles。"
      backTo="/principals"
      backLabel={`返回${domainNavLabel('principal')}列表`}
      fileHints={PRINCIPAL_FILE_HINTS}
      emptyHint={`在 data/profiles/principals/${id}/ 下添加 USER.md 或 SOUL.md。`}
      load={load}
      save={save}
    />
  )
}
