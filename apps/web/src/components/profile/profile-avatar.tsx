import { cn } from '@/lib/utils'
import { profileAvatarTone, profileInitials } from '@/lib/profile-avatar'

type ProfileAvatarProps = {
  id: string
  name: string
  avatarUrl?: string
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const sizeClass = {
  sm: 'size-10 text-sm',
  md: 'size-12 text-base',
  lg: 'size-14 text-lg',
} as const

export function ProfileAvatar({
  id,
  name,
  avatarUrl,
  size = 'md',
  className,
}: ProfileAvatarProps) {
  const initials = profileInitials(name)

  return (
    <span
      className={cn(
        'relative flex shrink-0 items-center justify-center overflow-hidden rounded-xl ring-1 ring-inset',
        sizeClass[size],
        !avatarUrl && profileAvatarTone(id),
        className,
      )}
      aria-hidden
    >
      {avatarUrl ? (
        <img src={avatarUrl} alt="" className="size-full object-cover" />
      ) : (
        <span className="font-semibold tracking-tight">{initials}</span>
      )}
    </span>
  )
}
