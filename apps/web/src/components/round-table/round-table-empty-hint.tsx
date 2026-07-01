import { Link } from 'react-router-dom'

import { useI18n } from '@/hooks/use-i18n'
import { cn } from '@/lib/utils'

interface RoundTableEmptyHintProps {
  loading: boolean
  rosterFromApi: boolean
  rosterTotal: number
  seatedExpertCount: number
  className?: string
}

export function RoundTableEmptyHint({
  loading,
  rosterFromApi,
  rosterTotal,
  seatedExpertCount,
  className,
}: RoundTableEmptyHintProps) {
  const { t } = useI18n()

  if (seatedExpertCount > 0) return null
  if (rosterFromApi && rosterTotal > 0) return null

  let message: string
  if (loading) {
    message = t('roundTable.empty.loading')
  } else if (!rosterFromApi) {
    message = t('roundTable.empty.noRoster')
  } else {
    message = t('roundTable.empty.rosterEmpty')
  }

  return (
    <div
      className={cn(
        'pointer-events-none absolute inset-x-4 bottom-4 z-20 rounded-xl bg-surface/95 px-4 py-3 text-center shadow-sm ring-1 ring-black/[0.06] backdrop-blur-sm',
        className,
      )}
    >
      <p className="text-[12px] leading-relaxed text-text-secondary">{message}</p>
      {!loading && rosterFromApi && (
        <p className="pointer-events-auto mt-2 text-[11px]">
          <Link to="/participants" className="text-brand hover:underline">
            {t('roundTable.empty.goParticipants')}
          </Link>
          <span className="text-text-tertiary">{t('roundTable.empty.discordHint')}</span>
        </p>
      )}
    </div>
  )
}
