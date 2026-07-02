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

type BriefTemplateCreateDialogProps = {
  open: boolean
  onClose: () => void
  onSubmit: (title: string) => void
}

export function BriefTemplateCreateDialog({
  open,
  onClose,
  onSubmit,
}: BriefTemplateCreateDialogProps) {
  const { t } = useI18n()
  const [title, setTitle] = useState('')

  useEffect(() => {
    if (!open) return
    setTitle('')
  }, [open])

  function handleSubmit() {
    const trimmed = title.trim()
    if (!trimmed) return
    onSubmit(trimmed)
  }

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogContent
        size="sm"
        className="space-y-5"
        aria-labelledby="brief-template-create-title"
      >
        <DialogHeader>
          <DialogTitle id="brief-template-create-title">
            {t('brief.page.createDialogTitle')}
          </DialogTitle>
          <DialogDescription>{t('brief.page.createDialogDescription')}</DialogDescription>
        </DialogHeader>

        <SettingsFieldRow
          label={t('brief.meta.titleLabel')}
          htmlFor="brief-template-create-title-input"
          required
          hint={t('brief.page.createTitleHint')}
        >
          <Input
            id="brief-template-create-title-input"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={t('brief.meta.titlePlaceholder')}
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
            disabled={!title.trim()}
            onClick={handleSubmit}
            className={cn(hePressable, 'rounded-xl px-5')}
          >
            {t('brief.page.create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
