import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'sonner'

import { fetchBriefTemplate, saveBriefTemplate } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import {
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { Button } from '@/components/ui/button'
import {
  heEyebrowBrand,
  heFieldSurface,
  hePageDesc,
  hePageTitle,
  hePanelShell,
  hePressable,
  heSpring,
  heTextarea,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { BriefTemplateDetail } from '@/types/brief-template'

function formatLaunchPreview(detail: BriefTemplateDetail): string {
  const lines: string[] = []
  if (detail.launch.topic) lines.push(`主题：${detail.launch.topic}`)
  if (detail.launch.brief.goal) lines.push(`目标：${detail.launch.brief.goal}`)
  if (detail.launch.brief.agenda?.length) {
    lines.push(`议程：${detail.launch.brief.agenda.join(' · ')}`)
  }
  if (detail.launch.meeting.mode) lines.push(`模式：${detail.launch.meeting.mode}`)
  return lines.join('\n')
}

export function BriefTemplateDetailPage() {
  const { id: rawId } = useParams()
  const id = rawId ? decodeURIComponent(rawId) : ''

  const [detail, setDetail] = useState<BriefTemplateDetail | null>(null)
  const [draft, setDraft] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    const data = await fetchBriefTemplate(id)
    setDetail(data)
    setDraft(data.content)
    setError(null)
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

  const readonly = detail?.source === 'builtin'
  const dirty = detail != null && draft !== detail.content

  async function handleSave() {
    if (!id || readonly) return
    setSaving(true)
    try {
      await saveBriefTemplate(id, draft)
      setDetail((prev) => (prev ? { ...prev, content: draft } : prev))
      toast.success('已保存 BRIEF.yaml')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  if (!id) return null

  return (
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

      <header className="space-y-3">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
          <h1 className={hePageTitle}>{detail?.title ?? id}</h1>
          <span className={heEyebrowBrand}>Meeting Brief</span>
          {detail && (
            <span className="rounded-full bg-black/[0.04] px-2.5 py-0.5 font-mono text-[11px] text-text-tertiary ring-1 ring-inset ring-black/[0.05]">
              {detail.source === 'builtin' ? '内置 · 只读' : '自定义'}
            </span>
          )}
        </div>
        {detail?.description && <p className={hePageDesc}>{detail.description}</p>}
        <p className="font-mono text-xs text-text-tertiary">{id} · BRIEF.yaml</p>
      </header>

      {loading && (
        <ProfileStatePanel title="加载中" description="正在读取 BRIEF.yaml…" />
      )}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && detail && (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,280px)]">
          <div className={cn(hePanelShell, 'flex flex-col gap-4 p-6 sm:p-8')}>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-medium text-text-primary">BRIEF.yaml</p>
              {readonly && (
                <span className="text-xs text-text-tertiary">
                  内置模板不可编辑；可复制内容后保存为自定义模板 ID
                </span>
              )}
            </div>
            <div className={heFieldSurface}>
              <textarea
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                readOnly={readonly}
                spellCheck={false}
                className={cn(heTextarea, readonly && 'opacity-80')}
              />
            </div>
            {!readonly && (
              <div className="flex flex-wrap items-center gap-3 border-t border-border-subtle/80 pt-4">
                <Button
                  onClick={handleSave}
                  disabled={!dirty || saving}
                  className={cn(hePressable, 'gap-2 rounded-full px-5')}
                >
                  <Save className="size-4" />
                  {saving ? '保存中…' : '保存模板'}
                </Button>
                {dirty && (
                  <span className="text-xs font-medium text-warning">有未保存的修改</span>
                )}
              </div>
            )}
          </div>

          <aside className={cn(hePanelShell, 'space-y-3 p-5')}>
            <p className="text-[13px] font-semibold text-text-primary">预填预览</p>
            <pre className="whitespace-pre-wrap text-[12px] leading-relaxed text-text-secondary">
              {formatLaunchPreview(detail) || '解析后将用于 MeetingCreated 预填'}
            </pre>
          </aside>
        </div>
      )}
    </div>
  )
}
