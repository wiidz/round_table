import { useMemo, useEffect, useState } from 'react'

import { fetchBriefTemplates } from '@/api/brief-templates'
import { ApiError } from '@/api/client'
import {
  BriefTemplateGridCard,
  BriefTemplateGridSkeleton,
} from '@/components/brief/brief-template-grid-card'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useI18n } from '@/hooks/use-i18n'
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
  const i18n = useI18n()
  const { t } = i18n
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
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('brief.page.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

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
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={i18n.briefTemplatePageEyebrow()}
          title={i18n.briefTemplatePageTitle()}
          description={t('brief.page.description')}
        />
      }
    >
    <div className="space-y-8">
      {loading && <BriefTemplateGridSkeleton />}

      {!loading && error && (
        <ProfileStatePanel
          variant="danger"
          title={t('common.error.loadFailed')}
          description={error}
        />
      )}

      {!loading && !error && templates.length === 0 && (
        <ProfileStatePanel
          title={t('brief.page.emptyTitle')}
          description={t('brief.page.emptyDescription')}
        />
      )}

      {!loading && !error && templates.length > 0 && (
        <div className="space-y-10">
          <TemplateSection
            title={t('brief.page.sectionBuiltin')}
            hint={t('brief.page.sectionBuiltinHint')}
            templates={builtin}
          />
          <TemplateSection
            title={t('brief.page.sectionCustom')}
            hint={t('brief.page.sectionCustomHint')}
            templates={custom}
          />
        </div>
      )}
    </div>
    </PageLayout>
  )
}
