import { useEffect, useState } from 'react'

import { fetchBriefTemplates } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import {
  ProfileListCard,
  ProfileListSkeleton,
} from '@/components/profile/profile-list-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'

import type { BriefTemplateIndex } from '@/types/brief-template'

function sourceLabel(source: BriefTemplateIndex['source']): string {
  return source === 'builtin' ? '内置' : '自定义'
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

  return (
    <div className="space-y-8">
      <ProfilePageHeader
        role="principal"
        eyebrow="Meeting Brief"
        title="简报模板 · Brief Template"
        description={
          <>
            可复用的会议简报（ADR-0014）：主题、目标、议程与范围。创建 Meeting 前选用模板预填，或从历史{' '}
            <code className="rounded-md bg-black/[0.04] px-1.5 py-0.5 font-mono text-[12px] ring-1 ring-inset ring-black/[0.05]">
              MEETING.md
            </code>{' '}
            克隆。内置模板只读；自定义模板保存在{' '}
            <code className="font-mono text-xs">data/briefs/</code>。
          </>
        }
      />

      {loading && <ProfileListSkeleton />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && templates.length === 0 && (
        <ProfileStatePanel
          title="暂无简报模板"
          description="请确认 data/_templates/briefs/ 下存在 BRIEF.yaml，或在此创建自定义模板。"
        />
      )}

      {!loading && !error && templates.length > 0 && (
        <ul className="space-y-4">
          {templates.map((item) => (
            <li key={item.id}>
              <ProfileListCard
                role="principal"
                href={`/brief-templates/${encodeURIComponent(item.id)}`}
                title={item.title}
                subtitle={item.description || item.id}
                files={[{ name: 'BRIEF.yaml', present: true }]}
                meta={`${sourceLabel(item.source)} · ${item.id}`}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
