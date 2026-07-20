import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { PrimaryButton, SecondaryButton, GhostButton } from '../ui/Button'
import { MineralCard, StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'

type HomeProps = { mode: 'home' }
type RoomProps = { mode: 'room'; onGuest: () => void }
type Props = HomeProps | RoomProps

function extractRoomId(input: string): string | null {
  const raw = input.trim()
  if (!raw) return null
  try {
    const u = new URL(raw)
    const m = /^\/room\/([^/]+)$/.exec(u.pathname)
    if (m?.[1]) return m[1]
  } catch {
    // not a URL
  }
  return raw
}

export function AuthGate(props: Props) {
  const auth = useAuth()
  const nav = useNavigate()
  const [inviteInput, setInviteInput] = useState('')
  const [joinMode, setJoinMode] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const redirectUri = typeof window !== 'undefined' ? window.location.href : undefined

  function onJoinByLink(e?: FormEvent) {
    e?.preventDefault()
    setError(null)
    const id = extractRoomId(inviteInput)
    if (!id) {
      setError('Вставь ссылку вида /room/<id> или roomId')
      return
    }
    nav(`/room/${id}`)
  }

  return (
    <div
      style={{
        position: 'relative',
        zIndex: 1,
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '32px 16px',
      }}
    >
      <MineralCard
        withCircuit
        circuitInterval={props.mode === 'room' ? 4800 : 5500}
        style={{ width: '100%', maxWidth: 440, padding: '52px 48px 44px' }}
      >
        {props.mode === 'room' && (
          <div style={{ textAlign: 'center', marginBottom: 20 }}>
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
                padding: '5px 14px',
                borderRadius: 20,
                background: 'color-mix(in srgb, var(--voco-accent) 10%, transparent)',
                border: '1px solid color-mix(in srgb, var(--voco-accent) 28%, transparent)',
                fontFamily: 'Outfit, sans-serif',
                fontSize: 11,
                fontWeight: 500,
                color: 'color-mix(in srgb, var(--voco-accent) 85%, transparent)',
                letterSpacing: '0.08em',
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  background: 'var(--voco-accent)',
                  boxShadow: '0 0 6px color-mix(in srgb, var(--voco-accent) 80%, transparent)',
                }}
              />
              КОМНАТА АКТИВНА
            </span>
          </div>
        )}

        <div style={{ textAlign: 'center', marginBottom: 10 }}>
          <h1
            style={{
              fontFamily: 'Syncopate, sans-serif',
              fontWeight: 700,
              fontSize: props.mode === 'room' ? 46 : 54,
              letterSpacing: '0.18em',
              color: 'transparent',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              backgroundImage: 'var(--voco-hero-grad)',
              lineHeight: 1.05,
              margin: 0,
            }}
          >
            VOCO
          </h1>
          <div
            style={{
              height: 1,
              background: 'linear-gradient(90deg, transparent, color-mix(in srgb, var(--voco-accent) 40%, transparent), transparent)',
              margin: '10px 0 24px',
            }}
          />
        </div>

        <p
          style={{
            fontFamily: 'Outfit, sans-serif',
            fontSize: 14,
            color: 'var(--voco-text-muted)',
            textAlign: 'center',
            lineHeight: 1.6,
            marginBottom: 36,
          }}
        >
          {props.mode === 'home'
            ? 'Войдите, чтобы создавать комнаты и сохранять профиль.'
            : 'Войдите с аккаунтом или продолжите как гость.'}
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <PrimaryButton fullWidth onClick={() => void auth.login(redirectUri)}>
            Войти
          </PrimaryButton>
          <SecondaryButton fullWidth onClick={() => void auth.register(redirectUri)}>
            Зарегистрироваться
          </SecondaryButton>
          {props.mode === 'room' && (
            <>
              <div style={{ height: 4 }} />
              <GhostButton fullWidth onClick={props.onGuest}>
                Позже — войти как гость
              </GhostButton>
            </>
          )}
        </div>

        {props.mode === 'home' && (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '28px 0' }}>
              <div style={{ flex: 1, height: 1, background: 'var(--voco-border-soft)' }} />
              <span style={{ fontFamily: 'Outfit, sans-serif', fontSize: 11, color: 'var(--voco-text-faint)', letterSpacing: '0.1em' }}>
                или войти по ссылке
              </span>
              <div style={{ flex: 1, height: 1, background: 'var(--voco-border-soft)' }} />
            </div>

            {!joinMode ? (
              <GhostButton fullWidth onClick={() => setJoinMode(true)}>
                Войти в комнату по ссылке →
              </GhostButton>
            ) : (
              <form style={{ display: 'flex', flexDirection: 'column', gap: 10 }} onSubmit={onJoinByLink}>
                <GlassInput
                  placeholder="Вставьте ссылку или ID комнаты"
                  value={inviteInput}
                  onChange={(e) => {
                    setInviteInput(e.target.value)
                    setError(null)
                  }}
                  autoFocus
                />
                <SecondaryButton fullWidth type="submit" disabled={!inviteInput.trim()}>
                  Войти в комнату
                </SecondaryButton>
              </form>
            )}
          </>
        )}

        {error && (
          <div style={{ marginTop: 16 }}>
            <StatusMessage type="error">{error}</StatusMessage>
          </div>
        )}
      </MineralCard>

      {props.mode === 'home' && (
        <p
          style={{
            marginTop: 28,
            fontFamily: 'Outfit, sans-serif',
            fontSize: 11,
            color: 'var(--voco-text-faint)',
            letterSpacing: '0.08em',
            textTransform: 'uppercase',
          }}
        >
          MVP video rooms
        </p>
      )}
    </div>
  )
}
