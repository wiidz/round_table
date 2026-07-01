import { Outlet } from 'react-router-dom'

/** 常规单栏页面：仅约束中间内容区宽度 */
export function PageMainLayout() {
  return (
    <main className="mx-auto w-full max-w-6xl">
      <Outlet />
    </main>
  )
}
