import { useRoomContext } from '@livekit/components-react'
import { ConnectionState, RoomEvent } from 'livekit-client'
import { useEffect, useState } from 'react'

function formatCallDuration(totalSeconds: number) {
  const s = Math.max(0, totalSeconds)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  return `${m}:${String(sec).padStart(2, '0')}`
}

/**
 * Время в комнате с момента первого успешного подключения до выхода.
 * Рендерить только внутри LiveKitRoom.
 */
export function RoomDurationClock() {
  const room = useRoomContext()
  const [startedAtMs, setStartedAtMs] = useState<number | null>(null)
  const [, setTick] = useState(0)

  useEffect(() => {
    const markConnected = () => setStartedAtMs((prev) => prev ?? Date.now())

    if (room.state === ConnectionState.Connected) markConnected()

    room.on(RoomEvent.Connected, markConnected)

    const interval = window.setInterval(() => setTick((x) => x + 1), 1000)

    return () => {
      room.off(RoomEvent.Connected, markConnected)
      window.clearInterval(interval)
    }
  }, [room])

  const seconds = startedAtMs == null ? 0 : Math.floor((Date.now() - startedAtMs) / 1000)
  const text = startedAtMs == null ? '—:—' : formatCallDuration(seconds)

  return (
    <div
      className="roomDurationClock"
      role="timer"
      aria-live="off"
      aria-label={startedAtMs == null ? 'Подключение' : `В комнате ${formatCallDuration(seconds)}`}
    >
      <svg className="roomDurationClock__icon" width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden>
        <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.75" />
        <path d="M12 7v6l4 3" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
      <span className="roomDurationClock__value">{text}</span>
    </div>
  )
}
