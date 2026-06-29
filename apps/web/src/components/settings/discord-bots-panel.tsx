import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Bot, Hash, Plus, RefreshCw, Save, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import { refreshDiscordBotProfiles, saveDiscordBots } from '@/api/settings'
import { fetchParticipants } from '@/api/participants'
import { FieldHintPopover, SettingsFieldRow, SettingsSwitch } from '@/components/settings/field-hint-popover'
import { SearchableSelect } from '@/components/settings/searchable-select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  heColumnTitleAI,
  heColumnTitleBrand,
  heFileBadge,
  heFormEmbed,
  heFieldReadonly,
  hePressable,
  heSectionDesc,
  heSectionTitle,
  heSubsectionTitleNeutral,
  sideTabButtonMotion,
  sideTabIconMotion,
  sideTabInactiveBorderClass,
  sideTabLabelMotion,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import { formatDateTimeYMDHMS } from '@/lib/format-date'
import type { DiscordBotInput, DiscordBotsUpdate, DiscordBotState, SettingsResponse } from '@/types/settings'

type ParticipantDraft = {
  application_id: string
  token: string
  configured: boolean
  avatar_url?: string
  discord_application_id?: string
  discord_username?: string
  display_name?: string
  bound_participant_id: string
}

type ExpertOption = {
  id: string
  display_name?: string
  expertise?: string
}

type ExpertSelectOption = ExpertOption & {
  disabled?: boolean
}

import { resolveDiscordBotTab, type DiscordBotTabKey } from '@/lib/discord-bot-nav'

const APP_ID_PATTERN = /^\d{17,20}$/

const BOT_SIDE_TAB_WIDTH = '10rem'

const botSideTabListClass = cn(
  'flex shrink-0 flex-col gap-2 self-start overflow-visible',
  'max-h-[min(32rem,calc(100vh-14rem))] overflow-y-auto bg-transparent pt-8 pb-1',
)

const botTabAvatar = cn(
  'flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xs p-0.5',
)

/** 右侧表单 panel：四角圆角，左侧无 ring，与激活 Tab 衔接 */
const botFormPanelShell = cn(
  'relative z-0 min-w-0 flex-1 overflow-hidden rounded-xl bg-canvas',
  'shadow-[var(--field-inset-shadow)]',
  'ring-1 ring-inset ring-t ring-r ring-b ring-[var(--field-ring)]',
  '-ml-px',
)

/** 未激活 ml-3 内缩；激活 ml-0 全宽突出（不用 -ml-3，避免滚动容器裁切） */
function botSideTabButtonClass(selected: boolean, configured: boolean) {
  return cn(
    sideTabButtonMotion,
    'flex min-h-[3rem] w-full max-w-full flex-row items-center gap-2.5 rounded-l-lg rounded-r-none',
    'border border-r-0 border-l-[3px] cursor-pointer py-2 pl-2 text-left',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 focus-visible:ring-offset-2',
    selected
      ? cn(
          'relative z-10 ml-0 min-h-[3.25rem] pl-2 pr-2',
          'border border-r-0 border-l-[3px] border-l-primary border-t-black/[0.12] border-b-black/[0.12] !bg-canvas font-semibold',
        )
      : cn(
          'z-0 ml-3 w-[calc(100%-0.75rem)]',
          sideTabInactiveBorderClass,
          'border-l-transparent bg-black/[0.04] font-medium text-[13px] text-text-secondary',
          'hover:bg-black/[0.06] hover:text-text-primary',
        ),
    !configured && !selected && 'opacity-75',
  )
}

const botSideTabAddClass = cn(
  sideTabButtonMotion,
  'z-0 ml-3 flex min-h-[3rem] w-[calc(100%-0.75rem)] flex-row items-center gap-2.5 rounded-l-lg rounded-r-none',
  'border border-r-0 border-l-transparent bg-black/[0.04] px-2 py-2',
  sideTabInactiveBorderClass,
  'text-[13px] text-text-tertiary hover:bg-black/[0.06] hover:text-text-secondary',
)

function tokenForSave(draft: string, configured: boolean): string | undefined {
  const value = draft.trim()
  if (!value) {
    if (configured) {
      return undefined
    }
    throw new Error('请填写 Bot Token')
  }
  return value
}

