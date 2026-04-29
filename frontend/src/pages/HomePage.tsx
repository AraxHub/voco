import { useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { createRoom } from '../lib/api'
import './home.css'

function extractRoomId(input: string): string | null {
  const raw = input.trim()
  if (!raw) return null

  // If user pasted full invite URL: https://host/room/<id>
  try {
    const u = new URL(raw)
    const m = /^\/room\/([^/]+)$/.exec(u.pathname)
    if (m?.[1]) return m[1]
  } catch {
    // not a URL
  }

  // Otherwise treat as roomId
  return raw
}

export function HomePage() {
  const nav = useNavigate()
  const [inviteInput, setInviteInput] = useState('')
  const [showJoin, setShowJoin] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const normalizedInvite = useMemo(() => inviteInput.trim(), [inviteInput])

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
    <div className="home">
      <header className="home__header">
        <div className="home__brand">VOCO</div>
        <div className="home__muted">MVP video rooms</div>
      </header>

      <main className="home__main">
        <h1 className="home__hero">VOCO</h1>

        <div className="home__cta">
          <button className="btn btn--primary btn--xl btn--xlPlus" disabled={busy} onClick={onCreate}>
            {busy ? 'Создаю…' : 'Создать комнату'}
          </button>

          <button
            className="btn btn--wide btn--slim btn--joinPlus"
            disabled={busy}
            onClick={() => {
              setError(null)
              setShowJoin((v) => !v)
            }}
          >
            Присоединиться по ссылке
          </button>

          <div className={`home__dropdown ${showJoin ? 'is-open' : ''}`}>
            <form onSubmit={onJoin} className="home__joinStack" aria-hidden={!showJoin}>
              <input
                className="input input--join"
                value={inviteInput}
                onChange={(e) => setInviteInput(e.target.value)}
                placeholder="Вставь invite ссылку или roomId"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck={false}
              />
              <button className="btn btn--wide" type="submit" disabled={busy || !showJoin}>
                Войти
              </button>
            </form>
          </div>
        </div>

        {error && <div className="home__error">{error}</div>}
      </main>

      <footer className="home__footer">
        <div className="home__muted">Backend: /api/v1/rooms • LiveKit token: /api/v1/rooms/:id/token</div>
      </footer>
    </div>
  )
}

