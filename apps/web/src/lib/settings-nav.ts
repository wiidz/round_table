const STORAGE_KEY = 'roundtable-settings-nav'

export type SettingsNavState = {
  tab: string
  subsection: string
}

export function readSettingsNav(): SettingsNavState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as SettingsNavState
    if (typeof parsed.tab !== 'string') return null
    const tab = parsed.tab === '传输' ? 'IM' : parsed.tab
    return {
      tab,
      subsection: typeof parsed.subsection === 'string' ? parsed.subsection : '',
    }
  } catch {
    return null
  }
}

export function writeSettingsNav(state: SettingsNavState): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    // ignore quota / private mode
  }
}
