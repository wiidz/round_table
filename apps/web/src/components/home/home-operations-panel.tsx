import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import { ArrowRight, ChevronRight, MessagesSquare, Server, Settings2 } from 'lucide-react'

import { SettingsNavLink } from '@/components/settings/settings-nav-link'
import { useI18n } from '@/hooks/use-i18n'
import type { Translator } from '@/lib/i18n/translate'
import {
  heFileBadge,
  hePanelShell,
  hePressable,
  heSpring,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { buildProcessRuntimeMetrics, type ProcessRuntimeMetric } from '@/lib/format-runtime'
import { SETTINGS_IM_DISCORD, settingsNavForDiscordBot } from '@/lib/settings-nav'
import { cn } from '@/lib/utils'

import type { DiscordBotState, DiscordTransportPhase } from '@/types/settings'
import type { ProcessSnapshot } from '@/types/runtime'

function ServiceStatusPill({
  tone,
  label,
  tooltip,
}: {
  tone: 'success' | 'warning' | 'neutral' | 'danger' | 'info'
  label: string
  tooltip?: string
}) {
  const pill = (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset',
        tone === 'success' && 'bg-success-soft text-success ring-success/20',
        tone === 'warning' && 'bg-warning-soft text-warning ring-warning/25',
        tone === 'danger' && 'bg-danger-soft text-danger ring-danger/20',
        tone === 'info' && 'bg-ai-soft text-ai ring-ai/15',
        tone === 'neutral' && 'bg-black/[0.04] text-text-tertiary ring-black/[0.06]',
        tooltip && 'cursor-help',
      )}
    >
      <span
        aria-hidden
        className={cn(
          'size-1.5 rounded-full',
          tone === 'success' && 'bg-success',
          tone === 'warning' && 'bg-warning',
          tone === 'danger' && 'bg-danger',
          tone === 'info' && 'bg-ai/70',
          tone === 'neutral' && 'bg-text-tertiary/40',
        )}
      />
      {label}
    </span>
  )

  if (!tooltip?.trim()) return pill

  return (
    <span
      className="group/pill relative inline-flex shrink-0 rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ai/30"
      tabIndex={0}
      aria-label={`${label}：${tooltip}`}
    >
      {pill}
      <span
        role="tooltip"
        className={cn(
          'pointer-events-none absolute bottom-[calc(100%+6px)] right-0 z-50 w-56',
          'rounded-xs bg-surface px-3 py-2.5 text-[13px] leading-relaxed text-text-secondary',
          'shadow-[var(--panel-shell-shadow)] ring-1 ring-[var(--panel-shell-ring)]',
          'invisible opacity-0 transition-opacity duration-150',
          'group-hover/pill:visible group-hover/pill:opacity-100',
          'group-focus-within/pill:visible group-focus-within/pill:opacity-100',
        )}
      >
        {tooltip}
      </span>
    </span>
  )
}

function isGatewayHost(bot: DiscordBotState): boolean {
  return bot.primary
}

function ServiceRoleTag({ variant, label }: { variant: 'main' | 'im'; label: string }) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-[10px] font-medium ring-1 ring-inset',
        variant === 'main' && 'bg-brand-soft text-brand ring-brand/25',
        variant === 'im' && 'bg-ai-soft text-ai ring-ai/20',
      )}
    >
      {label}
    </span>
  )
}

function RuntimeMetricGrid({ metrics }: { metrics: ProcessRuntimeMetric[] }) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {metrics.map((metric) => (
        <div
          key={metric.key}
          className="inline-flex min-w-[3.75rem] flex-col gap-0.5 rounded-lg bg-black/[0.025] px-2.5 py-1.5 ring-1 ring-inset ring-black/[0.05]"
        >
          <span className="text-[10px] font-medium tracking-wide text-text-tertiary">{metric.label}</span>
          <span className="font-mono text-[12px] font-semibold tabular-nums leading-none text-text-primary">
            {metric.value}
          </span>
        </div>
      ))}
    </div>
  )
}