function BotTab({
  label,
  roleId,
  avatarUrl,
  configured,
  selected,
  isModerator,
  onSelect,
}: {
  label: string
  roleId: string
  avatarUrl?: string
  configured: boolean
  selected: boolean
  isModerator?: boolean
  onSelect: () => void
}) {
  const displayLabel = label.trim() || '新 Bot'
  const displayRoleId = roleId.trim() || '待填写'
  const fallback = (displayLabel || '?').slice(0, 1).toUpperCase()

  return (
    <button
      type="button"
      title={`${displayLabel} (${displayRoleId})`}
      aria-label={displayLabel}
      aria-selected={selected}
      onClick={onSelect}
      className={botSideTabButtonClass(selected, configured)}
    >
      <span className="relative shrink-0">
        <span
          className={cn(
            botTabAvatar,
            sideTabIconMotion,
            selected
              ? 'border-2 border-primary/50 bg-brand-soft/70'
              : 'border border-black/[0.08] bg-black/[0.02]',
          )}
        >
          {avatarUrl ? (
            <img src={avatarUrl} alt="" className="size-full rounded-[3px] object-cover" />
          ) : (
            <span className="flex size-full items-center justify-center rounded-[3px] bg-black/[0.04] text-sm font-medium text-text-secondary">
              {fallback === '?' ? <Bot className="size-5 text-text-tertiary" /> : fallback}
            </span>
          )}
        </span>

        {isModerator && (
          <span
            aria-hidden
            className="pointer-events-none absolute -right-1 -top-1 flex size-4 items-center justify-center rounded-full bg-red-500 text-[9px] font-bold leading-none text-white ring-2 ring-canvas"
          >
            M
          </span>
        )}
      </span>

      <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-none">
        <span
          className={cn(
            'max-w-[6rem] truncate text-[13px]',
            sideTabLabelMotion,
            selected ? 'font-bold text-brand' : 'font-medium text-text-secondary',
          )}
        >
          {displayLabel}
        </span>
        <span
          className={cn(
            'max-w-[6rem] truncate font-mono text-[10px] text-text-tertiary',
            sideTabLabelMotion,
            selected && 'text-brand/70',
          )}
        >
          {displayRoleId}
        </span>
      </span>
    </button>
  )
}

function botApplicationId(b: Pick<DiscordBotState, 'id' | 'discord_application_id'>): string {
  const appId = (b.discord_application_id ?? b.id).trim()
  return APP_ID_PATTERN.test(appId) ? appId : ''
}

function toParticipantDrafts(bots: DiscordBotState[]): ParticipantDraft[] {
  return bots
    .filter((b) => b.deletable)
    .map((b) => ({
      application_id: botApplicationId(b),
      token: b.token ?? '',
      configured: b.configured,
      avatar_url: b.avatar_url,
      discord_application_id: b.discord_application_id ?? botApplicationId(b),
      discord_username: b.discord_username,
      display_name: b.display_name,
      bound_participant_id: b.bound_participant_id ?? '',
    }))
}

function newParticipant(): ParticipantDraft {
  return {
    application_id: '',
    token: '',
    configured: false,
    bound_participant_id: '',
  }
}

function StatusPill({
  children,
  tone = 'neutral',
}: {
  children: ReactNode
  tone?: 'success' | 'neutral' | 'accent'
}) {
  return (
    <span
      className={cn(
        heFileBadge,
        tone === 'success' && 'bg-success-soft text-success ring-success/20',
        tone === 'accent' &&
          'bg-red-500 font-semibold text-white ring-2 ring-red-500/25 ring-inset',
      )}
    >
      {children}
    </span>
  )
}

type FormHeaderTag = string | { label: string; tone?: 'success' | 'neutral' | 'accent' }

function SettingsBlock({
  title,
  description,
  accent = 'brand',
  children,
}: {
  title: string
  description?: string
  accent?: 'brand' | 'ai' | 'neutral'
  children: ReactNode
}) {
  const titleClass =
    accent === 'ai'
      ? heColumnTitleAI
      : accent === 'neutral'
        ? heSubsectionTitleNeutral
        : heColumnTitleBrand

  return (
    <section className="space-y-5">
      <div className="flex items-center gap-1.5">
        <h3 className={titleClass}>{title}</h3>
        {description && (
          <FieldHintPopover content={description} ariaLabel={`${title} 说明`} />
        )}
      </div>
      <div className="space-y-6">{children}</div>
    </section>
  )
}

