const STORAGE_KEY = 'roundtable-settings-nav'

export type SettingsNavState = {
  tab: string
  subsection: string
  /** DiscordBotsPanel 侧栏 Bot id（如 moderator 或 Application ID） */
  discordBotId?: string
}

/** Legacy tab names (Chinese / old keys) → stable tab keys */
const LEGACY_TAB_MAP: Record<string, string> = {
  服务: 'service',
  存储: 'storage',
  LLM: 'llm',
  会议: 'meeting',
  IM: 'im',
  传输: 'im',
}

export function normalizeSettingsTab(tab: string): string {
  return LEGACY_TAB_MAP[tab] ?? tab
}

/** 设置 → IM → Discord Bot */
export const SETTINGS_IM_DISCORD: SettingsNavState = {
  tab: 'im',
  subsection: 'discord',
}

export function settingsNavForDiscordBot(botId: string): SettingsNavState {
  return {
    ...SETTINGS_IM_DISCORD,
    discordBotId: botId,
  }
}

/** 设置 → 服务 */
export const SETTINGS_SERVICE: SettingsNavState = {
  tab: 'service',
  subsection: '',
}

export function primeSettingsNav(state: SettingsNavState): void {
  writeSettingsNav(state)
}

export function readSettingsNav(): SettingsNavState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as SettingsNavState
    if (typeof parsed.tab !== 'string') return null
    const tab = normalizeSettingsTab(parsed.tab)
    return {
      tab,
      subsection: typeof parsed.subsection === 'string' ? parsed.subsection : '',
      discordBotId:
        typeof parsed.discordBotId === 'string' && parsed.discordBotId.trim()
          ? parsed.discordBotId.trim()
          : undefined,
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
