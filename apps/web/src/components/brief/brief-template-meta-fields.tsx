import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useI18n } from '@/hooks/use-i18n'
import { heColumnTitleBrand, heFieldHint, hePageDesc, hePageTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { BriefTemplateDocument } from '@/types/brief-template'

interface BriefTemplatePageHeaderProps {
  document: BriefTemplateDocument
  className?: string
}

export function BriefTemplatePageHeader({ document, className }: BriefTemplatePageHeaderProps) {
  const { t } = useI18n()
  const title = document.meta.title?.trim()
  const description = document.meta.description?.trim()

  return (
    <div className={cn('space-y-2', className)}>
      <h1 className={cn(hePageTitle, 'font-bold text-text-primary')}>
        {title || t('brief.meta.unnamed')}
      </h1>
      <p className={hePageDesc}>
        <span className="font-medium text-text-secondary">{t('brief.meta.descriptionPrefix')}</span>
        <span className={description ? 'text-text-secondary' : 'text-text-tertiary'}>
          {description || t('brief.meta.noDescription')}
        </span>
      </p>
    </div>
  )
}

interface BriefTemplateMetaFieldsProps {
  document: BriefTemplateDocument
  readonly?: boolean
  onChange: (next: BriefTemplateDocument) => void
}

export function BriefTemplateMetaFields({
  document,
  readonly,
  onChange,
}: BriefTemplateMetaFieldsProps) {
  const { t } = useI18n()

  function patchMeta(partial: Partial<BriefTemplateDocument['meta']>) {
    onChange({ ...document, meta: { ...document.meta, ...partial } })
  }

  return (
    <section className="space-y-4">
      <div className="space-y-1">
        <p className={heColumnTitleBrand}>{t('brief.meta.sectionTitle')}</p>
        <p className={heFieldHint}>{t('brief.meta.sectionHint')}</p>
      </div>

      <div className="space-y-2">
        <label htmlFor="brief-meta-title" className="text-[13px] font-medium text-text-secondary">
          {t('brief.meta.titleLabel')} <span className="text-destructive">*</span>
        </label>
        <Input
          id="brief-meta-title"
          value={document.meta.title}
          readOnly={readonly}
          placeholder={t('brief.meta.titlePlaceholder')}
          onChange={(e) => patchMeta({ title: e.target.value })}
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="brief-meta-desc" className="text-[13px] font-medium text-text-secondary">
          {t('brief.meta.descriptionLabel')}
        </label>
        <Textarea
          id="brief-meta-desc"
          value={document.meta.description ?? ''}
          readOnly={readonly}
          rows={2}
          className="min-h-[5rem] font-sans text-[15px] leading-relaxed"
          placeholder={t('brief.meta.descriptionPlaceholder')}
          onChange={(e) => patchMeta({ description: e.target.value })}
        />
      </div>
    </section>
  )
}
