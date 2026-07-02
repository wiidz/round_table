import type { ReactNode } from 'react'

import { heScrollbar } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export const pageSideColumnClass = 'w-[16.5rem] shrink-0'
export const pageSideColumnWideClass = 'w-[18rem] shrink-0'

/** 侧栏吸顶：grid 列需配合 items-start，避免 stretch 吃掉 sticky */
export const pageStickyAsideClass = 'sticky top-20 self-start'

export const pageStickyAsideScrollClass =
  'max-h-[calc(100vh-7rem)] overflow-y-auto overscroll-contain'

export type PageSidebarBreakpoint = 'xl' | '96rem'

/** compact：16.5rem；wide：18rem（会议详情配置侧栏）；gutter：侧栏填满 main 两侧留白 */
export type PageSideColumnWidth = 'compact' | 'wide' | 'gutter'

function sidebarVisibleClass(from: PageSidebarBreakpoint): string {
  return from === '96rem' ? 'hidden min-[96rem]:block' : 'hidden xl:block'
}

function wideGridClass(from: PageSidebarBreakpoint, sideWidth: PageSideColumnWidth): string {
  if (sideWidth === 'gutter') {
    return from === '96rem'
      ? 'min-[96rem]:grid min-[96rem]:h-full min-[96rem]:w-full min-[96rem]:max-w-none min-[96rem]:items-stretch min-[96rem]:grid-cols-[minmax(12rem,1fr)_minmax(0,72rem)_minmax(12rem,1fr)]'
      : 'xl:grid xl:h-full xl:w-full xl:max-w-none xl:items-stretch xl:grid-cols-[minmax(12rem,1fr)_minmax(0,72rem)_minmax(12rem,1fr)]'
  }

  if (sideWidth === 'wide') {
    return from === '96rem'
      ? 'min-[96rem]:grid min-[96rem]:mx-auto min-[96rem]:w-fit min-[96rem]:max-w-full min-[96rem]:max-w-none min-[96rem]:items-start min-[96rem]:grid-cols-[18rem_minmax(0,72rem)_18rem]'
      : 'xl:grid xl:mx-auto xl:w-fit xl:max-w-full xl:max-w-none xl:items-start xl:grid-cols-[18rem_minmax(0,72rem)_18rem]'
  }

  return from === '96rem'
    ? 'min-[96rem]:grid min-[96rem]:mx-auto min-[96rem]:w-fit min-[96rem]:max-w-full min-[96rem]:max-w-none min-[96rem]:items-start min-[96rem]:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
    : 'xl:grid xl:mx-auto xl:w-fit xl:max-w-full xl:max-w-none xl:items-start xl:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
}

interface PageThreeColumnLayoutProps {
  left?: ReactNode
  right?: ReactNode
  children: ReactNode
  /** 宽屏下显示左右栏的断点；窄屏由页面在 main 内自行堆叠 */
  sidebarFrom?: PageSidebarBreakpoint
  /** 侧栏宽度策略；圆桌三栏请用 gutter */
  sideColumnWidth?: PageSideColumnWidth
  /** main 纵向撑满父级 flex 容器（如聊天页全高壳） */
  fillHeight?: boolean
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
  fillHeight = false,
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
    fillHeight && 'flex min-h-0 flex-1 flex-col',
  )

  const asideColumnClass = gutterSide
    ? 'min-w-0 h-full'
    : sideColumnWidth === 'wide'
      ? pageSideColumnWideClass
      : pageSideColumnClass

  const asideInnerClass = gutterSide ? 'h-full min-h-0' : 'space-y-5'

  const asideStickyClass = gutterSide
    ? undefined
    : cn(pageStickyAsideClass, pageStickyAsideScrollClass, heScrollbar)

  const mainClass = cn(
    'min-w-0 w-full',
    (gutterSide && hasRight) || fillHeight ? 'flex min-h-0 flex-1 flex-col' : undefined,
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
