import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { SearchableSelect } from '@/components/settings/searchable-select'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { useI18n } from '@/hooks/use-i18n'
import { hePressable, heSubsectionTitleNeutral } from '@/lib/highend-styles'
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
  const { t } = useI18n()
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

  function validateClient(): ParticipantRosterInput | null {
    const trimmedId = id.trim()
    const trimmedName = displayName.trim()
    const trimmedExp = expertise.trim()

    if (!trimmedId) {
      toast.error(t('profile.form.error.idRequired'))
      return null
    }
    if (!ID_PATTERN.test(trimmedId)) {
      toast.error(t('profile.form.error.idPattern'))
      return null
    }
    if (trimmedId === 'moderator') {
      toast.error(t('profile.form.error.idReserved'))
      return null
    }
    if (!trimmedName) {
      toast.error(t('profile.form.error.nameRequired'))
      return null
    }

    const peerIds = peers.filter((p) => p.id !== initial?.id)
    if (peerIds.some((p) => p.id === trimmedId)) {
      toast.error(t('profile.form.error.idDuplicate', { id: trimmedId }))
      return null
    }

    const nameKey = normalizeNameKey(trimmedName)
    const dupName = peerIds.find(
      (p) => normalizeNameKey(p.display_name?.trim() || p.id) === nameKey,
    )
    if (dupName) {
      toast.error(
        t('profile.form.error.nameDuplicate', {
          name: dupName.display_name || dupName.id,
        }),
      )
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
      toast.error(err instanceof Error ? err.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  const discordNone = t('profile.form.discordNone')

  return (
    <Dialog
      open={open}
      onClose={saving ? undefined : onClose}
      closeOnOverlayClick={!saving}
      closeOnEscape={!saving}
    >
      <DialogContent
        size="sm"
        className="space-y-5"
        aria-labelledby="participant-form-title"
      >
        <DialogHeader>
          <DialogTitle id="participant-form-title">
            {mode === 'create' ? t('profile.form.createTitle') : t('profile.form.editTitle')}
          </DialogTitle>
          <DialogDescription>{t('profile.form.description')}</DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          <SettingsFieldRow
            label={t('profile.form.idLabel')}
            htmlFor="participant-form-id"
            required
            hint={t('profile.form.idHint')}
          >
            <Input
              id="participant-form-id"
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder={t('profile.form.idPlaceholder')}
              className="font-mono"
              autoComplete="off"
              readOnly={mode === 'edit'}
            />
          </SettingsFieldRow>

          <SettingsFieldRow
            label={t('profile.form.nameLabel')}
            htmlFor="participant-form-name"
            required
            hint={t('profile.form.nameHint')}
          >
            <Input
              id="participant-form-name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder={t('profile.form.namePlaceholder')}
            />
          </SettingsFieldRow>

          <SettingsFieldRow
            label={t('profile.form.expertiseLabel')}
            htmlFor="participant-form-expertise"
            hint={t('profile.form.expertiseHint')}
          >
            <Input
              id="participant-form-expertise"
              value={expertise}
              onChange={(e) => setExpertise(e.target.value)}
              placeholder="research"
            />
          </SettingsFieldRow>

          <fieldset className="space-y-4 border-t border-black/[0.05] pt-5">
            <legend className={cn(heSubsectionTitleNeutral, 'mb-1')}>
              {t('profile.form.imLegend')}
            </legend>
            <SettingsFieldRow
              label={t('profile.form.discordBotLabel')}
              htmlFor="participant-form-discord"
              hint={t('profile.form.discordBotHint')}
            >
              <SearchableSelect
                id="participant-form-discord"
                value={discordBotId}
                placeholder={discordNone}
                searchPlaceholder={t('profile.form.discordSearchPlaceholder')}
                emptyOption={{
                  value: '',
                  label: discordNone,
                }}
                options={discordBots.map((bot) => ({
                  value: bot.id,
                  label: bot.label.trim() || bot.id,
                  hint: bot.id,
                }))}
                onChange={setDiscordBotId}
              />
            </SettingsFieldRow>
            <p className="text-[11px] text-text-tertiary">{t('profile.form.imComingSoon')}</p>
          </fieldset>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
            {t('common.cancel')}
          </Button>
          <Button
            type="button"
            disabled={saving}
            onClick={() => void handleSubmit()}
            className={cn(hePressable, 'rounded-xl px-5')}
          >
            {saving ? t('common.saving') : t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
