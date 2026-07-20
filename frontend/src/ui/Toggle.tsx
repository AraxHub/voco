import { useState, type ReactNode } from 'react'

interface HardwareToggleProps {
  icon: ReactNode
  iconOff?: ReactNode
  label: string
  active?: boolean
  onChange?: (active: boolean) => void
  size?: number
  danger?: boolean
}

export function HardwareToggle({
  icon,
  iconOff,
  label,
  active = true,
  onChange,
  size = 62,
  danger = false,
}: HardwareToggleProps) {
  const [pressed, setPressed] = useState(false)
  const [flash, setFlash] = useState(false)

  function handlePress() {
    setPressed(true)
    setFlash(true)
    setTimeout(() => setFlash(false), 600)
    setTimeout(() => {
      setPressed(false)
      onChange?.(!active)
    }, 140)
  }

  const activeColor = danger ? 'var(--voco-danger)' : 'var(--voco-toggle-on)'
  const activeCore = danger
    ? 'color-mix(in srgb, var(--voco-danger) 28%, transparent)'
    : 'var(--voco-toggle-on-core)'

  return (
    <div className="flex flex-col items-center gap-2">
      <button
        type="button"
        onPointerDown={handlePress}
        style={{
          width: size,
          height: size,
          borderRadius: '50%',
          background: active
            ? `radial-gradient(ellipse at 35% 35%, ${activeCore} 0%, var(--voco-card-strong) 70%)`
            : `radial-gradient(ellipse at 35% 35%, var(--voco-toggle-off) 0%, var(--voco-card-strong) 70%)`,
          border: `1px solid ${active ? activeColor : 'var(--voco-border)'}`,
          boxShadow: active
            ? `0 0 18px color-mix(in srgb, var(--voco-accent) 35%, transparent), inset 0 1px 0 rgba(255,255,255,0.2)`
            : 'inset 0 1px 4px rgba(0,0,0,0.12)',
          transform: pressed ? 'scale(0.92)' : 'scale(1)',
          transition: 'transform 0.14s cubic-bezier(0.22,1,0.36,1), box-shadow 0.2s ease, background 0.2s ease',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          outline: 'none',
          overflow: 'hidden',
        }}
      >
        <span
          style={{
            position: 'absolute',
            top: '12%',
            left: '20%',
            width: '40%',
            height: '28%',
            background: 'radial-gradient(ellipse, rgba(255,255,255,0.22) 0%, transparent 80%)',
            borderRadius: '50%',
            pointerEvents: 'none',
          }}
        />
        {flash && (
          <svg className="absolute inset-0 h-full w-full pointer-events-none rounded-full" overflow="visible">
            <circle
              cx="50%"
              cy="50%"
              r="48%"
              fill="none"
              stroke={activeColor}
              strokeWidth="1.5"
              strokeDasharray="10 200"
              strokeDashoffset="210"
              filter={`drop-shadow(0 0 5px ${activeColor})`}
              style={{ animation: 'circuit-lap 0.55s ease-out forwards' }}
            />
          </svg>
        )}
        <span
          style={{
            color: active ? (danger ? 'var(--voco-danger)' : 'var(--voco-accent)') : 'var(--voco-text-faint)',
            fontSize: size * 0.36,
            lineHeight: 1,
            transition: 'color 0.2s ease',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {active ? icon : (iconOff ?? icon)}
        </span>
      </button>
      <span
        style={{
          fontFamily: 'Outfit, sans-serif',
          fontSize: 11,
          fontWeight: 400,
          color: active ? 'var(--voco-text-muted)' : 'var(--voco-text-faint)',
          letterSpacing: '0.06em',
          textTransform: 'uppercase',
          transition: 'color 0.2s ease',
        }}
      >
        {label}
      </span>
    </div>
  )
}