function BotSettingsForm({
  title,
  subtitle,
  tags,
  id,
  kind,
  configured,
  discordApplicationId,
  discordUsername,
  token,
  isModerator = false,
  onIsModeratorChange,
  saving,
  onTokenChange,
  onSubmit,
  onDelete,
  embedded = false,
  expertOptions,
  boundParticipantId,
  onBoundParticipantIdChange,
}: {
  title: string
  subtitle?: string
  tags?: FormHeaderTag[]
  id: string
  kind: 'moderator' | 'participant'
  configured: boolean
  discordApplicationId?: string
  discordUsername?: string
  token: string
  isModerator?: boolean
  onIsModeratorChange?: (checked: boolean) => void
  saving: boolean
  onTokenChange: (value: string) => void
  onSubmit: () => void | Promise<void>
  onDelete?: () => void | Promise<void>
  embedded?: boolean
  expertOptions?: ExpertSelectOption[]
  boundParticipantId?: string
  onBoundParticipantIdChange?: (id: string) => void
}) {
  const discordProfileHint = '保存 Token 后点「同步信息」获取'
  const formKey = id || 'new'

  return (
    <div className={cn(embedded ? 'space-y-8' : cn(heFormEmbed, 'space-y-8 p-5 sm:p-6'))}>
      <header className="space-y-3 border-b border-black/[0.05] pb-6">
        <div className="space-y-1.5">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className={heSectionTitle}>{title}</h2>
            {tags?.map((tag) => {
              const label = typeof tag === 'string' ? tag : tag.label
              const tone = typeof tag === 'string' ? 'neutral' : (tag.tone ?? 'neutral')
              return (
                <StatusPill key={label} tone={tone}>
                  {label}
                </StatusPill>
              )
            })}
            {configured && <StatusPill tone="success">已配置</StatusPill>}
          </div>
          {subtitle && <p className={heSectionDesc}>{subtitle}</p>}
        </div>
      </header>

      <SettingsBlock
        title="Discord Developer"
        description="Application ID 与用户名通过顶部「同步信息」从 Discord API 拉取（只读）。"
        accent="ai"
      >
        <SettingsFieldRow
          label="Bot Token"
          htmlFor={`bot-token-${formKey}`}
          required
          hint="修改后需重启 discord 服务"
        >
          <Input
            id={`bot-token-${formKey}`}
            type="text"
            value={token}
            placeholder={configured ? '留空则保留现有 Token' : '填入 Discord Bot Token'}
            autoComplete="off"
            className="!rounded-xs font-mono text-sm"
            onChange={(e) => onTokenChange(e.target.value)}
          />
        </SettingsFieldRow>
        <SettingsFieldRow
          label="Application ID"
          htmlFor={`discord-app-id-${formKey}`}
          hint="Discord 开发者后台 Application ID"
        >
          <Input
            id={`discord-app-id-${formKey}`}
            value={discordApplicationId ?? ''}
            placeholder={discordProfileHint}
            disabled
            className={cn(heFieldReadonly, 'font-mono text-sm')}
          />
        </SettingsFieldRow>
        <SettingsFieldRow label="Bot 用户名" htmlFor={`discord-name-${formKey}`}>
          <Input
            id={`discord-name-${formKey}`}
            value={discordUsername ?? ''}
            placeholder={discordProfileHint}
            disabled
            className={cn(heFieldReadonly, 'text-sm')}
          />
        </SettingsFieldRow>
      </SettingsBlock>

      {onIsModeratorChange && (
        <SettingsBlock title="角色" accent="neutral">
          <SettingsSwitch
            id={`bot-moderator-${formKey}`}
            label="是否设为主持人"
            checked={isModerator}
            disabled={kind === 'participant' && !configured && !id.trim()}
            onCheckedChange={onIsModeratorChange}
            hint={
              kind === 'participant' && !configured && !id.trim()
                ? '请先保存 Token，系统将根据 Discord 资料自动注册 Bot'
                : '主持 Bot 负责指令、进度与会议流程；每位 Bot 均可绑定专家档案'
            }
          />
        </SettingsBlock>
      )}

      {expertOptions && onBoundParticipantIdChange && (
        <SettingsBlock
          title="绑定专家"
          description="每个 Discord Bot 仅绑定一位专家；绑定后 Bot 展示名称与头像跟随专家，并使用其 SOUL / AGENTS / TOOLS 档案。"
          accent="brand"
        >
          {expertOptions.length === 0 ? (
            <p className="text-[13px] text-text-tertiary">请先在「专家」页添加专家。</p>
          ) : (
            <SettingsFieldRow
              label="选择专家"
              htmlFor={`bot-expert-${formKey}`}
              hint="每位专家在同一平台只能绑定一个 Bot；已被其他 Bot 绑定的专家不可选"
            >
              <SearchableSelect
                id={`bot-expert-${formKey}`}
                value={boundParticipantId ?? ''}
                placeholder="不绑定专家"
                searchPlaceholder="输入名称或代号…"
                emptyOption={{
                  value: '',
                  label: '不绑定专家',
                }}
                options={expertOptions.map((expert) => {
                  const name = expert.display_name?.trim() || expert.id
                  return {
                    value: expert.id,
                    label: name,
                    hint: expert.disabled ? '已绑定其他 Bot' : expert.id,
                    disabled: expert.disabled,
                  }
                })}
                onChange={onBoundParticipantIdChange}
              />
            </SettingsFieldRow>
          )}
        </SettingsBlock>
      )}

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-black/[0.05] pt-6">
        {onDelete ? (
          <Button
            type="button"
            variant="outline"
            disabled={saving}
            className={cn(
              hePressable,
              'gap-2 rounded-xl px-5 text-destructive hover:bg-destructive/10 hover:text-destructive',
            )}
            onClick={() => void onDelete()}
          >
            <Trash2 className="size-4" />
            删除 Bot
          </Button>
        ) : (
          <span aria-hidden className="shrink-0" />
        )}
        <Button
          type="button"
          disabled={saving}
          className={cn(hePressable, 'gap-2 rounded-xl px-5')}
          onClick={() => void onSubmit()}
        >
          <Save className="size-4" />
          {saving ? '保存中…' : '保存 Bot 配置'}
        </Button>
      </div>
    </div>
  )
}