function RuntimeMetricSkeleton() {
  return (
    <div className="flex flex-wrap gap-1.5">
      {Array.from({ length: 3 }, (_, i) => (
        <div
          key={i}
          className="h-[42px] w-[4.5rem] animate-pulse rounded-lg bg-black/[0.04]"
        />
      ))}
    </div>
  )
}

function ServiceProcessCard({
  variant,
  icon: Icon,
  title,
  roleLabel,
  description,
  metrics,
  loading,
  status,
  footer,
}: {
  variant: 'main' | 'im'
  icon: LucideIcon
  title: string
  roleLabel: string
  description: string
  metrics: ProcessRuntimeMetric[]
  loading: boolean
  status: ReactNode
  footer?: ReactNode
}) {
  return (
    <article
      className={cn(
        hePanelShell,
        'relative overflow-hidden',
        'before:pointer-events-none before:absolute before:inset-y-3 before:left-0 before:w-[3px] before:rounded-r-full',
        variant === 'main' && 'before:bg-primary/45',
        variant === 'im' && 'before:bg-ai/45',
      )}
    >
      <div
        aria-hidden
        className={cn(
          'pointer-events-none absolute inset-x-0 top-0 h-20 bg-gradient-to-b to-transparent',
          variant === 'main' && 'from-brand-soft/70',
          variant === 'im' && 'from-ai-soft/60',
        )}
      />
      <div className="relative space-y-3 px-4 py-3.5 pl-5">
        <div className="flex items-start gap-3">
          <span
            className={cn(
              'inline-flex size-9 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset',
              variant === 'main' && 'bg-brand-soft text-brand ring-brand/20',
              variant === 'im' && 'bg-ai-soft text-ai ring-ai/15',
            )}
          >
            <Icon className="size-4" strokeWidth={1.75} aria-hidden />
          </span>
          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-3">
              <p className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <span className="text-[13px] font-semibold tracking-[-0.01em] text-text-primary">
                  {title}
                </span>
                <ServiceRoleTag variant={variant} label={roleLabel} />
              </p>
              <div className="shrink-0">{status}</div>
            </div>
          </div>
        </div>

        {loading ? <RuntimeMetricSkeleton /> : metrics.length > 0 ? <RuntimeMetricGrid metrics={metrics} /> : null}

        <div className="border-t border-black/[0.05] pt-3">
          <p className="text-[12px] leading-relaxed text-text-tertiary">{description}</p>
          {footer}
        </div>
      </div>
    </article>
  )
}

function botConnectionState(
  bot: DiscordBotState,
  phase: DiscordTransportPhase,
  t: Translator,
): { tone: 'success' | 'warning' | 'neutral' | 'info'; label: string; tooltip?: string } {
  if (!bot.configured) {
    return {
      tone: 'neutral',
      label: t('home.bots.unconfigured'),
      tooltip: t('home.bots.unconfiguredHint'),
    }
  }
  if (phase === 'starting') {
    if (isGatewayHost(bot)) {
      return {
        tone: 'warning',
        label: t('home.bots.gatewayConnecting'),
        tooltip: t('home.bots.gatewayConnectingHint'),
      }
    }
    return {
      tone: 'warning',
      label: t('home.bots.waitingTransport'),
      tooltip: t('home.bots.waitingTransportHint'),
    }
  }
  if (phase !== 'ready') {
    return {
      tone: 'neutral',
      label: t('home.bots.unavailable'),
      tooltip: isGatewayHost(bot)
        ? t('home.bots.unavailableHostHint')
        : t('home.bots.unavailableParticipantHint'),
    }
  }
  if (isGatewayHost(bot)) {
    return {
      tone: 'success',
      label: t('home.bots.gatewayOnline'),
      tooltip: t('home.bots.gatewayOnlineHint'),
    }
  }
  return {
    tone: 'info',
    label: t('home.bots.restSend'),
    tooltip: t('home.bots.restSendHint'),
  }
}

