import { Link } from 'react-router-dom'

import { PageLayout } from '@/components/layout/page-main-layout'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'

export function NotFoundPage() {
  const { t } = useI18n()

  return (
    <PageLayout>
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
        <p className="text-sm text-text-tertiary">404</p>
        <h1 className="text-2xl font-semibold">{t('pages.notFound.title')}</h1>
        <Button asChild>
          <Link to="/">{t('pages.notFound.backHome')}</Link>
        </Button>
      </div>
    </PageLayout>
  )
}