function optionalParticipantToken(p: ParticipantDraft): string | undefined {
  const value = (p.token ?? '').trim()
  return value || undefined
}

function assembleDiscordBotsUpdate({
  primaryRoleId,
  moderatorToken,
  moderatorConfigured,
  moderatorBoundParticipantId,
  participants,
  requireTokenForId,
  requireTokenForIndex,
}: {
  primaryRoleId: string
  moderatorToken: string
  moderatorConfigured: boolean
  moderatorBoundParticipantId: string
  participants: ParticipantDraft[]
  requireTokenForId?: string
  requireTokenForIndex?: number
}): DiscordBotsUpdate {
  validateDiscordBotExpertDrafts(participants, moderatorBoundParticipantId)

  const payload: DiscordBotInput[] = []

  participants.forEach((p, index) => {
    const appId = p.application_id.trim()
    const hasToken = Boolean((p.token ?? '').trim())
    if (!appId && !hasToken && !(p.bound_participant_id ?? '').trim()) {
      return
    }
    if (!appId && !hasToken) {
      throw new Error('新 Bot 需要填写 Token')
    }

    let token: string | undefined
    const needsToken =
      requireTokenForIndex === index ||
      (requireTokenForId && appId === requireTokenForId)
    if (appId !== primaryRoleId) {
      if (needsToken) {
        token = tokenForSave(p.token ?? '', p.configured)
      } else {
        token = optionalParticipantToken(p)
      }
    }

    payload.push({
      application_id: appId || undefined,
      token,
      bound_participant_id: (p.bound_participant_id ?? '').trim(),
    })
  })

  const update: DiscordBotsUpdate = {
    participants: payload,
    moderator_role_id: primaryRoleId,
    moderator_bound_participant_id: moderatorBoundParticipantId.trim(),
  }

  if (primaryRoleId === 'moderator') {
    if (requireTokenForId === 'moderator') {
      const tok = tokenForSave(moderatorToken, moderatorConfigured)
      if (tok !== undefined) update.moderator_token = tok
    } else {
      const tok = optionalParticipantToken({ token: moderatorToken, configured: moderatorConfigured } as ParticipantDraft)
      if (tok !== undefined) update.moderator_token = tok
    }
  } else {
    const primaryParticipant = participants.find(
      (p) => p.application_id.trim() === primaryRoleId,
    )
    if (primaryParticipant) {
      if (requireTokenForId === primaryRoleId) {
        const tok = tokenForSave(primaryParticipant.token ?? '', primaryParticipant.configured)
        if (tok !== undefined) update.moderator_token = tok
      } else {
        const tok = optionalParticipantToken(primaryParticipant)
        if (tok !== undefined) update.moderator_token = tok
      }
    }
    if (requireTokenForId === 'moderator') {
      const tok = tokenForSave(moderatorToken, moderatorConfigured)
      if (tok !== undefined) update.moderator_role_token = tok
    } else {
      const tok = optionalParticipantToken({ token: moderatorToken, configured: moderatorConfigured } as ParticipantDraft)
      if (tok !== undefined) update.moderator_role_token = tok
    }
  }

  return update
}

