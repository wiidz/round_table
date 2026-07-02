import { useEffect, useState } from 'react'

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
import { hePressable } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

type DiscordBotCreateDialogProps = {
  open: boolean
  onClose: () => void
  onSubmit: (displayName: string) => void
}

export function DiscordBotCreateDialog({
  open,
  onClose,
  onSubmit,
}: DiscordBotCreateDialogProps) {
  const { t } = useI18n()
  const [displayName, setDisplayName] = useState('')

  useEffect(() => {
    if (!open) return
    setDisplayName('')
  }, [open])

  function handleSubmit() {
    const trimmed = displayName.trim()
    if (!trimmed) return
    onSubmit(trimmed)
  }

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogContent
        size="sm"
        className="space-y-5"
        aria-labelledby="discord-bot-create-title"
      >
        <DialogHeader>
          <DialogTitle id="discord-bot-create-title">
            {t('settings.discord.addBotDialogTitle')}
          </DialogTitle>
          <DialogDescription>
            {t('settings.discord.addBotDialogDescription')}
          </DialogDescription>
        </DialogHeader>

        <SettingsFieldRow
          label={t('settings.discord.addBotNameLabel')}
          htmlFor="discord-bot-create-name"
          required
          hint={t('settings.discord.addBotNameHint')}
        >
          <Input
            id="discord-bot-create-name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder={t('settings.discord.addBotNamePlaceholder')}
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSubmit()
            }}
          />
        </SettingsFieldRow>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            type="button"
            disabled={!displayName.trim()}
            onClick={handleSubmit}
            className={cn(hePressable, 'rounded-xl px-5')}
          >
            {t('settings.discord.addBot')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
