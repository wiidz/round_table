import { PrincipalUserEditor } from '@/components/profile/principal-user-editor'
import { useParams } from 'react-router-dom'

export function PrincipalDetailPage() {
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  if (!id) {
    return null
  }

  return <PrincipalUserEditor id={id} />
}
