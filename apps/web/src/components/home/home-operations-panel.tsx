import { ArrowRight, ChevronRight, MessagesSquare, Server, Settings2 } from 'lucide-react'

import { DiscordTransportControl } from '@/components/settings/discord-transport-control'
import { DiscordTransportStatusBadge } from '@/components/settings/discord-transport-status-badge'
import { SettingsNavLink } from '@/components/settings/settings-nav-link'
import {
  heFileBadge,
  hePanelShell,
  hePressable,
  heSpring,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { formatProcessRuntime } from '@/lib/format-runtime'
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

function botConnectionState(
  bot: DiscordBotState,
  phase: DiscordTransportPhase,
): { tone: 'success' | 'warning' | 'neutral' | 'info'; label: string; tooltip?: string } {
  if (!bot.configured) {
    return {
      tone: 'neutral',
      label: '未配置',
      tooltip: '尚未填写 Bot Token，无法在 Discord 中发话或接收指令。',
    }
  }
  if (phase === 'starting') {
    if (isGatewayHost(bot)) {
      return {
        tone: 'warning',
        label: 'Gateway 连接中',
        tooltip:
          '司仪 Bot 正在连接 Discord Gateway。连接成功后会在 Discord 成员列表显示在线，并负责接收频道指令。',
      }
    }
    return {
      tone: 'warning',
      label: '等待 Transport',
      tooltip: 'Discord Transport 启动中，参与 Bot 需等待服务就绪后才可 REST 发话。',
    }
  }
  if (phase !== 'ready') {
    return {
      tone: 'neutral',
      label: '不可用',
      tooltip: isGatewayHost(bot)
        ? 'Discord Transport 未运行，司仪 Bot 未连接 Gateway，Discord 中将显示离线。'
        : 'Discord Transport 未运行，参与 Bot 暂时无法发话。',
    }
  }
  if (isGatewayHost(bot)) {
    return {
      tone: 'success',
      label: 'Gateway 在线',
      tooltip:
        '司仪 Bot 已连接 Discord Gateway，在成员列表显示在线；负责接收指令、推进会议流程并协调发话。',
    }
  }
  return {
    tone: 'info',
    label: 'REST 发话',
    tooltip:
      '参与 Bot 仅通过 REST 发消息，不会在 Discord 成员列表显示在线，但会议中可正常发言。',
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

interface HomeOperationsPanelProps {
  loading: boolean
  apiOnline: boolean
  apiError?: string | null
  transportPhase: DiscordTransportPhase
  transportLoading: boolean
  transportPid?: number
  transportReadyAt?: string
  transportUnavailable?: boolean
  discordBots: DiscordBotState[]
  serverRuntime?: ProcessSnapshot
  discordRuntime?: ProcessSnapshot
  onTransportRefresh: () => void
}

export function HomeOperationsPanel({
  loading,
  apiOnline,
  apiError,
  transportPhase,
  transportLoading,
  transportPid,
  transportReadyAt,
  transportUnavailable,
  discordBots,
  serverRuntime,
  discordRuntime,
  onTransportRefresh,
}: HomeOperationsPanelProps) {
  const configuredBots = discordBots.filter((b) => b.configured)
  const participantBots = configuredBots.filter((b) => !isGatewayHost(b))
  const gatewayOnline = transportPhase === 'ready' && configuredBots.some(isGatewayHost)
  const sendOnlyReady = transportPhase === 'ready' ? participantBots.length : 0
  const serverRuntimeLabel = formatProcessRuntime(serverRuntime)
  const discordRuntimeLabel = formatProcessRuntime(discordRuntime)

  return (
    <section className="space-y-4">
      <h2 className={heSubsectionTitleNeutral}>运行与 IM</h2>

      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]">
        {/* 左列：API + Transport 控制 */}
        <div className="space-y-3">
          <article
            className={cn(
              hePanelShell,
              'flex flex-wrap items-center gap-x-4 gap-y-3 px-4 py-3.5',
            )}
          >
            <span className="inline-flex size-8 shrink-0 items-center justify-center rounded-lg bg-black/[0.04] text-text-secondary">
              <Server className="size-3.5" strokeWidth={1.75} aria-hidden />
            </span>
            <div className="min-w-0 flex-1">
              <p className="text-[13px] font-semibold text-text-primary">RoundTable API</p>
              {serverRuntimeLabel && !loading && (
                <p className="mt-0.5 font-mono text-[11px] tabular-nums text-text-tertiary">
                  {serverRuntimeLabel}
                </p>
              )}
            </div>
            {loading ? (
              <div className="h-6 w-16 animate-pulse rounded-full bg-black/[0.04]" />
            ) : apiOnline ? (
              <ServiceStatusPill tone="success" label="运行正常" />
            ) : (
              <ServiceStatusPill tone="danger" label="连接异常" />
            )}
            {!loading && apiError && (
              <p className="w-full text-[12px] text-text-tertiary">{apiError}</p>
            )}
          </article>

          <article className={cn(hePanelShell, 'space-y-4 p-5')}>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="min-w-0 space-y-2">
                <div className="flex items-center gap-2">
                  <span className="inline-flex size-8 shrink-0 items-center justify-center rounded-lg bg-black/[0.04] text-text-secondary">
                    <MessagesSquare className="size-3.5" strokeWidth={1.75} aria-hidden />
                  </span>
                  <div>
                    <p className="text-[13px] font-semibold text-text-primary">Discord Transport</p>
                    <p className="text-[12px] text-text-tertiary">IM 接入 · 启停与日志</p>
                  </div>
                </div>
                {loading ? (
                  <div className="h-6 w-24 animate-pulse rounded-full bg-black/[0.04]" />
                ) : transportUnavailable ? (
                  <ServiceStatusPill tone="neutral" label="监管不可用" />
                ) : (
                  <DiscordTransportStatusBadge
                    phase={transportPhase}
                    pid={transportPid}
                    readyAt={transportReadyAt}
                  />
                )}
              </div>
              {!transportUnavailable && (
                <DiscordTransportControl
                  phase={transportPhase}
                  loading={transportLoading}
                  onRefresh={onTransportRefresh}
                />
              )}
            </div>

            {!loading && !transportUnavailable && (
              <p className="text-[12px] leading-relaxed text-text-tertiary">
                {transportPhase === 'ready'
                  ? gatewayOnline
                    ? sendOnlyReady > 0
                      ? `司仪 Gateway 已连接 · ${sendOnlyReady} 个参与 Bot REST 发话可用`
                      : '司仪 Gateway 已连接'
                    : 'Transport 已就绪'
                  : transportPhase === 'starting'
                    ? '正在拉起进程并连接 Gateway…'
                    : 'Transport 未启动时无法在 Discord 频道收发消息'}
              </p>
            )}

            {!loading && discordRuntimeLabel && (
              <p className="font-mono text-[11px] tabular-nums text-text-tertiary">
                {discordRuntimeLabel}
                {discordRuntime?.pid != null && discordRuntime.pid > 0 && (
                  <span className="text-text-tertiary/60">{` · PID ${discordRuntime.pid}`}</span>
                )}
              </p>
            )}

            {transportUnavailable && (
              <p className="text-[12px] text-text-tertiary">
                当前环境未启用 Transport 监管，请直接在终端运行 Discord 服务。
              </p>
            )}
          </article>
        </div>

        {/* 右列：Bot 列表 → 设置 */}
        <article className={cn(hePanelShell, 'flex flex-col p-5')}>
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="text-[13px] font-semibold text-text-primary">Discord Bot</p>
              <p className="text-[12px] text-text-tertiary">Token、绑定专家与 Gateway 角色</p>
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
              管理 Bot
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
              <p className="text-[13px] text-text-tertiary">尚未配置 Discord Bot</p>
              <SettingsNavLink
                to="/settings"
                nav={SETTINGS_IM_DISCORD}
                className={cn(
                  'inline-flex items-center gap-1 text-[13px] text-brand',
                  heSpring,
                  'hover:underline',
                )}
              >
                前往设置添加
                <ArrowRight className="size-3.5" />
              </SettingsNavLink>
            </div>
          )}

          {!loading && discordBots.length > 0 && (
            <ul className="space-y-2">
              {discordBots.map((bot) => {
                const state = botConnectionState(bot, transportPhase)
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
                            <span className="ml-1.5 text-[10px] font-normal text-brand">司仪</span>
                          ) : (
                            <span className="ml-1.5 text-[10px] font-normal text-text-tertiary">参与</span>
                          )}
                        </p>
                        <p className="truncate font-mono text-[11px] text-text-tertiary">
                          {bot.discord_username
                            ? `@${bot.discord_username.replace(/^@/, '')}`
                            : bot.configured
                              ? bot.id
                              : 'Token 未配置'}
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
              参与 Bot 为 REST 发话，Discord 成员列表显示离线属正常；点击 Bot 可跳转设置编辑 Token 与绑定。
            </p>
          )}
        </article>
      </div>
    </section>
  )
}
