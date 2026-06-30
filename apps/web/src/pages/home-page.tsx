import { useCallback, useEffect, useMemo, useState } from 'react'
import type { LucideIcon } from 'lucide-react'
import { ArrowRight, Bot, FileStack, LayoutList, UserRound, Users } from 'lucide-react'
import { Link } from 'react-router-dom'

import { fetchBriefTemplates } from '@/api/brief-templates'
import { fetchAllMeetings, fetchMeetings } from '@/api/meetings'
import { fetchParticipants } from '@/api/participants'
import { fetchPrincipals } from '@/api/principals'
import { fetchSettings, fetchDiscordTransportStatus } from '@/api/settings'
import { fetchRuntime } from '@/api/runtime'
import { ApiError } from '@/api/client'
import { HomeOperationsPanel } from '@/components/home/home-operations-panel'
import {
  MeetingGridCard,
  MeetingGridSkeleton,
} from '@/components/meeting/meeting-list-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { useDiscordTransportStatus } from '@/hooks/use-discord-transport-status'
import {
  hePanelShell,
  hePanelShellHover,
  hePressable,
  heSpring,
  heSubsectionTitleNeutral,
} from '@/lib/highend-styles'
import { meetingStatusLabel } from '@/lib/meeting-labels'
import { domainNavLabel } from '@/lib/ui-labels'
import { cn } from '@/lib/utils'

import type { MeetingIndex } from '@/types/meeting'
import type { DiscordBotState } from '@/types/settings'
import type { RuntimeResponse } from '@/types/runtime'

const RECENT_MEETINGS = 6

function formatTokenCount(value: number): string {
  if (value >= 1_000_000) {
    const compact = value / 1_000_000
    return `${compact >= 10 ? Math.round(compact) : compact.toFixed(1)}M`
  }
  if (value >= 10_000) {
    const compact = value / 1_000
    return `${compact >= 100 ? Math.round(compact) : compact.toFixed(1)}k`
  }
  return value.toLocaleString('zh-CN')
}

interface DashboardStatProps {
  label: string
  value: string
  hint?: string
  icon: LucideIcon
  accent?: 'brand' | 'ai' | 'neutral'
  to?: string
}

