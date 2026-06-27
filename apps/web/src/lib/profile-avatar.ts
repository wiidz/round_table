/** Deterministic avatar tint from profile id */
const AVATAR_TONES = [
  'bg-ai/15 text-ai ring-ai/20',
  'bg-brand-soft text-brand ring-primary/25',
  'bg-emerald-500/10 text-emerald-700 ring-emerald-500/20',
  'bg-amber-500/10 text-amber-800 ring-amber-500/20',
  'bg-violet-500/10 text-violet-700 ring-violet-500/20',
  'bg-sky-500/10 text-sky-800 ring-sky-500/20',
] as const

function hashString(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = (h * 31 + s.charCodeAt(i)) | 0
  }
  return Math.abs(h)
}

export function profileAvatarTone(id: string): string {
  return AVATAR_TONES[hashString(id) % AVATAR_TONES.length]
}

/** Display initials for avatar fallback (名称首字/首字母). */
export function profileInitials(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) return '?'

  const words = trimmed.split(/\s+/).filter(Boolean)
  if (words.length >= 2 && /^[\x00-\x7F]+$/.test(trimmed)) {
    return (words[0].charAt(0) + words[1].charAt(0)).toUpperCase()
  }

  const chars = [...trimmed]
  if (chars.length >= 2 && /[\u4e00-\u9fff]/.test(chars[0])) {
    return chars.slice(0, 2).join('')
  }
  return chars[0].toUpperCase()
}
