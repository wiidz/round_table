import { useEffect, useState } from 'react'

import { fetchPrincipals } from '@/api/principals'
import { ApiError } from '@/api/client'
import { PrincipalHub } from '@/components/profile/principal-hub'
import { ProfileListSkeleton } from '@/components/profile/profile-list-card'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useI18n } from '@/hooks/use-i18n'

export function PrincipalsPage() {
  const i18n = useI18n()
  const { t } = i18n
  const [principalId, setPrincipalId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    fetchPrincipals()
      .then((data) => {
        if (cancelled) return
        const list = data.principals ?? []
        setPrincipalId(list[0]?.id ?? null)
        setError(null)
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

  if (!loading && principalId) {
    return <PrincipalHub principalId={principalId} />
  }

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

        {!loading && !error && !principalId && (
          <ProfileStatePanel
            title={t('pages.principals.emptyTitle')}
            description={t('pages.principals.emptyDescription')}
          />
        )}
      </div>
    </PageLayout>
  )
}
