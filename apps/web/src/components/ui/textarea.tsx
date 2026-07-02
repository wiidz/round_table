import * as React from 'react'

import { FieldHintPopover } from '@/components/settings/field-hint-popover'
import {
  heInputEditable,
  heInputReadonly,
  heSpring,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  hint?: string
  hintAriaLabel?: string
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, hint, hintAriaLabel, readOnly, disabled, ...props }, ref) => {
    const locked = Boolean(readOnly || disabled)
    const hintText = hint?.trim()

    const textarea = (
      <textarea
        readOnly={readOnly}
        disabled={disabled}
        className={cn(
          locked ? heInputReadonly : heInputEditable,
          'min-h-[6rem] w-full resize-y px-3 py-2 text-sm leading-relaxed',
          locked ? 'text-text-tertiary' : 'text-text-primary placeholder:text-text-tertiary',
          'focus-visible:outline-none',
          heSpring,
          className,
        )}
        ref={ref}
        {...props}
      />
    )

    if (!hintText) return textarea

    return (
      <div className="flex min-w-0 items-start gap-1.5">
        <div className="min-w-0 flex-1">{textarea}</div>
        <FieldHintPopover content={hintText} ariaLabel={hintAriaLabel} />
      </div>
    )
  },
)
Textarea.displayName = 'Textarea'

export { Textarea }
