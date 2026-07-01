import { Link } from 'react-router-dom'
import { Pencil, Trash2 } from 'lucide-react'

import { ProfileAvatar } from '@/components/profile/profile-avatar'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import {
  hePanelShell,
  hePanelShellHover,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { PARTICIPANT_STANDARD_FILES } from '@/lib/i18n/profile-labels'
import { cn } from '@/lib/utils'

import type { ParticipantIndex } from '@/types/participant'

export function ParticipantGridCard({
  participant: p,
  onEdit,
  onDelete,
}: {
  participant: ParticipantIndex
  onEdit: (p: ParticipantIndex) => void
  onDelete: (p: ParticipantIndex) => void
}) {
  const { t, profileFileCaption } = useI18n()
  const name = p.display_name?.trim() || p.id
  const discordBind = p.im_bindings?.find((b) => b.platform === 'discord')
  const discordBot = discordBind?.application_id || discordBind?.bot_id
  const fileCount = p.files.length
  const totalFiles = PARTICIPANT_STANDARD_FILES.length

  return (
    <article
      className={cn(
        hePanelShell,
        hePanelShellHover,
        heSpring,
        'flex h-full flex-col gap-4 p-5',
      )}
    >
      <div className="flex items-start gap-3">
        <Link
          to={`/participants/${encodeURIComponent(p.id)}`}
          className={cn('shrink-0', hePressable)}
        >
          <ProfileAvatar id={p.id} name={name} size="lg" />
        </Link>
        <div className="min-w-0 flex-1 pt-0.5">
          <Link
            to={`/participants/${encodeURIComponent(p.id)}`}
            className="block hover:text-brand"
          >
            <p className="truncate text-[15px] font-semibold tracking-[-0.02em] text-text-primary">
              {name}
            </p>
            <p className="mt-0.5 truncate font-mono text-[11px] text-text-tertiary">
              {p.id}
            </p>
            {p.expertise && (
              <p className="mt-1.5 truncate text-[12px] text-text-secondary">
                {p.expertise}
              </p>
            )}
            {discordBot && (
              <p className="mt-1 truncate font-mono text-[10px] text-text-tertiary">
                Discord · {discordBot}
              </p>
            )}
          </Link>
        </div>
        <div className="flex shrink-0 gap-0.5">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 rounded-lg text-text-tertiary hover:text-text-primary"
            aria-label={t('profile.filesEditor.editAria', { name })}
            onClick={() => onEdit(p)}
          >
            <Pencil className="size-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 rounded-lg text-text-tertiary hover:text-destructive"
            aria-label={t('profile.filesEditor.deleteAria', { name })}
            onClick={() => onDelete(p)}
          >
            <Trash2 className="size-3.5" />
          </Button>
        </div>
      </div>

      <Link
        to={`/participants/${encodeURIComponent(p.id)}`}
        className={cn('mt-auto block', hePressable)}
      >
        <div className="space-y-2 border-t border-black/[0.05] pt-3">
          <div className="flex flex-wrap gap-1.5">
            {PARTICIPANT_STANDARD_FILES.map((file) => {
              const present = p.files.includes(file)
              const short = file.replace('.md', '')
              return (
                <span
                  key={file}
                  title={profileFileCaption(file)}
                  className={cn(
                    'rounded-md px-2 py-0.5 font-mono text-[10px] ring-1 ring-inset',
                    present
                      ? 'bg-ai/8 text-ai ring-ai/15'
                      : 'bg-black/[0.03] text-text-tertiary ring-black/[0.06] line-through',
                  )}
                >
                  {short}
                </span>
              )
            })}
          </div>
          <p className="text-[11px] tabular-nums text-text-tertiary">
            {t('profile.filesEditor.filesCount', { present: fileCount, total: totalFiles })}
          </p>
        </div>
      </Link>
    </article>
  )
}

export function ParticipantGridSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: 8 }, (_, i) => (
        <div
          key={i}
          className={cn(hePanelShell, 'h-[168px] animate-pulse bg-black/[0.02] p-5')}
        />
      ))}
    </div>
  )
}
