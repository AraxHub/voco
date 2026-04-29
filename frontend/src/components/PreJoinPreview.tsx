import './preJoinPreview.css'

type Props = {
  videoRef: React.RefObject<HTMLVideoElement | null>
  cameraOn: boolean
  micOn: boolean
  hasPreview: boolean
  onToggleCamera: () => void
  onToggleMic: () => void
}

function IconCamera({ on }: { on: boolean }) {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M14.5 8.5H6.8C5.81 8.5 5 9.31 5 10.3v3.4c0 .99.81 1.8 1.8 1.8h7.7c.99 0 1.8-.81 1.8-1.8v-3.4c0-.99-.81-1.8-1.8-1.8Z"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinejoin="round"
      />
      <path
        d="M16.3 11.2 19 9.7c.66-.37 1.5.11 1.5.86v3.86c0 .75-.84 1.23-1.5.86l-2.7-1.5"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {!on && (
        <path
          d="M4 4l16 16"
          stroke="currentColor"
          strokeWidth="2.2"
          strokeLinecap="round"
        />
      )}
    </svg>
  )
}

function IconMic({ on }: { on: boolean }) {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M12 14.2a3 3 0 0 0 3-3V7.8a3 3 0 1 0-6 0v3.4a3 3 0 0 0 3 3Z"
        stroke="currentColor"
        strokeWidth="1.8"
      />
      <path
        d="M6.8 11.2a5.2 5.2 0 1 0 10.4 0"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
      <path d="M12 16.4v3" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      <path d="M9.2 19.4h5.6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      {!on && (
        <path
          d="M4 4l16 16"
          stroke="currentColor"
          strokeWidth="2.2"
          strokeLinecap="round"
        />
      )}
    </svg>
  )
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
    <div className="pjPreview">
      <div className="pjPreview__frame">
        <video ref={videoRef} autoPlay playsInline muted className="pjPreview__video" />

        {!hasPreview && (
          <div className="pjPreview__placeholder">
            <div className="pjPreview__placeholderTitle">Preview</div>
            <div className="pjPreview__placeholderText">
              Камера выключена или нет доступа
            </div>
          </div>
        )}

        <div className="pjPreview__controls" role="group" aria-label="Устройства">
          <button
            type="button"
            className={`iconBtn3d ${cameraOn ? 'iconBtn3d--on' : 'iconBtn3d--off'}`}
            onClick={onToggleCamera}
            aria-pressed={cameraOn}
            aria-label={cameraOn ? 'Выключить камеру' : 'Включить камеру'}
          >
            <IconCamera on={cameraOn} />
          </button>
          <button
            type="button"
            className={`iconBtn3d ${micOn ? 'iconBtn3d--on' : 'iconBtn3d--off'}`}
            onClick={onToggleMic}
            aria-pressed={micOn}
            aria-label={micOn ? 'Выключить микрофон' : 'Включить микрофон'}
          >
            <IconMic on={micOn} />
          </button>
        </div>
      </div>
    </div>
  )
}

