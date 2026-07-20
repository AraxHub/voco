import { useState, type InputHTMLAttributes } from 'react'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function GlassInput({ label, error, className = '', ...props }: InputProps) {
  const [focused, setFocused] = useState(false)

  return (
    <div className={`flex flex-col gap-1.5 ${className}`}>
      {label && (
        <label
          style={{
            fontFamily: 'Outfit, sans-serif',
            fontSize: 12,
            fontWeight: 500,
            letterSpacing: '0.08em',
            color: 'var(--voco-text-muted)',
            textTransform: 'uppercase',
          }}
        >
          {label}
        </label>
      )}
      <div className="relative">
        <input
          {...props}
          onFocus={(e) => {
            setFocused(true)
            props.onFocus?.(e)
          }}
          onBlur={(e) => {
            setFocused(false)
            props.onBlur?.(e)
          }}
          style={{
            width: '100%',
            padding: '13px 16px',
            borderRadius: 10,
            background: focused ? 'var(--voco-input-bg-focus)' : 'var(--voco-input-bg)',
            border: `1px solid ${
              error ? 'var(--voco-error-border)' : focused ? 'var(--voco-input-border-focus)' : 'var(--voco-input-border)'
            }`,
            color: 'var(--voco-text)',
            fontFamily: 'Outfit, sans-serif',
            fontSize: 15,
            fontWeight: 400,
            outline: 'none',
            transition: 'all 0.2s ease',
            boxShadow: focused
              ? '0 0 0 3px color-mix(in srgb, var(--voco-accent) 12%, transparent), inset 0 1px 4px rgba(0,0,0,0.08)'
              : 'inset 0 1px 4px rgba(0,0,0,0.06)',
            backdropFilter: 'blur(12px)',
            ...props.style,
          }}
        />
        {focused && (
          <svg
            className="absolute pointer-events-none"
            style={{ top: -1, left: -1, width: 28, height: 28 }}
            overflow="visible"
          >
            <path
              d="M 0,14 L 0,0 L 14,0"
              fill="none"
              stroke="var(--voco-circuit)"
              strokeWidth="1.5"
              strokeDasharray="8 60"
              strokeDashoffset="68"
              style={{ animation: 'circuit-run 0.55s ease forwards' }}
              filter="drop-shadow(0 0 4px color-mix(in srgb, var(--voco-accent) 55%, transparent))"
            />
          </svg>
        )}
      </div>
      {error && (
        <p
          style={{
            fontFamily: 'Outfit, sans-serif',
            fontSize: 12,
            color: 'var(--voco-danger)',
            marginTop: 2,
          }}
        >
          {error}
        </p>
      )}
    </div>
  )
}
