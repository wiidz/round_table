import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'

import { fetchPrincipal, fetchPrincipalPersona, savePrincipalUserProfile, setActivePrincipalPersona } from '@/api/principals'
import { ApiError } from '@/api/client'
import { PrincipalPersonaWorkspace } from '@/components/profile/principal-persona-workspace'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { PageLayout } from '@/components/layout/page-main-layout'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useLocale } from '@/contexts/locale-context'
import { useI18n } from '@/hooks/use-i18n'
import {
  applyPrincipalFieldPreset,
  getPrincipalUserPresets,
  principalPresetApplied,
  type PrincipalUserPreset,
} from '@/lib/i18n/principal-user-presets'
import {
  heColumnTitleBrand,
  heFileBadge,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { localeToUserLanguage } from '@/lib/locale'
import { cn } from '@/lib/utils'
import {
  EMPTY_PRINCIPAL_USER_PROFILE,
  type PrincipalPersonaMeta,
  type PrincipalUserProfile,
} from '@/types/principal'

interface PrincipalUserEditorProps {
  id: string
  initialPersonaId?: string
}

function profileContentEqual(a: PrincipalUserProfile, b: PrincipalUserProfile): boolean {
  return (
    (a.confirmation ?? '') === (b.confirmation ?? '') &&
    (a.context ?? '') === (b.context ?? '')
  )
}

function PrincipalPresetButtons({
  presets,
  value,
  onApply,
}: {
  presets: readonly PrincipalUserPreset[]
  value: string
  onApply: (snippet: string) => void
}) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {presets.map((preset) => {
        const applied = principalPresetApplied(value, preset.value)
        return (
          <button
            key={preset.label}
            type="button"
            aria-pressed={applied}
            className={cn(
              'rounded-full px-2.5 py-1 text-[11px] font-medium ring-1 ring-inset',
              hePressable,
              heSpring,
              applied
                ? 'bg-surface/90 text-text-secondary ring-black/[0.08]'
                : 'bg-black/[0.02] text-text-secondary ring-black/[0.05] hover:bg-brand-soft/50 hover:text-brand hover:ring-primary/25',
            )}
            onClick={() => onApply(preset.value)}
          >
            {preset.label}
          </button>
        )
      })}
    </div>
  )
}

