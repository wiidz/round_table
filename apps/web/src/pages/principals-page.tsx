import { useEffect, useState } from 'react'

import { fetchPrincipals } from '@/api/principals'
import { ApiError } from '@/api/client'
import {
  ProfileListCard,
  ProfileListSkeleton,
} from '@/components/profile/profile-list-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'

import type { PrincipalIndex } from '@/types/principal'

export function PrincipalsPage() {
  const [principals, setPrincipals] = useState<PrincipalIndex[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    fetchPrincipals()
      .then((data) => {
        if (!cancelled) {
          setPrincipals(data.principals ?? [])
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载 Principal 列表')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <ProfilePageHeader
        role="principal"
        eyebrow="Decision Owner"
        title="Principal"
        description={
          <>
            管理{' '}
            <code className="rounded-md bg-black/[0.04] px-1.5 py-0.5 font-mono text-[12px] ring-1 ring-inset ring-black/[0.05]">
              data/profiles/principals/
            </code>{' '}
            下的身份档案。标准文件为 <strong className="font-medium text-text-primary">USER.md</strong>
            （ADR-0010）；亦可包含 SOUL.md 等 Markdown。
          </>
        }
      />

      {loading && <ProfileListSkeleton />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && principals.length === 0 && (
        <ProfileStatePanel
          title="暂无 Principal 档案"
          description={
            <>
              Discord 绑定后会自动创建{' '}
              <code className="font-mono text-xs">discord:{'{user_id}'}</code>{' '}
              目录，或手动在{' '}
              <code className="font-mono text-xs">data/profiles/principals/</code>{' '}
              下新建。
            </>
          }
        />
      )}

      {!loading && !error && principals.length > 0 && (
        <ul className="space-y-4">
          {principals.map((p) => (
            <li key={p.id}>
              <ProfileListCard
                role="principal"
                href={`/principals/${encodeURIComponent(p.id)}`}
                title={p.display_name || p.id}
                subtitle={p.display_name ? p.id : undefined}
                files={p.files}
                meta={`${p.files.length} 文件`}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
