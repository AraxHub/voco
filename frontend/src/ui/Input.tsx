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
            padding: '11px 14px',
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
            transition: 'border-color 0.18s ease, background 0.18s ease',
            boxShadow: 'none',
            ...props.style,
          }}
        />
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
