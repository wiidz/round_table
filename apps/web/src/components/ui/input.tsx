import * as React from 'react'

import { cn } from '@/lib/utils'
import { heFieldSurface, heSpring } from '@/lib/highend-styles'

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        heFieldSurface,
        'h-10 w-full bg-surface px-3 text-sm text-text-primary placeholder:text-text-tertiary',
        type === 'number' && '[appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none',
        heSpring,
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
)
Input.displayName = 'Input'

export { Input }