function DashboardStat({
  label,
  value,
  hint,
  icon: Icon,
  accent = 'neutral',
  to,
}: DashboardStatProps) {
  const inner = (
    <article
      className={cn(
        hePanelShell,
        to && hePanelShellHover,
        to && hePressable,
        heSpring,
        'flex h-full flex-col gap-3 p-5',
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-[11px] font-medium uppercase tracking-[0.14em] text-text-tertiary">
          {label}
        </p>
        <span
          className={cn(
            'inline-flex size-8 shrink-0 items-center justify-center rounded-lg',
            accent === 'brand' && 'bg-brand-soft text-brand',
            accent === 'ai' && 'bg-ai-soft text-ai',
            accent === 'neutral' && 'bg-black/[0.04] text-text-secondary',
          )}
        >
          <Icon className="size-3.5" strokeWidth={1.75} aria-hidden />
        </span>
      </div>
      <p
        className={cn(
          'text-[28px] font-semibold leading-none tracking-[-0.03em] tabular-nums',
          to ? 'text-text-primary group-hover:text-brand' : 'text-text-primary',
        )}
      >
        {value}
      </p>
      {hint && <p className="text-[12px] leading-relaxed text-text-tertiary">{hint}</p>}
      {to && (
        <span className="mt-auto inline-flex items-center gap-1 text-[12px] text-text-tertiary group-hover:text-brand">
          查看
          <ArrowRight className="size-3 opacity-60" />
        </span>
      )}
    </article>
  )

  if (!to) return inner

  return (
    <Link to={to} className="group block h-full">
      {inner}
    </Link>
  )
}

export function HomePage() {
  const [meetings, setMeetings] = useState<MeetingIndex[]>([])
  const [allMeetings, setAllMeetings] = useState<MeetingIndex[]>([])
  const [meetingTotal, setMeetingTotal] = useState(0)
  const [principalCount, setPrincipalCount] = useState(0)
  const [briefTemplateCount, setBriefTemplateCount] = useState(0)
  const [participantCount, setParticipantCount] = useState(0)
  const [discordBots, setDiscordBots] = useState<DiscordBotState[]>([])
  const [transportUnavailable, setTransportUnavailable] = useState(false)
  const [runtime, setRuntime] = useState<RuntimeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const discordTransport = useDiscordTransportStatus(true)

  const refreshRuntime = useCallback(() => {
    fetchRuntime()
      .then(setRuntime)
      .catch(() => setRuntime(null))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    Promise.all([
      fetchMeetings(1, RECENT_MEETINGS),
      fetchAllMeetings(),
      fetchPrincipals(),
      fetchBriefTemplates(),
      fetchParticipants(),
      fetchSettings().catch(() => null),
    ])
      .then(([meetingsRes, allMeetingsRes, principalsRes, briefsRes, participantsRes, settingsRes]) => {
        if (cancelled) return
        setMeetings(meetingsRes.meetings ?? [])
        setAllMeetings(allMeetingsRes)
        setMeetingTotal(meetingsRes.total ?? allMeetingsRes.length)
        setPrincipalCount(principalsRes.principals?.length ?? 0)
        setBriefTemplateCount(briefsRes.templates?.length ?? 0)
        setParticipantCount(participantsRes.participants?.length ?? 0)
        setDiscordBots(settingsRes?.discord_bots ?? [])
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载概览数据')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    fetchDiscordTransportStatus().catch((err: unknown) => {
      if (err instanceof ApiError && err.status === 503) {
        setTransportUnavailable(true)
      }
    })
  }, [])

  useEffect(() => {
    refreshRuntime()
    const timer = window.setInterval(refreshRuntime, 10_000)
    return () => window.clearInterval(timer)
  }, [refreshRuntime])

  const apiOnline = !loading && !error

  const recentStats = useMemo(() => {
    let completed = 0
    let running = 0

    for (const meeting of meetings) {
      const status = meetingStatusLabel(meeting.status)
      if (status === '已结束' || status === '已归档') completed += 1
      if (status === '进行中') running += 1
    }

    return { completed, running }
  }, [meetings])

  const usageStats = useMemo(() => {
    let tokenSum = 0
    let llmCalls = 0

    for (const meeting of allMeetings) {
      tokenSum += meeting.total_tokens ?? 0
      llmCalls += meeting.llm_call_count ?? 0
    }

    return { tokenSum, llmCalls }
  }, [allMeetings])

  return (
    <div className="space-y-10">
      <ProfilePageHeader
        role="principal"
        eyebrow="Overview"
        title="概览"
        description="仪表盘汇总会议与档案；运行与 IM 区可直接启停 Discord、查看日志并跳转 Bot 设置；下方进入最近 Meeting 复盘。"
      />

      {error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      <section className="space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <h2 className={heSubsectionTitleNeutral}>仪表盘</h2>
        </div>

        {loading ? (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
            {Array.from({ length: 5 }, (_, i) => (
              <div
                key={i}
                className={cn(hePanelShell, 'h-[148px] animate-pulse bg-black/[0.02]')}
              />
            ))}
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
            <DashboardStat
              label="会议"
              value={String(meetingTotal)}
              hint={
                recentStats.running > 0
                  ? `其中 ${recentStats.running} 场进行中（近 ${meetings.length} 场）`
                  : `近 ${meetings.length} 场 · ${recentStats.completed} 场已结束`
              }
              icon={LayoutList}
              accent="brand"
              to="/meetings"
            />
            <DashboardStat
              label="简报模板"
              value={String(briefTemplateCount)}
              hint="可复用 Meeting Brief（ADR-0014）"
              icon={FileStack}
              accent="brand"
              to="/brief-templates"
            />
            <DashboardStat
              label={domainNavLabel('principal')}
              value={String(principalCount)}
              hint="USER.md 偏好档案"
              icon={UserRound}
              to="/principals"
            />
            <DashboardStat
              label={domainNavLabel('participant')}
              value={String(participantCount)}
              hint="专家档案"
              icon={Users}
              to="/participants"
            />
            <DashboardStat
              label="Token 用量"
              value={usageStats.tokenSum > 0 ? formatTokenCount(usageStats.tokenSum) : '—'}
              hint={
                usageStats.llmCalls > 0
                  ? `${usageStats.llmCalls} 次 LLM 调用（共 ${meetingTotal} 场）`
                  : '暂无用量记录'
              }
              icon={Bot}
              accent="ai"
              to="/meetings"
            />
          </div>
        )}
      </section>

      <HomeOperationsPanel
        loading={loading}
        apiOnline={apiOnline}
        apiError={error}
        transportPhase={discordTransport.phase}
        transportPid={discordTransport.status?.pid}
        transportUnavailable={transportUnavailable}
        discordBots={discordBots}
        serverRuntime={runtime?.server}
        discordRuntime={runtime?.discord_transport}
      />

      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className={heSubsectionTitleNeutral}>最近会议</h2>
          <Link
            to="/meetings"
            className={cn(
              'inline-flex items-center gap-1 text-[13px] text-text-secondary',
              heSpring,
              'hover:text-brand',
            )}
          >
            查看全部
            <ArrowRight className="size-3.5 opacity-60" />
          </Link>
        </div>

        {loading && <MeetingGridSkeleton count={3} />}

        {!loading && meetings.length === 0 && !error && (
          <ProfileStatePanel
            title="暂无会议"
            description={
              <>
                在{' '}
                <code className="font-mono text-xs">data/workspaces/</code>{' '}
                下尚未发现 Meeting；启动 Discord Transport 后可在频道发起会议。
              </>
            }
          />
        )}

        {!loading && meetings.length > 0 && (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {meetings.slice(0, 3).map((meeting) => (
              <MeetingGridCard key={meeting.id} meeting={meeting} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
