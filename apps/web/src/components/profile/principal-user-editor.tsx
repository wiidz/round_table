import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'

import { fetchPrincipal, savePrincipalUserProfile } from '@/api/principals'
import { ApiError } from '@/api/client'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { PageLayout } from '@/components/layout/page-main-layout'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useI18n } from '@/hooks/use-i18n'
import {
  heColumnTitleBrand,
  hePanelShell,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import {
  EMPTY_PRINCIPAL_USER_PROFILE,
  type PrincipalUserProfile,
} from '@/types/principal'

interface PrincipalUserEditorProps {
  id: string
}

function profilesEqual(a: PrincipalUserProfile, b: PrincipalUserProfile): boolean {
  return (
    a.language === b.language &&
    (a.confirmation ?? '') === (b.confirmation ?? '') &&
    (a.context ?? '') === (b.context ?? '')
  )
}

export function PrincipalUserEditor({ id }: PrincipalUserEditorProps) {
  const { t, domainNavLabel, domainPageEyebrow } = useI18n()
  const [displayName, setDisplayName] = useState('')
  const [savedProfile, setSavedProfile] = useState<PrincipalUserProfile>(
    EMPTY_PRINCIPAL_USER_PROFILE,
  )
  const [form, setForm] = useState<PrincipalUserProfile>(EMPTY_PRINCIPAL_USER_PROFILE)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    const data = await fetchPrincipal(id)
    const profile = data.user_profile ?? EMPTY_PRINCIPAL_USER_PROFILE
    setDisplayName(data.display_name ?? '')
    setSavedProfile(profile)
    setForm(profile)
    setError(null)
  }, [id])

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

  const dirty = !profilesEqual(form, savedProfile)

  async function handleSave() {
    setSaving(true)
    try {
      const res = await savePrincipalUserProfile(id, form)
      setSavedProfile(res.user_profile)
      setForm(res.user_profile)
      toast.success(t('profile.principal.saveSuccess'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <PageLayout
      header={
        <div className="space-y-8">
          <Link
            to="/principals"
            className={cn(
              'inline-flex items-center gap-1.5 text-sm text-text-secondary',
              'hover:text-brand',
              heSpring,
            )}
          >
            <ArrowLeft className="size-4" />
            {t('profile.principal.backToList', { principal: domainNavLabel('principal') })}
          </Link>

          <ProfilePageHeader
            role="principal"
            eyebrow={domainPageEyebrow('principal')}
            title={displayName || id}
            description={
              <>
                {displayName && (
                  <span className="mb-1 block font-mono text-xs text-text-tertiary">{id}</span>
                )}
                {t('profile.principal.description')}
              </>
            }
          />
        </div>
      }
    >
    <div className="space-y-8">
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

      {!loading && !error && (
        <div className={cn(hePanelShell, 'flex flex-col gap-6 p-6 sm:p-8')}>
          <div className="space-y-6">
            <p className={heColumnTitleBrand}>{t('profile.principal.preferences')}</p>
            <SettingsFieldRow
              label={t('profile.principal.languageLabel')}
              htmlFor="user-language"
              hint={t('profile.principal.languageHint')}
            >
              <Input
                id="user-language"
                value={form.language}
                placeholder="zh-CN"
                onChange={(e) => setForm((prev) => ({ ...prev, language: e.target.value }))}
              />
            </SettingsFieldRow>
            <SettingsFieldRow
              label={t('profile.principal.confirmationLabel')}
              htmlFor="user-confirmation"
              hint={t('profile.principal.confirmationHint')}
            >
              <Input
                id="user-confirmation"
                value={form.confirmation ?? ''}
                placeholder="review numbered lists carefully"
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, confirmation: e.target.value }))
                }
              />
            </SettingsFieldRow>
            <SettingsFieldRow
              label={t('profile.principal.contextLabel')}
              htmlFor="user-context"
              hint={t('profile.principal.contextHint')}
            >
              <Textarea
                id="user-context"
                value={form.context ?? ''}
                rows={5}
                className="min-h-[8rem] font-sans text-sm"
                placeholder={t('profile.principal.contextPlaceholder')}
                onChange={(e) => setForm((prev) => ({ ...prev, context: e.target.value }))}
              />
            </SettingsFieldRow>
          </div>

          <div className="flex flex-wrap items-center gap-3 border-t border-border-subtle/80 pt-4">
            <Button
              onClick={handleSave}
              disabled={!dirty || saving}
              className={cn(hePressable, 'gap-2 rounded-full px-5')}
            >
              <Save className="size-4" />
              {saving ? t('common.saving') : t('profile.principal.save')}
            </Button>
            {dirty && (
              <span className="text-xs font-medium text-warning">
                {t('profile.filesEditor.unsaved')}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
    </PageLayout>
  )
}