function botLabel(bot: DiscordBotState): string {
  return (
    bot.display_name?.trim() ||
    bot.label?.trim() ||
    bot.discord_username?.trim() ||
    bot.id
  )
}

function transportServiceStatus(
  phase: DiscordTransportPhase,
  t: Translator,
): {
  tone: 'success' | 'warning' | 'neutral'
  label: string
} {
  if (phase === 'ready') {
    return { tone: 'success', label: t('home.transport.running') }
  }
  if (phase === 'starting') {
    return { tone: 'warning', label: t('home.transport.starting') }
  }
  return { tone: 'neutral', label: t('home.transport.stopped') }
}

interface HomeOperationsPanelProps {
  loading: boolean
  apiOnline: boolean
  apiError?: string | null
  transportPhase: DiscordTransportPhase
  transportPid?: number
  transportUnavailable?: boolean
  discordBots: DiscordBotState[]
  serverRuntime?: ProcessSnapshot
  discordRuntime?: ProcessSnapshot
}

export function HomeOperationsPanel({
  loading,
  apiOnline,
  apiError,
  transportPhase,
  transportPid,
  transportUnavailable,
  discordBots,
  serverRuntime,
  discordRuntime,
}: HomeOperationsPanelProps) {
  const { t } = useI18n()
  const configuredBots = discordBots.filter((b) => b.configured)
  const serverMetrics = buildProcessRuntimeMetrics(serverRuntime)
  const discordMetrics = buildProcessRuntimeMetrics(
    discordRuntime ?? (transportPid != null && transportPid > 0 ? { pid: transportPid } : undefined),
  )
  const transportStatus = transportServiceStatus(transportPhase, t)

  return (
    <section className="space-y-4">
      <h2 className={heSubsectionTitleNeutral}>{t('home.operationsTitle')}</h2>

      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]">
        <div className="space-y-3">
          <ServiceProcessCard
            variant="main"
            icon={Server}
            title={t('home.api.title')}
            roleLabel={t('home.api.role')}
            description={t('home.api.description')}
            metrics={serverMetrics}
            loading={loading}
            status={
              loading ? (
                <div className="h-6 w-16 animate-pulse rounded-full bg-black/[0.04]" />
              ) : apiOnline ? (
                <ServiceStatusPill tone="success" label={t('home.api.running')} />
              ) : (
                <ServiceStatusPill tone="danger" label={t('home.api.connectionError')} />
              )
            }
            footer={
              !loading && apiError ? (
                <p className="mt-2 text-[12px] text-text-tertiary">{apiError}</p>
              ) : undefined
            }
          />

          <ServiceProcessCard
            variant="im"
            icon={MessagesSquare}
            title={t('home.transport.title')}
            roleLabel={t('home.transport.role')}
            description={t('home.transport.description')}
            metrics={discordMetrics}
            loading={loading}
            status={
              loading ? (
                <div className="h-6 w-16 animate-pulse rounded-full bg-black/[0.04]" />
              ) : transportUnavailable ? (
                <ServiceStatusPill tone="neutral" label={t('home.transport.supervisionUnavailable')} />
              ) : (
                <SettingsNavLink
                  to="/settings"
                  nav={SETTINGS_IM_DISCORD}
                  aria-label={t('home.transport.goToSettings', { status: transportStatus.label })}
                  className={cn(
                    hePressable,
                    heSpring,
                    'inline-flex rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ai/30',
                  )}
                >
                  <ServiceStatusPill tone={transportStatus.tone} label={transportStatus.label} />
                </SettingsNavLink>
              )
            }
            footer={
              transportUnavailable ? (
                <p className="mt-2 text-[12px] text-text-tertiary">
                  {t('home.transport.supervisionHint')}
                </p>
              ) : undefined
            }
          />
        </div>

        <article className={cn(hePanelShell, 'flex flex-col p-5')}>
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="text-[13px] font-semibold text-text-primary">{t('home.bots.title')}</p>
              <p className="text-[12px] text-text-tertiary">{t('home.bots.subtitle')}</p>
            </div>
            <SettingsNavLink
              to="/settings"
              nav={SETTINGS_IM_DISCORD}
              className={cn(
                heFileBadge,
                hePressable,
                heSpring,
                'inline-flex items-center gap-1 px-2.5 py-1 text-[12px] hover:bg-brand-soft/70 hover:text-brand hover:ring-primary/20',
              )}
            >
              <Settings2 className="size-3.5 opacity-70" aria-hidden />
              {t('home.bots.manage')}
              <ChevronRight className="size-3 opacity-50" aria-hidden />
            </SettingsNavLink>
          </div>

          {loading && (
            <div className="space-y-2">
              {Array.from({ length: 3 }, (_, i) => (
                <div key={i} className="h-12 animate-pulse rounded-xl bg-black/[0.03]" />
              ))}
            </div>
          )}

          {!loading && discordBots.length === 0 && (
            <div className="flex flex-1 flex-col items-start justify-center gap-3 py-4">
              <p className="text-[13px] text-text-tertiary">{t('home.bots.empty')}</p>
              <SettingsNavLink
                to="/settings"
                nav={SETTINGS_IM_DISCORD}
                className={cn(
                  'inline-flex items-center gap-1 text-[13px] text-brand',
                  heSpring,
                  'hover:underline',
                )}
              >
                {t('home.bots.goSettings')}
                <ArrowRight className="size-3.5" />
              </SettingsNavLink>
            </div>
          )}

          {!loading && discordBots.length > 0 && (
            <ul className="space-y-2">
              {discordBots.map((bot) => {
                const state = botConnectionState(bot, transportPhase, t)
                return (
                  <li key={bot.id}>
                    <SettingsNavLink
                      to="/settings"
                      nav={settingsNavForDiscordBot(bot.id)}
                      className={cn(
                        hePressable,
                        heSpring,
                        'flex items-center gap-3 rounded-xl bg-black/[0.02] px-3 py-2.5 ring-1 ring-inset ring-black/[0.04]',
                        'hover:bg-brand-soft/40 hover:ring-primary/15',
                      )}
                    >
                      {bot.avatar_url ? (
                        <img
                          src={bot.avatar_url}
                          alt=""
                          className="size-9 shrink-0 rounded-full bg-black/[0.04] object-cover"
                        />
                      ) : (
                        <span className="inline-flex size-9 shrink-0 items-center justify-center rounded-full bg-black/[0.05] text-[11px] font-semibold text-text-tertiary">
                          {botLabel(bot).slice(0, 1).toUpperCase()}
                        </span>
                      )}
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-[13px] font-medium text-text-primary">
                          {botLabel(bot)}
                          {isGatewayHost(bot) ? (
                            <span className="ml-1.5 text-[10px] font-normal text-brand">
                              {t('home.bots.hostTag')}
                            </span>
                          ) : (
                            <span className="ml-1.5 text-[10px] font-normal text-text-tertiary">
                              {t('home.bots.participantTag')}
                            </span>
                          )}
                        </p>
                        <p className="truncate font-mono text-[11px] text-text-tertiary">
                          {bot.discord_username
                            ? `@${bot.discord_username.replace(/^@/, '')}`
                            : bot.configured
                              ? bot.id
                              : t('home.bots.tokenMissing')}
                        </p>
                      </div>
                      <ServiceStatusPill
                        tone={state.tone}
                        label={state.label}
                        tooltip={state.tooltip}
                      />
                      <ChevronRight className="size-3.5 shrink-0 text-text-tertiary/40" aria-hidden />
                    </SettingsNavLink>
                  </li>
                )
              })}
            </ul>
          )}

          {!loading && configuredBots.some((b) => !isGatewayHost(b)) && (
            <p className="mt-3 text-[11px] leading-relaxed text-text-tertiary">
              {t('home.bots.restHint')}
            </p>
          )}
        </article>
      </div>
    </section>
  )
}
