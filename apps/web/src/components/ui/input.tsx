import * as React from 'react'

import { FieldHintPopover } from '@/components/settings/field-hint-popover'
import {
  heInputControlTypography,
  heInputEditable,
  heInputReadonly,
  heSpring,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  /** 控件右侧 ? 提示（与 SettingsFieldRow 的 hint 二选一，避免重复） */
  hint?: string
  hintAriaLabel?: string
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, hint, hintAriaLabel, readOnly, disabled, ...props }, ref) => {
    const locked = Boolean(readOnly || disabled)
    const hintText = hint?.trim()

    const input = (
      <input
        type={type}
        readOnly={readOnly}
        disabled={disabled}
        className={cn(
          locked ? heInputReadonly : heInputEditable,
          heInputControlTypography,
          locked && 'text-text-tertiary',
          type === 'number' &&
            '[appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none',
          heSpring,
          className,
        )}
        ref={ref}
        {...props}
      />
    )

    if (!hintText) return input

    return (
      <div className="flex min-w-0 items-center gap-1.5">
        <div className="min-w-0 flex-1">{input}</div>
        <FieldHintPopover content={hintText} ariaLabel={hintAriaLabel} />
      </div>
    )
  },
)
Input.displayName = 'Input'

export { Input }
