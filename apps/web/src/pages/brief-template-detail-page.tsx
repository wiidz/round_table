import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Pencil, Save, X } from 'lucide-react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

import { createBriefTemplate, fetchBriefTemplate, saveBriefTemplate } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import { BriefTemplateFormFields } from '@/components/brief/brief-template-form-fields'
import { BriefTemplatePageHeader } from '@/components/brief/brief-template-meta-fields'
import { BriefTemplatePreview } from '@/components/brief/brief-template-preview'
import { PageLayout } from '@/components/layout/page-main-layout'
import { ProfileStatePanel } from '@/components/profile/profile-page-header'
import { Button } from '@/components/ui/button'
import {
  heEyebrowBrand,
  hePanelShell,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import {
  documentsEqual,
  emptyBriefDocument,
  normalizeBriefDocument,
} from '@/lib/brief-template-document'
import { cn } from '@/lib/utils'

import type { BriefTemplateDetail, BriefTemplateDocument } from '@/types/brief-template'

export function BriefTemplateDetailPage() {
  const navigate = useNavigate()
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

  const isBuiltin = detail?.source === 'builtin'

  const load = useCallback(async () => {
    const data = await fetchBriefTemplate(id)
    const doc = normalizeBriefDocument(data.document ?? emptyBriefDocument(data.title))
    setDetail(data)
    setSavedDocument(doc)
    setFormDocument(doc)
    setError(null)
  }, [id])

  useEffect(() => {
    setMode('view')
  }, [id])

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    load()
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载模板')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id, load])

  const dirty = !documentsEqual(formDocument, savedDocument)
  const headerDocument = mode === 'edit' ? formDocument : savedDocument

  function handleStartEdit() {
    setFormDocument(savedDocument)
    setMode('edit')
  }

  function handleCancelEdit() {
    setFormDocument(savedDocument)
    setMode('view')
  }

  async function handleSave() {
    setSaving(true)
    try {
      const normalized = normalizeBriefDocument(formDocument)
      if (!normalized.meta.title) {
        toast.error('请填写模板名称')
        return
      }

      if (isBuiltin) {
        const res = await createBriefTemplate({ document: normalized })
        toast.success(`已另存为自定义模板（${res.id}）`)
        navigate(`/brief-templates/${encodeURIComponent(res.id)}`)
        return
      }

      await saveBriefTemplate(id, { document: normalized })
      setSavedDocument(normalized)
      setFormDocument(normalized)
      setMode('view')
      await load()
      toast.success('已保存模板')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
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
        返回简报模板列表
      </Link>

      {!loading && !error && detail && (
        <header className="space-y-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-wrap items-center gap-2">
              <span className={heEyebrowBrand}>Meeting Brief · 简报模板</span>
              <span className="rounded-full bg-black/[0.04] px-2.5 py-0.5 font-mono text-[11px] text-text-tertiary ring-1 ring-inset ring-black/[0.05]">
                {detail.source === 'builtin' ? '内置' : '自定义'}
              </span>
              {!isBuiltin && (
                <span className="font-mono text-[11px] text-text-tertiary/80">{id}</span>
              )}
            </div>

            {mode === 'view' && (
              <Button
                type="button"
                variant="outline"
                onClick={handleStartEdit}
                className={cn(hePressable, 'shrink-0 gap-2 !rounded-xs px-4')}
              >
                <Pencil className="size-4" />
                编辑
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
        <ProfileStatePanel title="加载中" description="正在读取模板…" />
      )}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && detail && (
        <>
          <div className={cn(hePanelShell, 'overflow-visible')}>
            {mode === 'view' ? (
              <div className="p-5 sm:p-7">
                <BriefTemplatePreview document={savedDocument} />
              </div>
            ) : (
              <div className="flex flex-col gap-8 p-6 sm:p-8">
                <p className="text-[12px] leading-relaxed text-text-tertiary">
                  编辑模板信息与会议预填字段；保存时由服务端生成 BRIEF.yaml。
                  {isBuiltin && ' 内置模板另存时将按模板名称自动生成 ID。'}
                </p>

                <BriefTemplateFormFields
                  document={formDocument}
                  readonly={false}
                  onChange={setFormDocument}
                />

                <div className="flex flex-wrap items-center gap-3 border-t border-black/[0.05] pt-4">
                  <Button
                    onClick={() => void handleSave()}
                    disabled={saving || (!isBuiltin && !dirty)}
                    className={cn(hePressable, 'gap-2 rounded-full px-5')}
                  >
                    <Save className="size-4" />
                    {saving ? '保存中…' : isBuiltin ? '另存为自定义模板' : '保存模板'}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={saving}
                    onClick={handleCancelEdit}
                    className={cn(hePressable, 'gap-2 !rounded-xs px-4')}
                  >
                    <X className="size-4" />
                    取消
                  </Button>
                  {dirty && !isBuiltin && (
                    <span className="text-xs font-medium text-warning">有未保存的修改</span>
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
