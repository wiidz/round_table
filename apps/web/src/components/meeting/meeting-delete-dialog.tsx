import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import { hePressable } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

interface MeetingDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  topic: string
  meetingId: string
  deleting?: boolean
  onConfirm: () => void
}

export function MeetingDeleteDialog({
  open,
  onOpenChange,
  topic,
  meetingId,
  deleting = false,
  onConfirm,
}: MeetingDeleteDialogProps) {
  const { t } = useI18n()

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('meetingUi.delete.title')}</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>
                {t('meetingUi.delete.confirm', { topic })}
              </p>
              <p>{t('meetingUi.delete.warning')}</p>
              <p className="font-mono text-[11px] text-text-tertiary">{meetingId}</p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting}>{t('common.cancel')}</AlertDialogCancel>
          <Button
            type="button"
            variant="destructive"
            disabled={deleting}
            className={cn(hePressable, 'rounded-xl px-5')}
            onClick={onConfirm}
          >
            {deleting ? t('meetingUi.delete.deleting') : t('meetingUi.delete.confirmButton')}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
