import type { ReactNode } from 'react'

import { heScrollbar } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

export const pageSideColumnClass = 'w-[16.5rem] shrink-0'

type PageSidebarBreakpoint = 'xl' | '96rem'

function sidebarVisibleClass(from: PageSidebarBreakpoint): string {
  return from === '96rem' ? 'hidden min-[96rem]:block' : 'hidden xl:block'
}

function wideGridClass(from: PageSidebarBreakpoint): string {
  return from === '96rem'
    ? 'min-[96rem]:grid min-[96rem]:mx-auto min-[96rem]:w-fit min-[96rem]:max-w-full min-[96rem]:max-w-none min-[96rem]:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
    : 'xl:grid xl:mx-auto xl:w-fit xl:max-w-full xl:max-w-none xl:grid-cols-[16.5rem_minmax(0,72rem)_16.5rem]'
}

interface PageThreeColumnLayoutProps {
  header?: ReactNode
  left?: ReactNode
  right?: ReactNode
  children: ReactNode
  /** 宽屏下显示左右栏的断点；窄屏由页面在 main 内自行堆叠 */
  sidebarFrom?: PageSidebarBreakpoint
  className?: string
}

/**
 * 三栏布局（对称 grid，main 始终相对页面居中）：
 * - 仅 right：占位 | main | right（左侧等宽占位平衡右侧栏）
 * - left + right：left | main | right
 * main 列宽 max 72rem；header 通过内层 max-w-6xl 与 main 左缘对齐。
 */
export function PageThreeColumnLayout({
  header,
  left,
  right,
  children,
  sidebarFrom = '96rem',
  className,
}: PageThreeColumnLayoutProps) {
  const sideVisible = sidebarVisibleClass(sidebarFrom)
  const hasLeft = Boolean(left)
  const hasRight = Boolean(right)

  const containerClass = cn(
    'mx-auto w-full max-w-6xl gap-5',
    hasRight && wideGridClass(sidebarFrom),
  )

  const asideScrollClass = cn(
    'max-h-[calc(100vh-7rem)] overflow-y-auto',
    heScrollbar,
  )

  return (
    <div className={cn('flex flex-col gap-6', className)}>
      <div className={containerClass}>
        {header && (
          <div className="col-span-full min-w-0">
            <div className="mx-auto w-full max-w-6xl">{header}</div>
          </div>
        )}

        {!hasLeft && hasRight && (
          <div className={cn(pageSideColumnClass, sideVisible)} aria-hidden />
        )}

        {left && (
          <aside className={cn(pageSideColumnClass, 'lg:sticky lg:top-20', sideVisible)}>
            <div className={cn('space-y-5', asideScrollClass)}>{left}</div>
          </aside>
        )}

        <main className="min-w-0 w-full">{children}</main>

        {right && (
          <aside className={cn(pageSideColumnClass, 'lg:sticky lg:top-20', sideVisible)}>
            <div className={asideScrollClass}>{right}</div>
          </aside>
        )}
      </div>
    </div>
  )
}