function validateDiscordBotExpertDrafts(
  participants: ParticipantDraft[],
  moderatorBoundParticipantId: string,
) {
  const expertBot = new Map<string, string>()
  const botExpert = new Map<string, string>()

  function track(botKey: string, expertId: string) {
    const prevBot = expertBot.get(expertId)
    if (prevBot && prevBot !== botKey) {
      throw new Error(`专家 ${expertId} 不能绑定多个 Discord Bot`)
    }
    expertBot.set(expertId, botKey)
    const prevExpert = botExpert.get(botKey)
    if (prevExpert && prevExpert !== expertId) {
      throw new Error(`Bot ${botKey} 只能绑定一位专家`)
    }
    botExpert.set(botKey, expertId)
  }

  const moderatorExpert = moderatorBoundParticipantId.trim()
  if (moderatorExpert) {
    track('主持人', moderatorExpert)
  }

  for (const [index, p] of participants.entries()) {
    const botKey = p.application_id.trim() || `(新 Bot ${index + 1})`
    const expertId = (p.bound_participant_id ?? '').trim()
    if (!expertId) continue
    track(botKey, expertId)
  }
}

type BotBindingSlot = {
  key: string
  bound_participant_id: string
}

function botBindingSlots(
  moderatorBound: string,
  participants: ParticipantDraft[],
): BotBindingSlot[] {
  return [
    { key: 'moderator', bound_participant_id: moderatorBound },
    ...participants.map((p, index) => ({
      key: p.application_id.trim() || `draft-${index}`,
      bound_participant_id: p.bound_participant_id ?? '',
    })),
  ]
}

function expertSelectOptions(
  roster: ExpertOption[],
  slots: BotBindingSlot[],
  activeKey: string,
): ExpertSelectOption[] {
  const taken = new Map<string, string>()
  slots.forEach((slot) => {
    if (slot.key === activeKey) return
    const expertId = slot.bound_participant_id.trim()
    if (expertId) {
      taken.set(expertId, slot.key === 'moderator' ? '主持人' : slot.key)
    }
  })
  const current =
    slots.find((slot) => slot.key === activeKey)?.bound_participant_id.trim() ?? ''
  return roster
    .map((expert) => ({
      ...expert,
      disabled: taken.has(expert.id) && expert.id !== current,
    }))
    .sort((a, b) => Number(a.disabled) - Number(b.disabled))
}

function participantTabLabel(
  p: ParticipantDraft,
  expertRoster: ExpertOption[],
  index: number,
): string {
  const bound = (p.bound_participant_id ?? '').trim()
  const expert = expertRoster.find((e) => e.id === bound)
  return (
    expert?.display_name?.trim() ||
    p.display_name?.trim() ||
    p.discord_username?.trim() ||
    p.application_id.trim() ||
    `新 Bot ${index + 1}`
  )
}

function participantTabSubtitle(p: ParticipantDraft): string {
  const appId = p.application_id.trim() || p.discord_application_id?.trim() || ''
  if (appId) {
    return appId.length > 12 ? `${appId.slice(0, 6)}…${appId.slice(-4)}` : appId
  }
  return '待保存'
}

function handlePrimaryChange(
  applicationId: string,
  checked: boolean,
  setPrimaryRoleId: (id: string) => void,
  participants: ParticipantDraft[],
) {
  if (checked) {
    const appId = applicationId.trim()
    if (appId === 'moderator') {
      setPrimaryRoleId('moderator')
      return
    }
    if (!appId) {
      toast.error('请先保存 Bot Token 以获取 Application ID')
      return
    }
    setPrimaryRoleId(appId)
    return
  }

  if (applicationId.trim() === 'moderator' || applicationId.trim() === '') {
    const fallback = participants.find((p) => p.application_id.trim())?.application_id.trim()
    if (fallback) {
      setPrimaryRoleId(fallback)
      return
    }
    toast.error('至少需要一位主持 Bot')
    return
  }

  setPrimaryRoleId('moderator')
}

const GUILD_ID_KEY = 'ROUND_TABLE_DISCORD_GUILD_ID'
const AUTO_START_KEY = 'ROUND_TABLE_DISCORD_AUTO_START'

