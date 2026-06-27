import type { ReactNode } from 'react'
import { CircleHelp } from 'lucide-react'

import { heFieldLabel, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export const settingsFieldRowGrid =
  'grid grid-cols-1 gap-x-6 gap-y-2 sm:grid-cols-[minmax(8rem,11rem)_1fr] sm:items-start'

export function FieldHintPopover({
  content,
  ariaLabel = '字段说明',
}: {
  content: string
  ariaLabel?: string
}) {
  const text = content.trim()
  if (!text) return null

  return (
    <span className="group/hint relative inline-flex shrink-0 self-center">
      <button
        type="button"
        tabIndex={0}
        aria-label={ariaLabel}
        className={cn(
          'rounded-full p-0.5 text-text-tertiary hover:bg-black/[0.04] hover:text-text-secondary',
          heSpring,
        )}
      >
        <CircleHelp className="size-3.5" strokeWidth={1.75} aria-hidden />
      </button>
      <span
        role="tooltip"
        className={cn(
          'pointer-events-none absolute bottom-[calc(100%+6px)] right-0 z-50 w-56',
          'rounded-xs bg-surface px-3 py-2.5 text-[13px] leading-relaxed text-text-secondary',
          'whitespace-pre-line shadow-[var(--panel-shell-shadow)] ring-1 ring-[var(--panel-shell-ring)]',
          'invisible opacity-0 transition-opacity duration-150',
          'group-hover/hint:visible group-hover/hint:opacity-100',
          'group-focus-within/hint:visible group-focus-within/hint:opacity-100',
        )}
      >
        {text}
      </span>
    </span>
  )
}

export function SettingsFieldLabel({
  htmlFor,
  label,
  required,
}: {
  htmlFor?: string
  label: string
  required?: boolean
}) {
  return (
    <label htmlFor={htmlFor} className={cn(heFieldLabel, 'block pt-2 sm:pt-2.5')}>
      {label}
      {required && <span className="normal-case text-destructive"> *</span>}
    </label>
  )
}

export function SettingsFieldRow({
  label,
  htmlFor,
  required,
  hint,
  labelExtra,
  children,
}: {
  label: string
  htmlFor?: string
  required?: boolean
  hint?: string
  labelExtra?: ReactNode
  children: ReactNode
}) {
  return (
    <div className={settingsFieldRowGrid}>
      <div className="space-y-2">
        <SettingsFieldLabel htmlFor={htmlFor} label={label} required={required} />
        {labelExtra}
      </div>
      <div className="flex min-w-0 items-start gap-1.5">
        <div className="min-w-0 flex-1">{children}</div>
        {hint && <FieldHintPopover content={hint} ariaLabel={`${label} 说明`} />}
      </div>
    </div>
  )
}

export function SettingsToggle({
  id,
  checked,
  disabled,
  ariaLabel,
  onCheckedChange,
}: {
  id: string
  checked: boolean
  disabled?: boolean
  ariaLabel: string
  onCheckedChange?: (checked: boolean) => void
}) {
  return (
    <button
      id={id}
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={ariaLabel}
      disabled={disabled}
      onClick={() => {
        if (!disabled) {
          onCheckedChange?.(!checked)
        }
      }}
      className={cn(
        'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full pt-0.5',
        heSpring,
        checked ? 'bg-primary' : 'bg-black/10',
        disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
      )}
    >
      <span
        aria-hidden
        className={cn(
          'block size-5 rounded-full bg-white shadow-sm ring-1 ring-black/[0.06]',
          heSpring,
          checked ? 'translate-x-[1.375rem]' : 'translate-x-0.5',
        )}
      />
    </button>
  )
}

export function SettingsSwitch({
  id,
  checked,
  disabled,
  onCheckedChange,
  label,
  hint,
}: {
  id: string
  checked: boolean
  disabled?: boolean
  onCheckedChange?: (checked: boolean) => void
  label: string
  hint?: string
}) {
  return (
    <SettingsFieldRow label={label} hint={hint}>
      <SettingsToggle
        id={id}
        checked={checked}
        disabled={disabled}
        ariaLabel={label}
        onCheckedChange={onCheckedChange}
      />
    </SettingsFieldRow>
  )
}
