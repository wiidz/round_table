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

/** Distribute expert seats on left (150°–210°) and right (-30°–30°) arcs. */
export function participantAngles(count: number): number[] {
  if (count <= 0) return []

  const leftCount = Math.ceil(count / 2)
  const rightCount = count - leftCount
  const left = spreadAngles(leftCount, 150, 210)
  const right = spreadAngles(rightCount, -30, 30)

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
  const radiusX = options?.radiusX ?? 32
  const radiusY = options?.radiusY ?? 30

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
  if (seat.kind === 'moderator' || seat.kind === 'principal') return 'right'
  return seat.x < 50 ? 'left' : 'right'
}

/** Top/bottom seats: Live bubble sits beside avatar (not above/below). */
export function isPoleSeat(seat: SeatLayout): boolean {
  return seat.kind === 'moderator' || seat.kind === 'principal'
}

/** @deprecated use isPoleSeat */
export function isVerticalLiveSeat(seat: SeatLayout): boolean {
  return isPoleSeat(seat)
}

/** Position anchor so Live bubbles grow inward, not off-screen. */
export function seatAnchorTransform(seat: SeatLayout, hasLive: boolean): string {
  if (!hasLive || isPoleSeat(seat)) return 'translate(-50%, -50%)'
  if (seat.x < 50) return 'translate(0, -50%)'
  return 'translate(-100%, -50%)'
}

/** Flex direction: pole seats use horizontal bubble beside avatar. */
export function seatContentLayoutClass(seat: SeatLayout, hasLive: boolean): string {
  if (!hasLive) return 'flex-col items-center gap-1'
  if (isPoleSeat(seat)) return 'flex-row-reverse items-start gap-2'
  if (seat.x < 50) return 'flex-row items-start gap-2'
  return 'flex-row-reverse items-start gap-2'
}