function DiscordGeneralSection({
  guildId,
  onGuildIdChange,
  autoStart,
  onAutoStartChange,
}: {
  guildId: string
  onGuildIdChange: (value: string) => void
  autoStart: boolean
  onAutoStartChange: (checked: boolean) => void
}) {
  return (
    <section className="space-y-4">
      <div className="flex items-center gap-2">
        <Hash className="size-4 shrink-0 text-info" strokeWidth={2} aria-hidden />
        <h2 className={heSectionTitle}>General · 通用</h2>
      </div>
      <SettingsSwitch
        id={AUTO_START_KEY}
        label="自动启动"
        checked={autoStart}
        onCheckedChange={onAutoStartChange}
        hint="开启后，启动 roundtable 主服务时会自动拉起 Discord transport；保存后需重启主服务生效"
      />
      <SettingsFieldRow
        label="Guild ID"
        htmlFor={GUILD_ID_KEY}
        hint="留空表示不限制 Discord 服务器；Bot 仅在该 Guild 内响应指令"
      >
        <Input
          id={GUILD_ID_KEY}
          name={GUILD_ID_KEY}
          type="text"
          value={guildId}
          placeholder="留空表示不限制服务器"
          autoComplete="off"
          className="!rounded-xs font-mono text-sm"
          onChange={(e) => onGuildIdChange(e.target.value)}
        />
      </SettingsFieldRow>
    </section>
  )
}

