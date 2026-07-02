import { forwardRef, type ComponentProps, type CSSProperties, type ReactNode } from 'react'

import {
  SIDE_TAB_WORKSPACE_WIDTH,
  sideTabWorkspaceAddButtonClass,
  sideTabWorkspaceButtonClass,
  sideTabWorkspaceListClass,
  sideTabWorkspacePanelClass,
  sideTabWorkspaceRowClass,
  type SideTabWorkspaceTone,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

type SideTabWorkspaceProps = {
  children: ReactNode
  className?: string
}

export function SideTabWorkspace({ children, className }: SideTabWorkspaceProps) {
  return <div className={cn(sideTabWorkspaceRowClass, className)}>{children}</div>
}

type SideTabWorkspaceNavProps = ComponentProps<'nav'> & {
  width?: string
}

export const SideTabWorkspaceNav = forwardRef<HTMLElement, SideTabWorkspaceNavProps>(
  function SideTabWorkspaceNav(
    { children, className, width = SIDE_TAB_WORKSPACE_WIDTH, style, ...props },
    ref,
  ) {
    return (
      <nav
        ref={ref}
        {...props}
        className={cn(
          sideTabWorkspaceListClass,
          'relative z-10 min-w-40 shrink-0 grow-0',
          className,
        )}
        style={{ width, minWidth: width, flexShrink: 0, ...style }}
      >
        {children}
      </nav>
    )
  },
)

type SideTabWorkspaceTabProps = ComponentProps<'button'> & {
  selected: boolean
  tone?: SideTabWorkspaceTone
}

export function SideTabWorkspaceTab({
  selected,
  tone = 'canvas',
  className,
  type = 'button',
  ...props
}: SideTabWorkspaceTabProps) {
  return (
    <button
      {...props}
      type={type}
      aria-selected={selected}
      className={sideTabWorkspaceButtonClass(selected, className, tone)}
    />
  )
}

type SideTabWorkspaceAddTabProps = ComponentProps<'button'>

export function SideTabWorkspaceAddTab({
  className,
  type = 'button',
  ...props
}: SideTabWorkspaceAddTabProps) {
  return (
    <button
      {...props}
      type={type}
      className={cn(sideTabWorkspaceAddButtonClass, className)}
    />
  )
}

type SideTabWorkspacePanelProps = {
  children: ReactNode
  className?: string
  style?: CSSProperties
  tone?: SideTabWorkspaceTone
}

export function SideTabWorkspacePanel({
  children,
  className,
  style,
  tone = 'canvas',
}: SideTabWorkspacePanelProps) {
  return (
    <div
      className={cn(sideTabWorkspacePanelClass(tone), 'relative z-0 min-w-0', className)}
      style={style}
    >
      {children}
    </div>
  )
}
