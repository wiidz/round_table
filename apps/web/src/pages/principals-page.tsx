import { useEffect, useState } from 'react'

import { fetchPrincipals } from '@/api/principals'
import { ApiError } from '@/api/client'
import {
  ProfileListCard,
  ProfileListSkeleton,
} from '@/components/profile/profile-list-card'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useI18n } from '@/hooks/use-i18n'
import { PRINCIPAL_STANDARD_FILES } from '@/lib/profile-labels'

import type { PrincipalIndex } from '@/types/principal'

export function PrincipalsPage() {
  const i18n = useI18n()
  const { t } = i18n
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
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('pages.principals.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={i18n.domainPageEyebrow('principal')}
          title={i18n.domainPageTitle('principal')}
          description={t('pages.principals.description')}
        />
      }
    >
    <div className="space-y-8">
      {loading && <ProfileListSkeleton />}

      {!loading && error && (
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.loadFailed')}
          description={error}
        />
      )}

      {!loading && !error && principals.length === 0 && (
        <ProfileStatePanel
          title={t('pages.principals.emptyTitle')}
          description={t('pages.principals.emptyDescription')}
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
                meta={
                  p.files.includes('USER.md')
                    ? t('profile.list.userConfigured')
                    : t('profile.list.userPending')
                }
              />
            </li>
          ))}
        </ul>
      )}
    </div>
    </PageLayout>
  )
}
