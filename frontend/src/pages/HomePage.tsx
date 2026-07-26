import { useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthGate } from '../components/AuthGate'
import { useAuth } from '../context/AuthContext'
import { createRoom } from '../lib/api'
import { PrimaryButton, SecondaryButton, GhostButton } from '../ui/Button'
import { StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'

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

export function HomePage() {
  const nav = useNavigate()
  const auth = useAuth()
  const [inviteInput, setInviteInput] = useState('')
  const [showJoin, setShowJoin] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const normalizedInvite = useMemo(() => inviteInput.trim(), [inviteInput])

  if (auth.enabled && !auth.ready) {
    return <div className="authBoot">Загрузка…</div>
  }

  if (!auth.enabled && import.meta.env.DEV) {
    return (
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'grid',
          placeItems: 'center',
          padding: 24,
        }}
      >
        <div
          style={{
            maxWidth: 440,
            padding: 32,
            borderRadius: 20,
            background: 'rgba(12,20,30,0.72)',
            border: '1px solid rgba(255,255,255,0.09)',
          }}
        >
          <h1 style={{ fontFamily: 'Syncopate, sans-serif', fontSize: 22, letterSpacing: '0.12em', marginTop: 0 }}>
            Keycloak не настроен
          </h1>
          <p style={{ color: '#6b7a8d', lineHeight: 1.6 }}>
            Нужен <code>frontend/.env</code> с <code>VITE_KEYCLOAK_URL=http://localhost:8180</code> и перезапуск Vite.
          </p>
        </div>
      </div>
    )
  }

  if (auth.enabled && auth.ready && !auth.authenticated) {
    return <AuthGate mode="home" />
  }

  async function onCreate() {
    setError(null)
    setBusy(true)
    try {
      const res = await createRoom({})
      nav(`/room/${res.roomId}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'create failed')
    } finally {
      setBusy(false)
    }
  }

  function onJoin(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const id = extractRoomId(normalizedInvite)
    if (!id) {
      setError('Вставь invite ссылку вида /room/<id> или roomId')
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
      }}
    >
      <header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '20px 40px',
          borderBottom: '1px solid var(--voco-border-soft)',
          gap: 16,
          flexWrap: 'wrap',
        }}
      >
        <span
          style={{
            fontFamily: 'Syncopate, sans-serif',
            fontWeight: 700,
            fontSize: 18,
            letterSpacing: '0.2em',
            color: 'transparent',
            backgroundClip: 'text',
            WebkitBackgroundClip: 'text',
            backgroundImage: 'var(--voco-brand-grad)',
          }}
        >
          VOCO
        </span>

        <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
          <span
            style={{
              fontFamily: 'Outfit, sans-serif',
              fontSize: 11,
              color: 'var(--voco-text-faint)',
              letterSpacing: '0.1em',
              textTransform: 'uppercase',
            }}
          >
            MVP video rooms
          </span>
          {auth.enabled && auth.authenticated && (
            <>
              <span
                style={{
                  fontFamily: 'Outfit, sans-serif',
                  fontSize: 13,
                  fontWeight: 500,
                  color: 'var(--voco-text-muted)',
                  padding: '6px 14px',
                  borderRadius: 8,
                  background: 'var(--voco-chip-bg)',
                  border: '1px solid var(--voco-border-soft)',
                }}
              >
                {auth.username ?? 'user'}
              </span>
              <Link
                to="/account"
                aria-label="Аккаунт"
                title="Аккаунт"
                className="voco-nav-btn voco-nav-btn--icon homeSettings"
              >
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" aria-hidden>
                  <path
                    fill="currentColor"
                    d="M19.14 12.94c.04-.31.06-.63.06-.94s-.02-.63-.06-.94l2.03-1.58a.5.5 0 0 0 .12-.64l-1.92-3.32a.5.5 0 0 0-.6-.22l-2.39.96a7.1 7.1 0 0 0-1.63-.94l-.36-2.54A.5.5 0 0 0 13.9 2h-3.8a.5.5 0 0 0-.49.42l-.36 2.54c-.59.24-1.13.55-1.63.94l-2.39-.96a.5.5 0 0 0-.6.22L2.71 8.48a.5.5 0 0 0 .12.64l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58a.5.5 0 0 0-.12.64l1.92 3.32c.14.24.43.34.68.22l2.39-.96c.5.39 1.04.7 1.63.94l.36 2.54c.05.24.25.42.49.42h3.8c.24 0 .44-.18.49-.42l.36-2.54c.59-.24 1.13-.55 1.63-.94l2.39.96c.25.1.54 0 .68-.22l1.92-3.32a.5.5 0 0 0-.12-.64l-2.03-1.58ZM12 15.5A3.5 3.5 0 1 1 12 8.5a3.5 3.5 0 0 1 0 7Z"
                  />
                </svg>
              </Link>
            </>
          )}
        </div>
      </header>

      <main
        className="fade-in-up"
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '60px 24px 120px',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: 20 }}>
          <h1
            className="gradient-breathe"
            style={{
              fontFamily: 'Syncopate, sans-serif',
              fontWeight: 700,
              fontSize: 'clamp(64px, 12vw, 140px)',
              letterSpacing: '0.18em',
              color: 'transparent',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              backgroundImage: 'var(--voco-hero-grad)',
              lineHeight: 1,
              margin: 0,
            }}
          >
            VOCO
          </h1>
          <p
            style={{
              marginTop: 20,
              fontFamily: 'Outfit, sans-serif',
              fontSize: 16,
              fontWeight: 300,
              color: 'var(--voco-text-dim)',
              letterSpacing: '0.06em',
              maxWidth: 360,
              margin: '16px auto 0',
              lineHeight: 1.6,
            }}
          >
            Мгновенные видеокомнаты — только по ссылке.
          </p>
        </div>

        <div
          style={{
            marginTop: 52,
            width: '100%',
            maxWidth: 520,
            display: 'flex',
            flexDirection: 'column',
            gap: 14,
          }}
        >
          <PrimaryButton
            fullWidth
            loading={busy}
            onClick={() => void onCreate()}
            style={{ fontSize: 15, padding: '16px 36px' }}
          >
            {busy ? 'Создаю…' : 'Создать комнату'}
          </PrimaryButton>

          {!showJoin ? (
            <SecondaryButton
              fullWidth
              disabled={busy}
              onClick={() => {
                setError(null)
                setShowJoin(true)
              }}
            >
              Присоединиться по ссылке
            </SecondaryButton>
          ) : (
            <form
              onSubmit={onJoin}
              style={{
                padding: 20,
                borderRadius: 14,
                background: 'var(--voco-chip-bg)',
                border: '1px solid var(--voco-border)',
                display: 'flex',
                flexDirection: 'column',
                gap: 12,
              }}
            >
              <GlassInput
                placeholder="Вставьте ссылку или ID комнаты"
                value={inviteInput}
                onChange={(e) => {
                  setInviteInput(e.target.value)
                  setError(null)
                }}
                autoFocus
              />
              <div style={{ display: 'flex', gap: 10 }}>
                <SecondaryButton type="submit" style={{ flex: 1 }} disabled={busy}>
                  Войти в комнату
                </SecondaryButton>
                <GhostButton
                  type="button"
                  onClick={() => {
                    setShowJoin(false)
                    setError(null)
                  }}
                >
                  Отмена
                </GhostButton>
              </div>
            </form>
          )}

          {error && <StatusMessage type="error">{error}</StatusMessage>}
        </div>
      </main>
    </div>
  )
}
