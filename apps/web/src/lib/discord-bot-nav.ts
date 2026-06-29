import type { DiscordBotState } from '@/types/settings'

export type DiscordBotTabKey = 'moderator' | `participant-${number}`

/** 将 DiscordBotState.id 解析为 DiscordBotsPanel 侧栏 Tab */
export function resolveDiscordBotTab(
  botId: string | undefined,
  bots: DiscordBotState[],
): DiscordBotTabKey {
  if (!botId || botId === 'moderator') {
    return 'moderator'
  }
  const participantBots = bots.filter((b) => b.deletable)
  const index = participantBots.findIndex(
    (b) => b.id === botId || b.discord_application_id === botId,
  )
  if (index >= 0) {
    return `participant-${index}`
  }
  return 'moderator'
}
