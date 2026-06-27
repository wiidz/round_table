import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { SearchableSelect } from '@/components/settings/searchable-select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { heFieldSurface, hePressable, heSpring } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type {
  ParticipantIMBinding,
  ParticipantIndex,
  ParticipantRosterInput,
} from '@/types/participant'

const ID_PATTERN = /^[a-z][a-z0-9_-]*$/

type DiscordBotOption = {
  id: string
  label: string
}

type ParticipantFormDialogProps = {
  open: boolean
  mode: 'create' | 'edit'
  initial?: ParticipantIndex | null
  peers: ParticipantIndex[]
  discordBots: DiscordBotOption[]
  onClose: () => void
  onSubmit: (input: ParticipantRosterInput) => Promise<void>
}

export function ParticipantFormDialog({
  open,
  mode,
  initial,
  peers,
  discordBots,
  onClose,
  onSubmit,
}: ParticipantFormDialogProps) {
  const [id, setId] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [expertise, setExpertise] = useState('')
  const [discordBotId, setDiscordBotId] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setId(initial?.id ?? '')
    setDisplayName(initial?.display_name?.trim() ?? '')
    setExpertise(initial?.expertise?.trim() ?? '')
    setDiscordBotId(discordBindingFromParticipant(initial?.im_bindings))
  }, [open, initial])

  if (!open) return null

  function validateClient(): ParticipantRosterInput | null {
    const trimmedId = id.trim()
    const trimmedName = displayName.trim()
    const trimmedExp = expertise.trim()

    if (!trimmedId) {
      toast.error('请填写代号')
      return null
    }
    if (!ID_PATTERN.test(trimmedId)) {
      toast.error('代号须小写字母开头，仅含 a-z、0-9、_、-')
      return null
    }
    if (trimmedId === 'moderator') {
      toast.error('代号 moderator 为系统保留')
      return null
    }
    if (!trimmedName) {
      toast.error('请填写名称')
      return null
    }

    const peerIds = peers.filter((p) => p.id !== initial?.id)
    if (peerIds.some((p) => p.id === trimmedId)) {
      toast.error(`代号 ${trimmedId} 已存在`)
      return null
    }

    const nameKey = normalizeNameKey(trimmedName)
    const dupName = peerIds.find(
      (p) => normalizeNameKey(p.display_name?.trim() || p.id) === nameKey,
    )
    if (dupName) {
      toast.error(`名称与 ${dupName.display_name || dupName.id} 重复`)
      return null
    }

    return {
      id: trimmedId,
      display_name: trimmedName,
      expertise: trimmedExp || undefined,
      im_bindings: buildIMBindings(discordBotId),
    }
  }

  async function handleSubmit() {
    const payload = validateClient()
    if (!payload) return
    setSaving(true)
    try {
      await onSubmit(payload)
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/25 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-labelledby="participant-form-title"
      onClick={onClose}
    >
      <div
        className={cn(
          heFieldSurface,
          'w-full max-w-md space-y-5 rounded-2xl bg-surface p-6 shadow-xl !rounded-xs',
          heSpring,
        )}
        onClick={(e) => e.stopPropagation()}
      >
        <header className="space-y-1">
          <h2 id="participant-form-title" className="text-base font-semibold text-text-primary">
            {mode === 'create' ? '添加专家' : '编辑专家'}
          </h2>
          <p className="text-[13px] text-text-secondary">
            代号用于档案目录；名称不可重复。每位专家可绑定 Discord、Telegram 等 IM（每平台一个 Bot）。
          </p>
        </header>

        <div className="space-y-4">
          <label className="block space-y-1.5">
            <span className="text-xs font-medium text-text-tertiary">代号</span>
            <Input
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder="例如 analyst"
              className="!rounded-xs font-mono"
              autoComplete="off"
            />
          </label>
          <label className="block space-y-1.5">
            <span className="text-xs font-medium text-text-tertiary">名称</span>
            <Input
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="例如 数据分析师"
              className="!rounded-xs"
            />
          </label>
          <label className="block space-y-1.5">
            <span className="text-xs font-medium text-text-tertiary">专长（可选）</span>
            <Input
              value={expertise}
              onChange={(e) => setExpertise(e.target.value)}
              placeholder="例如 research"
              className="!rounded-xs"
            />
          </label>

          <fieldset className="space-y-2 border-t border-black/[0.05] pt-4">
            <legend className="text-xs font-medium text-text-tertiary">IM 绑定（1 对多平台）</legend>
            <label className="block space-y-1.5">
              <span className="text-[11px] text-text-tertiary">Discord Bot</span>
              <SearchableSelect
                value={discordBotId}
                placeholder="不绑定（使用主持人 Bot 发言）"
                searchPlaceholder="搜索 Bot 名称或 ID…"
                emptyOption={{
                  value: '',
                  label: '不绑定（使用主持人 Bot 发言）',
                }}
                options={discordBots.map((bot) => ({
                  value: bot.id,
                  label: bot.label.trim() || bot.id,
                  hint: bot.id,
                }))}
                onChange={setDiscordBotId}
              />
            </label>
            <p className="text-[11px] text-text-tertiary">
              Telegram、Slack 等平台绑定即将支持
            </p>
          </fieldset>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
            取消
          </Button>
          <Button
            type="button"
            disabled={saving}
            onClick={() => void handleSubmit()}
            className={cn(hePressable, 'rounded-xl px-5')}
          >
            {saving ? '保存中…' : '保存'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function normalizeNameKey(name: string): string {
  const trimmed = name.trim()
  if (/^[\x00-\x7F]+$/.test(trimmed)) {
    return trimmed.toLowerCase()
  }
  return trimmed
}

function discordBindingFromParticipant(binds?: ParticipantIMBinding[]): string {
  const bind = binds?.find((b) => b.platform === 'discord')
  return bind?.application_id?.trim() || bind?.bot_id?.trim() || ''
}

function buildIMBindings(applicationId: string): ParticipantIMBinding[] {
  const id = applicationId.trim()
  if (!id) return []
  return [{ platform: 'discord', application_id: id }]
}
