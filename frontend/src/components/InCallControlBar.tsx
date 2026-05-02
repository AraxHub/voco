import { DisconnectButton, StartAudio, TrackToggle } from '@livekit/components-react'
import { Track } from 'livekit-client'
import './inCallControlBar.css'

type Props = {
  onCopyLink?: () => void
  copied?: boolean
}

function IconClip() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M9.5 12.5l6.1-6.1a3 3 0 1 1 4.2 4.2l-7.5 7.5a5 5 0 0 1-7.1-7.1l7.6-7.6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function InCallControlBar({ onCopyLink, copied }: Props) {
  return (
    <div className="callBar" role="toolbar" aria-label="Управление звонком">
      <StartAudio className="callBar__btn callBar__btn--ghost" label="Разрешить звук" />

      {onCopyLink && (
        <button
          type="button"
          className={`callBar__btn ${copied ? 'is-copied' : ''}`}
          onClick={onCopyLink}
          aria-label="Скопировать ссылку"
        >
          <IconClip />
        </button>
      )}

      <TrackToggle source={Track.Source.Microphone} className="callBar__btn" aria-label="Микрофон" />

      <TrackToggle source={Track.Source.Camera} className="callBar__btn" aria-label="Камера" />

      <TrackToggle source={Track.Source.ScreenShare} className="callBar__btn" aria-label="Демонстрация экрана" />

      <DisconnectButton className="callBar__btn callBar__btn--leave">Выйти</DisconnectButton>
    </div>
  )
}