export function PrincipalUserEditor({ id, initialPersonaId }: PrincipalUserEditorProps) {
  const { t, domainPageEyebrow } = useI18n()
  const { locale } = useLocale()
  const presets = useMemo(() => getPrincipalUserPresets(locale), [locale])
  const languageCode = localeToUserLanguage(locale)
  const languageDisplay =
    locale === 'en'
      ? `${t('pages.settings.localeEn')} (${languageCode})`
      : `${t('pages.settings.localeZh')} (${languageCode})`

  const [displayName, setDisplayName] = useState('')
  const [personas, setPersonas] = useState<PrincipalPersonaMeta[]>([])
  const [editingPersonaId, setEditingPersonaId] = useState('')
  const [activePersonaId, setActivePersonaId] = useState('')
  const [savedProfile, setSavedProfile] = useState<PrincipalUserProfile>(
    EMPTY_PRINCIPAL_USER_PROFILE,
  )
  const [form, setForm] = useState<PrincipalUserProfile>(EMPTY_PRINCIPAL_USER_PROFILE)
  const [loading, setLoading] = useState(true)
  const [personaLoading, setPersonaLoading] = useState(false)
  const [activating, setActivating] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadPersonaProfile = useCallback(async (personaId: string) => {
    const data = await fetchPrincipalPersona(id, personaId)
    const profile = data.user_profile ?? EMPTY_PRINCIPAL_USER_PROFILE
    setSavedProfile(profile)
    setForm(profile)
  }, [id])

  const load = useCallback(async () => {
    const data = await fetchPrincipal(id)
    setDisplayName(data.display_name ?? '')
    setPersonas(data.personas ?? [])
    setActivePersonaId(data.active_persona_id)
    const personaId =
      initialPersonaId && data.personas.some((p) => p.id === initialPersonaId)
        ? initialPersonaId
        : data.active_persona_id
    setEditingPersonaId(personaId)
    await loadPersonaProfile(personaId)
    setError(null)
  }, [id, initialPersonaId, loadPersonaProfile])

  useEffect(() => {
    setForm((prev) => ({ ...prev, language: languageCode }))
  }, [languageCode])

  const dirty =
    !profileContentEqual(form, savedProfile) || languageCode !== (savedProfile.language ?? '')

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

  async function handleSave() {
    if (!editingPersonaId) return
    setSaving(true)
    try {
      const payload: PrincipalUserProfile = {
        ...form,
        language: languageCode,
      }
      const res = await savePrincipalUserProfile(id, payload, editingPersonaId)
      const next = { ...res.user_profile, language: languageCode }
      setSavedProfile(next)
      setForm(next)
      toast.success(t('profile.principal.saveSuccess'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  async function handleSelectEditingPersona(personaId: string) {
    if (personaId === editingPersonaId) return
    if (dirty) {
      const ok = window.confirm(t('profile.filesEditor.switchConfirm'))
      if (!ok) return
    }
    setPersonaLoading(true)
    try {
      setEditingPersonaId(personaId)
      await loadPersonaProfile(personaId)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('profile.principal.loadFailed'))
    } finally {
      setPersonaLoading(false)
    }
  }

  async function handleActivatePersona() {
    if (!editingPersonaId || editingPersonaId === activePersonaId) return
    setActivating(true)
    try {
      const res = await setActivePrincipalPersona(id, editingPersonaId)
      setActivePersonaId(res.active_persona_id)
      setPersonas(res.personas)
      const title =
        res.personas.find((p) => p.id === editingPersonaId)?.title ?? editingPersonaId
      toast.success(t('profile.principal.persona.activated', { title }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setActivating(false)
    }
  }

  const editingPersonaTitle =
    personas.find((p) => p.id === editingPersonaId)?.title ?? editingPersonaId

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
            {t('profile.principal.backToPreview')}
          </Link>

          <ProfilePageHeader
            role="principal"
            eyebrow={domainPageEyebrow('principal')}
            title={editingPersonaTitle || displayName || id}
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
          <PrincipalPersonaWorkspace
            personas={personas}
            selectedPersonaId={editingPersonaId}
            activePersonaId={activePersonaId}
            disabled={personaLoading || saving || activating}
            onSelect={(personaId) => void handleSelectEditingPersona(personaId)}
            panelActions={
              editingPersonaId !== activePersonaId ? (
                <Button
                  type="button"
                  variant="outline"
                  disabled={activating || personaLoading || saving}
                  onClick={() => void handleActivatePersona()}
                  className={cn(hePressable, 'rounded-xl px-4')}
                >
                  {activating
                    ? t('profile.principal.persona.activating')
                    : t('profile.principal.persona.activate')}
                </Button>
              ) : null
            }
          >
            <div className="space-y-8">
              <p className={heColumnTitleBrand}>{t('profile.principal.preferences')}</p>

              <SettingsFieldRow
                label={t('profile.principal.languageLabel')}
                htmlFor="user-language"
                hint={t('profile.principal.languageHint')}
                labelExtra={<span className={heFileBadge}>{t('common.readonly')}</span>}
              >
                <Input
                  id="user-language"
                  value={languageDisplay}
                  readOnly
                  tabIndex={-1}
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label={t('profile.principal.confirmationLabel')}
                htmlFor="user-confirmation"
                hint={t('profile.principal.confirmationHint')}
              >
                <div className="space-y-2.5">
                  <Input
                    id="user-confirmation"
                    value={form.confirmation ?? ''}
                    placeholder={t('profile.principal.confirmationPlaceholder')}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, confirmation: e.target.value }))
                    }
                  />
                  <PrincipalPresetButtons
                    presets={presets.confirmation}
                    value={form.confirmation ?? ''}
                    onApply={(snippet) =>
                      setForm((prev) => ({
                        ...prev,
                        confirmation: applyPrincipalFieldPreset(prev.confirmation ?? '', snippet),
                      }))
                    }
                  />
                </div>
              </SettingsFieldRow>

              <SettingsFieldRow
                label={t('profile.principal.contextLabel')}
                htmlFor="user-context"
                hint={t('profile.principal.contextHint')}
              >
                <div className="space-y-2.5">
                  <Textarea
                    id="user-context"
                    value={form.context ?? ''}
                    rows={5}
                    className="min-h-[8rem] font-sans text-sm"
                    placeholder={t('profile.principal.contextPlaceholder')}
                    onChange={(e) => setForm((prev) => ({ ...prev, context: e.target.value }))}
                  />
                  <PrincipalPresetButtons
                    presets={presets.context}
                    value={form.context ?? ''}
                    onApply={(snippet) =>
                      setForm((prev) => ({
                        ...prev,
                        context: applyPrincipalFieldPreset(prev.context ?? '', snippet),
                      }))
                    }
                  />
                </div>
              </SettingsFieldRow>

              <div className="flex flex-wrap items-center gap-3 border-t border-border-subtle/80 pt-4">
              <Button
                type="button"
                variant="outline"
                asChild
                className={cn(hePressable, 'rounded-full px-5')}
              >
                <Link
                  to="/principals"
                  onClick={(e) => {
                    if (dirty && !window.confirm(t('profile.filesEditor.switchConfirm'))) {
                      e.preventDefault()
                    }
                  }}
                >
                  {t('common.cancel')}
                </Link>
              </Button>
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
          </PrincipalPersonaWorkspace>
        )}
      </div>
    </PageLayout>
  )
}
