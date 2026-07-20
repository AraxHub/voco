import { useEffect, useState } from 'react'

interface ToastProps {
  message: string
  visible: boolean
  onDone?: () => void
}

export function Toast({ message, visible, onDone }: ToastProps) {
  const [show, setShow] = useState(false)

  useEffect(() => {
    if (visible) {
      setShow(true)
      const t = setTimeout(() => { setShow(false); onDone?.() }, 2600)
      return () => clearTimeout(t)
    }
  }, [visible])

  if (!show) return null

  return (
    <div
      className="toast-in fixed z-50 pointer-events-none"
      style={{
        bottom: 110,
        left: '50%',
        transform: 'translateX(-50%)',
      }}
    >
      <div
        style={{
          padding: '10px 20px',
          borderRadius: 40,
          background: 'rgba(10,24,38,0.88)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(29,233,182,0.35)',
          color: '#c8eee6',
          fontFamily: 'Outfit, sans-serif',
          fontSize: 13,
          fontWeight: 500,
          letterSpacing: '0.04em',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          boxShadow: '0 4px 32px rgba(0,0,0,0.5), 0 0 18px rgba(29,233,182,0.1)',
          whiteSpace: 'nowrap',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Edge light run */}
        <svg
          className="absolute inset-0 w-full h-full rounded-full pointer-events-none"
          overflow="visible"
        >
          <rect
            x="0" y="0" width="100%" height="100%"
            rx="9999" ry="9999"
            fill="none"
            stroke="rgba(29,233,182,0.8)"
            strokeWidth="1.5"
            strokeDasharray="14 400"
            strokeDashoffset="414"
            filter="drop-shadow(0 0 4px rgba(29,233,182,0.5))"
            style={{ animation: 'circuit-lap 0.75s ease-out forwards' }}
          />
        </svg>
        <span style={{ color: '#1de9b6' }}>✓</span>
        {message}
      </div>
    </div>
  )
}
