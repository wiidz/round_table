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
import { useLocale } from '@/contexts/locale-context'
import { useI18n } from '@/hooks/use-i18n'
import { settingsFieldDescription, settingsFieldLabel } from '@/lib/i18n/settings-fields'
import type { Translator } from '@/lib/i18n/translate'
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

const TAB_ORDER = ['service', 'storage', 'llm', 'meeting', 'im'] as const

const TAB_META: Record<string, { icon: LucideIcon }> = {
  service: { icon: Server },
  storage: { icon: Database },
  llm: { icon: Bot },
  meeting: { icon: Users },
  im: { icon: MessagesSquare },
}

/** Server-side section name for meeting limits (locale-independent) */
const MEETING_LIMITS_SECTION = '设定上限'

const SERVER_GROUP_FALLBACK: Record<string, string> = {
  service: '服务',
  storage: '存储',
  llm: 'LLM',
  meeting: '会议',
  im: 'IM',
}

const SUBSECTION_FALLBACK_ICONS: Record<string, typeof Bot> = {}

function subsectionFallbackIcon(id: string) {
  return SUBSECTION_FALLBACK_ICONS[id] ?? Settings2
}

function groupFields(
  fields: SettingsFieldState[],
  settingsTabKey: (serverGroup: string) => string,
) {
  const map = new Map<string, SettingsFieldState[]>()
  for (const f of fields) {
    const tabKey = settingsTabKey(f.group)
    const list = map.get(tabKey) ?? []
    list.push(f)
    map.set(tabKey, list)
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

function serverGroupForTabKey(
  tabKey: string,
  fields: SettingsFieldState[],
  settingsTabKey: (serverGroup: string) => string,
): string {
  const groups = new Set(fields.map((f) => f.group))
  for (const g of groups) {
    if (settingsTabKey(g) === tabKey) return g
  }
  return SERVER_GROUP_FALLBACK[tabKey] ?? tabKey
}

function tabIcon(tab: string) {
  return TAB_META[tab]?.icon ?? Settings2
}

function fieldSections(fields: SettingsFieldState[], t: Translator) {
  const runtime = fields.filter((f) => !f.secret)
  const secrets = fields.filter((f) => f.secret)
  if (runtime.length === 0 || secrets.length === 0) {
    return [{ title: '', fields }]
  }
  return [
    { title: t('pages.settings.sections.runtimeParams'), fields: runtime },
    { title: t('pages.settings.sections.apiKeys'), fields: secrets },
  ]
}

function subsectionRailTitle(tab: string, t: Translator) {
  if (tab === 'llm') return t('pages.settings.rail.models')
  if (tab === 'im') return t('pages.settings.rail.platforms')
  return t('pages.settings.rail.category')
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
  comingSoonSuffix,
}: {
  items: SettingsSubsectionMeta[]
  activeId: string
  onSelect: (id: string) => void
  groupLabel: string
  title: string
  discordRunning?: boolean
  comingSoonSuffix: string
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
            title={sub.available ? sub.label : `${sub.label}${comingSoonSuffix}`}
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

const MEETING_SECTION_ORDER = [MEETING_LIMITS_SECTION] as const

function meetingSectionDescription(section: string, t: Translator): string | undefined {
  if (section === MEETING_LIMITS_SECTION) {
    return t('pages.settings.sections.limitsDescription')
  }
  return undefined
}

function meetingSectionTitle(section: string, t: Translator): string {
  if (section === MEETING_LIMITS_SECTION) {
    return t('pages.settings.sections.limits')
  }
  return section
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

function groupMeetingFields(fields: SettingsFieldState[], t: Translator) {
  const sections: { title: string; description?: string; fields: SettingsFieldState[] }[] = []
  for (const serverSection of MEETING_SECTION_ORDER) {
    const sectionFields = fields.filter((f) => f.section === serverSection)
    if (sectionFields.length > 0) {
      sections.push({
        title: meetingSectionTitle(serverSection, t),
        description: meetingSectionDescription(serverSection, t),
        fields: sectionFields,
      })
    }
  }
  const rest = fields.filter(
    (f) => !f.section || !MEETING_SECTION_ORDER.includes(f.section as typeof MEETING_SECTION_ORDER[number]),
  )
  if (rest.length > 0) {
    sections.push({ title: '', fields: rest })
  }
  return sections
}

function settingsFieldHint(field: SettingsFieldState, t: Translator) {
  const description = settingsFieldDescription(t, field)
  const parts = [description]
  if (!field.input_type && field.key) {
    parts.push(t('pages.settings.field.keyName', { key: field.key }))
  }
  return parts.filter(Boolean).join('\n\n')
}

const SELECT_CHEVRON =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E\")"

function SettingsFieldInput({
  field,
  value,
  onChange,
  t,
}: {
  field: SettingsFieldState
  value: string
  onChange: (value: string) => void
  t: Translator
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
        ariaLabel={settingsFieldLabel(t, field)}
        onCheckedChange={(next) => onChange(next ? 'required' : 'skip')}
      />
    )
  }

  if (isRadioField(field) && field.options?.length) {
    return (
      <fieldset className="flex flex-wrap gap-2 sm:gap-3">
        <legend className="sr-only">{settingsFieldLabel(t, field)}</legend>
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
              <span className="text-sm text-text-primary">
                {field.key === 'ROUND_TABLE_LOCALE'
                  ? opt.value === 'en'
                    ? t('pages.settings.localeEn')
                    : t('pages.settings.localeZh')
                  : opt.label}
              </span>
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
  t,
}: {
  field: SettingsFieldState
  draft: Record<string, string>
  onChange: (key: string, value: string) => void
  t: Translator
}) {
  const hint = settingsFieldHint(field, t)

  const badges = (
    <div className="flex flex-wrap gap-1.5">
      {field.secret && field.configured && (
        <span className={cn(heFileBadge, 'bg-success-soft text-success ring-success/20')}>
          {t('common.configured')}
        </span>
      )}
      {field.restart_required && (
        <span className={heFileBadge}>{t('pages.settings.field.restartRequired')}</span>
      )}
      {field.secret && <span className={heFileBadge}>{t('common.readonly')}</span>}
    </div>
  )

  return (
    <SettingsFieldRowLayout
      label={settingsFieldLabel(t, field)}
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
            ? t('pages.settings.field.secretConfigured')
            : t('pages.settings.field.secretMissing')}
        </p>
      ) : field.editable ? (
        <SettingsFieldInput
          field={field}
          value={draft[field.key] ?? ''}
          onChange={(next) => onChange(field.key, next)}
          t={t}
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

function localeOptions(t: Translator): { value: string; label: string }[] {
  return [
    { value: 'zh', label: t('pages.settings.localeZh') },
    { value: 'en', label: t('pages.settings.localeEn') },
  ]
}

function normalizeSettingsField(field: SettingsFieldState, t: Translator): SettingsFieldState {
  if (field.key === LOCALE_KEY) {
    const options = localeOptions(t)
    return {
      ...field,
      input_type: 'radio',
      options: field.options?.length ? field.options : options,
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

function normalizeSettingsFields(fields: SettingsFieldState[], t: Translator): SettingsFieldState[] {
  return fields.map((field) => normalizeSettingsField(field, t))
}

const DISCORD_GENERAL_KEYS = new Set([DISCORD_GUILD_ID_KEY, DISCORD_AUTO_START_KEY])

export function SettingsPage() {
  const i18n = useI18n()
  const { t, settingsTabKey, settingsTabLabel } = i18n
  const { applyLocaleFromFields } = useLocale()
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

  const grouped = useMemo(
    () => groupFields(fields, settingsTabKey),
    [fields, settingsTabKey],
  )
  const tabs = useMemo(() => orderedTabs(grouped), [grouped])
  const activeServerGroup = useMemo(
    () => serverGroupForTabKey(activeTab, fields, settingsTabKey),
    [activeTab, fields, settingsTabKey],
  )
  const tabSubsections = subsections[activeServerGroup] ?? []
  const hasSubsections = tabSubsections.length > 0

  const activeSubMeta = tabSubsections.find((s) => s.id === activeSubsection)
  const subsectionAvailable = !hasSubsections || (activeSubMeta?.available ?? true)

  const showDiscordBots = activeTab === 'im' && activeSubsection === 'discord'
  const pollDiscordTransport = activeTab === 'im'
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
    activeTab === 'meeting' && activeFields.some((f) => f.editable)

  const showDiscordGeneralSave = showDiscordBots && discordBots.length > 0 && editableCount > 0

  function applySettingsResponse(data: SettingsResponse) {
    setFields(normalizeSettingsFields(data.fields ?? [], t))
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
      setActiveSubsection(defaultSubsection(activeServerGroup, subsections))
    }
  }, [loading, activeTab, hasSubsections, tabSubsections, activeSubsection, subsections, activeServerGroup])

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
          setError(t('common.error.requestFailed', { status: err.status, message: err.message }))
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError(t('pages.settings.loadFailed'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

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
      applyLocaleFromFields(resp.fields ?? [])
      setSubsections(resp.subsections ?? {})
      toast.success(t('pages.settings.saved'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  const panelTitle = hasSubsections && activeSubMeta
    ? `${settingsTabLabel(activeTab)} · ${activeSubMeta.label}`
    : settingsTabLabel(activeTab)

  return (
    <PageLayout
      header={
        <ProfilePageHeader
          role="principal"
          eyebrow={t('pages.settings.eyebrow')}
          title={t('pages.settings.title')}
          description={t('pages.settings.description')}
        />
      }
    >
    <div className="space-y-8">
      {loading && (
        <ProfileStatePanel
          className="!rounded-xs"
          title={t('common.loading')}
          description={t('pages.settings.loadingDescription')}
        />
      )}

      {!loading && error && (
        <ProfileStatePanel
          className="!rounded-xs"
          variant="danger"
          title={t('common.error.loadFailed')}
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
              aria-label={t('pages.settings.navAriaLabel')}
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
                          {settingsTabLabel(tab)}
                        </span>
                        {count > 0 ? (
                          <span
                            className={cn(
                              'truncate text-[10px] font-normal tabular-nums text-text-tertiary',
                              sideTabLabelMotion,
                              selected && 'text-brand/70',
                            )}
                          >
                            {t('pages.settings.navItemCount', { count })}
                          </span>
                        ) : null}
                      </span>
                    </span>
                  </button>
                )
              })}
            </nav>

            <nav
              aria-label={t('pages.settings.navAriaLabel')}
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
                        <span className="truncate">{settingsTabLabel(tab)}</span>
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
                  groupLabel={settingsTabLabel(activeTab)}
                  title={subsectionRailTitle(activeTab, t)}
                  comingSoonSuffix={t('pages.settings.comingSoonSuffix')}
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
                      title={t('pages.settings.comingSoonTitle')}
                      description={
                        activeTab === 'llm'
                          ? t('pages.settings.comingSoonLlm')
                          : t('pages.settings.comingSoonIm')
                      }
                    />
                  </div>
                )}

                {subsectionAvailable && activeFields.length > 0 && (
                  <div className="mt-8 space-y-10">
                    {activeTab === 'meeting' ? (
                      groupMeetingFields(activeFields, t).map((section) =>
                        section.title ? (
                          <SettingsSectionBlock
                            key={section.title}
                            title={section.title}
                            description={section.description}
                          >
                            {section.fields.map((field) => (
                              <SettingsFieldRow
                                key={field.key}
                                field={field}
                                draft={draft}
                                onChange={updateField}
                                t={t}
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
                                t={t}
                              />
                            ))}
                          </div>
                        ),
                      )
                    ) : (
                      fieldSections(activeFields, t).map((section) => (
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
                                t={t}
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
                      {saving ? t('common.saving') : t('pages.settings.save')}
                    </Button>
                  </div>
                )}

                {subsectionAvailable && activeTab === 'meeting' && meetPresets.length > 0 && (
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

                {subsectionAvailable && activeTab === 'meeting' && (
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
                            {saving ? t('common.saving') : t('pages.settings.save')}
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
                      title={t('pages.settings.emptyNoFieldsTitle')}
                      description={t('pages.settings.emptyNoFieldsDescription')}
                    />
                  </div>
                )}

                {subsectionAvailable && showDiscordBots && activeFields.length === 0 && discordBots.length === 0 && (
                  <div className="mt-8">
                    <ProfileStatePanel
                      className="!rounded-xs"
                      title={t('pages.settings.emptyNoBotsTitle')}
                      description={t('pages.settings.emptyNoBotsDescription')}
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
                      {saving ? t('common.saving') : t('pages.settings.save')}
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
