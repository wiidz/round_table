import { useMemo, useEffect, useState } from 'react'

import { fetchBriefTemplates } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import {
  BriefTemplateGridCard,
  BriefTemplateGridSkeleton,
} from '@/components/brief/brief-template-grid-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { heSubsectionTitleNeutral } from '@/lib/highend-styles'

import type { BriefTemplateIndex } from '@/types/brief-template'

function TemplateSection({
  title,
  hint,
  templates,
}: {
  title: string
  hint?: string
  templates: BriefTemplateIndex[]
}) {
  if (templates.length === 0) return null

  return (
    <section className="space-y-4">
      <div>
        <h2 className={heSubsectionTitleNeutral}>{title}</h2>
        {hint && <p className="mt-1 text-[12px] text-text-tertiary">{hint}</p>}
      </div>
      <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {templates.map((item) => (
          <li key={item.id} className="min-h-0">
            <BriefTemplateGridCard template={item} />
          </li>
        ))}
      </ul>
    </section>
  )
}

export function BriefTemplatesPage() {
  const [templates, setTemplates] = useState<BriefTemplateIndex[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    fetchBriefTemplates()
      .then((data) => {
        if (!cancelled) {
          setTemplates(data.templates ?? [])
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载简报模板')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const { builtin, custom } = useMemo(() => {
    const builtinItems: BriefTemplateIndex[] = []
    const customItems: BriefTemplateIndex[] = []
    for (const item of templates) {
      if (item.source === 'builtin') builtinItems.push(item)
      else customItems.push(item)
    }
    return { builtin: builtinItems, custom: customItems }
  }, [templates])

  return (
    <div className="space-y-8">
      <ProfilePageHeader
        role="principal"
        eyebrow="Meeting Brief"
        title="简报模板 · Brief Template"
        description="可复用的会议意图：主题、目标、议程与范围。内置模板可另存为自定义副本后再改。"
      />

      {loading && <BriefTemplateGridSkeleton />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && templates.length === 0 && (
        <ProfileStatePanel
          title="暂无简报模板"
          description="请确认 data/_templates/briefs/ 下存在 BRIEF.yaml。"
        />
      )}

      {!loading && !error && templates.length > 0 && (
        <div className="space-y-10">
          <TemplateSection
            title="内置模板"
            hint="只读起点；编辑后请另存为自定义模板。"
            templates={builtin}
          />
          <TemplateSection
            title="自定义模板"
            hint="保存在 data/briefs/，可直接修改并保存。"
            templates={custom}
          />
        </div>
      )}
    </div>
  )
}
