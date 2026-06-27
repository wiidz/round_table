import { useEffect, useMemo, useRef, useState } from 'react'
import { Hash, RotateCcw, Save } from 'lucide-react'
import { toast } from 'sonner'

import { resetMeetPresets, saveMeetPresets } from '@/api/settings'
import {
  FieldHintPopover,
  SettingsFieldRow,
  SettingsToggle,
} from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  heFieldSurface,
  hePressable,
  heSectionDesc,
  heSectionTitle,
  heSpring,
  sideTabButtonMotion,
  sideTabInactiveBorderClass,
  sideTabLabelMotion,
} from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { MeetPresetConfig, SettingsResponse } from '@/types/settings'

type PresetDraft = MeetPresetConfig

type PresetTabKey = string

const PRESET_SIDE_TAB_WIDTH = '10rem'
const PANEL_MIN_EXTRA_REM = 4

const presetSideTabListClass = cn(
  'flex shrink-0 flex-col gap-2 self-start overflow-visible bg-transparent pt-8 pb-1',
)

/** 与 discord-bots-panel botFormPanelShell 一致 */
const presetFormPanelShell = cn(
  'relative z-0 min-w-0 flex-1 overflow-hidden rounded-xl bg-canvas',
  'shadow-[var(--field-inset-shadow)]',
  'ring-1 ring-inset ring-t ring-r ring-b ring-[var(--field-ring)]',
  '-ml-px',
)

/** 与 discord-bots-panel botSideTabButtonClass 一致 */
function presetSideTabButtonClass(selected: boolean) {
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
  )
}

const MEET_MODE_OPTIONS: { value: string; label: string }[] = [
  { value: 'deliberation', label: '研讨型（方案共建）' },
  { value: 'decision', label: '裁决型（是否批准）' },
]

function normalizeCommandKey(s: string): string {
  const trimmed = s.trim().replace(/\s+/g, '')
  if (!trimmed) return ''
  if (/^[\x00-\x7F]+$/.test(trimmed)) {
    return trimmed.toUpperCase()
  }
  return trimmed
}

function clonePresets(presets: MeetPresetConfig[]): PresetDraft[] {
  return presets.map((p) => ({
    ...p,
    command: (p.command?.trim() || p.id).trim(),
  }))
}

function PresetTab({
  preset,
  selected,
  onSelect,
}: {
  preset: PresetDraft
  selected: boolean
  onSelect: () => void
}) {
  const label = preset.name_zh || preset.id

  return (
    <button
      type="button"
      title={`${label} (${preset.id})`}
      aria-label={label}
      aria-selected={selected}
      onClick={onSelect}
      className={presetSideTabButtonClass(selected)}
    >
      <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-none pl-1">
        <span
          className={cn(
            'max-w-[8.5rem] truncate text-[13px]',
            sideTabLabelMotion,
            selected ? 'font-bold text-brand' : 'font-medium text-text-secondary',
          )}
        >
          {label}
        </span>
        <span
          className={cn(
            'max-w-[8.5rem] truncate font-mono text-[10px] text-text-tertiary',
            sideTabLabelMotion,
            selected && 'text-brand/70',
          )}
        >
          {preset.command || preset.id}
        </span>
      </span>
    </button>
  )
}

