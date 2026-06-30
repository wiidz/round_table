import { cn } from '@/lib/utils'

/** Speaker name under seat avatar — shared by round-table live seats and transcript list. */
export function transcriptSpeakerLabelClass({
  highlighted = false,
  focused = false,
  muted = false,
}: {
  highlighted?: boolean
  focused?: boolean
  muted?: boolean
}): string {
  return cn(
    'pointer-events-none max-w-[5.5rem] truncate text-center text-[10px] font-medium',
    highlighted && 'text-ai',
    focused && !highlighted && 'text-brand',
    muted && !highlighted && 'text-text-tertiary',
    !highlighted && !focused && !muted && 'text-text-secondary',
  )
}

/** Top padding when a #{n} badge sits on the bubble corner. */
export const transcriptBubbleBadgePaddingTop = '!pt-5'
