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

type PrincipalPersonaCreateDialogProps = {
  open: boolean
  creating?: boolean
  onClose: () => void
  onSubmit: (title: string) => Promise<void>
}

export function PrincipalPersonaCreateDialog({
  open,
  creating,
  onClose,
  onSubmit,
}: PrincipalPersonaCreateDialogProps) {
  const { t } = useI18n()
  const [title, setTitle] = useState('')

  useEffect(() => {
    if (!open) return
    setTitle('')
  }, [open])

  async function handleSubmit() {
    const trimmed = title.trim()
    if (!trimmed) return
    await onSubmit(trimmed)
  }

  return (
    <Dialog
      open={open}
      onClose={creating ? undefined : onClose}
      closeOnOverlayClick={!creating}
      closeOnEscape={!creating}
    >
      <DialogContent
        size="sm"
        className="space-y-5"
        aria-labelledby="principal-persona-create-title"
      >
        <DialogHeader>
          <DialogTitle id="principal-persona-create-title">
            {t('profile.principal.persona.newDialogTitle')}
          </DialogTitle>
          <DialogDescription>
            {t('profile.principal.persona.newDialogDescription')}
          </DialogDescription>
        </DialogHeader>

        <SettingsFieldRow
          label={t('profile.principal.persona.newTitleLabel')}
          htmlFor="principal-persona-create-title-input"
          required
          hint={t('profile.principal.persona.newTitleHint')}
        >
          <Input
            id="principal-persona-create-title-input"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={t('profile.principal.persona.newPlaceholder')}
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') void handleSubmit()
            }}
          />
        </SettingsFieldRow>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={creating}>
            {t('common.cancel')}
          </Button>
          <Button
            type="button"
            disabled={creating || !title.trim()}
            onClick={() => void handleSubmit()}
            className={cn(hePressable, 'rounded-xl px-5')}
          >
            {creating ? t('common.saving') : t('profile.principal.persona.create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
