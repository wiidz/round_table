import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { ChevronDown, Search, X } from 'lucide-react'

import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

export type SearchableSelectOption = {
  value: string
  label: string
  hint?: string
  disabled?: boolean
}

type SearchableSelectProps = {
  id?: string
  options: SearchableSelectOption[]
  placeholder?: string
  searchPlaceholder?: string
  emptyOption?: SearchableSelectOption
  disabled?: boolean
} & (
  | {
      multiple?: false
      value: string
      onChange: (value: string) => void
    }
  | {
      multiple: true
      value: string[]
      onChange: (value: string[]) => void
    }
)

export function SearchableSelect(props: SearchableSelectProps) {
  const autoId = useId()
  const fieldId = props.id ?? autoId
  const listboxId = `${fieldId}-listbox`
  const rootRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')

  const selectedSet = useMemo(() => {
    if (props.multiple) {
      return new Set(props.value)
    }
    return props.value ? new Set([props.value]) : new Set<string>()
  }, [props.multiple, props.value])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    const list = props.options.filter((opt) => {
      if (!q) return true
      const hay = `${opt.label} ${opt.value} ${opt.hint ?? ''}`.toLowerCase()
      return hay.includes(q)
    })
    if (!props.multiple && props.emptyOption && !q) {
      return [props.emptyOption, ...list]
    }
    return list
  }, [props.options, props.emptyOption, props.multiple, query])

  useEffect(() => {
    if (!open) return
    function onPointerDown(e: MouseEvent) {
      if (!rootRef.current?.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onPointerDown)
    return () => document.removeEventListener('mousedown', onPointerDown)
  }, [open])

  function optionLabel(opt: SearchableSelectOption) {
    if (opt.disabled && opt.hint) return `${opt.label}（${opt.hint}）`
    return opt.label
  }

  function selectOption(opt: SearchableSelectOption) {
    if (opt.disabled) return
    if (props.multiple) {
      const next = new Set(props.value)
      if (next.has(opt.value)) {
        next.delete(opt.value)
      } else {
        next.add(opt.value)
      }
      props.onChange(Array.from(next))
      setQuery('')
      inputRef.current?.focus()
      return
    }
    props.onChange(opt.value)
    setQuery('')
    setOpen(false)
  }

  function removeValue(value: string) {
    if (!props.multiple) return
    props.onChange(props.value.filter((v) => v !== value))
  }

  const selectedOptions = props.multiple
    ? props.value
        .map((v) => props.options.find((o) => o.value === v))
        .filter((o): o is SearchableSelectOption => Boolean(o))
    : []

  const singleSelected =
    !props.multiple && props.value
      ? props.options.find((o) => o.value === props.value) ??
        (props.emptyOption?.value === props.value ? props.emptyOption : undefined)
      : undefined

  const inputDisplayValue = open
    ? query
    : props.multiple
      ? query
      : singleSelected?.label ?? ''

  return (
    <div ref={rootRef} className="relative space-y-2">
      {props.multiple && selectedOptions.length > 0 && (
        <ul className="flex flex-wrap gap-1.5">
          {selectedOptions.map((opt) => (
            <li key={opt.value}>
              <span className="inline-flex max-w-full items-center gap-1 rounded-md bg-brand/8 py-0.5 pl-2 pr-1 text-[12px] text-text-primary ring-1 ring-inset ring-brand/15">
                <span className="truncate">{opt.label}</span>
                <button
                  type="button"
                  className="rounded p-0.5 text-text-tertiary hover:bg-black/[0.06] hover:text-text-primary"
                  aria-label={`移除 ${opt.label}`}
                  onClick={() => removeValue(opt.value)}
                >
                  <X className="size-3" />
                </button>
              </span>
            </li>
          ))}
        </ul>
      )}

      <div className="relative">
        <Search
          className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-tertiary"
          aria-hidden
        />
        <Input
          ref={inputRef}
          id={fieldId}
          role="combobox"
          aria-expanded={open}
          aria-controls={listboxId}
          aria-autocomplete="list"
          disabled={props.disabled}
          value={inputDisplayValue}
          placeholder={
            open
              ? props.searchPlaceholder ?? '搜索…'
              : props.multiple
                ? props.placeholder ?? '搜索并选择专家'
                : props.placeholder ?? '选择…'
          }
          className="!rounded-xs pl-8 pr-8 text-sm"
          onFocus={() => {
            setOpen(true)
            if (!props.multiple) setQuery('')
          }}
          onChange={(e) => {
            setQuery(e.target.value)
            setOpen(true)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Escape') {
              setOpen(false)
              setQuery('')
            }
          }}
        />
        <button
          type="button"
          tabIndex={-1}
          aria-label={open ? '收起选项' : '展开选项'}
          disabled={props.disabled}
          className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-1 text-text-tertiary hover:bg-black/[0.04] hover:text-text-primary"
          onClick={() => {
            setOpen((v) => !v)
            inputRef.current?.focus()
          }}
        >
          <ChevronDown className={cn('size-4 transition-transform', open && 'rotate-180')} />
        </button>
      </div>

      {open && (
        <ul
          id={listboxId}
          role="listbox"
          aria-multiselectable={props.multiple || undefined}
          className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded-xs border border-black/[0.08] bg-surface py-1 shadow-lg"
        >
          {filtered.length === 0 ? (
            <li className="px-3 py-2 text-[13px] text-text-tertiary">无匹配项</li>
          ) : (
            filtered.map((opt) => {
              const selected = selectedSet.has(opt.value)
              return (
                <li key={opt.value || '__empty'} role="option" aria-selected={selected}>
                  <button
                    type="button"
                    disabled={opt.disabled}
                    className={cn(
                      'flex w-full flex-col items-start gap-0.5 px-3 py-2 text-left text-sm',
                      opt.disabled
                        ? 'cursor-not-allowed text-text-tertiary'
                        : 'text-text-primary hover:bg-black/[0.04]',
                      selected && !opt.disabled && 'bg-brand/6 font-medium text-brand',
                    )}
                    onClick={() => selectOption(opt)}
                  >
                    <span>{optionLabel(opt)}</span>
                    {opt.hint && !opt.disabled && (
                      <span className="font-mono text-[10px] text-text-tertiary">{opt.hint}</span>
                    )}
                  </button>
                </li>
              )
            })
          )}
        </ul>
      )}
    </div>
  )
}
