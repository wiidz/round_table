import * as React from 'react'

import { cn } from '@/lib/utils'
import { heFieldSurface, heSpring } from '@/lib/highend-styles'

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => (
    <textarea
      className={cn(
        heFieldSurface,
        'min-h-[6rem] w-full resize-y bg-surface px-3 py-2 text-sm leading-relaxed text-text-primary',
        'placeholder:text-text-tertiary',
        'focus-visible:outline-none',
        'disabled:cursor-not-allowed disabled:opacity-50',
        heSpring,
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
)
Textarea.displayName = 'Textarea'

export { Textarea }
