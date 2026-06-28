import { apiFetch } from '@/api/client'
import type {
  DiscordBotsUpdate,
  DiscordTransportLogs,
  DiscordTransportStatus,
  MeetCastConfig,
  MeetPresetConfig,
  SettingsResponse,
  SettingsValues,
} from '@/types/settings'

export function fetchSettings() {
  return apiFetch<SettingsResponse>('/settings')
}

export function saveSettings(values: SettingsValues) {
  return apiFetch<SettingsResponse>('/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ values }),
  })
}

export function saveDiscordBots(update: DiscordBotsUpdate) {
  return apiFetch<SettingsResponse>('/settings/discord-bots', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(update),
  })
}

export function saveMeetPresets(presets: MeetPresetConfig[]) {
  return apiFetch<SettingsResponse>('/settings/meet-presets', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ presets }),
  })
}

export function resetMeetPresets() {
  return apiFetch<SettingsResponse>('/settings/meet-presets/reset', {
    method: 'POST',
  })
}

export function saveMeetCasts(casts: MeetCastConfig[]) {
  return apiFetch<SettingsResponse>('/settings/meet-casts', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ casts }),
  })
}

export function resetMeetCasts() {
  return apiFetch<SettingsResponse>('/settings/meet-casts/reset', {
    method: 'POST',
  })
}

export function refreshDiscordBotProfiles() {
  return apiFetch<SettingsResponse>('/settings/discord-bots/refresh-profiles', {
    method: 'POST',
  })
}

export function fetchDiscordTransportStatus() {
  return apiFetch<DiscordTransportStatus>('/settings/discord-transport/status')
}

export function startDiscordTransport() {
  return apiFetch<DiscordTransportStatus>('/settings/discord-transport/start', {
    method: 'POST',
  })
}

export function stopDiscordTransport() {
  return apiFetch<DiscordTransportStatus>('/settings/discord-transport/stop', {
    method: 'POST',
  })
}

export function fetchDiscordTransportLogs(lines = 200) {
  return apiFetch<DiscordTransportLogs>(`/settings/discord-transport/logs?lines=${lines}`)
}

export function clearDiscordTransportLogs() {
  return apiFetch<DiscordTransportLogs>('/settings/discord-transport/logs/clear', {
    method: 'POST',
  })
}
