import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { ChevronDown, Search, X } from 'lucide-react'

import { Input } from '@/components/ui/input'
import { useI18n } from '@/hooks/use-i18n'
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
  /** 多选时不在输入框上方展示标签 chips（由外部自定义已选展示） */
  hideSelectedChips?: boolean
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

const MENU_MAX_HEIGHT = 224 // max-h-56
const MENU_GAP = 4
const VIEWPORT_PADDING = 8

type MenuStyle = {
  left: number
  width: number
  top?: number
  bottom?: number
  maxHeight: number
}

function computeMenuStyle(trigger: DOMRect): MenuStyle {
  const spaceBelow = window.innerHeight - trigger.bottom - MENU_GAP - VIEWPORT_PADDING
  const spaceAbove = trigger.top - MENU_GAP - VIEWPORT_PADDING
  const openUp = spaceBelow < MENU_MAX_HEIGHT && spaceAbove > spaceBelow
  const maxHeight = Math.min(MENU_MAX_HEIGHT, Math.max(0, openUp ? spaceAbove : spaceBelow))

  if (openUp) {
    return {
      left: trigger.left,
      width: trigger.width,
      bottom: window.innerHeight - trigger.top + MENU_GAP,
      maxHeight,
    }
  }

  return {
    left: trigger.left,
    width: trigger.width,
    top: trigger.bottom + MENU_GAP,
    maxHeight,
  }
}

export function SearchableSelect(props: SearchableSelectProps) {
  const { t } = useI18n()
  const autoId = useId()
  const fieldId = props.id ?? autoId
  const listboxId = `${fieldId}-listbox`
  const rootRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLDivElement>(null)
  const listboxRef = useRef<HTMLUListElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [menuStyle, setMenuStyle] = useState<MenuStyle>()

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

    function updateMenuPosition() {
      const el = triggerRef.current
      if (!el) return
      setMenuStyle(computeMenuStyle(el.getBoundingClientRect()))
    }

    updateMenuPosition()
    window.addEventListener('scroll', updateMenuPosition, true)
    window.addEventListener('resize', updateMenuPosition)

    function onPointerDown(e: MouseEvent) {
      const target = e.target as Node
      if (rootRef.current?.contains(target)) return
      if (listboxRef.current?.contains(target)) return
      setOpen(false)
    }
    document.addEventListener('mousedown', onPointerDown)

    return () => {
      window.removeEventListener('scroll', updateMenuPosition, true)
      window.removeEventListener('resize', updateMenuPosition)
      document.removeEventListener('mousedown', onPointerDown)
    }
  }, [open, filtered.length])

  useEffect(() => {
    if (!open) return
    const listbox = listboxRef.current
    if (!listbox) return

    const el = listbox

    function onWheel(e: WheelEvent) {
      e.preventDefault()
      e.stopPropagation()
      el.scrollTop += e.deltaY
    }

    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
  }, [open, menuStyle, filtered.length])

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

  const searchPlaceholder = props.searchPlaceholder ?? t('settings.searchableSelect.search')
  const defaultPlaceholder = props.multiple
    ? t('settings.searchableSelect.placeholderExperts')
    : t('settings.searchableSelect.placeholderDefault')

  return (
    <div ref={rootRef} className="relative space-y-2">
      {props.multiple && !props.hideSelectedChips && selectedOptions.length > 0 && (
        <ul className="flex flex-wrap gap-1.5">
          {selectedOptions.map((opt) => (
            <li key={opt.value}>
              <span className="inline-flex max-w-full items-center gap-1 rounded-md bg-brand/8 py-0.5 pl-2 pr-1 text-[12px] text-text-primary ring-1 ring-inset ring-brand/15">
                <span className="truncate">{opt.label}</span>
                <button
                  type="button"
                  className="rounded p-0.5 text-text-tertiary hover:bg-black/[0.06] hover:text-text-primary"
                  aria-label={t('settings.searchableSelect.remove', { label: opt.label })}
                  onClick={() => removeValue(opt.value)}
                >
                  <X className="size-3" />
                </button>
              </span>
            </li>
          ))}
        </ul>
      )}

      <div ref={triggerRef} className="relative">
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
          placeholder={open ? searchPlaceholder : props.placeholder ?? defaultPlaceholder}
          className="pl-8 pr-8 text-sm"
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
          aria-label={
            open
              ? t('settings.searchableSelect.collapse')
              : t('settings.searchableSelect.expand')
          }
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

      {open &&
        menuStyle &&
        createPortal(
          <ul
            ref={listboxRef}
            id={listboxId}
            role="listbox"
            aria-multiselectable={props.multiple || undefined}
            style={{
              position: 'fixed',
              left: menuStyle.left,
              width: menuStyle.width,
              maxHeight: menuStyle.maxHeight,
              ...(menuStyle.top != null
                ? { top: menuStyle.top }
                : { bottom: menuStyle.bottom }),
            }}
            className="z-[100] overflow-y-auto overscroll-contain rounded-xs border border-black/[0.08] bg-surface py-1 shadow-lg"
          >
            {filtered.length === 0 ? (
              <li className="px-3 py-2 text-[13px] text-text-tertiary">
                {t('settings.searchableSelect.noResults')}
              </li>
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
          </ul>,
          document.body,
        )}
    </div>
  )
}
