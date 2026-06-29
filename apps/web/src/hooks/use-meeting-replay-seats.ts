import { useEffect, useMemo, useState } from 'react'

import { fetchParticipants } from '@/api/participants'
import {
  parseParticipantsFromMeetingMd,
  participantsFromMessages,
  resolveMeetingLineup,
} from '@/lib/meeting-participants'
import { computeRoundTableSeats, type RosterSeatInput } from '@/lib/round-table-layout'
import type { ChatMessage } from '@/types/chat'

function enrichLabels(lineup: RosterSeatInput[], roster: RosterSeatInput[]): RosterSeatInput[] {
  if (lineup.length === 0) return []
  const rosterById = new Map(roster.map((p) => [p.id, p.label]))
  return lineup.map((p) => ({
    id: p.id,
    label: rosterById.get(p.id) ?? p.label,
  }))
}

/** Roster for archived meeting replay (MEETING.md + transcript messages). */
export function useMeetingReplaySeats(
  meetingMd: string,
  messages: ChatMessage[],
) {
  const [roster, setRoster] = useState<RosterSeatInput[]>([])
  const [loading, setLoading] = useState(true)

  const meetingLineup = useMemo(
    () => parseParticipantsFromMeetingMd(meetingMd),
    [meetingMd],
  )

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

  const spoken = useMemo(() => participantsFromMessages(messages), [messages])

  const participants = useMemo(() => {
    const lineup = resolveMeetingLineup('post', {
      roster,
      meetingMdParticipants: meetingLineup,
      messageParticipants: [],
      spokenParticipants: spoken,
    })
    return enrichLabels(lineup, roster)
  }, [roster, meetingLineup, spoken])

  const seats = useMemo(() => computeRoundTableSeats(participants), [participants])

  return {
    seats,
    participants,
    rosterTotal: roster.length,
    loading,
    rosterFromApi: roster.length > 0,
  }
}
