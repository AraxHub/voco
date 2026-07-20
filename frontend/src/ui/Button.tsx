import { useState, type ReactNode, type ButtonHTMLAttributes } from 'react'
import { CircuitRim } from './CircuitTrace'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  loading?: boolean
  fullWidth?: boolean
}

export function PrimaryButton({ children, loading, fullWidth, disabled, className = '', ...props }: ButtonProps) {
  const [rimActive, setRimActive] = useState(false)
  const [pressed, setPressed] = useState(false)

  function handlePress() {
    setPressed(true)
    setRimActive(true)
    setTimeout(() => setRimActive(false), 1000)
    setTimeout(() => setPressed(false), 160)
  }

  return (
    <button
      {...props}
      disabled={disabled || loading}
      onPointerDown={handlePress}
      className={`relative overflow-hidden select-none outline-none ${fullWidth ? 'w-full' : ''} ${className}`}
      style={{
        padding: '14px 36px',
        borderRadius: 9999,
        background: 'var(--voco-primary-grad)',
        backgroundSize: '200% 200%',
        animation: 'gradient-pos 6s ease infinite',
        color: 'var(--voco-btn-fg)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 600,
        fontSize: 14,
        letterSpacing: '0.07em',
        boxShadow: pressed
          ? '0 2px 12px color-mix(in srgb, var(--voco-accent) 28%, transparent), inset 0 2px 6px rgba(0,0,0,0.12)'
          : 'var(--voco-primary-glow)',
        transform: pressed ? 'scale(0.975) translateY(1px)' : 'scale(1)',
        transition: 'transform 0.14s cubic-bezier(0.22,1,0.36,1), box-shadow 0.14s ease',
        opacity: disabled || loading ? 0.5 : 1,
        cursor: disabled || loading ? 'not-allowed' : 'pointer',
        border: 'none',
        ...props.style,
      }}
    >
      <span
        className="sheen-loop absolute inset-0 pointer-events-none"
        style={{
          background: 'var(--voco-sheen)',
          borderRadius: 9999,
        }}
      />
      <CircuitRim active={rimActive} color="rgba(255,255,255,0.95)" />
      {loading ? (
        <span className="flex items-center justify-center gap-2">
          <Spinner /> {children}
        </span>
      ) : (
        children
      )}
    </button>
  )
}

export function SecondaryButton({ children, loading, fullWidth, disabled, className = '', ...props }: ButtonProps) {
  const [hovered, setHovered] = useState(false)

  return (
    <button
      {...props}
      disabled={disabled || loading}
      onPointerEnter={() => setHovered(true)}
      onPointerLeave={() => setHovered(false)}
      className={`relative overflow-hidden select-none outline-none ${fullWidth ? 'w-full' : ''} ${className}`}
      style={{
        padding: '13px 36px',
        borderRadius: 9999,
        background: hovered
          ? 'color-mix(in srgb, var(--voco-accent) 10%, transparent)'
          : 'var(--voco-input-bg)',
        border: `1px solid ${hovered ? 'color-mix(in srgb, var(--voco-accent) 50%, transparent)' : 'var(--voco-border)'}`,
        color: hovered ? 'var(--voco-accent)' : 'var(--voco-text-muted)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 500,
        fontSize: 14,
        letterSpacing: '0.06em',
        transition: 'all 0.22s ease',
        opacity: disabled || loading ? 0.5 : 1,
        cursor: disabled || loading ? 'not-allowed' : 'pointer',
        ...props.style,
      }}
    >
      {loading ? (
        <span className="flex items-center justify-center gap-2">
          <Spinner color="var(--voco-accent)" /> {children}
        </span>
      ) : (
        children
      )}
    </button>
  )
}

export function GhostButton({ children, loading, fullWidth, disabled, className = '', ...props }: ButtonProps) {
  const [hovered, setHovered] = useState(false)

  return (
    <button
      {...props}
      disabled={disabled || loading}
      onPointerEnter={() => setHovered(true)}
      onPointerLeave={() => setHovered(false)}
      className={`select-none outline-none ${fullWidth ? 'w-full' : ''} ${className}`}
      style={{
        padding: '13px 36px',
        borderRadius: 9999,
        background: 'transparent',
        border: 'none',
        color: hovered ? 'var(--voco-text-muted)' : 'var(--voco-text-faint)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 400,
        fontSize: 14,
        letterSpacing: '0.04em',
        transition: 'color 0.18s ease',
        textDecoration: hovered ? 'underline' : 'none',
        textUnderlineOffset: 3,
        opacity: disabled || loading ? 0.5 : 1,
        cursor: disabled || loading ? 'not-allowed' : 'pointer',
        ...props.style,
      }}
    >
      {children}
    </button>
  )
}

/** Top-bar actions: always outlined so they stay visible on both themes */
export function NavButton({ children, loading, fullWidth, disabled, className = '', ...props }: ButtonProps) {
  return (
    <button
      {...props}
      disabled={disabled || loading}
      className={`voco-nav-btn ${fullWidth ? 'w-full' : ''} ${className}`}
      style={props.style}
    >
      {children}
    </button>
  )
}

export function DangerButton({ children, loading, fullWidth, disabled, className = '', ...props }: ButtonProps) {
  const [rimActive, setRimActive] = useState(false)
  const [pressed, setPressed] = useState(false)

  function handlePress() {
    setPressed(true)
    setRimActive(true)
    setTimeout(() => setRimActive(false), 900)
    setTimeout(() => setPressed(false), 140)
  }

  return (
    <button
      {...props}
      disabled={disabled || loading}
      onPointerDown={handlePress}
      className={`relative overflow-hidden select-none outline-none ${fullWidth ? 'w-full' : ''} ${className}`}
      style={{
        padding: '13px 36px',
        borderRadius: 9999,
        background: 'var(--voco-danger-bg)',
        border: '1px solid var(--voco-danger-border)',
        color: 'var(--voco-danger)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 500,
        fontSize: 14,
        letterSpacing: '0.05em',
        transform: pressed ? 'scale(0.976)' : 'scale(1)',
        transition: 'all 0.14s ease',
        boxShadow: '0 0 18px color-mix(in srgb, var(--voco-danger) 20%, transparent)',
        opacity: disabled || loading ? 0.5 : 1,
        cursor: disabled || loading ? 'not-allowed' : 'pointer',
        ...props.style,
      }}
    >
      <CircuitRim active={rimActive} color="rgba(220,100,100,0.9)" />
      {loading ? (
        <span className="flex items-center justify-center gap-2">
          <Spinner color="var(--voco-danger)" /> {children}
        </span>
      ) : (
        children
      )}
    </button>
  )
}

function Spinner({ color = 'var(--voco-btn-fg)' }: { color?: string }) {
  return (
    <span
      style={{
        display: 'inline-block',
        width: 14,
        height: 14,
        border: `2px solid color-mix(in srgb, ${color} 30%, transparent)`,
        borderTopColor: color,
        borderRadius: '50%',
        animation: 'spin-slow 0.7s linear infinite',
      }}
    />
  )
}
