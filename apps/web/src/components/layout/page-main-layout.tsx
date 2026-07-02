import type { ReactNode } from 'react'
import { Outlet } from 'react-router-dom'

import {
  PageThreeColumnLayout,
  type PageSidebarBreakpoint,
  type PageSideColumnWidth,
} from '@/components/layout/page-three-column-layout'
import { cn } from '@/lib/utils'

/** 路由级页面容器：全宽 Outlet，不加 max-w-6xl */
export function PageMainLayout() {
  return <Outlet />
}

interface PageLayoutProps {
  /** 页头：仅在此处约束 max-w-6xl，不参与三栏 grid */
  header?: ReactNode
  left?: ReactNode
  right?: ReactNode
  children: ReactNode
  sidebarFrom?: PageSidebarBreakpoint
  sideColumnWidth?: PageSideColumnWidth
  /** main 纵向撑满（聊天等全高页） */
  fillHeight?: boolean
  className?: string
  bodyClassName?: string
}

/**
 * 全站页面壳：header（单栏 max-w-6xl）+ PageThreeColumnLayout（主内容 / 侧栏）。
 * 所有页面应使用本组件，而非直接向 PageThreeColumnLayout 传 header。
 */
export function PageLayout({
  header,
  left,
  right,
  children,
  sidebarFrom,
  sideColumnWidth,
  fillHeight,
  className,
  bodyClassName,
}: PageLayoutProps) {
  return (
    <div className={cn('flex flex-col gap-6', className)}>
      {header ? <div className="mx-auto w-full max-w-6xl shrink-0">{header}</div> : null}
      <PageThreeColumnLayout
        left={left}
        right={right}
        sidebarFrom={sidebarFrom}
        sideColumnWidth={sideColumnWidth}
        fillHeight={fillHeight}
        className={bodyClassName}
      >
        {children}
      </PageThreeColumnLayout>
    </div>
  )
}
