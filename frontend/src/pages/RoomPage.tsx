import '@livekit/components-styles'
import { LiveKitRoom, RoomAudioRenderer } from '@livekit/components-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { AuthGate } from '../components/AuthGate'
import { InCallLayout } from '../components/InCallLayout'
import { PreJoinPreview } from '../components/PreJoinPreview'
import { useAuth } from '../context/AuthContext'
import { issueToken } from '../lib/api'
import { allowGuestAccess, hasGuestAccess } from '../lib/guestSession'
import { PrimaryButton, NavButton } from '../ui/Button'
import { StatusMessage } from '../ui/Card'
import { GlassInput } from '../ui/Input'

type JoinState =
  | { phase: 'prejoin' }
  | { phase: 'joining' }
  | { phase: 'joined'; token: string; livekitUrl: string }
  | { phase: 'error'; message: string }

export function RoomPage() {
  const nav = useNavigate()
  const auth = useAuth()
  const { roomId } = useParams()
  const id = (roomId || '').trim()

  const [guestAllowed, setGuestAllowed] = useState(() => hasGuestAccess(id))
  const [name, setName] = useState('')
  const [cameraOn, setCameraOn] = useState(false)
  const [micOn, setMicOn] = useState(true)
  const [joinState, setJoinState] = useState<JoinState>({ phase: 'prejoin' })
  const [chatOpen, setChatOpen] = useState(false)

  const [previewStream, setPreviewStream] = useState<MediaStream | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const canJoin = useMemo(() => id.length > 0 && name.trim().length > 0, [id, name])

  const showAuthGate =
    auth.enabled && auth.ready && !auth.authenticated && !guestAllowed && id.length > 0

  useEffect(() => {
    setGuestAllowed(hasGuestAccess(id))
  }, [id])

  useEffect(() => {
    if (!auth.authenticated || !auth.username) return
    setName((prev) => (prev.trim() ? prev : auth.username ?? ''))
  }, [auth.authenticated, auth.username])

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

  useEffect(() => {
    return () => {
      stopPreview()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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

  if (joinState.phase === 'joined') {
    return (
      <div style={{ position: 'relative', zIndex: 1, height: '100vh' }}>
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
            setJoinState({ phase: 'prejoin' })
          }}
          data-lk-theme="default"
          style={{ height: '100vh' }}
        >
          <InCallLayout
            chatOpen={chatOpen}
            onToggleChat={() => setChatOpen((v) => !v)}
            onCloseChat={() => setChatOpen(false)}
            roomId={id}
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
        onClick={() => nav('/')}
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
