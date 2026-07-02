import { useEffect, useMemo, useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'

import {
  ProfilePageHeader,
  ProfileStatePanel,
  type ProfileRole,
} from '@/components/profile/profile-page-header'
import { PageLayout } from '@/components/layout/page-main-layout'
import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { MarkdownReader } from '@/components/markdown/markdown-reader'
import {
  MarkdownViewToggle,
  type MarkdownViewMode,
} from '@/components/markdown/markdown-view-toggle'
import { ApiError } from '@/api/client'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { useI18n } from '@/hooks/use-i18n'
import {
  heColumnTitleAI,
  heColumnTitleBrand,
  heInputReadonly,
  heFilePill,
  heFilePillSelected,
  hePageDesc,
  hePageTitle,
  hePanelShell,
  hePressable,
  heSpring,
  heEyebrowAI,
  heEyebrowBrand,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import { profileFileHasTitle } from '@/lib/i18n/profile-labels'

interface ProfileLoadData {
  id: string
  files: Record<string, string>
  display_name?: string
  expertise?: string
  avatar_url?: string
}

interface ProfileFilesEditorProps {
  role: ProfileRole
  eyebrow: string
  pageTitle: string
  pageDescription: string
  title?: string
  subtitle?: string
  backTo: string
  backLabel: string
  fileHints: Record<string, string>
  standardFiles?: readonly string[]
  emptyHint: string
  load: () => Promise<ProfileLoadData>
  save: (filename: string, content: string) => Promise<void>
  resolveTitle?: (data: ProfileLoadData) => string
  resolveSubtitle?: (data: ProfileLoadData) => string | undefined
  resolveAvatar?: (data: ProfileLoadData) => {
    id: string
    name: string
    avatarUrl?: string
  }
}

export function ProfileFilesEditor({
  role,
  eyebrow,
  pageTitle,
  pageDescription,
  title,
  backTo,
  backLabel,
  fileHints,
  standardFiles,
  emptyHint,
  load,
  save,
  resolveTitle,
  resolveSubtitle,
  resolveAvatar,
}: ProfileFilesEditorProps) {
  const { t, profileFileCaption } = useI18n()
  const [entityId, setEntityId] = useState('')
  const [heading, setHeading] = useState(title ?? '')
  const [subheading, setSubheading] = useState<string>()
  const [avatar, setAvatar] = useState<{
    id: string
    name: string
    avatarUrl?: string
  }>()
  const [files, setFiles] = useState<Record<string, string>>({})
  const [activeFile, setActiveFile] = useState('')
  const [draft, setDraft] = useState('')
  const [viewMode, setViewMode] = useState<MarkdownViewMode>('source')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fileNames = useMemo(() => {
    const names = new Set(Object.keys(files))
    if (standardFiles) {
      for (const name of standardFiles) {
        names.add(name)
      }
    }
    return [...names].sort()
  }, [files, standardFiles])
  const columnTitleClass =
    role === 'principal' ? heColumnTitleBrand : heColumnTitleAI

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    load()
      .then((data) => {
        if (cancelled) return
        setEntityId(data.id)
        setHeading(resolveTitle?.(data) ?? data.display_name ?? title ?? data.id)
        setSubheading(resolveSubtitle?.(data))
        setAvatar(resolveAvatar?.(data))
        setFiles(data.files ?? {})
        const names = standardFiles?.length
          ? [...standardFiles]
          : Object.keys(data.files ?? {}).sort()
        const first = names[0] ?? ''
        setActiveFile(first)
        setDraft(first ? (data.files[first] ?? '') : '')
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('profile.state.loadProfileFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [load, title, standardFiles, resolveTitle, resolveSubtitle, resolveAvatar])

  function selectFile(name: string) {
    const isDirty =
      activeFile !== '' && draft !== (files[activeFile] ?? '')
    if (isDirty && !window.confirm(t('profile.filesEditor.switchConfirm'))) {
      return
    }
    setActiveFile(name)
    setDraft(files[name] ?? '')
  }

  async function handleSave() {
    if (!activeFile) return
    setSaving(true)
    try {
      await save(activeFile, draft)
      setFiles((prev) => ({ ...prev, [activeFile]: draft }))
      toast.success(t('profile.filesEditor.saveSuccess', { filename: activeFile }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  const dirty = activeFile !== '' && draft !== (files[activeFile] ?? '')

  const pageHeader = (
    <>
      <Link
        to={backTo}
        className={cn(
          'inline-flex items-center gap-1.5 text-sm text-text-secondary',
          'hover:text-brand',
          heSpring,
        )}
      >
        <ArrowLeft className="size-4" />
        {backLabel}
      </Link>

      {avatar && !loading && !error ? (
        <header className="space-y-3">
          <div className="flex flex-wrap items-start gap-4">
            <ProfileAvatar
              id={avatar.id}
              name={avatar.name}
              avatarUrl={avatar.avatarUrl}
              size="lg"
            />
            <div className="min-w-0 flex-1 space-y-3">
              <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
                <h1 className={hePageTitle}>{heading}</h1>
                <span className={role === 'principal' ? heEyebrowBrand : heEyebrowAI}>
                  {eyebrow}
                </span>
              </div>
              {subheading && (
                <p className="font-mono text-xs text-text-tertiary">{subheading}</p>
              )}
              <p className={hePageDesc}>{pageDescription}</p>
            </div>
          </div>
        </header>
      ) : (
        <ProfilePageHeader
          role={role}
          eyebrow={eyebrow}
          title={loading ? pageTitle : heading}
          description={
            loading ? (
              pageDescription
            ) : (
              <>
                {heading !== entityId && (
                  <span className="mb-1 block font-mono text-xs text-text-tertiary">
                    {entityId}
                  </span>
                )}
                {pageDescription}
              </>
            )
          }
        />
      )}
    </>
  )

  return (
    <PageLayout header={<div className="space-y-8">{pageHeader}</div>}>
    <div className="space-y-8">
      {loading && (
        <ProfileStatePanel
          title={t('common.loading')}
          description={t('profile.filesEditor.loadingMarkdown')}
        />
      )}

      {!loading && error && (
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.loadFailed')}
          description={error}
        />
      )}

      {!loading && !error && fileNames.length === 0 && (
        <ProfileStatePanel
          title={t('profile.filesEditor.emptyTitle')}
          description={emptyHint}
        />
      )}

      {!loading && !error && fileNames.length > 0 && (
        <div className={cn(hePanelShell, 'p-6 sm:p-8')}>
          <div className="grid gap-8 lg:grid-cols-[minmax(0,220px)_minmax(0,1fr)]">
            <aside className="space-y-4">
              <p className={columnTitleClass}>{t('profile.filesEditor.sectionTitle')}</p>
              <nav className="flex flex-row flex-wrap gap-2 lg:flex-col lg:items-start">
                {fileNames.map((name) => {
                  const exists = Object.hasOwn(files, name)
                  return (
                  <button
                    key={name}
                    type="button"
                    onClick={() => selectFile(name)}
                    className={cn(
                      hePressable,
                      activeFile === name ? heFilePillSelected : heFilePill,
                      !exists && 'opacity-60',
                    )}
                  >
                    {profileFileHasTitle(name) ? (
                      <span className="flex min-w-0 flex-col gap-0.5 text-left">
                        <span className="text-[13px]">{t(`profile.files.${name}`)}</span>
                        <span className="font-mono text-[10px] text-text-tertiary/90">
                          {name}
                        </span>
                      </span>
                    ) : (
                      name
                    )}
                    {!exists && (
                      <span className="ml-1 text-[10px] text-text-tertiary">
                        {t('profile.filesEditor.notCreated')}
                      </span>
                    )}
                  </button>
                  )
                })}
              </nav>
            </aside>

            <div className="flex min-w-0 flex-col gap-4">
              <SettingsFieldRow
                label={profileFileCaption(activeFile)}
                hint={
                  (fileHints[activeFile] ?? t('profile.filesEditor.defaultHint')) +
                  (!Object.hasOwn(files, activeFile) ? t('profile.filesEditor.createOnSave') : '')
                }
              >
                <div className="space-y-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <MarkdownViewToggle mode={viewMode} onChange={setViewMode} />
                    {dirty && (
                      <span className="text-xs font-medium text-warning">
                        {viewMode === 'preview'
                          ? t('profile.filesEditor.unsavedPreview')
                          : t('profile.filesEditor.unsaved')}
                      </span>
                    )}
                  </div>

                  {viewMode === 'preview' ? (
                    <div className={cn(heInputReadonly, 'p-5 sm:p-6')}>
                      <MarkdownReader content={draft} constrained={false} />
                    </div>
                  ) : (
                    <Textarea
                      value={draft}
                      onChange={(e) => setDraft(e.target.value)}
                      spellCheck={false}
                      className="min-h-[420px] font-mono text-[14px] leading-[1.75]"
                    />
                  )}
                </div>
              </SettingsFieldRow>

              {viewMode === 'source' && (
                <div className="flex flex-wrap items-center gap-3 border-t border-border-subtle/80 pt-4">
                  <Button
                    onClick={handleSave}
                    disabled={!dirty || saving}
                    className={cn(hePressable, 'gap-2 rounded-full px-5')}
                  >
                    <Save className="size-4" />
                    {saving ? t('common.saving') : t('profile.filesEditor.save')}
                  </Button>
                  {dirty && (
                    <span className="text-xs font-medium text-warning">
                      {t('profile.filesEditor.unsaved')}
                    </span>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
    </PageLayout>
  )
}
