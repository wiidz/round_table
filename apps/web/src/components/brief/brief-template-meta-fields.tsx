import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { heColumnTitleBrand, heFieldHint, hePageDesc, hePageTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { BriefTemplateDocument } from '@/types/brief-template'

interface BriefTemplatePageHeaderProps {
  document: BriefTemplateDocument
  className?: string
}

/** 页头：模板名称 + 模板说明（卡片外，只读） */
export function BriefTemplatePageHeader({ document, className }: BriefTemplatePageHeaderProps) {
  const title = document.meta.title?.trim()
  const description = document.meta.description?.trim()

  return (
    <div className={cn('space-y-2', className)}>
      <h1 className={cn(hePageTitle, 'font-bold text-text-primary')}>
        {title || '未命名模板'}
      </h1>
      <p className={hePageDesc}>
        <span className="font-medium text-text-secondary">模板说明：</span>
        <span className={description ? 'text-text-secondary' : 'text-text-tertiary'}>
          {description || '暂无说明'}
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
  function patchMeta(partial: Partial<BriefTemplateDocument['meta']>) {
    onChange({ ...document, meta: { ...document.meta, ...partial } })
  }

  return (
    <section className="space-y-4">
      <div className="space-y-1">
        <p className={heColumnTitleBrand}>模板信息</p>
        <p className={heFieldHint}>列表展示用；不影响会议预填字段</p>
      </div>

      <div className="space-y-2">
        <label htmlFor="brief-meta-title" className="text-[13px] font-medium text-text-secondary">
          模板名称 <span className="text-destructive">*</span>
        </label>
        <Input
          id="brief-meta-title"
          value={document.meta.title}
          readOnly={readonly}
          placeholder="例如：裁决型评审"
          onChange={(e) => patchMeta({ title: e.target.value })}
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="brief-meta-desc" className="text-[13px] font-medium text-text-secondary">
          模板说明
        </label>
        <Textarea
          id="brief-meta-desc"
          value={document.meta.description ?? ''}
          readOnly={readonly}
          rows={2}
          className="min-h-[5rem] font-sans text-[15px] leading-relaxed"
          placeholder="说明这套模板的适用场景与用途，例如：围绕 Topic 形成可执行共识，适合是否上线类议题"
          onChange={(e) => patchMeta({ description: e.target.value })}
        />
      </div>
    </section>
  )
}
