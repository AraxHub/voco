import { type CSSProperties, type ReactNode } from 'react'
import { CircuitCorners } from './CircuitTrace'

interface MineralCardProps {
  children: ReactNode
  className?: string
  withCircuit?: boolean
  circuitInterval?: number
  style?: CSSProperties
}

export function MineralCard({
  children,
  className = '',
  withCircuit = false,
  circuitInterval = 5000,
  style,
}: MineralCardProps) {
  return (
    <div
      className={`relative ${className}`}
      style={{
        background: 'var(--voco-card)',
        backdropFilter: 'blur(28px)',
        WebkitBackdropFilter: 'blur(28px)',
        border: '1px solid var(--voco-border)',
        borderRadius: 20,
        boxShadow: 'var(--voco-shadow), inset 0 1px 0 rgba(255,255,255,0.08)',
        ...style,
      }}
    >
      {withCircuit && <CircuitCorners armLength={8} color="var(--voco-circuit)" interval={circuitInterval} />}
      {children}
    </div>
  )
}

interface StatusMessageProps {
  type: 'error' | 'success'
  children: ReactNode
}

export function StatusMessage({ type, children }: StatusMessageProps) {
  const isError = type === 'error'
  return (
    <div
      style={{
        padding: '11px 16px',
        borderRadius: 10,
        background: isError ? 'var(--voco-error-bg)' : 'var(--voco-success-bg)',
        border: `1px solid ${isError ? 'var(--voco-error-border)' : 'var(--voco-success-border)'}`,
        color: isError ? 'var(--voco-danger)' : 'var(--voco-success)',
        fontFamily: 'Outfit, sans-serif',
        fontSize: 13,
        fontWeight: 400,
        lineHeight: 1.5,
        display: 'flex',
        alignItems: 'flex-start',
        gap: 10,
      }}
    >
      <span style={{ flexShrink: 0, marginTop: 1 }}>{isError ? '⚠' : '✓'}</span>
      <span>{children}</span>
    </div>
  )
}
