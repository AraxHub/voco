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
      className={`gradient-breathe relative overflow-hidden select-none outline-none ${fullWidth ? 'w-full' : ''} ${className}`}
      style={{
        padding: '10px 20px',
        borderRadius: 10,
        background: 'var(--voco-primary-grad)',
        backgroundSize: '100% 100%',
        color: 'var(--voco-btn-fg)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 600,
        fontSize: 14,
        letterSpacing: '0.01em',
        boxShadow: pressed
          ? '0 1px 4px color-mix(in srgb, var(--voco-accent) 20%, transparent), inset 0 1px 3px rgba(0,0,0,0.1)'
          : 'var(--voco-primary-glow)',
        transform: pressed ? 'scale(0.98)' : 'scale(1)',
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
          borderRadius: 10,
        }}
      />
      <CircuitRim active={rimActive} color="rgba(255,255,255,0.7)" />
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
        padding: '10px 20px',
        borderRadius: 10,
        background: hovered
          ? 'color-mix(in srgb, var(--voco-accent) 10%, transparent)'
          : 'var(--voco-input-bg)',
        border: `1px solid ${hovered ? 'color-mix(in srgb, var(--voco-accent) 40%, transparent)' : 'var(--voco-border)'}`,
        color: hovered ? 'var(--voco-accent)' : 'var(--voco-text-muted)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 500,
        fontSize: 14,
        letterSpacing: '0.01em',
        transition: 'all 0.18s ease',
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
        padding: '10px 16px',
        borderRadius: 10,
        background: 'transparent',
        border: 'none',
        color: hovered ? 'var(--voco-accent)' : 'var(--voco-text-muted)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 500,
        fontSize: 14,
        letterSpacing: '0.01em',
        transition: 'color 0.18s ease',
        textDecoration: 'none',
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
        padding: '10px 20px',
        borderRadius: 10,
        background: 'var(--voco-danger-bg)',
        border: '1px solid var(--voco-danger-border)',
        color: 'var(--voco-danger)',
        fontFamily: 'Outfit, sans-serif',
        fontWeight: 500,
        fontSize: 14,
        letterSpacing: '0.01em',
        transform: pressed ? 'scale(0.98)' : 'scale(1)',
        transition: 'all 0.14s ease',
        boxShadow: 'none',
        opacity: disabled || loading ? 0.5 : 1,
        cursor: disabled || loading ? 'not-allowed' : 'pointer',
        ...props.style,
      }}
    >
      <CircuitRim active={rimActive} color="rgba(220,100,100,0.7)" />
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
