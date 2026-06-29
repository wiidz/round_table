import { describe, expect, it } from 'vitest'

import {
  computeRoundTableSeats,
  participantAngles,
  seatAnchorTransform,
  seatBubbleTailClass,
  seatContentLayoutClass,
} from '@/lib/round-table-layout'

describe('participantAngles', () => {
  it('returns empty for zero experts', () => {
    expect(participantAngles(0)).toEqual([])
  })

  it('places single expert on left arc', () => {
    expect(participantAngles(1)).toEqual([180])
  })

  it('alternates left and right for multiple experts', () => {
    const angles = participantAngles(4)
    expect(angles).toHaveLength(4)
    expect(angles[0]).toBeGreaterThan(90)
    expect(angles[1]).toBeLessThan(90)
  })
})

describe('computeRoundTableSeats', () => {
  it('includes moderator, experts, and principal', () => {
    const seats = computeRoundTableSeats([
      { id: 'dev', label: '开发' },
      { id: 'design', label: '方案' },
    ])
    expect(seats.map((s) => s.kind)).toEqual([
      'moderator',
      'participant',
      'participant',
      'principal',
    ])
    expect(seats.find((s) => s.kind === 'moderator')?.label).toBe('司仪')
    expect(seats.find((s) => s.kind === 'principal')?.label).toBe('我')
  })
})

describe('seatBubbleTailClass', () => {
  it('uses horizontal tail for pole seats', () => {
    const principal = computeRoundTableSeats([]).find((s) => s.kind === 'principal')!
    const moderator = computeRoundTableSeats([]).find((s) => s.kind === 'moderator')!
    expect(seatBubbleTailClass(moderator)).toBe('right')
    expect(seatBubbleTailClass(principal)).toBe('right')
  })
})

describe('seatAnchorTransform', () => {
  it('anchors side seats at outer edge when live', () => {
    const seats = computeRoundTableSeats([{ id: 'a', label: 'A' }, { id: 'b', label: 'B' }])
    const left = seats.find((s) => s.kind === 'participant' && s.x < 50)!
    const right = seats.find((s) => s.kind === 'participant' && s.x >= 50)!
    expect(seatAnchorTransform(left, true)).toBe('translate(0, -50%)')
    expect(seatAnchorTransform(right, true)).toBe('translate(-100%, -50%)')
    expect(seatAnchorTransform(left, false)).toBe('translate(-50%, -50%)')
  })
})

describe('seatContentLayoutClass', () => {
  it('places bubble inward on left arc', () => {
    const seats = computeRoundTableSeats([{ id: 'a', label: 'A' }])
    const left = seats.find((s) => s.kind === 'participant')!
    expect(seatContentLayoutClass(left, true)).toContain('flex-row')
    expect(seatContentLayoutClass(left, true)).not.toContain('reverse')
  })
})
