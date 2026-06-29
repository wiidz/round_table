import { useEffect, useMemo, useState } from 'react'

import { fetchParticipants } from '@/api/participants'
import { computeRoundTableSeats, type RosterSeatInput, type SeatLayout } from '@/lib/round-table-layout'
import type { ChatMessage } from '@/types/chat'

function participantsFromMessages(messages: ChatMessage[]): RosterSeatInput[] {
  const seen = new Map<string, string>()
  for (const message of messages) {
    if (message.role !== 'participant') continue
    const id = message.authorId?.trim()
    if (!id || seen.has(id)) continue
    seen.set(id, message.authorName?.trim() || id)
  }
  return [...seen.entries()].map(([id, label]) => ({ id, label }))
}

export function useRosterSeats(messages: ChatMessage[]) {
  const [roster, setRoster] = useState<RosterSeatInput[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetchParticipants()
      .then((response) => {
        if (cancelled) return
        const list = (response.participants ?? [])
          .filter((p) => p.in_roster !== false)
          .map((p) => ({
            id: p.id,
            label: p.display_name?.trim() || p.id,
          }))
        setRoster(list)
      })
      .catch(() => {
        if (!cancelled) setRoster([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const fallback = useMemo(() => participantsFromMessages(messages), [messages])

  const participants = roster.length > 0 ? roster : fallback

  const seats = useMemo(() => computeRoundTableSeats(participants), [participants])

  return { seats, participants, loading, rosterFromApi: roster.length > 0 }
}

export type { SeatLayout, RosterSeatInput }
