import type { ReactNode } from 'react'
import { Ban, CheckCircle2, Scan } from 'lucide-react'

import {
  applyScopePreset,
  BRIEF_SCOPE_PRESETS,
  scopePresetApplied,
  type BriefScopePreset,
} from '@/components/brief/brief-scope-presets'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import {
  BRIEF_TEMPLATE_SECTIONS,
  BRIEF_SCOPE_EMPTY_COPY,
  briefFieldLabelClass,
  briefScopeBlockShell,
  briefScopeBlockTone,
  briefScopeIconShell,
} from '@/components/brief/brief-template-sections'
import { Textarea } from '@/components/ui/textarea'
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
  return (
    <div className="space-y-4">
      <BriefSectionHeading
        title={BRIEF_TEMPLATE_SECTIONS.scope.title}
        description={BRIEF_TEMPLATE_SECTIONS.scope.description}
        as="h3"
      />
      <div className="space-y-4">
        <ScopeField
          id="brief-in-scope"
          label="讨论范围"
          icon={<Scan className="size-3.5 text-brand" aria-hidden />}
          blockClassName={briefScopeBlockTone.inScope}
          chipClassName={scopePresetChipTone.inScope}
          presets={BRIEF_SCOPE_PRESETS.inScope}
          value={inScope}
          readonly={readonly}
          placeholder="本次会议要讨论什么"
          onChange={(v) => onChange({ in_scope: v })}
        />
        <ScopeField
          id="brief-out-scope"
          label="不在范围"
          icon={<Ban className="size-3.5 text-destructive" aria-hidden />}
          blockClassName={briefScopeBlockTone.outOfScope}
          chipClassName={scopePresetChipTone.outOfScope}
          presets={BRIEF_SCOPE_PRESETS.outOfScope}
          value={outOfScope}
          readonly={readonly}
          placeholder="明确排除、不在本次会议讨论的内容"
          onChange={(v) => onChange({ out_of_scope: v })}
        />
        <ScopeField
          id="brief-done"
          label="完成标准"
          icon={<CheckCircle2 className="size-3.5 text-success" aria-hidden />}
          blockClassName={briefScopeBlockTone.done}
          chipClassName={scopePresetChipTone.done}
          presets={BRIEF_SCOPE_PRESETS.doneCriteria}
          value={doneCriteria}
          readonly={readonly}
          placeholder="怎样算这场会开完了"
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
  return (
    <section className="space-y-4">
      <BriefSectionHeading
        title={BRIEF_TEMPLATE_SECTIONS.scope.title}
        description={BRIEF_TEMPLATE_SECTIONS.scope.description}
      />
      <div className="space-y-4">
        <ScopePreviewBlock
          label="讨论范围"
          icon={<Scan className="size-3.5 text-brand" aria-hidden />}
          blockClassName={briefScopeBlockTone.inScope}
          empty={BRIEF_SCOPE_EMPTY_COPY.inScope}
          children={inScope}
        />
        <ScopePreviewBlock
          label="不在范围"
          icon={<Ban className="size-3.5 text-destructive" aria-hidden />}
          blockClassName={briefScopeBlockTone.outOfScope}
          empty={BRIEF_SCOPE_EMPTY_COPY.outOfScope}
          children={outOfScope}
        />
        <ScopePreviewBlock
          label="完成标准"
          icon={<CheckCircle2 className="size-3.5 text-success" aria-hidden />}
          blockClassName={briefScopeBlockTone.done}
          empty={BRIEF_SCOPE_EMPTY_COPY.doneCriteria}
          children={doneCriteria}
        />
      </div>
    </section>
  )
}
