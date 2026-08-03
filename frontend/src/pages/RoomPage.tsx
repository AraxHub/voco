import '@livekit/components-styles'
import { LiveKitRoom, RoomAudioRenderer } from '@livekit/components-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { AuthGate } from '../components/AuthGate'
import { InCallLayout } from '../components/InCallLayout'
import { PreJoinPreview } from '../components/PreJoinPreview'
import { useAuth } from '../context/AuthContext'
import { cancelCall, issueToken, wsURL } from '../lib/api'
import { getAccessToken } from '../lib/keycloak'
import { allowGuestAccess, hasGuestAccess } from '../lib/guestSession'
import { PrimaryButton, NavButton } from '../ui/Button'
import { StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'

type JoinState =
  | { phase: 'prejoin' }
  | { phase: 'joining' }
  | { phase: 'joined'; token: string; livekitUrl: string }
  | { phase: 'error'; message: string }
  | { phase: 'ended'; message: string }

export function RoomPage() {
  const nav = useNavigate()
  const auth = useAuth()
  const { roomId } = useParams()
  const [search] = useSearchParams()
  const id = (roomId || '').trim()
  const autoJoin = search.get('join') === '1'
  const awaitingPeer = search.get('awaiting') === '1'

  const [guestAllowed, setGuestAllowed] = useState(() => hasGuestAccess(id))
  const [name, setName] = useState('')
  const [cameraOn, setCameraOn] = useState(false)
  const [micOn, setMicOn] = useState(true)
  const [joinState, setJoinState] = useState<JoinState>({ phase: 'prejoin' })
  const [chatOpen, setChatOpen] = useState(false)
  const [awaitBanner, setAwaitBanner] = useState(awaitingPeer)

  const [previewStream, setPreviewStream] = useState<MediaStream | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const lastPresetNameRef = useRef<string | null>(null)
  const autoJoinedRef = useRef(false)
  const callActiveRef = useRef(awaitingPeer)
  const intentionalLeaveRef = useRef(false)
  const ringingRef = useRef(awaitingPeer)

  const canJoin = useMemo(() => id.length > 0 && name.trim().length > 0, [id, name])

  const showAuthGate =
    auth.enabled && auth.ready && !auth.authenticated && !guestAllowed && id.length > 0

  useEffect(() => {
    setGuestAllowed(hasGuestAccess(id))
  }, [id])

  useEffect(() => {
    const preset = auth.displayName || auth.username
    if (!auth.authenticated || !preset) return
    setName((prev) => {
      if (!prev.trim() || prev === lastPresetNameRef.current) return preset
      return prev
    })
    lastPresetNameRef.current = preset
  }, [auth.authenticated, auth.displayName, auth.username])

  function stopPreview() {
    setPreviewStream((s) => {
      if (s) s.getTracks().forEach((t) => t.stop())
      return null
    })
    if (videoRef.current) videoRef.current.srcObject = null
  }

  useEffect(() => {
    if (!cameraOn) {
      stopPreview()
      return
    }

    let cancelled = false
    navigator.mediaDevices
      .getUserMedia({ video: true, audio: false })
      .then((stream) => {
        if (cancelled) {
          stream.getTracks().forEach((t) => t.stop())
          return
        }
        setPreviewStream(stream)
      })
      .catch(() => setPreviewStream(null))

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cameraOn])

  useEffect(() => {
    if (!videoRef.current) return
    videoRef.current.srcObject = previewStream
  }, [previewStream])

  // Do NOT cancelCall in effect cleanup — React StrictMode remounts and would
  // abort an outgoing call before the callee ever sees call.incoming.
  // Cancel only on explicit leave (back button) or tab close.
  useEffect(() => {
    return () => {
      stopPreview()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!awaitingPeer || !id) return
    const onPageHide = () => {
      if (!ringingRef.current || intentionalLeaveRef.current) return
      ringingRef.current = false
      callActiveRef.current = false
      const token = getAccessToken()
      if (!token) return
      void fetch(`/api/v1/calls/${encodeURIComponent(id)}/cancel`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        keepalive: true,
      })
    }
    window.addEventListener('pagehide', onPageHide)
    return () => window.removeEventListener('pagehide', onPageHide)
  }, [awaitingPeer, id])

  function endCallLocally(message: string, cancel: boolean) {
    intentionalLeaveRef.current = true
    ringingRef.current = false
    callActiveRef.current = false
    stopPreview()
    setCameraOn(false)
    setMicOn(false)
    setAwaitBanner(false)
    if (cancel) {
      void cancelCall(id).catch(() => undefined)
    }
    setJoinState({ phase: 'ended', message })
    nav('/chats')
  }

  function onLeaveCall() {
    // Red hangup: always leave. Cancel remote ring only while still waiting.
    endCallLocally('Звонок завершён', ringingRef.current || awaitBanner)
  }

  async function onJoin() {
    if (!canJoin) return
    setJoinState({ phase: 'joining' })
    try {
      const res = await issueToken(id, { name: name.trim() })
      const livekitUrl = res.livekitUrl || ''
      if (!res.token) throw new Error('empty token from backend')
      if (!livekitUrl) throw new Error('empty livekitUrl from backend')
      setJoinState({ phase: 'joined', token: res.token, livekitUrl })
    } catch (e) {
      setJoinState({ phase: 'error', message: e instanceof Error ? e.message : 'join failed' })
    }
  }

  useEffect(() => {
    if (!autoJoin || autoJoinedRef.current || intentionalLeaveRef.current) return
    if (!canJoin || showAuthGate) return
    if (joinState.phase !== 'prejoin') return
    autoJoinedRef.current = true
    void onJoin()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoJoin, canJoin, showAuthGate, joinState.phase, name])

  useEffect(() => {
    if (!awaitingPeer || !auth.token || !id) return
    let ws: WebSocket
    try {
      ws = new WebSocket(wsURL(auth.token))
    } catch {
      return
    }
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as { event?: string; payload?: { roomId?: string } }
        if (!msg.event || msg.payload?.roomId !== id) return
        if (msg.event === 'call.accepted') {
          ringingRef.current = false
          callActiveRef.current = false
          setAwaitBanner(false)
        }
        if (msg.event === 'call.declined' || msg.event === 'call.missed' || msg.event === 'call.cancelled') {
          ringingRef.current = false
          callActiveRef.current = false
          stopPreview()
          setJoinState({
            phase: 'ended',
            message:
              msg.event === 'call.declined'
                ? 'Звонок отклонён'
                : msg.event === 'call.missed'
                  ? 'Нет ответа — звонок завершён'
                  : 'Звонок отменён',
          })
        }
      } catch {
        /* ignore */
      }
    }
    return () => ws.close()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [awaitingPeer, auth.token, id])

  function onContinueAsGuest() {
    allowGuestAccess(id)
    setGuestAllowed(true)
  }

  useEffect(() => {
    if (joinState.phase !== 'joined') return
    setChatOpen(false)
  }, [joinState.phase])

  if (!id) {
    return (
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'grid',
          placeItems: 'center',
          gap: 16,
        }}
      >
        <StatusMessage type="error">Нет roomId в URL</StatusMessage>
        <NavButton onClick={() => nav('/')}>На главную</NavButton>
      </div>
    )
  }

  if (auth.enabled && !auth.ready) {
    return <div className="authBoot">Загрузка…</div>
  }

  if (showAuthGate) {
    return <AuthGate mode="room" onGuest={onContinueAsGuest} />
  }

  if (joinState.phase === 'ended') {
    return (
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100vh',
          display: 'grid',
          placeItems: 'center',
          gap: 16,
          padding: 24,
        }}
      >
        <StatusMessage type="error">{joinState.message}</StatusMessage>
        <NavButton
          onClick={() => {
            endCallLocally('Звонок завершён', false)
          }}
        >
          К чатам
        </NavButton>
      </div>
    )
  }

  if (joinState.phase === 'joined') {
    return (
      <div style={{ position: 'relative', zIndex: 1, height: '100vh' }}>
        {awaitBanner && (
          <div
            style={{
              position: 'absolute',
              top: 16,
              left: '50%',
              transform: 'translateX(-50%)',
              zIndex: 5,
              padding: '10px 16px',
              borderRadius: 12,
              background: 'var(--voco-card)',
              border: '1px solid var(--voco-border)',
              color: 'var(--voco-text)',
            }}
          >
            Ожидаем ответа… (до 1 мин)
          </div>
        )}
        <LiveKitRoom
          token={joinState.token}
          serverUrl={joinState.livekitUrl}
          connect={true}
          video={cameraOn}
          audio={micOn}
          onDisconnected={() => {
            stopPreview()
            setCameraOn(false)
            setMicOn(false)
            // User already handled leave via red button / back.
            if (intentionalLeaveRef.current) {
              return
            }
            // Transient LiveKit drop while still ringing — stay on prejoin,
            // but do NOT clear autoJoinedRef (that re-joins and traps the user).
            if (ringingRef.current) {
              setJoinState({ phase: 'prejoin' })
              return
            }
            setJoinState({ phase: 'ended', message: 'Звонок завершён' })
            nav('/chats')
          }}
          data-lk-theme="default"
          style={{ height: '100vh' }}
        >
          <InCallLayout
            chatOpen={chatOpen}
            onToggleChat={() => setChatOpen((v) => !v)}
            onCloseChat={() => setChatOpen(false)}
            roomId={id}
            onLeave={onLeaveCall}
          />
          <RoomAudioRenderer />
        </LiveKitRoom>
      </div>
    )
  }

  const joinLabel = auth.authenticated ? 'Присоединиться к звонку' : 'Подключиться'
  const joining = joinState.phase === 'joining'

  return (
    <div
      className="fade-in-up"
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
      <NavButton
        type="button"
        onClick={() => {
          endCallLocally('Звонок завершён', ringingRef.current || callActiveRef.current)
        }}
        style={{ position: 'absolute', top: 28, left: 32 }}
      >
        ← Назад
      </NavButton>

      <div style={{ width: '100%', maxWidth: 680, display: 'flex', flexDirection: 'column', gap: 28 }}>
        <PreJoinPreview
          videoRef={videoRef}
          cameraOn={cameraOn}
          micOn={micOn}
          hasPreview={Boolean(previewStream)}
          onToggleCamera={() => setCameraOn((v) => !v)}
          onToggleMic={() => setMicOn((v) => !v)}
        />

        <div
          style={{
            padding: '28px 32px',
            borderRadius: 16,
            background: 'var(--voco-card)',
            backdropFilter: 'blur(20px)',
            border: '1px solid var(--voco-border-soft)',
            display: 'flex',
            flexDirection: 'column',
            gap: 24,
          }}
        >
          <GlassInput
            label="Ваше имя"
            placeholder="Введите ваше имя"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />

          <PrimaryButton
            fullWidth
            loading={joining}
            disabled={!canJoin || joining}
            onClick={() => void onJoin()}
            style={{ fontSize: 15, padding: '15px 36px' }}
          >
            {joining ? 'Подключаемся…' : joinLabel}
          </PrimaryButton>

          {joinState.phase === 'error' && <StatusMessage type="error">{joinState.message}</StatusMessage>}
        </div>
      </div>
    </div>
  )
}
