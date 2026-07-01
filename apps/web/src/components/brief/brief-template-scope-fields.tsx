import type { ReactNode } from 'react'
import { Ban, CheckCircle2, Scan } from 'lucide-react'

import {
  applyScopePreset,
  getBriefScopePresets,
  scopePresetApplied,
  type BriefScopePreset,
} from '@/components/brief/brief-scope-presets'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  briefFieldLabelClass,
  briefScopeBlockShell,
  briefScopeBlockTone,
  briefScopeIconShell,
} from '@/components/brief/brief-template-sections'
import { Textarea } from '@/components/ui/textarea'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefScopeEmptyCopy, getBriefSections } from '@/lib/i18n/brief-sections'
import { hePressable, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

function ScopePresetButtons({
  presets,
  value,
  chipClassName,
  onApply,
}: {
  presets: readonly BriefScopePreset[]
  value: string
  chipClassName?: string
  onApply: (snippet: string) => void
}) {
  return (
    <div className="flex flex-wrap gap-1.5">
        {presets.map((preset) => {
          const applied = scopePresetApplied(value, preset.value)
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
                  : chipClassName,
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

interface ScopeFieldProps {
  id: string
  label: string
  icon: ReactNode
  blockClassName?: string
  chipClassName?: string
  presets?: readonly BriefScopePreset[]
  value: string
  readonly?: boolean
  placeholder?: string
  onChange: (value: string) => void
}

function ScopeField({
  id,
  label,
  icon,
  blockClassName,
  chipClassName,
  presets,
  value,
  readonly,
  placeholder,
  onChange,
}: ScopeFieldProps) {
  return (
    <div className={cn(briefScopeBlockShell, blockClassName, 'space-y-3.5')}>
      <div className="flex items-center gap-2">
        <span className={briefScopeIconShell}>{icon}</span>
        <p className={briefFieldLabelClass}>{label}</p>
      </div>
      <Textarea
        id={id}
        value={value}
        readOnly={readonly}
        rows={3}
        className="min-h-[5.5rem] border-0 bg-surface/70 font-sans text-sm shadow-none ring-0 focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-primary/35"
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
      {!readonly && presets && presets.length > 0 && (
        <ScopePresetButtons
          presets={presets}
          value={value}
          chipClassName={chipClassName}
          onApply={(snippet) => onChange(applyScopePreset(value, snippet))}
        />
      )}
    </div>
  )
}

interface BriefTemplateScopeFieldsProps {
  inScope: string
  outOfScope: string
  doneCriteria: string
  readonly?: boolean
  onChange: (patch: { in_scope?: string; out_of_scope?: string; done_criteria?: string }) => void
}

const scopePresetChipTone = {
  inScope: 'bg-surface/80 text-brand ring-primary/20 hover:bg-surface hover:ring-primary/30',
  outOfScope:
    'bg-surface/80 text-destructive ring-destructive/20 hover:bg-surface hover:ring-destructive/30',
  done: 'bg-surface/80 text-success ring-success/25 hover:bg-surface hover:ring-success/35',
} as const

export function BriefTemplateScopeFields({
  inScope,
  outOfScope,
  doneCriteria,
  readonly,
  onChange,
}: BriefTemplateScopeFieldsProps) {
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const presets = getBriefScopePresets(locale)

  return (
    <div className="space-y-4">
      <BriefSectionHeading
        title={sections.scope.title}
        description={sections.scope.description}
        as="h3"
      />
      <div className="space-y-4">
        <ScopeField
          id="brief-in-scope"
          label={t('brief.scope.inScope')}
          icon={<Scan className="size-3.5 text-brand" aria-hidden />}
          blockClassName={briefScopeBlockTone.inScope}
          chipClassName={scopePresetChipTone.inScope}
          presets={presets.inScope}
          value={inScope}
          readonly={readonly}
          placeholder={t('brief.scope.inScopePlaceholder')}
          onChange={(v) => onChange({ in_scope: v })}
        />
        <ScopeField
          id="brief-out-scope"
          label={t('brief.scope.outOfScope')}
          icon={<Ban className="size-3.5 text-destructive" aria-hidden />}
          blockClassName={briefScopeBlockTone.outOfScope}
          chipClassName={scopePresetChipTone.outOfScope}
          presets={presets.outOfScope}
          value={outOfScope}
          readonly={readonly}
          placeholder={t('brief.scope.outOfScopePlaceholder')}
          onChange={(v) => onChange({ out_of_scope: v })}
        />
        <ScopeField
          id="brief-done"
          label={t('brief.scope.doneCriteria')}
          icon={<CheckCircle2 className="size-3.5 text-success" aria-hidden />}
          blockClassName={briefScopeBlockTone.done}
          chipClassName={scopePresetChipTone.done}
          presets={presets.doneCriteria}
          value={doneCriteria}
          readonly={readonly}
          placeholder={t('brief.scope.donePlaceholder')}
          onChange={(v) => onChange({ done_criteria: v })}
        />
      </div>
    </div>
  )
}

interface BriefTemplateScopePreviewProps {
  inScope?: string
  outOfScope?: string
  doneCriteria?: string
}

function ScopePreviewBlock({
  label,
  icon,
  blockClassName,
  children,
  empty = '—',
}: {
  label: string
  icon: ReactNode
  blockClassName: string
  children?: string
  empty?: string
}) {
  const text = children?.trim()
  return (
    <div className={cn(briefScopeBlockShell, blockClassName, 'space-y-3')}>
      <div className="flex items-center gap-2">
        <span className={briefScopeIconShell}>{icon}</span>
        <p className={briefFieldLabelClass}>{label}</p>
      </div>
      <p
        className={cn(
          'text-[14px] leading-relaxed',
          text ? 'font-medium text-text-primary' : 'text-text-tertiary',
        )}
      >
        {text || empty}
      </p>
    </div>
  )
}

export function BriefTemplateScopePreview({
  inScope,
  outOfScope,
  doneCriteria,
}: BriefTemplateScopePreviewProps) {
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const emptyCopy = getBriefScopeEmptyCopy(locale)

  return (
    <section className="space-y-4">
      <BriefSectionHeading
        title={sections.scope.title}
        description={sections.scope.description}
      />
      <div className="space-y-4">
        <ScopePreviewBlock
          label={t('brief.scope.inScope')}
          icon={<Scan className="size-3.5 text-brand" aria-hidden />}
          blockClassName={briefScopeBlockTone.inScope}
          empty={emptyCopy.inScope}
          children={inScope}
        />
        <ScopePreviewBlock
          label={t('brief.scope.outOfScope')}
          icon={<Ban className="size-3.5 text-destructive" aria-hidden />}
          blockClassName={briefScopeBlockTone.outOfScope}
          empty={emptyCopy.outOfScope}
          children={outOfScope}
        />
        <ScopePreviewBlock
          label={t('brief.scope.doneCriteria')}
          icon={<CheckCircle2 className="size-3.5 text-success" aria-hidden />}
          blockClassName={briefScopeBlockTone.done}
          empty={emptyCopy.doneCriteria}
          children={doneCriteria}
        />
      </div>
    </section>
  )
}
