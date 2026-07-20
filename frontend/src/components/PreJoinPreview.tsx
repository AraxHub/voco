import type { RefObject } from 'react'
import { CircuitCorners } from '../ui/CircuitTrace'
import { HardwareToggle } from '../ui/Toggle'

type Props = {
  videoRef: RefObject<HTMLVideoElement | null>
  cameraOn: boolean
  micOn: boolean
  hasPreview: boolean
  onToggleCamera: () => void
  onToggleMic: () => void
}

export function PreJoinPreview({
  videoRef,
  cameraOn,
  micOn,
  hasPreview,
  onToggleCamera,
  onToggleMic,
}: Props) {
  return (
    <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: 28 }}>
      <div
        style={{
          position: 'relative',
          width: '100%',
          paddingBottom: '56.25%',
          borderRadius: 16,
          overflow: 'hidden',
          background: 'var(--voco-bg-soft)',
          border: '1px solid var(--voco-border)',
          boxShadow: 'var(--voco-shadow)',
        }}
      >
        <div style={{ position: 'absolute', inset: 0 }}>
          {cameraOn && hasPreview ? (
            <video
              ref={videoRef}
              autoPlay
              playsInline
              muted
              style={{ width: '100%', height: '100%', objectFit: 'cover', transform: 'scaleX(-1)' }}
            />
          ) : (
            <div
              style={{
                width: '100%',
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 16,
                background: cameraOn
                  ? 'radial-gradient(ellipse at 40% 40%, color-mix(in srgb, var(--voco-accent) 10%, transparent) 0%, var(--voco-bg) 80%)'
                  : undefined,
              }}
            >
              <div
                style={{
                  width: 72,
                  height: 72,
                  borderRadius: '50%',
                  background:
                    'radial-gradient(ellipse at 35% 35%, color-mix(in srgb, var(--voco-accent) 22%, transparent) 0%, var(--voco-card-strong) 70%)',
                  border: '1px solid color-mix(in srgb, var(--voco-accent) 35%, transparent)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 28,
                }}
              >
                👤
              </div>
              <p
                style={{
                  fontFamily: 'Outfit, sans-serif',
                  fontSize: 13,
                  color: 'var(--voco-text-faint)',
                  letterSpacing: '0.04em',
                  margin: 0,
                }}
              >
                {cameraOn ? 'Нет доступа к камере' : 'Камера выключена'}
              </p>
            </div>
          )}

          <CircuitCorners armLength={8} color="var(--voco-circuit)" interval={4000} />

          {micOn && (
            <div
              style={{
                position: 'absolute',
                bottom: 14,
                left: 14,
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '4px 10px',
                borderRadius: 20,
                background: 'color-mix(in srgb, var(--voco-bg) 70%, transparent)',
                backdropFilter: 'blur(8px)',
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: 'var(--voco-accent)',
                  boxShadow: '0 0 8px color-mix(in srgb, var(--voco-accent) 80%, transparent)',
                }}
              />
              <span style={{ fontFamily: 'Outfit, sans-serif', fontSize: 11, color: 'var(--voco-text-muted)' }}>
                Микро активно
              </span>
            </div>
          )}
        </div>
      </div>

      <div style={{ display: 'flex', gap: 24, justifyContent: 'center' }}>
        <HardwareToggle
          icon={<MicIcon />}
          iconOff={<MicOffIcon />}
          label="Микрофон"
          active={micOn}
          onChange={() => onToggleMic()}
          size={58}
        />
        <HardwareToggle
          icon={<CamIcon />}
          iconOff={<CamOffIcon />}
          label="Камера"
          active={cameraOn}
          onChange={() => onToggleCamera()}
          size={58}
        />
      </div>
    </div>
  )
}

const MicIcon = () => (
  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
    <rect x="9" y="2" width="6" height="12" rx="3" />
    <path d="M5 10a7 7 0 0 0 14 0" strokeLinecap="round" />
    <line x1="12" y1="17" x2="12" y2="21" strokeLinecap="round" />
    <line x1="9" y1="21" x2="15" y2="21" strokeLinecap="round" />
  </svg>
)
const MicOffIcon = () => (
  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
    <line x1="2" y1="2" x2="22" y2="22" />
    <path d="M12 2a3 3 0 0 1 3 3v5a3 3 0 0 1-.03.44" />
    <path d="M9 9v3a3 3 0 0 0 5.12 2.12" />
    <path d="M5 10a7 7 0 0 0 12 4.9" />
    <line x1="12" y1="17" x2="12" y2="21" />
    <line x1="9" y1="21" x2="15" y2="21" />
  </svg>
)
const CamIcon = () => (
  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
    <rect x="2" y="6" width="14" height="12" rx="2" />
    <polygon points="22,7 16,12 22,17" />
  </svg>
)
const CamOffIcon = () => (
  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
    <line x1="2" y1="2" x2="22" y2="22" />
    <path d="M10.66 6H14a2 2 0 0 1 2 2v2.34l1 1L22 7v10" />
    <path d="M16 16a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8" />
  </svg>
)
