import { describe, expect, it } from 'vitest'

import { assignTurnForRole } from '@/lib/assign-turn'

describe('assignTurnForRole', () => {
  it('assigns incrementing turns for moderator and participant', () => {
    let next = 1
    const mod = assignTurnForRole('moderator', next)
    expect(mod.turn).toBe(1)
    next = mod.nextTurn

    const expert = assignTurnForRole('participant', next)
    expect(expert.turn).toBe(2)
    next = expert.nextTurn

    const user = assignTurnForRole('user', next)
    expect(user.turn).toBe(3)
    expect(user.nextTurn).toBe(4)
  })

  it('skips turn for system messages', () => {
    const result = assignTurnForRole('system', 5)
    expect(result).toEqual({ nextTurn: 5 })
  })
})
