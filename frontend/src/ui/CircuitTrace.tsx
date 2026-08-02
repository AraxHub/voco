import { useEffect, useState } from 'react'
import { useFancyEffects } from '../hooks/useFancyEffects'

interface CircuitCornersProps {
  armLength?: number
  color?: string
  active?: boolean
  interval?: number
  className?: string
}

/** Corner PCB traces with occasional “train on rails” pulse. Uses viewBox % coords (valid SVG). */
export function CircuitCorners({
  armLength = 8,
  color = 'var(--voco-circuit)',
  active = false,
  interval = 4200,
  className = '',
}: CircuitCornersProps) {
  const fancy = useFancyEffects()
  const [running, setRunning] = useState(active)
  const a = Math.min(armLength, 24)

  useEffect(() => {
    if (active) {
      setRunning(true)
      return
    }
    // Idle PCB pulses only on capable desktops — interval re-renders heat weak GPUs.
    if (!fancy) return

    const id = setInterval(() => {
      setRunning(true)
      setTimeout(() => setRunning(false), 1400)
    }, interval + Math.random() * 2000)
    return () => clearInterval(id)
  }, [active, interval, fancy])

  const corners = [
    `M 0,${a} L 0,0 L ${a},0`,
    `M ${100 - a},0 L 100,0 L 100,${a}`,
    `M 0,${100 - a} L 0,100 L ${a},100`,
    `M ${100 - a},100 L 100,100 L 100,${100 - a}`,
  ]

  const glow = 'color-mix(in srgb, var(--voco-accent) 55%, transparent)'
  const base = 'color-mix(in srgb, var(--voco-accent) 22%, transparent)'

  return (
    <svg
      className={`absolute inset-0 h-full w-full pointer-events-none ${className}`}
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      overflow="visible"
      aria-hidden
    >
      {corners.map((d, i) => (
        <path key={`b${i}`} d={d} fill="none" stroke={base} strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
      ))}
      {running &&
        corners.map((d, i) => (
          <path
            key={`t${i}`}
            d={d}
            fill="none"
            stroke={color}
            strokeWidth="1.1"
            vectorEffect="non-scaling-stroke"
            strokeDasharray="8 40"
            strokeDashoffset={48}
            filter={`drop-shadow(0 0 3px ${glow})`}
            style={{
              animation: 'circuit-run 1.1s cubic-bezier(0.4,0,0.2,1) forwards',
              animationDelay: `${i * 0.08}s`,
            }}
          />
        ))}
    </svg>
  )
}

interface CircuitRimProps {
  active?: boolean
  color?: string
}

export function CircuitRim({ active = false, color = 'var(--voco-circuit)' }: CircuitRimProps) {
  return (
    <svg className="pointer-events-none absolute inset-0 h-full w-full rounded-full" overflow="visible" aria-hidden>
      <rect
        x="0"
        y="0"
        width="100%"
        height="100%"
        rx="9999"
        ry="9999"
        fill="none"
        stroke="color-mix(in srgb, var(--voco-accent) 25%, transparent)"
        strokeWidth="1"
        vectorEffect="non-scaling-stroke"
      />
      {active && (
        <rect
          x="0"
          y="0"
          width="100%"
          height="100%"
          rx="9999"
          ry="9999"
          fill="none"
          stroke={color}
          strokeWidth="1.5"
          vectorEffect="non-scaling-stroke"
          strokeDasharray="14 400"
          strokeDashoffset="414"
          filter="drop-shadow(0 0 5px color-mix(in srgb, var(--voco-accent) 55%, transparent))"
          style={{ animation: 'circuit-lap 0.9s cubic-bezier(0.4,0,0.2,1) forwards' }}
        />
      )}
    </svg>
  )
}
