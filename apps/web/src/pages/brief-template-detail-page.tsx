import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Pencil, Save, X } from 'lucide-react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import { createBriefTemplate, fetchBriefTemplate, saveBriefTemplate } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import { BriefTemplateFormFields } from '@/components/brief/brief-template-form-fields'
import { BriefTemplatePageHeader } from '@/components/brief/brief-template-meta-fields'
import { BriefTemplatePreview } from '@/components/brief/brief-template-preview'
import { PageLayout } from '@/components/layout/page-main-layout'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import {
  heEyebrowBrand,
  hePanelShell,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import {
  briefTemplateHasSubstantiveContent,
  documentsEqual,
  emptyBriefDocument,
  normalizeBriefDocument,
} from '@/lib/brief-template-document'
import { cn } from '@/lib/utils'

import type { BriefTemplateDetail, BriefTemplateDocument } from '@/types/brief-template'

export function BriefTemplateDetailPage() {
  const i18n = useI18n()
  const { t } = i18n
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const [detail, setDetail] = useState<BriefTemplateDetail | null>(null)
  const [savedDocument, setSavedDocument] = useState<BriefTemplateDocument>(
    emptyBriefDocument(),
  )
  const [formDocument, setFormDocument] = useState<BriefTemplateDocument>(emptyBriefDocument())
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [mode, setMode] = useState<'view' | 'edit'>('view')

  const isNewDraft = id === 'new'
  const isBuiltin = !isNewDraft && detail?.source === 'builtin'

  const load = useCallback(async () => {
    const data = await fetchBriefTemplate(id)
    const doc = normalizeBriefDocument(data.document ?? emptyBriefDocument(data.title))
    setDetail(data)
    setSavedDocument(doc)
    setFormDocument(doc)
    setError(null)
  }, [id])

  useEffect(() => {
    if (isNewDraft) return
    setMode('view')
  }, [id, isNewDraft])

  useEffect(() => {
    if (isNewDraft || loading || error || !detail) return
    if (searchParams.get('edit') !== '1') return
    setFormDocument(savedDocument)
    setMode('edit')
    navigate(`/brief-templates/${encodeURIComponent(id)}`, { replace: true })
  }, [isNewDraft, loading, error, detail, searchParams, savedDocument, navigate, id])

  useEffect(() => {
    if (!id) return
    if (isNewDraft) {
      const title = searchParams.get('title')?.trim() ?? ''
      const doc = emptyBriefDocument(title)
      setDetail({
        id: '',
        title: title || t('brief.meta.unnamed'),
        source: 'custom',
        content: '',
        document: doc,
        launch: { topic: '', brief: doc.brief, meeting: {} },
        updated_at: new Date().toISOString(),
      })
      setSavedDocument(doc)
      setFormDocument(doc)
      setError(null)
      setLoading(false)
      setMode('edit')
      return
    }

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
          setError(t('brief.page.detailLoadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id, isNewDraft, load, searchParams, t])

  const dirty = !documentsEqual(formDocument, savedDocument)
  const headerDocument = mode === 'edit' ? formDocument : savedDocument

  function handleStartEdit() {
    setFormDocument(savedDocument)
    setMode('edit')
  }

  function handleCancelEdit() {
    if (isNewDraft) {
      navigate('/brief-templates')
      return
    }
    setFormDocument(savedDocument)
    setMode('view')
  }

  async function handleSave() {
    setSaving(true)
    try {
      const normalized = normalizeBriefDocument(formDocument)
      if (!normalized.meta.title) {
        toast.error(t('brief.page.titleRequired'))
        return
      }
      if (!briefTemplateHasSubstantiveContent(normalized)) {
        toast.error(t('brief.page.contentRequired'))
        return
      }

      if (isNewDraft || isBuiltin) {
        const res = await createBriefTemplate({ document: normalized })
        toast.success(
          isNewDraft ? t('brief.page.created') : t('brief.page.savedAsCustom', { id: res.id }),
        )
        navigate(`/brief-templates/${encodeURIComponent(res.id)}`)
        return
      }

      await saveBriefTemplate(id, { document: normalized })
      setSavedDocument(normalized)
      setFormDocument(normalized)
      setMode('view')
      await load()
      toast.success(t('brief.page.saved'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  if (!id) return null

  const pageHeader = (
    <div className="space-y-8">
      <Link
        to="/brief-templates"
        className={cn(
          'inline-flex items-center gap-1.5 text-sm text-text-secondary',
          'hover:text-brand',
          heSpring,
        )}
      >
        <ArrowLeft className="size-4" />
        {t('brief.page.detailBack')}
      </Link>

      {!loading && !error && detail && (
        <header className="space-y-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-wrap items-center gap-2">
              <span className={heEyebrowBrand}>{i18n.briefTemplatePageEyebrow()}</span>
              <span className="rounded-full bg-black/[0.04] px-2.5 py-0.5 font-mono text-[11px] text-text-tertiary ring-1 ring-inset ring-black/[0.05]">
                {isNewDraft ? t('common.custom') : detail.source === 'builtin' ? t('common.builtin') : t('common.custom')}
              </span>
              {!isBuiltin && !isNewDraft && (
                <span className="font-mono text-[11px] text-text-tertiary/80">{id}</span>
              )}
            </div>

            {mode === 'view' && !isNewDraft && (
              <Button
                type="button"
                variant="outline"
                onClick={handleStartEdit}
                className={cn(hePressable, 'shrink-0 gap-2 !rounded-xs px-4')}
              >
                <Pencil className="size-4" />
                {t('common.edit')}
              </Button>
            )}
          </div>

          <BriefTemplatePageHeader document={headerDocument} />
        </header>
      )}
    </div>
  )

  return (
    <PageLayout header={pageHeader}>
    <div className="space-y-8">
      {loading && (
        <ProfileStatePanel
          title={t('common.loading')}
          description={t('brief.page.loadingDescription')}
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
          <div className={cn(hePanelShell, 'overflow-visible')}>
            {mode === 'view' ? (
              !isNewDraft && (
              <div className="p-5 sm:p-7">
                <BriefTemplatePreview document={savedDocument} />
              </div>
              )
            ) : (
              <div className="flex flex-col gap-8 p-6 sm:p-8">
                <p className="text-[12px] leading-relaxed text-text-tertiary">
                  {t('brief.page.editHint')}
                  {isBuiltin && t('brief.page.editHintBuiltin')}
                  {isNewDraft && t('brief.page.editHintNew')}
                </p>

                <BriefTemplateFormFields
                  document={formDocument}
                  readonly={false}
                  onChange={setFormDocument}
                />

                <div className="flex flex-wrap items-center gap-3 border-t border-black/[0.05] pt-4">
                  <Button
                    onClick={() => void handleSave()}
                    disabled={saving || (!isBuiltin && !isNewDraft && !dirty)}
                    className={cn(hePressable, 'gap-2 rounded-full px-5')}
                  >
                    <Save className="size-4" />
                    {saving
                      ? t('common.saving')
                      : isNewDraft
                        ? t('brief.page.create')
                        : isBuiltin
                          ? t('brief.page.saveAsCustom')
                          : t('brief.page.save')}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={saving}
                    onClick={handleCancelEdit}
                    className={cn(hePressable, 'gap-2 !rounded-xs px-4')}
                  >
                    <X className="size-4" />
                    {t('common.cancel')}
                  </Button>
                  {dirty && !isBuiltin && (
                    <span className="text-xs font-medium text-warning">
                      {t('brief.page.unsavedChanges')}
                    </span>
                  )}
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
    </PageLayout>
  )
}
