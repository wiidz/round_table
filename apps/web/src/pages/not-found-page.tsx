import { Link } from 'react-router-dom'

import { PageLayout } from '@/components/layout/page-main-layout'
import { Button } from '@/components/ui/button'

export function NotFoundPage() {
  return (
    <PageLayout>
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
        <p className="text-sm text-text-tertiary">404</p>
        <h1 className="text-2xl font-semibold">页面不存在</h1>
        <Button asChild>
          <Link to="/">返回首页</Link>
        </Button>
      </div>
    </PageLayout>
  )
}
