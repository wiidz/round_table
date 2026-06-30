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
import { domainPageEyebrow, domainPageTitle } from '@/lib/ui-labels'
import { PRINCIPAL_STANDARD_FILES } from '@/lib/profile-labels'

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
          setError('无法加载委托人列表')
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
        eyebrow={domainPageEyebrow('principal')}
        title={domainPageTitle('principal')}
        description={
          <>
            管理委托人（Principal）身份与{' '}
            <strong className="font-medium text-text-primary">USER.md</strong>{' '}
            偏好档案（ADR-0010）。每人仅一份 USER.md，描述语言、验收习惯与背景；单次会议意图请用{' '}
            <strong className="font-medium text-text-primary">简报模板</strong>。
          </>
        }
      />

      {loading && <ProfileListSkeleton />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && principals.length === 0 && (
        <ProfileStatePanel
          title="暂无委托人档案"
          description={
            <>
              Discord 绑定后会自动创建{' '}
              <code className="font-mono text-xs">discord:{'{user_id}'}</code>{' '}
              目录；进入详情页可编辑 USER.md。
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
                files={PRINCIPAL_STANDARD_FILES.map((name) => ({
                  name,
                  present: p.files.includes(name),
                }))}
                meta={p.files.includes('USER.md') ? 'USER.md 已配置' : '待编辑 USER.md'}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
