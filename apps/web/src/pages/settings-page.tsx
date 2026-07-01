import { useEffect, useMemo, useState, type ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import {
  Bot,
  Database,
  Hash,
  MessagesSquare,
  Save,
  Server,
  Settings2,
  Users,
} from 'lucide-react'
import { toast } from 'sonner'

import { fetchSettings, saveSettings } from '@/api/settings'
import { ApiError } from '@/api/client'
import { BrandIcon, hasBrandIcon } from '@/components/brand-icon'
import { PageLayout } from '@/components/layout/page-main-layout'
import { ProfilePageHeader, ProfileStatePanel } from '@/components/profile/profile-page-header'
import { DiscordBotsPanel } from '@/components/settings/discord-bots-panel'
import { MeetCastsPanel } from '@/components/settings/meet-casts-panel'
import { MeetPresetsPanel } from '@/components/settings/meet-presets-panel'
import { DiscordRunningRing } from '@/components/settings/discord-running-ring'
import { DiscordTransportControl } from '@/components/settings/discord-transport-control'
import { SettingsFieldRow as SettingsFieldRowLayout, SettingsToggle } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  heColumnTitleBrand,
  heFieldReadonly,
  heFieldSurface,
  heFileBadge,
  heFilePill,
  heFilePillSelected,
  hePanelShell,
  hePressable,
  heSectionDesc,
  heSectionTitle,
  heSpring,
  SETTINGS_SIDE_TAB_WIDTH,
  sideTabLabelMotion,
  settingsSideTabButtonClass,
  settingsSideTabListClass,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import { useDiscordTransportStatus } from '@/hooks/use-discord-transport-status'
import { readSettingsNav, writeSettingsNav } from '@/lib/settings-nav'
import type { DiscordBotState, MeetCastConfig, MeetPresetConfig, SettingsFieldState, SettingsResponse, SettingsSubsectionMeta } from '@/types/settings'

const settingsTabPill = cn(heFilePill, '!rounded-xs px-4 py-2.5 text-[15px]')
const settingsTabPillSelected = cn(heFilePillSelected, '!rounded-xs px-4 py-2.5 text-[15px]')
const settingsPanelShell = cn(hePanelShell, '!rounded-lg')

const TAB_ORDER = ['服务', '存储', 'LLM', '会议', 'IM'] as const

const TAB_META: Record<string, { icon: LucideIcon }> = {
  服务: { icon: Server },
  存储: { icon: Database },
  LLM: { icon: Bot },
  会议: { icon: Users },
  IM: { icon: MessagesSquare },
}

const SUBSECTION_FALLBACK_ICONS: Record<string, typeof Bot> = {}

function subsectionFallbackIcon(id: string) {
  return SUBSECTION_FALLBACK_ICONS[id] ?? Settings2
}

function groupFields(fields: SettingsFieldState[]) {
  const map = new Map<string, SettingsFieldState[]>()
  for (const f of fields) {
    const list = map.get(f.group) ?? []
    list.push(f)
    map.set(f.group, list)
  }
  return map
}

function orderedTabs(grouped: Map<string, SettingsFieldState[]>) {
  const tabs: string[] = []
  for (const name of TAB_ORDER) {
    if (grouped.has(name) && (grouped.get(name)?.length ?? 0) > 0) {
      tabs.push(name)
    }
  }
  for (const name of grouped.keys()) {
    if (!tabs.includes(name)) {
      tabs.push(name)
    }
  }
  return tabs
}

function tabIcon(tab: string) {
  return TAB_META[tab]?.icon ?? Settings2
}

function fieldSections(fields: SettingsFieldState[]) {
  const runtime = fields.filter((f) => !f.secret)
  const secrets = fields.filter((f) => f.secret)
  if (runtime.length === 0 || secrets.length === 0) {
    return [{ title: '', fields }]
  }
  return [
    { title: '运行参数', fields: runtime },
    { title: 'API 密钥', fields: secrets },
  ]
}

const SUBSECTION_RAIL_TITLE: Record<string, string> = {
  LLM: '模型',
  IM: '平台',
}

function subsectionRailTitle(tab: string) {
  return SUBSECTION_RAIL_TITLE[tab] ?? '分类'
}

function SubsectionMark({ id, className }: { id: string; className?: string }) {
  if (hasBrandIcon(id)) {
    return <BrandIcon id={id} className={className ?? 'size-7'} />
  }
  const Icon = subsectionFallbackIcon(id)
  return <Icon className={className ?? 'size-6'} strokeWidth={1.75} aria-hidden />
}

function subsectionRailButtonClass(selected: boolean, discordActive: boolean) {
  if (selected && discordActive) {
    return 'bg-brand-soft'
  }
  if (selected) {
    return 'bg-brand-soft ring-2 ring-primary/40 shadow-[var(--field-focus-shadow)]'
  }
  // Running but not selected: very faint green wash; transparent ring so dash gaps stay clean.
  if (discordActive) {
    return 'bg-success-soft ring-1 ring-inset ring-transparent hover:bg-success-soft'
  }
  return 'bg-black/[0.02] ring-1 ring-inset ring-black/[0.06] hover:bg-black/[0.04]'
}

function defaultSubsection(
  group: string,
  subsections: Record<string, SettingsSubsectionMeta[]>,
): string {
  const list = subsections[group]
  if (!list?.length) return ''
  const firstAvailable = list.find((s) => s.available)
  return firstAvailable?.id ?? list[0].id
}

function SubsectionIconRail({
  items,
  activeId,
  onSelect,
  groupLabel,
  title,
  discordRunning,
}: {
  items: SettingsSubsectionMeta[]
  activeId: string
  onSelect: (id: string) => void
  groupLabel: string
  title: string
  discordRunning?: boolean
}) {
  return (
    <nav
      aria-label={`${groupLabel} ${title}`}
      className="flex max-h-[min(32rem,calc(100vh-12rem))] w-[4.5rem] shrink-0 flex-col items-center gap-2 overflow-y-auto border-b border-black/[0.05] px-2 py-4 sm:border-b-0 sm:border-r"
    >
      <p className="mb-0.5 w-full text-center text-[10px] font-medium tracking-wide text-text-tertiary">
        {title}
      </p>
      {items.map((sub) => {
        const selected = sub.id === activeId
        const discordActive = sub.id === 'discord' && discordRunning === true
        return (
          <button
            key={sub.id}
            type="button"
            title={sub.available ? sub.label : `${sub.label}（即将推出）`}
            aria-label={sub.label}
            aria-selected={selected}
            onClick={() => onSelect(sub.id)}
            className={[
              'relative flex size-12 shrink-0 items-center justify-center rounded-xs',
              heSpring,
              subsectionRailButtonClass(selected, discordActive),
              discordActive &&
                '[--dr-ring-tail:#45dea0] [--dr-ring-mid:#5ef0a8] [--dr-ring-head:#7dffc8] [--dr-track:rgba(47,182,124,0.14)]',
              !sub.available && 'opacity-45',
            ].join(' ')}
          >
            <SubsectionMark id={sub.id} className="size-7" />
            {discordActive && <DiscordRunningRing emphasis={selected} />}
          </button>
        )
      })}
    </nav>
  )
}

const MEETING_SECTION_ORDER = ['设定上限'] as const

const MEETING_SECTION_DESC: Record<(typeof MEETING_SECTION_ORDER)[number], string> = {
  设定上限:
    '单场会议的 Engine 硬约束（写入 MeetingCreated）。预设里的辩论轮次、确认关开关不得突破此处上限；「确认关轮次上限」指 Principal 最多审阅几次方案（含首次）。',
}

function SettingsSectionBlock({
  title,
  description,
  children,
}: {
  title: string
  description?: string
  children: ReactNode
}) {
  return (
    <section className="space-y-6">
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <Hash className="size-4 shrink-0 text-info" strokeWidth={2} aria-hidden />
          <h2 className={heSectionTitle}>{title}</h2>
        </div>
        {description && <p className={heSectionDesc}>{description}</p>}
      </div>
      <div className="space-y-8">{children}</div>
    </section>
  )
}

function groupMeetingFields(fields: SettingsFieldState[]) {
  const sections: { title: string; fields: SettingsFieldState[] }[] = []
  for (const title of MEETING_SECTION_ORDER) {
    const sectionFields = fields.filter((f) => f.section === title)
    if (sectionFields.length > 0) {
      sections.push({ title, fields: sectionFields })
    }
  }
  const rest = fields.filter((f) => !f.section || !MEETING_SECTION_ORDER.includes(f.section as typeof MEETING_SECTION_ORDER[number]))
  if (rest.length > 0) {
    sections.push({ title: '', fields: rest })
  }
  return sections
}

function settingsFieldHint(field: SettingsFieldState) {
  const parts = [field.description]
  if (!field.input_type && field.key) {
    parts.push(`键名：${field.key}`)
  }
  return parts.filter(Boolean).join('\n\n')
}

const SELECT_CHEVRON =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E\")"

function SettingsFieldInput({
  field,
  value,
  onChange,
}: {
  field: SettingsFieldState
  value: string
  onChange: (value: string) => void
}) {
  const inputClass = cn(
    heFieldSurface,
    'h-10 w-full bg-surface px-3 text-sm text-text-primary placeholder:text-text-tertiary',
    heSpring,
    '!rounded-xs',
  )

  if (isSwitchField(field)) {
    const checked = value === 'required'
    return (
      <SettingsToggle
        id={field.key}
        checked={checked}
        ariaLabel={field.label}
        onCheckedChange={(next) => onChange(next ? 'required' : 'skip')}
      />
    )
  }

  if (isRadioField(field) && field.options?.length) {
    return (
      <fieldset className="flex flex-wrap gap-2 sm:gap-3">
        <legend className="sr-only">{field.label}</legend>
        {field.options.map((opt) => {
          const optionId = `${field.key}-${opt.value}`
          const selected = value === opt.value
          return (
            <label
              key={opt.value}
              htmlFor={optionId}
              className={cn(
                heFieldSurface,
                'flex min-w-[calc(50%-0.25rem)] flex-1 cursor-pointer items-center gap-2 bg-surface px-3 py-2.5 sm:min-w-0',
                heSpring,
                '!rounded-xs',
                selected && 'ring-2 ring-inset ring-primary/45 shadow-[var(--field-focus-shadow)]',
              )}
            >
              <input
                id={optionId}
                type="radio"
                name={field.key}
                value={opt.value}
                checked={selected}
                className="size-4 shrink-0 accent-primary"
                onChange={() => onChange(opt.value)}
              />
              <span className="text-sm text-text-primary">{opt.label}</span>
            </label>
          )
        })}
      </fieldset>
    )
  }

  if (field.input_type === 'select' && field.options?.length) {
    return (
      <select
        id={field.key}
        name={field.key}
        value={value}
        autoComplete="off"
        className={cn(inputClass, 'cursor-pointer appearance-none pr-9')}
        style={{
          backgroundImage: SELECT_CHEVRON,
          backgroundRepeat: 'no-repeat',
          backgroundPosition: 'right 0.65rem center',
          backgroundSize: '1rem',
        }}
        onChange={(e) => onChange(e.target.value)}
      >
        {field.options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    )
  }

  if (field.input_type === 'number') {
    return (
      <Input
        id={field.key}
        name={field.key}
        type="number"
        inputMode="numeric"
        min={field.min}
        max={field.max}
        step={1}
        value={value}
        placeholder={field.placeholder}
        autoComplete="off"
        className="!rounded-xs"
        onChange={(e) => onChange(e.target.value)}
      />
    )
  }

  return (
    <Input
      id={field.key}
      name={field.key}
      type="text"
      value={value}
      placeholder={field.placeholder}
      autoComplete="off"
      className="!rounded-xs"
      onChange={(e) => onChange(e.target.value)}
    />
  )
}

function SettingsFieldRow({
  field,
  draft,
  onChange,
}: {
  field: SettingsFieldState
  draft: Record<string, string>
  onChange: (key: string, value: string) => void
}) {
  const hint = settingsFieldHint(field)

  const badges = (
    <div className="flex flex-wrap gap-1.5">
      {field.secret && field.configured && (
        <span className={cn(heFileBadge, 'bg-success-soft text-success ring-success/20')}>
          已配置
        </span>
      )}
      {field.restart_required && <span className={heFileBadge}>需重启</span>}
      {field.secret && <span className={heFileBadge}>只读</span>}
    </div>
  )

  return (
    <SettingsFieldRowLayout
      label={field.label}
      htmlFor={isRadioField(field) ? undefined : field.key}
      hint={hint || undefined}
      labelExtra={badges}
    >
      {field.secret ? (
        <p
          className={cn(
            heFieldReadonly,
            'px-3 py-2.5 text-sm ring-1 ring-inset',
          )}
        >
          {field.configured
            ? '已在 deploy/.env 中配置，修改后请重启服务。'
            : '请在 deploy/.env 中配置对应密钥。'}
        </p>
      ) : field.editable ? (
        <SettingsFieldInput
          field={field}
          value={draft[field.key] ?? ''}
          onChange={(next) => onChange(field.key, next)}
        />
      ) : (
        <Input
          id={field.key}
          name={field.key}
          type="text"
          value={field.value ?? ''}
          readOnly
          className={heFieldReadonly}
        />
      )}
    </SettingsFieldRowLayout>
  )
}

const DISCORD_GUILD_ID_KEY = 'ROUND_TABLE_DISCORD_GUILD_ID'
const DISCORD_AUTO_START_KEY = 'ROUND_TABLE_DISCORD_AUTO_START'
const LOCALE_KEY = 'ROUND_TABLE_LOCALE'

const LOCALE_OPTIONS: { value: string; label: string }[] = [
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
]

function normalizeSettingsField(field: SettingsFieldState): SettingsFieldState {
  if (field.key === LOCALE_KEY) {
    return {
      ...field,
      input_type: 'radio',
      options: field.options?.length ? field.options : LOCALE_OPTIONS,
    }
  }
  return field
}

function isSwitchField(field: SettingsFieldState) {
  return field.input_type === 'switch'
}

function isRadioField(field: SettingsFieldState) {
  return field.input_type === 'radio' || field.key === LOCALE_KEY
}

function normalizeSettingsFields(fields: SettingsFieldState[]): SettingsFieldState[] {
  return fields.map(normalizeSettingsField)
}

const DISCORD_GENERAL_KEYS = new Set([DISCORD_GUILD_ID_KEY, DISCORD_AUTO_START_KEY])

export function SettingsPage() {
  const [fields, setFields] = useState<SettingsFieldState[]>([])
  const [discordBots, setDiscordBots] = useState<DiscordBotState[]>([])
  const [meetPresets, setMeetPresets] = useState<MeetPresetConfig[]>([])
  const [meetPresetsDefaults, setMeetPresetsDefaults] = useState<MeetPresetConfig[]>([])
  const [meetCasts, setMeetCasts] = useState<MeetCastConfig[]>([])
  const [subsections, setSubsections] = useState<Record<string, SettingsSubsectionMeta[]>>({})
  const [draft, setDraft] = useState<Record<string, string>>({})
  const [activeTab, setActiveTab] = useState(
    () => readSettingsNav()?.tab ?? TAB_ORDER[0],
  )
  const [activeSubsection, setActiveSubsection] = useState(
    () => readSettingsNav()?.subsection ?? '',
  )
  const [initialDiscordBotId] = useState(
    () => readSettingsNav()?.discordBotId,
  )
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const nav = readSettingsNav()
    if (!nav?.discordBotId) return
    writeSettingsNav({ tab: nav.tab, subsection: nav.subsection })
  }, [])

  const grouped = useMemo(() => groupFields(fields), [fields])
  const tabs = useMemo(() => orderedTabs(grouped), [grouped])
  const tabSubsections = subsections[activeTab] ?? []
  const hasSubsections = tabSubsections.length > 0

  const activeSubMeta = tabSubsections.find((s) => s.id === activeSubsection)
  const subsectionAvailable = !hasSubsections || (activeSubMeta?.available ?? true)

  const showDiscordBots = activeTab === 'IM' && activeSubsection === 'discord'
  const pollDiscordTransport = activeTab === 'IM'
  const discordTransport = useDiscordTransportStatus(pollDiscordTransport)

  const activeFields = useMemo(() => {
    const all = grouped.get(activeTab) ?? []
    const filtered = !hasSubsections
      ? all
      : all.filter((f) => f.subsection === activeSubsection)
    if (showDiscordBots) {
      return filtered.filter((f) => !DISCORD_GENERAL_KEYS.has(f.key))
    }
    return filtered
  }, [grouped, activeTab, hasSubsections, activeSubsection, showDiscordBots])

  const ActiveTabIcon = tabIcon(activeTab)
  const ActiveSubIcon = activeSubsection ? subsectionFallbackIcon(activeSubsection) : ActiveTabIcon
  const showBrandInTitle = activeSubsection !== '' && hasBrandIcon(activeSubsection)

  const editableCount = useMemo(
    () => fields.filter((f) => f.editable).length,
    [fields],
  )

  const showMeetingLimitsSave =
    activeTab === '会议' && activeFields.some((f) => f.editable)

  const showDiscordGeneralSave = showDiscordBots && discordBots.length > 0 && editableCount > 0

  function applySettingsResponse(data: SettingsResponse) {
    setFields(normalizeSettingsFields(data.fields ?? []))
    setDiscordBots(data.discord_bots ?? [])
    setMeetPresets(data.meet_presets ?? [])
    setMeetPresetsDefaults(data.meet_presets_defaults ?? [])
    setMeetCasts(data.meet_casts ?? [])
    setSubsections(data.subsections ?? {})
    const initial: Record<string, string> = {}
    for (const f of data.fields ?? []) {
      if (f.editable) {
        initial[f.key] = f.value ?? ''
      }
    }
    setDraft(initial)
  }

  useEffect(() => {
    if (loading) return
    if (tabs.length > 0 && !tabs.includes(activeTab)) {
      setActiveTab(tabs[0])
    }
  }, [loading, tabs, activeTab])

  useEffect(() => {
    if (loading) return
    if (!hasSubsections) {
      if (activeSubsection !== '') {
        setActiveSubsection('')
      }
      return
    }
    const ids = tabSubsections.map((s) => s.id)
    if (!ids.includes(activeSubsection)) {
      setActiveSubsection(defaultSubsection(activeTab, subsections))
    }
  }, [loading, activeTab, hasSubsections, tabSubsections, activeSubsection, subsections])

  useEffect(() => {
    if (loading) return
    writeSettingsNav({ tab: activeTab, subsection: activeSubsection })
  }, [loading, activeTab, activeSubsection])

  useEffect(() => {
    let cancelled = false
    fetchSettings()
      .then((data) => {
        if (cancelled) return
        applySettingsResponse(data)
        setError(null)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载设置')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  function updateField(key: string, value: string) {
    setDraft((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSave() {
    setSaving(true)
    try {
      const payload: Record<string, string> = {}
      for (const f of fields) {
        if (f.editable) {
          payload[f.key] = draft[f.key] ?? ''
        }
      }
      const resp = await saveSettings(payload)
      applySettingsResponse(resp)
      setSubsections(resp.subsections ?? {})
      toast.success('已保存，新会议立即生效')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const panelTitle = hasSubsections && activeSubMeta
    ? `${activeTab} · ${activeSubMeta.label}`
    : activeTab

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow="Configuration"
          title="设置"
          description="在此调整 RoundTable 运行参数；保存后对新发起的会议立即生效。API 密钥等敏感项请在 deploy/.env 中配置。"
        />
      }
    >
    <div className="space-y-8">
      {loading && (
        <ProfileStatePanel
          className="!rounded-xs"
          title="加载中"
          description="正在读取运行时配置…"
        />
      )}

      {!loading && error && (
        <ProfileStatePanel
          className="!rounded-xs"
          variant="danger"
          title="加载失败"
          description={error}
        />
      )}

      {!loading && !error && (
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void handleSave()
          }}
        >
          <div className="flex flex-col overflow-visible lg:flex-row lg:items-start lg:pl-3">
            <nav
              aria-label="设置分类"
              className={settingsSideTabListClass}
              style={{ width: SETTINGS_SIDE_TAB_WIDTH }}
            >
              {tabs.map((tab) => {
                const selected = tab === activeTab
                const count = grouped.get(tab)?.length ?? 0
                const Icon = tabIcon(tab)
                return (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => setActiveTab(tab)}
                    className={settingsSideTabButtonClass(selected)}
                  >
                    <span className="flex min-w-0 flex-1 items-center gap-2.5">
                      <span
                        data-tab-icon
                        className="relative flex size-7 shrink-0 items-center justify-center"
                      >
                        <Icon strokeWidth={selected ? 2 : 1.75} aria-hidden />
                      </span>
                      <span className="flex min-w-0 flex-1 flex-col items-start justify-center gap-0.5 leading-none">
                        <span
                          className={cn(
                            'truncate whitespace-nowrap text-[13px] font-medium',
                            sideTabLabelMotion,
                            selected && 'font-bold text-brand',
                            !selected && 'text-text-secondary',
                          )}
                        >
                          {tab}
                        </span>
                        {count > 0 ? (
                          <span
                            className={cn(
                              'truncate text-[10px] font-normal tabular-nums text-text-tertiary',
                              sideTabLabelMotion,
                              selected && 'text-brand/70',
                            )}
                          >
                            {count} 项
                          </span>
                        ) : null}
                      </span>
                    </span>
                  </button>
                )
              })}
            </nav>

            <nav
              aria-label="设置分类"
              className="flex shrink-0 gap-1.5 overflow-x-auto border-b border-black/[0.05] p-3 lg:hidden"
            >
              {tabs.map((tab) => {
                const selected = tab === activeTab
                const count = grouped.get(tab)?.length ?? 0
                const Icon = tabIcon(tab)
                return (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => setActiveTab(tab)}
                    className={cn(
                      'shrink-0 text-left',
                      selected ? settingsTabPillSelected : settingsTabPill,
                      heSpring,
                    )}
                  >
                    <span className="flex items-center justify-between gap-2">
                      <span className="flex min-w-0 items-center gap-2.5">
                        <Icon
                          className={cn(
                            'size-5 shrink-0',
                            selected ? 'text-brand' : 'text-text-tertiary',
                          )}
                          strokeWidth={1.75}
                          aria-hidden
                        />
                        <span className="truncate">{tab}</span>
                      </span>
                      <span className="text-xs font-normal tabular-nums opacity-60">
                        {count}
                      </span>
                    </span>
                  </button>
                )
              })}
            </nav>

            <div
              className={cn(
                settingsPanelShell,
                'flex min-w-0 flex-1 flex-col lg:flex-row',
              )}
            >
            <div className="flex min-h-0 min-w-0 flex-1 flex-row">
              {hasSubsections && (
                <SubsectionIconRail
                  items={tabSubsections}
                  activeId={activeSubsection}
                  onSelect={setActiveSubsection}
                  groupLabel={activeTab}
                  title={subsectionRailTitle(activeTab)}
                  discordRunning={
                    discordTransport.status != null ? discordTransport.ready : undefined
                  }
                />
              )}

              <div className="min-w-0 flex-1 p-6 sm:p-8">
                <div className="flex items-center justify-between gap-4">
                  <div className="flex min-w-0 items-center gap-2.5">
                    {showBrandInTitle ? (
                      <BrandIcon id={activeSubsection} className="size-5" />
                    ) : (
                      <ActiveSubIcon className="size-4 text-brand" aria-hidden />
                    )}
                    <p className={heColumnTitleBrand + ' !border-l-0 !pl-0'}>{panelTitle}</p>
                  </div>
                  {showDiscordBots && (
                    <DiscordTransportControl
                      phase={discordTransport.phase}
                      loading={discordTransport.status == null}
                      onRefresh={() => void discordTransport.refresh()}
                    />
                  )}
                </div>

                {!subsectionAvailable && (
                  <div className="mt-8">
                    <ProfileStatePanel
                      className="!rounded-xs"
                      title="即将推出"
                      description={
                        activeTab === 'LLM'
                          ? '该 LLM Provider 尚未接入，敬请期待。'
                          : '该 IM 平台尚未接入，敬请期待。'
                      }
                    />
                  </div>
                )}

                {subsectionAvailable && activeFields.length > 0 && (
                  <div className="mt-8 space-y-10">
                    {activeTab === '会议' ? (
                      groupMeetingFields(activeFields).map((section) =>
                        section.title ? (
                          <SettingsSectionBlock
                            key={section.title}
                            title={section.title}
                            description={
                              MEETING_SECTION_DESC[
                                section.title as (typeof MEETING_SECTION_ORDER)[number]
                              ]
                            }
                          >
                            {section.fields.map((field) => (
                              <SettingsFieldRow
                                key={field.key}
                                field={field}
                                draft={draft}
                                onChange={updateField}
                              />
                            ))}
                          </SettingsSectionBlock>
                        ) : (
                          <div key="other" className="space-y-8">
                            {section.fields.map((field) => (
                              <SettingsFieldRow
                                key={field.key}
                                field={field}
                                draft={draft}
                                onChange={updateField}
                              />
                            ))}
                          </div>
                        ),
                      )
                    ) : (
                      fieldSections(activeFields).map((section) => (
                        <div key={section.title || 'default'}>
                          {section.title && (
                            <h3 className={cn(heColumnTitleBrand, 'mb-6')}>{section.title}</h3>
                          )}
                          <div className="space-y-8">
                            {section.fields.map((field) => (
                              <SettingsFieldRow
                                key={field.key}
                                field={field}
                                draft={draft}
                                onChange={updateField}
                              />
                            ))}
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                )}

                {subsectionAvailable && showMeetingLimitsSave && (
                  <div
                    className={cn(
                      'flex flex-wrap items-center gap-3 border-t border-black/[0.05] pt-6',
                      activeFields.length > 0 ? 'mt-10' : 'mt-8',
                    )}
                  >
                    <Button
                      type="submit"
                      disabled={saving}
                      className={cn(hePressable, 'gap-2 rounded-xs px-5')}
                    >
                      <Save className="size-4" />
                      {saving ? '保存中…' : '保存配置'}
                    </Button>
                  </div>
                )}

                {subsectionAvailable && activeTab === '会议' && meetPresets.length > 0 && (
                  <div className={cn(showMeetingLimitsSave || activeFields.length > 0 ? 'mt-10' : 'mt-8')}>
                    <MeetPresetsPanel
                      presets={meetPresets}
                      defaults={meetPresetsDefaults}
                      maxRoundsCap={Number.parseInt(
                        draft['ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT'] ?? '20',
                        10,
                      )}
                      onSaved={(resp) => {
                        applySettingsResponse(resp)
                      }}
                    />
                  </div>
                )}

                {subsectionAvailable && activeTab === '会议' && (
                  <div className="mt-10 border-t border-black/[0.05] pt-10">
                    <MeetCastsPanel
                      casts={meetCasts}
                      onSaved={(resp) => {
                        applySettingsResponse(resp)
                      }}
                    />
                  </div>
                )}

                {subsectionAvailable && showDiscordBots && discordBots.length > 0 && (
                  <div className={activeFields.length > 0 ? 'mt-10 border-t border-black/[0.05] pt-10' : 'mt-8'}>
                    <DiscordBotsPanel
                      bots={discordBots}
                      initialBotId={initialDiscordBotId}
                      guildId={draft[DISCORD_GUILD_ID_KEY] ?? ''}
                      onGuildIdChange={(value) => updateField(DISCORD_GUILD_ID_KEY, value)}
                      autoStart={(draft[DISCORD_AUTO_START_KEY] ?? 'false') === 'true'}
                      onAutoStartChange={(checked) =>
                        updateField(DISCORD_AUTO_START_KEY, checked ? 'true' : 'false')
                      }
                      generalFooter={
                        showDiscordGeneralSave ? (
                          <Button
                            type="submit"
                            disabled={saving}
                            className={cn(hePressable, 'gap-2 rounded-xs px-5')}
                          >
                            <Save className="size-4" />
                            {saving ? '保存中…' : '保存配置'}
                          </Button>
                        ) : undefined
                      }
                      onSaved={(resp) => {
                        applySettingsResponse(resp)
                      }}
                    />
                  </div>
                )}

                {subsectionAvailable && !showDiscordBots && activeFields.length === 0 && (
                  <div className="mt-8">
                    <ProfileStatePanel
                      className="!rounded-xs"
                      title="暂无配置项"
                      description="当前分类下没有可展示的设置。"
                    />
                  </div>
                )}

                {subsectionAvailable && showDiscordBots && activeFields.length === 0 && discordBots.length === 0 && (
                  <div className="mt-8">
                    <ProfileStatePanel
                      className="!rounded-xs"
                      title="暂无 Bot 配置"
                      description="无法加载 Discord Bot 列表。"
                    />
                  </div>
                )}

                {editableCount > 0 && !showDiscordGeneralSave && !showMeetingLimitsSave && (
                  <div className="mt-10 flex flex-wrap items-center gap-3 border-t border-black/[0.05] pt-6">
                    <Button
                      type="submit"
                      disabled={saving}
                      className={cn(hePressable, 'gap-2 rounded-xs px-5')}
                    >
                      <Save className="size-4" />
                      {saving ? '保存中…' : '保存配置'}
                    </Button>
                  </div>
                )}
              </div>
            </div>
            </div>
          </div>
        </form>
      )}
    </div>
    </PageLayout>
  )
}
