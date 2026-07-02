import { describe, expect, it } from 'vitest'

import {
  parseMeetingParticipantsFromMessages,
  parseParticipantSummaryText,
  parseParticipantsFromMeetingMd,
  resolveMeetingLineup,
} from '@/lib/meeting-participants'
import type { ChatMessage } from '@/types/chat'

describe('parseParticipantSummaryText', () => {
  it('parses id·display pairs', () => {
    expect(parseParticipantSummaryText('designer·游戏策划, player·RO 老玩家代表')).toEqual([
      { id: 'designer', label: '游戏策划' },
      { id: 'player', label: 'RO 老玩家代表' },
    ])
  })
})

describe('parseParticipantsFromMeetingMd', () => {
  it('parses 参会人员 table', () => {
    const md = `## 参会人员

| 参会者 | 角色 | 专长 | 参会目标 |
|--------|------|------|----------|
| design | design | general | — |
| player | player | experience | — |
`
    expect(parseParticipantsFromMeetingMd(md)).toEqual([
      { id: 'design', label: 'design' },
      { id: 'player', label: 'player' },
    ])
  })
})

describe('parseMeetingParticipantsFromMessages', () => {
  it('reads lineup from setup confirm message', () => {
    const messages: ChatMessage[] = [
      {
        id: '1',
        role: 'moderator',
        content: '👥 designer·游戏策划, player·RO 老玩家代表',
        createdAt: 1,
      },
    ]
    const roster = [
      { id: 'designer', label: '游戏策划' },
      { id: 'player', label: 'RO 老玩家代表' },
    ]
    expect(parseMeetingParticipantsFromMessages(messages, roster)).toEqual(roster)
  })
})

describe('resolveMeetingLineup', () => {
  const roster = [
    { id: 'a', label: 'A' },
    { id: 'b', label: 'B' },
    { id: 'c', label: 'C' },
  ]
  const lineup = [
    { id: 'a', label: 'A' },
    { id: 'b', label: 'B' },
  ]

  it('seats full lineup when meeting is running', () => {
    expect(
      resolveMeetingLineup('running', {
        roster,
        meetingMdParticipants: lineup,
        messageParticipants: [],
        spokenParticipants: [{ id: 'a', label: 'A' }],
      }),
    ).toEqual(lineup)
  })

  it('seats confirmed lineup during setup', () => {
    expect(
      resolveMeetingLineup('setup', {
        roster,
        meetingMdParticipants: lineup,
        messageParticipants: lineup,
        spokenParticipants: [{ id: 'a', label: 'A' }],
      }),
    ).toEqual(lineup)
  })

  it('seats nobody during setup before lineup confirm', () => {
    expect(
      resolveMeetingLineup('setup', {
        roster,
        meetingMdParticipants: lineup,
        messageParticipants: [],
        spokenParticipants: [],
      }),
    ).toEqual([])
  })
})
