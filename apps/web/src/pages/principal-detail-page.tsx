import { useCallback } from 'react'
import { useParams } from 'react-router-dom'

import { fetchPrincipal, savePrincipalFile } from '@/api/principals'
import { ProfileFilesEditor } from '@/components/profile-files-editor'
import { PRINCIPAL_FILE_HINTS } from '@/lib/profile-labels'

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
      eyebrow="Decision Owner"
      pageTitle="Principal 档案"
      pageDescription="编辑 Principal 偏好与背景（USER.md 等），保存后立即写入 data/profiles。"
      backTo="/principals"
      backLabel="返回 Principal 列表"
      fileHints={PRINCIPAL_FILE_HINTS}
      emptyHint={`在 data/profiles/principals/${id}/ 下添加 USER.md 或 SOUL.md。`}
      load={load}
      save={save}
    />
  )
}
