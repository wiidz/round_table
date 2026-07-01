import type { ReactNode } from 'react'

import { heScrollbar } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export const pageSideColumnClass = 'w-[16.5rem] shrink-0'

export type PageSidebarBreakpoint = 'xl' | '96rem'

/** compact：固定 16.5rem 侧栏；gutter：侧栏填满 main 两侧留白（保留 grid gap） */
export type PageSideColumnWidth = 'compact' | 'gutter'

function sidebarVisibleClass(from: PageSidebarBreakpoint): string {
  return from === '96rem' ? 'hidden min-[96rem]:block' : 'hidden xl:block'
}

function wideGridClass(from: PageSidebarBreakpoint, sideWidth: PageSideColumnWidth): string {
  if (sideWidth === 'gutter') {
    return from === '96rem'
      ? 'min-[96rem]:grid min-[96rem]:h-full min-[96rem]:w-full min-[96rem]:max-w-none min-[96rem]:items-stretch min-[96rem]:grid-cols-[minmax(12rem,1fr)_minmax(0,72rem)_minmax(12rem,1fr)]'
      : 'xl:grid xl:h-full xl:w-full xl:max-w-none xl:items-stretch xl:grid-cols-[minmax(12rem,1fr)_minmax(0,72rem)_minmax(12rem,1fr)]'
  }

  return from === '96rem'
    ? 'min-[96rem]:grid min-[96rem]:mx-auto min-[96rem]:w-fit min-[96rem]:max-w-full min-[96rem]:max-w-none min-[96rem]:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
    : 'xl:grid xl:mx-auto xl:w-fit xl:max-w-full xl:max-w-none xl:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
}

interface PageThreeColumnLayoutProps {
  left?: ReactNode
  right?: ReactNode
  children: ReactNode
  /** 宽屏下显示左右栏的断点；窄屏由页面在 main 内自行堆叠 */
  sidebarFrom?: PageSidebarBreakpoint
  /** 侧栏宽度策略；圆桌三栏请用 gutter */
  sideColumnWidth?: PageSideColumnWidth
  className?: string
}

/**
 * 三栏布局（对称 grid，main 相对页面居中）：
 * - 仅 right：占位 | main | right
 * - left + right：left | main | right
 * main 列宽 max 72rem。页头请放在外层 PageLayout，不要传入本组件。
 */
export function PageThreeColumnLayout({
  left,
  right,
  children,
  sidebarFrom = '96rem',
  sideColumnWidth = 'compact',
  className,
}: PageThreeColumnLayoutProps) {
  const sideVisible = sidebarVisibleClass(sidebarFrom)
  const gutterSide = sideColumnWidth === 'gutter'
  const hasLeft = Boolean(left)
  const hasRight = Boolean(right)
  const wideSide = hasRight && wideGridClass(sidebarFrom, sideColumnWidth)

  const containerClass = cn(
    'mx-auto w-full max-w-6xl gap-5',
    wideSide,
    gutterSide && hasRight && 'min-[96rem]:max-w-none xl:max-w-none',
  )

  const asideScrollClass = cn(
    'max-h-[calc(100vh-7rem)] overflow-y-auto',
    heScrollbar,
  )

  const asideColumnClass = gutterSide ? 'min-w-0 h-full' : pageSideColumnClass
  const asideStickyClass = gutterSide ? undefined : 'lg:sticky lg:top-20'

  const asideInnerClass = gutterSide
    ? 'h-full min-h-0'
    : cn('space-y-5', asideScrollClass)

  const mainClass = cn(
    'min-w-0 w-full',
    gutterSide && hasRight && 'h-full min-h-0',
  )

  const placeholderClass = gutterSide
    ? cn('min-w-0', sideVisible)
    : cn(pageSideColumnClass, sideVisible)

  return (
    <div className={cn(containerClass, className)}>
      {!hasLeft && hasRight && <div className={placeholderClass} aria-hidden />}

      {left && (
        <aside className={cn(asideColumnClass, asideStickyClass, sideVisible)}>
          <div className={asideInnerClass}>{left}</div>
        </aside>
      )}

      <main className={mainClass}>{children}</main>

      {right && (
        <aside className={cn(asideColumnClass, asideStickyClass, sideVisible)}>
          <div className={asideInnerClass}>{right}</div>
        </aside>
      )}
    </div>
  )
}
