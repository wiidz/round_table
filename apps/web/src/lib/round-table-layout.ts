export type SeatKind = 'moderator' | 'principal' | 'participant'

export interface RosterSeatInput {
  id: string
  label: string
}

export interface SeatLayout {
  id: string
  kind: SeatKind
  label: string
  /** 0–100, container left edge */
  x: number
  /** 0–100, container top edge */
  y: number
  angleDeg: number
}

export interface ComputeSeatsOptions {
  radiusX?: number
  radiusY?: number
}

const MODERATOR_ANGLE = -90
const PRINCIPAL_ANGLE = 90

function angleToPercent(angleDeg: number, radiusX: number, radiusY: number): Pick<SeatLayout, 'x' | 'y'> {
  const rad = (angleDeg * Math.PI) / 180
  return {
    x: 50 + radiusX * Math.cos(rad),
    y: 50 + radiusY * Math.sin(rad),
  }
}

function spreadAngles(count: number, startDeg: number, endDeg: number): number[] {
  if (count <= 0) return []
  if (count === 1) return [(startDeg + endDeg) / 2]
  return Array.from({ length: count }, (_, i) => startDeg + ((endDeg - startDeg) * i) / (count - 1))
}

/** Distribute expert seats on left (140°–220°) and right (-40°–40°) arcs. */
export function participantAngles(count: number): number[] {
  if (count <= 0) return []

  const leftCount = Math.ceil(count / 2)
  const rightCount = count - leftCount
  const left = spreadAngles(leftCount, 140, 220)
  const right = spreadAngles(rightCount, -40, 40)

  const out: number[] = []
  for (let i = 0; i < count; i++) {
    if (i % 2 === 0) {
      out.push(left[Math.floor(i / 2)]!)
    } else {
      out.push(right[Math.floor(i / 2)]!)
    }
  }
  return out
}

/** Fixed moderator (12h) + principal (6h); experts on side arcs. */
export function computeRoundTableSeats(
  participants: RosterSeatInput[],
  options?: ComputeSeatsOptions,
): SeatLayout[] {
  const radiusX = options?.radiusX ?? 41
  const radiusY = options?.radiusY ?? 38

  const seats: SeatLayout[] = []

  const modPos = angleToPercent(MODERATOR_ANGLE, radiusX, radiusY)
  seats.push({
    id: 'moderator',
    kind: 'moderator',
    label: '司仪',
    angleDeg: MODERATOR_ANGLE,
    ...modPos,
  })

  const angles = participantAngles(participants.length)
  for (let i = 0; i < participants.length; i++) {
    const participant = participants[i]!
    const angle = angles[i]!
    seats.push({
      id: participant.id,
      kind: 'participant',
      label: participant.label,
      angleDeg: angle,
      ...angleToPercent(angle, radiusX, radiusY),
    })
  }

  const principalPos = angleToPercent(PRINCIPAL_ANGLE, radiusX, radiusY)
  seats.push({
    id: 'user',
    kind: 'principal',
    label: '我',
    angleDeg: PRINCIPAL_ANGLE,
    ...principalPos,
  })

  return seats
}

/** Bubble tail points toward avatar from the Live bubble. */
export type BubbleTail = 'left' | 'right' | 'top' | 'bottom'

export function seatBubbleTailClass(seat: SeatLayout): BubbleTail {
  // Moderator at 12h: avatar above bubble → tail at top of bubble (↑ toward avatar)
  if (seat.kind === 'moderator') return 'top'
  // Principal at 6h: avatar below bubble → tail at bottom of bubble (↓ toward avatar)
  if (seat.kind === 'principal') return 'bottom'
  return seat.x < 50 ? 'left' : 'right'
}

/** Top/bottom seats: Live bubble is vertical (avatar + bubble stacked). */
export function isPoleSeat(seat: SeatLayout): boolean {
  return seat.kind === 'moderator' || seat.kind === 'principal'
}

/** @deprecated use isPoleSeat */
export function isVerticalLiveSeat(seat: SeatLayout): boolean {
  return isPoleSeat(seat)
}

/** Avatar center is pinned to the seat coordinate. */
export function seatAnchorTransform(_seat: SeatLayout, _hasLive?: boolean): string {
  return 'translate(-50%, -50%)'
}

export type SeatSide = 'left' | 'right' | 'pole'

export function seatSide(seat: SeatLayout): SeatSide {
  if (isPoleSeat(seat)) return 'pole'
  return seat.x < 50 ? 'left' : 'right'
}

/** Absolute slot for a live bubble beside a fixed avatar. */
export function liveBubbleSlotClass(seat: SeatLayout, expanded: boolean): string {
  const width = expanded ? 'w-[min(38vw,26rem)]' : 'w-[10.5rem]'

  if (seat.kind === 'moderator') {
    return `absolute left-1/2 top-[calc(100%+0.375rem)] z-20 -translate-x-1/2 ${width}`
  }
  if (seat.kind === 'principal') {
    return `absolute bottom-[calc(100%+0.375rem)] left-1/2 z-20 -translate-x-1/2 ${width}`
  }
  if (seat.x < 50) {
    return `absolute left-[calc(100%+0.5rem)] top-1/2 z-20 -translate-y-1/2 ${width}`
  }
  return `absolute right-[calc(100%+0.5rem)] top-1/2 z-20 -translate-y-1/2 ${width}`
}
