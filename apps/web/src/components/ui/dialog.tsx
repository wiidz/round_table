import * as React from 'react'
import { createPortal } from 'react-dom'
import { useEffect } from 'react'

import {
  heDialogContent,
  heDialogOverlay,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

type DialogProps = {
  open: boolean
  onClose?: () => void
  children: React.ReactNode
  closeOnOverlayClick?: boolean
  closeOnEscape?: boolean
}

export function Dialog({
  open,
  onClose,
  children,
  closeOnOverlayClick = true,
  closeOnEscape = true,
}: DialogProps) {
  useEffect(() => {
    if (!open || !closeOnEscape || !onClose) return
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open, closeOnEscape, onClose])

  if (!open) return null

  return createPortal(
    <div
      className={heDialogOverlay}
      role="presentation"
      onClick={closeOnOverlayClick && onClose ? onClose : undefined}
    >
      {children}
    </div>,
    document.body,
  )
}

type DialogContentProps = React.HTMLAttributes<HTMLDivElement> & {
  size?: 'sm' | 'md' | 'lg'
  padded?: boolean
}

const dialogSizeClass: Record<NonNullable<DialogContentProps['size']>, string> = {
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-[56rem]',
}

export const DialogContent = React.forwardRef<HTMLDivElement, DialogContentProps>(
  ({ className, size = 'sm', padded = true, onClick, ...props }, ref) => (
    <div
      ref={ref}
      role="dialog"
      aria-modal="true"
      className={cn(heDialogContent, dialogSizeClass[size], padded && 'p-6', className)}
      onClick={(event) => {
        event.stopPropagation()
        onClick?.(event)
      }}
      {...props}
    />
  ),
)
DialogContent.displayName = 'DialogContent'

export function DialogHeader({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('space-y-1 text-left', className)} {...props} />
}

export function DialogFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end', className)}
      {...props}
    />
  )
}

export function DialogTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h2 className={cn('text-base font-semibold text-text-primary', className)} {...props} />
}

export function DialogDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p className={cn('text-[13px] leading-relaxed text-text-secondary', className)} {...props} />
  )
}
