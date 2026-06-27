const SIZE = 48
const RADIUS = 4
const PERIMETER = 185
const DURATION = '1.3s'

/** All hues stay in the bright-green band — no dark/muddy strokes. */
const RING_TAIL = 'var(--dr-ring-tail, #45dea0)'
const RING_MID = 'var(--dr-ring-mid, #5ef0a8)'
const RING_HEAD = 'var(--dr-ring-head, #7dffc8)'
const RING_TRACK = 'var(--dr-track, rgba(47, 182, 124, 0.14))'

/** Closed rounded-rect orbit; pathLength keeps dash phase seamless at the seam. */
const ORBIT_PATH = `M ${RADIUS} 0 H ${SIZE - RADIUS} A ${RADIUS} ${RADIUS} 0 0 1 ${SIZE} ${RADIUS} V ${SIZE - RADIUS} A ${RADIUS} ${RADIUS} 0 0 1 ${SIZE - RADIUS} ${SIZE} H ${RADIUS} A ${RADIUS} ${RADIUS} 0 0 1 0 ${SIZE - RADIUS} V ${RADIUS} A ${RADIUS} ${RADIUS} 0 0 1 ${RADIUS} 0 Z`

function CometLayer({
  len,
  opacity,
  width,
  color,
  glow,
}: {
  len: number
  opacity: number
  width: number
  color: string
  glow?: boolean
}) {
  return (
    <path
      d={ORBIT_PATH}
      pathLength={PERIMETER}
      fill="none"
      stroke={color}
      strokeWidth={width}
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity={opacity}
      strokeDasharray={`${len} ${PERIMETER - len}`}
      filter={glow ? 'url(#discord-ring-glow)' : undefined}
    >
      <animate
        attributeName="stroke-dashoffset"
        from="0"
        to={`${PERIMETER}`}
        dur={DURATION}
        repeatCount="indefinite"
      />
    </path>
  )
}

type CometSpec = { len: number; opacity: number; width: number; color: string; glow?: boolean }

function cometLayers(emphasis: boolean): CometSpec[] {
  const boost = emphasis ? 1.15 : 1
  return [
    { len: 56, opacity: 0.14 * boost, width: 3.2, color: RING_TAIL },
    { len: 42, opacity: 0.28 * boost, width: 2.9, color: RING_MID },
    { len: 28, opacity: 0.48 * boost, width: 2.7, color: RING_MID },
    { len: 18, opacity: 0.72 * boost, width: 3.1, color: RING_HEAD },
    { len: 12, opacity: 0.95 * boost, width: 3.6, color: RING_HEAD, glow: true },
  ]
}

export function DiscordRunningRing({ emphasis = false }: { emphasis?: boolean }) {
  return (
    <svg
      className="pointer-events-none absolute inset-0 z-10 size-full overflow-visible motion-reduce:hidden"
      viewBox={`0 0 ${SIZE} ${SIZE}`}
      aria-hidden
    >
      <defs>
        <filter id="discord-ring-glow" x="-60%" y="-60%" width="220%" height="220%">
          <feGaussianBlur stdDeviation="1.4" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      <rect
        x="0.5"
        y="0.5"
        width={SIZE - 1}
        height={SIZE - 1}
        rx={RADIUS}
        ry={RADIUS}
        fill="none"
        stroke={RING_TRACK}
        strokeWidth="1"
      />

      {cometLayers(emphasis).map((layer) => (
        <CometLayer key={layer.len} {...layer} />
      ))}
    </svg>
  )
}
