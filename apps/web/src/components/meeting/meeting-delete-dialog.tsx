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
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除会议</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>
                确定删除会议「<span className="font-medium text-text-primary">{topic}</span>」？
              </p>
              <p>将永久删除 workspace 文件与事件记录，此操作不可恢复。</p>
              <p className="font-mono text-[11px] text-text-tertiary">{meetingId}</p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting}>取消</AlertDialogCancel>
          <Button
            type="button"
            variant="destructive"
            disabled={deleting}
            className={cn(hePressable, 'rounded-xl px-5')}
            onClick={onConfirm}
          >
            {deleting ? '删除中…' : '确认删除'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