function MeetModeRadio({
  id,
  value,
  onChange,
}: {
  id: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <fieldset className="flex flex-wrap gap-2 sm:gap-3">
      <legend className="sr-only">会议模式</legend>
      {MEET_MODE_OPTIONS.map((opt) => {
        const optionId = `${id}-mode-${opt.value}`
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
              name={`${id}-mode`}
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

export function MeetPresetsPanel({
  presets,
  defaults,
  maxRoundsCap,
  onSaved,
}: {
  presets: MeetPresetConfig[]
  defaults: MeetPresetConfig[]
  maxRoundsCap: number
  onSaved: (resp: SettingsResponse) => void
}) {
  const [drafts, setDrafts] = useState<PresetDraft[]>(() => clonePresets(presets))
  const [activeId, setActiveId] = useState<PresetTabKey>(() => presets[0]?.id ?? '1')
  const [saving, setSaving] = useState(false)
  const [resetting, setResetting] = useState(false)
  const [panelMinHeight, setPanelMinHeight] = useState<number>()
  const tabsRef = useRef<HTMLElement>(null)

  useEffect(() => {
    setDrafts(clonePresets(presets))
    if (presets.length > 0 && !presets.some((p) => p.id === activeId)) {
      setActiveId(presets[0].id)
    }
  }, [presets, activeId])

  const orderedPresets = useMemo(() => {
    const deliberation = drafts.filter((p) => p.group === 'deliberation')
    const decision = drafts.filter((p) => p.group === 'decision')
    return [...deliberation, ...decision]
  }, [drafts])

  useEffect(() => {
    const el = tabsRef.current
    if (!el) return

    const update = () => {
      const rootFontSize =
        Number.parseFloat(getComputedStyle(document.documentElement).fontSize) || 16
      setPanelMinHeight(el.offsetHeight + PANEL_MIN_EXTRA_REM * rootFontSize)
    }

    update()
    const observer = new ResizeObserver(update)
    observer.observe(el)
    window.addEventListener('resize', update)
    return () => {
      observer.disconnect()
      window.removeEventListener('resize', update)
    }
  }, [orderedPresets.length, activeId])

  const active = drafts.find((p) => p.id === activeId) ?? drafts[0]
  const cap = maxRoundsCap > 0 ? maxRoundsCap : 20

  function patchActive(patch: Partial<PresetDraft>) {
    if (!active) return
    setDrafts((prev) =>
      prev.map((p) => (p.id === active.id ? { ...p, ...patch } : p)),
    )
  }

  async function handleSave() {
    const missing = drafts.find((p) => !p.name_zh.trim())
    if (missing) {
      toast.error(`请填写预设 ${missing.id} 的菜单名称`)
      setActiveId(missing.id)
      return
    }
    const noCmd = drafts.find((p) => !p.command?.trim())
    if (noCmd) {
      toast.error(`预设 ${noCmd.id} 须填写绑定指令`)
      setActiveId(noCmd.id)
      return
    }
    const reserved = drafts.find((p) => normalizeCommandKey(p.command ?? '') === '0')
    if (reserved) {
      toast.error('0 为系统保留（自定义入口），不能绑定到预设')
      setActiveId(reserved.id)
      return
    }
    const keys = new Map<string, string>()
    for (const p of drafts) {
      const key = normalizeCommandKey(p.command ?? '')
      const prev = keys.get(key)
      if (prev) {
        toast.error(`指令「${p.command}」与预设 ${prev} 重复`)
        setActiveId(p.id)
        return
      }
      keys.set(key, p.id)
    }
    setSaving(true)
    try {
      const resp = await saveMeetPresets(drafts)
      onSaved(resp)
      toast.success('预设菜单已保存')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  function restoreActiveDefault() {
    const def = defaults.find((d) => d.id === active?.id)
    if (!def || !active) return
    setDrafts((prev) =>
      prev.map((p) => (p.id === active.id ? clonePresets([def])[0] : p)),
    )
    toast.success('已恢复此项默认（尚未保存）')
  }

  async function handleResetAll() {
    if (
      !window.confirm(
        '确定将全部预设恢复为系统默认？当前自定义会丢失，且立即写入配置。',
      )
    ) {
      return
    }
    setResetting(true)
    try {
      const resp = await resetMeetPresets()
      onSaved(resp)
      toast.success('已恢复默认预设菜单')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '恢复失败')
    } finally {
      setResetting(false)
    }
  }

  if (drafts.length === 0) {
    return null
  }

  return (
    <section className="space-y-6">
      <div className="space-y-1.5">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <Hash className="size-4 shrink-0 text-info" strokeWidth={2} aria-hidden />
            <h2 className={heSectionTitle}>预设菜单</h2>
          </div>
          <Button
            type="button"
            variant="outline"
            disabled={resetting || saving}
            onClick={() => void handleResetAll()}
            className={cn(hePressable, 'shrink-0 gap-2 rounded-xl px-4')}
          >
            <RotateCcw className="size-4" />
            {resetting ? '恢复中…' : '恢复全部默认'}
          </Button>
        </div>
        <p className={heSectionDesc}>
          Discord 发起主题后的选项菜单；每项可配置名称、绑定指令与会议参数。恢复默认可回到系统内置方案。
        </p>
      </div>

      <div className="flex min-h-0 min-w-0 flex-col gap-4 sm:flex-row sm:items-start sm:gap-0">
        <nav
          ref={tabsRef}
          aria-label="会议预设列表"
          className={presetSideTabListClass}
          style={{ width: PRESET_SIDE_TAB_WIDTH }}
        >
          {orderedPresets.map((p) => (
            <PresetTab
              key={p.id}
              preset={p}
              selected={p.id === activeId}
              onSelect={() => setActiveId(p.id)}
            />
          ))}
        </nav>

        {active && (
          <div
            className={presetFormPanelShell}
            style={panelMinHeight != null ? { minHeight: panelMinHeight } : undefined}
          >
            <div className="space-y-8 p-5 sm:p-6">
              <header className="space-y-3 border-b border-black/[0.05] pb-6">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="space-y-1.5">
                    <h2 className={heSectionTitle}>{active.name_zh}</h2>
                    <p className={heSectionDesc}>
                      内部编号 <span className="font-mono">{active.id}</span>
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={!defaults.some((d) => d.id === active.id)}
                    onClick={restoreActiveDefault}
                    className={cn(hePressable, 'gap-1.5 rounded-xl shrink-0')}
                  >
                    <RotateCcw className="size-3.5" />
                    恢复此项默认
                  </Button>
                </div>
              </header>

              <SettingsFieldRow
                label="绑定指令"
                htmlFor={`preset-${active.id}-command`}
                required
                hint="Discord 预设菜单中 Principal 发送此文字即选中本项。不可使用 0（保留给自定义）；各预设指令不能重复。"
              >
                <Input
                  id={`preset-${active.id}-command`}
                  type="text"
                  value={active.command ?? ''}
                  maxLength={24}
                  placeholder={active.id}
                  onChange={(e) => patchActive({ command: e.target.value })}
                  className="!rounded-xs max-w-[14rem] font-mono"
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="菜单名称"
                htmlFor={`preset-${active.id}-name-zh`}
                required
                hint="Discord 菜单展示文案，emoji 与标题写在一起，例如「⚡ 闪电研讨」。"
              >
                <Input
                  id={`preset-${active.id}-name-zh`}
                  type="text"
                  value={active.name_zh}
                  maxLength={40}
                  placeholder="例如 ⚡ 闪电研讨"
                  onChange={(e) => patchActive({ name_zh: e.target.value })}
                  className="!rounded-xs"
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="英文名称"
                htmlFor={`preset-${active.id}-name-en`}
                hint="Discord 英文界面下的菜单文案；留空则沿用默认种子值。"
              >
                <Input
                  id={`preset-${active.id}-name-en`}
                  type="text"
                  value={active.name_en}
                  maxLength={48}
                  placeholder="例如 Flash"
                  onChange={(e) => patchActive({ name_en: e.target.value })}
                  className="!rounded-xs"
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="会议模式"
                hint="单场会议采用研讨或裁决流程。"
              >
                <MeetModeRadio
                  id={`preset-${active.id}`}
                  value={active.mode}
                  onChange={(mode) => patchActive({ mode })}
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="辩论轮次"
                htmlFor={`preset-${active.id}-rounds`}
                hint={`本场最多 ${cap} 轮（受「设定的上限 → 辩论轮次上限」约束）。`}
              >
                <Input
                  id={`preset-${active.id}-rounds`}
                  type="number"
                  min={1}
                  max={cap}
                  value={String(active.max_rounds)}
                  onChange={(e) => {
                    const n = Number.parseInt(e.target.value, 10)
                    if (!Number.isNaN(n)) {
                      patchActive({ max_rounds: n })
                    }
                  }}
                  className="!rounded-xs max-w-[8rem]"
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="确认关"
                hint="开启后，会议结束前 Principal 须批准合成方案；驳回后可继续讨论。最多审阅几轮由上方「确认关轮次上限」控制。"
              >
                <SettingsToggle
                  id={`preset-${active.id}-confirm`}
                  checked={active.confirmation === 'required'}
                  ariaLabel="确认关"
                  onCheckedChange={(checked) =>
                    patchActive({ confirmation: checked ? 'required' : 'skip' })
                  }
                />
              </SettingsFieldRow>

              <SettingsFieldRow
                label="自由对话"
                htmlFor={`preset-${active.id}-free`}
                hint="每位 Participant 在辩论前可提问的轮数；0 表示关闭。"
              >
                <Input
                  id={`preset-${active.id}-free`}
                  type="number"
                  min={0}
                  max={5}
                  value={String(active.free_dialogue_questions)}
                  onChange={(e) => {
                    const n = Number.parseInt(e.target.value, 10)
                    if (!Number.isNaN(n)) {
                      patchActive({ free_dialogue_questions: n })
                    }
                  }}
                  className="!rounded-xs max-w-[8rem]"
                />
              </SettingsFieldRow>

              <div className="flex flex-wrap items-center justify-end gap-2 border-t border-black/[0.05] pt-6">
                <FieldHintPopover
                  content="保存后立即作用于 Discord 发起会议时的选项菜单；无需重启服务。"
                  ariaLabel="保存说明"
                />
                <Button
                  type="button"
                  disabled={saving}
                  onClick={() => void handleSave()}
                  className={cn(hePressable, 'gap-2 rounded-xl px-5')}
                >
                  <Save className="size-4" />
                  {saving ? '保存中…' : '保存预设菜单'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
