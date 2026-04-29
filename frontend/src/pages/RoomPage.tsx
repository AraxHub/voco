import '@livekit/components-styles'
import { LiveKitRoom, RoomAudioRenderer } from '@livekit/components-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { InCallLayout } from '../components/InCallLayout'
import { PreJoinPreview } from '../components/PreJoinPreview'
import { issueToken } from '../lib/api'
import './room.css'

type JoinState =
  | { phase: 'prejoin' }
  | { phase: 'joining' }
  | { phase: 'joined'; token: string; livekitUrl: string }
  | { phase: 'error'; message: string }

export function RoomPage() {
  const nav = useNavigate()
  const { roomId } = useParams()
  const id = (roomId || '').trim()

  const [name, setName] = useState('')
  const [cameraOn, setCameraOn] = useState(false)
  const [micOn, setMicOn] = useState(true)
  const [joinState, setJoinState] = useState<JoinState>({ phase: 'prejoin' })
  const [chatOpen, setChatOpen] = useState(false)

  const [previewStream, setPreviewStream] = useState<MediaStream | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const canJoin = useMemo(() => id.length > 0 && name.trim().length > 0, [id, name])

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

  useEffect(() => {
    if (joinState.phase !== 'joined') return
    setChatOpen(false)
  }, [joinState.phase])

  if (!id) {
    return (
      <div className="roomPage">
        <div className="roomPage__error">Нет roomId в URL</div>
        <button className="btn" onClick={() => nav('/')}>
          На главную
        </button>
      </div>
    )
  }

  if (joinState.phase === 'joined') {
    return (
      <div className="roomPage roomPage--inCall roomPage--noHighlight">
        <LiveKitRoom
          token={joinState.token}
          serverUrl={joinState.livekitUrl}
          connect={true}
          video={cameraOn}
          audio={micOn}
          onDisconnected={() => {
            // Privacy-by-default on leave: stop camera preview, reset toggles, go back to prejoin.
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

  return (
    <div className="roomPage">
      <main className="roomJoin">
        <div className="roomJoin__top">
          <button className="linkBtn" onClick={() => nav('/')}>
            ← Назад
          </button>
        </div>

        <PreJoinPreview
          videoRef={videoRef}
          cameraOn={cameraOn}
          micOn={micOn}
          hasPreview={Boolean(previewStream)}
          onToggleCamera={() => setCameraOn((v) => !v)}
          onToggleMic={() => setMicOn((v) => !v)}
        />

        <div className="roomJoin__stack">
          <input
            className="input roomJoin__name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Введите ваше имя"
            autoFocus
          />

          <button
            className="btn3d"
            disabled={!canJoin || joinState.phase === 'joining'}
            onClick={onJoin}
          >
            {joinState.phase === 'joining' ? 'Подключаемся…' : 'Подключиться'}
          </button>

          {joinState.phase === 'error' && <div className="errorBox">{joinState.message}</div>}
        </div>
      </main>
    </div>
  )
}

