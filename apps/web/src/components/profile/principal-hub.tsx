import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Pencil } from 'lucide-react'
import { toast } from 'sonner'

import {
  createPrincipalPersona,
  fetchPrincipal,
  fetchPrincipalPersona,
  setActivePrincipalPersona,
} from '@/api/principals'
import { ApiError } from '@/api/client'
import { PrincipalPersonaCreateDialog } from '@/components/profile/principal-persona-create-dialog'
import { PrincipalPersonaWorkspace } from '@/components/profile/principal-persona-workspace'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { PrincipalUserPreview } from '@/components/profile/principal-user-preview'
import { PageLayout } from '@/components/layout/page-main-layout'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import { hePressable, heSectionDesc, heSectionTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import { EMPTY_PRINCIPAL_USER_PROFILE } from '@/types/principal'

import type { PrincipalDetail, PrincipalUserProfile } from '@/types/principal'

interface PrincipalHubProps {
  principalId: string
}

export function PrincipalHub({ principalId }: PrincipalHubProps) {
  const { t, domainPageEyebrow } = useI18n()
  const [detail, setDetail] = useState<PrincipalDetail | null>(null)
  const [viewingPersonaId, setViewingPersonaId] = useState('')
  const [profile, setProfile] = useState<PrincipalUserProfile>(EMPTY_PRINCIPAL_USER_PROFILE)
  const [loading, setLoading] = useState(true)
  const [profileLoading, setProfileLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [activating, setActivating] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const loadPersonaProfile = useCallback(
    async (personaId: string) => {
      const data = await fetchPrincipalPersona(principalId, personaId)
      setProfile(data.user_profile ?? EMPTY_PRINCIPAL_USER_PROFILE)
      setViewingPersonaId(personaId)
    },
    [principalId],
  )

  const load = useCallback(async () => {
    const data = await fetchPrincipal(principalId)
    setDetail(data)
    setError(null)
    await loadPersonaProfile(data.active_persona_id)
  }, [principalId, loadPersonaProfile])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    load()
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('profile.principal.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [load, t])

  async function handleViewPersona(personaId: string) {
    if (!detail || personaId === viewingPersonaId) return
    setProfileLoading(true)
    try {
      await loadPersonaProfile(personaId)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('profile.principal.loadFailed'))
    } finally {
      setProfileLoading(false)
    }
  }

  async function handleActivatePersona() {
    if (!detail || !viewingPersonaId || viewingPersonaId === detail.active_persona_id) return
    setActivating(true)
    try {
      const res = await setActivePrincipalPersona(principalId, viewingPersonaId)
      const title =
        res.personas.find((p) => p.id === viewingPersonaId)?.title ?? viewingPersonaId
      setDetail((prev) =>
        prev ? { ...prev, active_persona_id: res.active_persona_id, personas: res.personas } : prev,
      )
      toast.success(t('profile.principal.persona.activated', { title }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setActivating(false)
    }
  }

  async function handleCreatePersona(title: string) {
    setCreating(true)
    try {
      await createPrincipalPersona(principalId, title)
      setCreateDialogOpen(false)
      await load()
      toast.success(t('profile.principal.persona.created'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setCreating(false)
    }
  }

  const displayName = detail?.display_name || principalId
  const activePersonaId = detail?.active_persona_id ?? ''
  const viewingTitle =
    detail?.personas.find((p) => p.id === viewingPersonaId)?.title ??
    t('profile.principal.preferences')
  const editHref = `/principals/edit?persona=${encodeURIComponent(viewingPersonaId)}`
  const isViewingActive = Boolean(detail && viewingPersonaId === detail.active_persona_id)

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={domainPageEyebrow('principal')}
          title={displayName}
          description={
            <>
              {detail?.display_name && (
                <span className="mb-1 block font-mono text-xs text-text-tertiary">
                  {principalId}
                </span>
              )}
              {t('pages.principals.description')}
            </>
          }
        />
      }
    >
      <div className="space-y-6">
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

        {!loading && !error && detail && (
          <>
            <PrincipalPersonaWorkspace
              personas={detail.personas}
              selectedPersonaId={viewingPersonaId}
              activePersonaId={activePersonaId}
              disabled={profileLoading || activating}
              onSelect={(personaId) => void handleViewPersona(personaId)}
              onAddPersona={() => setCreateDialogOpen(true)}
              panelHeader={
                <header className="flex items-start justify-between gap-4 border-b border-black/[0.05] pb-5">
                  <div className="min-w-0 space-y-2">
                    <h2 className={heSectionTitle}>{viewingTitle}</h2>
                    <p className={heSectionDesc}>{t('profile.principal.persona.switchHint')}</p>
                  </div>
                  <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                    {!isViewingActive && (
                      <Button
                        type="button"
                        variant="outline"
                        disabled={activating || profileLoading}
                        onClick={() => void handleActivatePersona()}
                        className={cn(hePressable, 'rounded-xl px-4')}
                      >
                        {activating
                          ? t('profile.principal.persona.activating')
                          : t('profile.principal.persona.activate')}
                      </Button>
                    )}
                    <Button
                      asChild
                      className={cn(hePressable, 'gap-2 rounded-xl px-4')}
                    >
                      <Link to={editHref}>
                        <Pencil className="size-4" />
                        {t('profile.principal.editProfile')}
                      </Link>
                    </Button>
                  </div>
                </header>
              }
            >
              <PrincipalUserPreview profile={profile} embedded />
            </PrincipalPersonaWorkspace>

            <PrincipalPersonaCreateDialog
              open={createDialogOpen}
              creating={creating}
              onClose={() => {
                if (!creating) setCreateDialogOpen(false)
              }}
              onSubmit={handleCreatePersona}
            />
          </>
        )}
      </div>
    </PageLayout>
  )
}
