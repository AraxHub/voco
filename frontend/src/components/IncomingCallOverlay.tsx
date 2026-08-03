import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { acceptCall, declineCall, wsURL } from '../lib/api'
import { useAuth } from '../context/AuthContext'
import { PrimaryButton, GhostButton } from '../ui/Button'
import './IncomingCallOverlay.css'

export type IncomingCall = {
  roomId: string
  conversationId: string
  callerId: string
  callerName: string
  expiresAt: string
}

function playRingtone(ctx: AudioContext) {
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()
  osc.type = 'sine'
  osc.frequency.value = 440
  gain.gain.value = 0.0001
  osc.connect(gain)
  gain.connect(ctx.destination)
  osc.start()

  let on = true
  const pulse = () => {
    const now = ctx.currentTime
    if (on) {
      osc.frequency.setValueAtTime(880, now)
      gain.gain.cancelScheduledValues(now)
      gain.gain.setValueAtTime(0.0001, now)
      gain.gain.exponentialRampToValueAtTime(0.12, now + 0.05)
      gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.45)
    }
    on = !on
  }
  pulse()
  const id = window.setInterval(pulse, 500)
  return () => {
    window.clearInterval(id)
    try {
      osc.stop()
      osc.disconnect()
      gain.disconnect()
    } catch {
      /* ignore */
    }
  }
}

export function IncomingCallHost() {
  const auth = useAuth()
  const nav = useNavigate()
  const [call, setCall] = useState<IncomingCall | null>(null)
  const [secondsLeft, setSecondsLeft] = useState(60)
  const stopRingRef = useRef<(() => void) | null>(null)
  const audioCtxRef = useRef<AudioContext | null>(null)
  const callRef = useRef<IncomingCall | null>(null)

  const clearCall = useCallback(() => {
    stopRingRef.current?.()
    stopRingRef.current = null
    void audioCtxRef.current?.close().catch(() => undefined)
    audioCtxRef.current = null
    callRef.current = null
    setCall(null)
  }, [])

  useEffect(() => {
    if (!auth.token || !auth.authenticated) return
    let ws: WebSocket
    try {
      ws = new WebSocket(wsURL(auth.token))
    } catch {
      return
    }
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as {
          event?: string
          payload?: IncomingCall & { roomId?: string; reason?: string }
        }
        if (msg.event === 'call.incoming' && msg.payload?.roomId) {
          const next: IncomingCall = {
            roomId: String(msg.payload.roomId),
            conversationId: String(msg.payload.conversationId || ''),
            callerId: String(msg.payload.callerId || ''),
            callerName: String(msg.payload.callerName || 'Абонент'),
            expiresAt: String(msg.payload.expiresAt || ''),
          }
          callRef.current = next
          setCall(next)
        }
        const active = callRef.current
        if (
          (msg.event === 'call.cancelled' || msg.event === 'call.missed' || msg.event === 'call.declined') &&
          active &&
          msg.payload?.roomId === active.roomId
        ) {
          clearCall()
        }
      } catch {
        /* ignore */
      }
    }
    return () => ws.close()
  }, [auth.token, auth.authenticated, clearCall])

  useEffect(() => {
    if (!call) return
    const expires = Date.parse(call.expiresAt)
    const tick = () => {
      const left = Number.isFinite(expires)
        ? Math.max(0, Math.ceil((expires - Date.now()) / 1000))
        : 60
      setSecondsLeft(left)
      if (left <= 0) clearCall()
    }
    tick()
    const id = window.setInterval(tick, 250)
    return () => window.clearInterval(id)
  }, [call, clearCall])

  useEffect(() => {
    if (!call) return
    const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    const ctx = new Ctx()
    audioCtxRef.current = ctx
    void ctx.resume()
    stopRingRef.current = playRingtone(ctx)
    return () => {
      stopRingRef.current?.()
      stopRingRef.current = null
      void ctx.close().catch(() => undefined)
    }
  }, [call])

  if (!call) return null

  return (
    <div className="incomingCall" role="dialog" aria-modal="true" aria-label="Входящий звонок">
      <div className="incomingCall__pulse" aria-hidden />
      <div className="incomingCall__card">
        <p className="incomingCall__eyebrow">Входящий звонок</p>
        <h1 className="incomingCall__name">{call.callerName}</h1>
        <p className="incomingCall__timer">{secondsLeft} с</p>
        <div className="incomingCall__actions">
          <GhostButton
            type="button"
            onClick={() => {
              const roomId = call.roomId
              clearCall()
              void declineCall(roomId).catch(() => undefined)
            }}
          >
            Отклонить
          </GhostButton>
          <PrimaryButton
            type="button"
            onClick={() => {
              const roomId = call.roomId
              clearCall()
              void acceptCall(roomId)
                .then(() => nav(`/room/${roomId}?join=1`))
                .catch(() => undefined)
            }}
          >
            Принять
          </PrimaryButton>
        </div>
      </div>
    </div>
  )
}