export function DiscordBotsPanel({
  bots,
  guildId,
  onGuildIdChange,
  autoStart,
  onAutoStartChange,
  onSaved,
  generalFooter,
  initialBotId,
}: {
  bots: DiscordBotState[]
  guildId: string
  onGuildIdChange: (value: string) => void
  autoStart: boolean
  onAutoStartChange: (checked: boolean) => void
  onSaved: (resp: SettingsResponse) => void
  generalFooter?: ReactNode
  /** 从概览等页跳转时预选 Bot 侧栏 Tab */
  initialBotId?: string
}) {
  const moderator = useMemo(() => bots.find((b) => b.id === 'moderator'), [bots])
  const [primaryRoleId, setPrimaryRoleId] = useState(
    () => bots.find((b) => b.primary)?.id ?? 'moderator',
  )
  const [moderatorToken, setModeratorToken] = useState('')
  const [moderatorBoundParticipantId, setModeratorBoundParticipantId] = useState('')
  const [participants, setParticipants] = useState<ParticipantDraft[]>(() => toParticipantDrafts(bots))
  const [expertRoster, setExpertRoster] = useState<ExpertOption[]>([])
  const [activeTab, setActiveTab] = useState<DiscordBotTabKey>(() =>
    resolveDiscordBotTab(initialBotId, bots),
  )
  const [saving, setSaving] = useState(false)
  const [refreshing, setRefreshing] = useState(false)

  const lastProfileFetchedAt = useMemo(() => {
    const times = bots
      .map((b) => b.profile_fetched_at)
      .filter((t): t is string => Boolean(t))
      .sort()
    return times.at(-1) ?? ''
  }, [bots])

  useEffect(() => {
    setParticipants(toParticipantDrafts(bots))
    setPrimaryRoleId(bots.find((b) => b.primary)?.id ?? 'moderator')
    const mod = bots.find((b) => b.id === 'moderator')
    setModeratorToken(mod?.token ?? '')
    setModeratorBoundParticipantId(mod?.bound_participant_id ?? '')
  }, [bots])

  useEffect(() => {
    if (!initialBotId) return
    setActiveTab(resolveDiscordBotTab(initialBotId, bots))
  }, [initialBotId, bots])

  useEffect(() => {
    let cancelled = false
    fetchParticipants()
      .then((data) => {
        if (cancelled) return
        setExpertRoster(
          (data.participants ?? []).map((p) => ({
            id: p.id,
            display_name: p.display_name,
            expertise: p.expertise,
          })),
        )
      })
      .catch(() => {
        if (!cancelled) setExpertRoster([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  function updateParticipant(index: number, patch: Partial<ParticipantDraft>) {
    setParticipants((prev) => prev.map((p, i) => (i === index ? { ...p, ...patch } : p)))
  }

  async function persistBots(update: DiscordBotsUpdate) {
    setSaving(true)
    try {
      const resp = await saveDiscordBots(update)
      onSaved(resp)
      toast.success('Bot 配置已保存，请重启 discord 服务')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
      throw err
    } finally {
      setSaving(false)
    }
  }

  async function handleSaveModerator() {
    if (!moderator) return
    let update: DiscordBotsUpdate
    try {
      update = assembleDiscordBotsUpdate({
        primaryRoleId,
        moderatorToken,
        moderatorConfigured: moderator.configured,
        moderatorBoundParticipantId,
        participants,
        requireTokenForId: 'moderator',
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
      return
    }
    try {
      await persistBots(update)
    } catch {
      // toast shown in persistBots
    }
  }

  async function handleSaveParticipant(index: number) {
    const p = participants[index]
    if (!p.application_id.trim() && !(p.token ?? '').trim() && !p.configured) {
      toast.error('请填写 Bot Token')
      return
    }
    let update: DiscordBotsUpdate
    try {
      update = assembleDiscordBotsUpdate({
        primaryRoleId,
        moderatorToken,
        moderatorConfigured: moderator?.configured ?? false,
        moderatorBoundParticipantId,
        participants,
        requireTokenForId: p.application_id.trim() || undefined,
        requireTokenForIndex: index,
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
      return
    }
    try {
      await persistBots(update)
    } catch {
      // toast shown in persistBots
    }
  }

  async function removeParticipant(index: number) {
    const removed = participants[index]
    const removedAppId = removed?.application_id.trim()
    const nextPrimary =
      removedAppId && primaryRoleId === removedAppId ? 'moderator' : primaryRoleId
    const next = participants.filter((_, i) => i !== index)
    setParticipants(next)
    setPrimaryRoleId(nextPrimary)
    setActiveTab('moderator')

    let update: DiscordBotsUpdate
    try {
      update = assembleDiscordBotsUpdate({
        primaryRoleId: nextPrimary,
        moderatorToken,
        moderatorConfigured: moderator?.configured ?? false,
        moderatorBoundParticipantId,
        participants: next,
      })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败')
      setParticipants(toParticipantDrafts(bots))
      setPrimaryRoleId(bots.find((b) => b.primary)?.id ?? 'moderator')
      return
    }

    setSaving(true)
    try {
      const resp = await saveDiscordBots(update)
      onSaved(resp)
      toast.success('已删除 Bot')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败')
      setParticipants(toParticipantDrafts(bots))
      setPrimaryRoleId(bots.find((b) => b.primary)?.id ?? 'moderator')
    } finally {
      setSaving(false)
    }
  }

  function addParticipant() {
    setParticipants((prev) => {
      const next = [...prev, newParticipant()]
      setActiveTab(`participant-${next.length - 1}`)
      return next
    })
  }

  async function handleRefreshProfiles() {
    setRefreshing(true)
    try {
      const resp = await refreshDiscordBotProfiles()
      onSaved(resp)
      const count = (resp.discord_bots ?? []).filter((b) => b.avatar_url).length
      if (count > 0) {
        toast.success(`已拉取 ${count} 个 Bot 头像并缓存`)
      } else {
        toast.warning('拉取完成，但未获取到头像（请检查 Token 或网络/代理）')
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '拉取头像失败')
    } finally {
      setRefreshing(false)
    }
  }

  if (!moderator) {
    return null
  }

  const activeParticipantIndex =
    activeTab.startsWith('participant-') ? Number(activeTab.slice('participant-'.length)) : -1
  const activeParticipant =
    activeParticipantIndex >= 0 ? participants[activeParticipantIndex] : undefined
  const bindingSlots = botBindingSlots(moderatorBoundParticipantId, participants)

  return (
    <div className="space-y-10">
      <DiscordGeneralSection
        guildId={guildId}
        onGuildIdChange={onGuildIdChange}
        autoStart={autoStart}
        onAutoStartChange={onAutoStartChange}
      />

      {generalFooter && (
        <div className="flex flex-wrap items-center gap-3 border-b border-black/[0.05] pb-10">
          {generalFooter}
        </div>
      )}

      <section className="space-y-6">
      <header>
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 space-y-1.5">
            <div className="flex items-center gap-2">
              <Hash className="size-4 shrink-0 text-info" strokeWidth={2} aria-hidden />
              <h2 className={heSectionTitle}>Discord Bots</h2>
            </div>
            <p className={heSectionDesc}>
              修改 Token 并保存后，可点「同步信息」获取 <a href="https://discord.com/developers/applications" target='blank'>Developer Portal</a>里的资料（
              {lastProfileFetchedAt
                ? `上次：${formatDateTimeYMDHMS(lastProfileFetchedAt)}`
                : '从未拉取'}
              ）
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={refreshing}
            className={cn(hePressable, 'shrink-0 gap-1.5 rounded-xl')}
            onClick={() => void handleRefreshProfiles()}
          >
            <RefreshCw className={cn('size-3.5', refreshing && 'animate-spin')} />
            {refreshing ? '同步中…' : '同步信息'}
          </Button>
        </div>
      </header>

      <div className="min-w-0">
      <div className="flex min-h-0 min-w-0 flex-col gap-4 sm:flex-row sm:items-start sm:gap-0">
        <nav
          aria-label="Discord Bot 列表"
          className={botSideTabListClass}
          style={{ width: BOT_SIDE_TAB_WIDTH }}
        >
          <BotTab
            label={moderator.label ?? moderator.display_name ?? '主持人'}
            roleId={moderator.id}
            avatarUrl={moderator.avatar_url}
            configured={moderator.configured}
            selected={activeTab === 'moderator'}
            isModerator={primaryRoleId === 'moderator'}
            onSelect={() => setActiveTab('moderator')}
          />

          {participants.map((p, index) => {
            const key: DiscordBotTabKey = `participant-${index}`
            const tabLabel = participantTabLabel(p, expertRoster, index)
            const appId = p.application_id.trim()
            return (
              <BotTab
                key={key}
                label={tabLabel}
                roleId={participantTabSubtitle(p)}
                avatarUrl={p.avatar_url}
                configured={p.configured}
                selected={activeTab === key}
                isModerator={Boolean(appId) && primaryRoleId === appId}
                onSelect={() => setActiveTab(key)}
              />
            )
          })}

          <button
            type="button"
            title="添加参与 Bot"
            aria-label="添加参与 Bot"
            onClick={addParticipant}
            className={botSideTabAddClass}
          >
            <span className="flex size-10 shrink-0 items-center justify-center rounded-xs border border-dashed border-black/[0.12] bg-black/[0.02]">
              <Plus className="size-4" />
            </span>
            <span className="truncate font-medium">添加 Bot</span>
          </button>
        </nav>

        {(activeTab === 'moderator' || activeParticipant) && (
          <div className={botFormPanelShell}>
            <div className="space-y-8 p-5 sm:p-6">
      {activeTab === 'moderator' && (
        <BotSettingsForm
          embedded
          kind="moderator"
          title="主持人"
          tags={[{ label: 'Moderator', tone: 'accent' }]}
          subtitle="填写 Token、绑定专家并指定是否为主持 Bot；主持 Bot 负责指令、进度与会议流程"
          id={moderator.id}
          configured={moderator.configured}
          discordApplicationId={moderator.discord_application_id}
          discordUsername={moderator.discord_username}
          token={moderatorToken}
          isModerator={primaryRoleId === 'moderator'}
          onIsModeratorChange={(checked) =>
            handlePrimaryChange('moderator', checked, setPrimaryRoleId, participants)
          }
          saving={saving}
          onTokenChange={setModeratorToken}
          onBoundParticipantIdChange={setModeratorBoundParticipantId}
          expertOptions={expertSelectOptions(expertRoster, bindingSlots, 'moderator')}
          boundParticipantId={moderatorBoundParticipantId}
          onSubmit={handleSaveModerator}
        />
      )}

      {activeParticipant && (
        <BotSettingsForm
          embedded
          kind="participant"
          title={participantTabLabel(activeParticipant, expertRoster, activeParticipantIndex)}
          subtitle="填写 Token 并选择绑定专家；展示名称与头像跟随专家，Discord 资料仅本地缓存预览"
          id={activeParticipant.application_id.trim() || 'new'}
          configured={activeParticipant.configured}
          discordApplicationId={activeParticipant.discord_application_id}
          discordUsername={activeParticipant.discord_username}
          token={activeParticipant.token ?? ''}
          isModerator={primaryRoleId === activeParticipant.application_id.trim()}
          onIsModeratorChange={(checked) =>
            handlePrimaryChange(
              activeParticipant.application_id,
              checked,
              setPrimaryRoleId,
              participants,
            )
          }
          saving={saving}
          onTokenChange={(value) => updateParticipant(activeParticipantIndex, { token: value })}
          onBoundParticipantIdChange={(id) =>
            updateParticipant(activeParticipantIndex, { bound_participant_id: id })
          }
          expertOptions={expertSelectOptions(
            expertRoster,
            bindingSlots,
            activeParticipant.application_id.trim() || `draft-${activeParticipantIndex}`,
          )}
          boundParticipantId={activeParticipant.bound_participant_id}
          onSubmit={() => handleSaveParticipant(activeParticipantIndex)}
          onDelete={() => void removeParticipant(activeParticipantIndex)}
        />
      )}
            </div>
          </div>
        )}
      </div>
      </div>
      </section>
    </div>
  )
}
