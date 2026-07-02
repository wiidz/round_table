import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'

import { fetchPrincipals } from '@/api/principals'
import { ApiError } from '@/api/client'
import { PrincipalUserEditor } from '@/components/profile/principal-user-editor'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useI18n } from '@/hooks/use-i18n'

export function PrincipalEditPage() {
  const { t, domainPageEyebrow, domainPageTitle } = useI18n()
  const [searchParams] = useSearchParams()
  const [principalId, setPrincipalId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const initialPersonaId = searchParams.get('persona') ?? undefined

  useEffect(() => {
    let cancelled = false
    fetchPrincipals()
      .then((data) => {
        if (cancelled) return
        setPrincipalId(data.principals?.[0]?.id ?? null)
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
    return (
      <PrincipalUserEditor
        id={principalId}
        initialPersonaId={initialPersonaId}
      />
    )
  }

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={domainPageEyebrow('principal')}
          title={domainPageTitle('principal')}
          description={t('pages.principals.description')}
        />
      }
    >
      {loading && (
        <ProfileStatePanel
          title={t('common.loading')}
          description={t('profile.state.loadingPrincipal')}
        />
      )}
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
    </PageLayout>
  )
}
