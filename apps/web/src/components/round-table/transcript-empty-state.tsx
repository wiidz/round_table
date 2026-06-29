import { cn } from '@/lib/utils'

type TranscriptEmptyVariant = 'list' | 'detail'

interface TranscriptEmptyStateProps {
  variant: TranscriptEmptyVariant
  title: string
  description: string
  className?: string
}

function TranscriptListIllustration() {
  return (
    <svg
      viewBox="0 0 120 120"
      className="size-[4.5rem] text-ai/70"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden
    >
      <rect x="22" y="20" width="76" height="80" rx="14" fill="currentColor" fillOpacity="0.1" />
      {[0, 1, 2].map((row) => {
        const y = 34 + row * 22
        return (
          <g key={row}>
            <circle cx="36" cy={y + 6} r="5" fill="currentColor" fillOpacity={0.28 - row * 0.05} />
            <rect x="48" y={y} width="28" height="4" rx="2" fill="currentColor" fillOpacity={0.32 - row * 0.06} />
            <rect x="48" y={y + 8} width="40" height="3" rx="1.5" fill="currentColor" fillOpacity={0.18 - row * 0.03} />
          </g>
        )
      })}
      <rect x="28" y="88" width="64" height="3" rx="1.5" fill="var(--brand-color)" fillOpacity="0.35" />
    </svg>
  )
}

function TranscriptDetailIllustration() {
  return (
    <svg
      viewBox="0 0 120 120"
      className="size-[4.5rem] text-ai/70"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden
    >
      <rect x="24" y="16" width="72" height="88" rx="12" fill="currentColor" fillOpacity="0.1" />
      <rect x="34" y="28" width="32" height="5" rx="2.5" fill="currentColor" fillOpacity="0.38" />
      <rect x="34" y="40" width="52" height="3" rx="1.5" fill="currentColor" fillOpacity="0.22" />
      <rect x="34" y="48" width="48" height="3" rx="1.5" fill="currentColor" fillOpacity="0.18" />
      <rect x="34" y="56" width="40" height="3" rx="1.5" fill="currentColor" fillOpacity="0.14" />
      <rect x="34" y="68" width="28" height="4" rx="2" fill="currentColor" fillOpacity="0.26" />
      <rect x="34" y="78" width="44" height="3" rx="1.5" fill="currentColor" fillOpacity="0.16" />
      <rect x="34" y="86" width="36" height="3" rx="1.5" fill="currentColor" fillOpacity="0.12" />
      <circle cx="84" cy="84" r="14" fill="var(--brand-color)" fillOpacity="0.18" />
      <circle cx="84" cy="84" r="9" stroke="var(--brand-color)" strokeWidth="2" strokeOpacity="0.75" />
      <path
        d="M88 88 L94 94"
        stroke="var(--brand-color)"
        strokeWidth="2"
        strokeLinecap="round"
        strokeOpacity="0.75"
      />
    </svg>
  )
}

/** Centered empty state with inline illustration for transcript side panels. */
export function TranscriptEmptyState({
  variant,
  title,
  description,
  className,
}: TranscriptEmptyStateProps) {
  return (
    <div
      className={cn(
        'flex min-h-0 flex-1 flex-col items-center justify-center px-6 py-10 text-center',
        className,
      )}
    >
      <div
        className={cn(
          'mb-5 flex size-[7.5rem] items-center justify-center rounded-[1.75rem] ring-1',
          variant === 'list'
            ? 'bg-ai-soft ring-ai/12'
            : 'bg-brand-soft ring-primary/15',
        )}
        aria-hidden
      >
        {variant === 'list' ? <TranscriptListIllustration /> : <TranscriptDetailIllustration />}
      </div>

      <h3 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">{title}</h3>
      <p className="mt-2 max-w-[15rem] text-[13px] leading-relaxed text-text-tertiary">{description}</p>
    </div>
  )
}

/** Compact panel header without left border accent. */
export function TranscriptPanelHeader({
  title,
  subtitle,
  className,
}: {
  title: string
  subtitle?: string
  className?: string
}) {
  return (
    <div className={cn('shrink-0 border-b border-black/[0.06] px-4 py-4 sm:px-5', className)}>
      <h2 className="text-[15px] font-semibold tracking-[-0.02em] text-text-primary">{title}</h2>
      {subtitle && <p className="mt-1 text-[12px] text-text-tertiary">{subtitle}</p>}
    </div>
  )
}
