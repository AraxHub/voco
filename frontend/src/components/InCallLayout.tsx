import { GridLayout, ParticipantTile, useTracks } from '@livekit/components-react'
import { RoomEvent, Track } from 'livekit-client'
import { useState } from 'react'
import { ChatDrawer } from './ChatDrawer'
import { InCallControlBar } from './InCallControlBar'
import './inCallLayout.css'

type Props = {
  chatOpen: boolean
  onToggleChat: () => void
  onCloseChat: () => void
  roomId: string
}

function IconChatHandle() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M7.2 18.2 4.5 20v-3.2A7.6 7.6 0 0 1 3.2 13c0-4.3 3.9-7.8 8.8-7.8s8.8 3.5 8.8 7.8-3.9 7.8-8.8 7.8c-1.6 0-3.2-.4-4.6-1.1Z"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function InCallLayout({ chatOpen, onToggleChat, onCloseChat, roomId }: Props) {
  const tracks = useTracks(
    [
      { source: Track.Source.Camera, withPlaceholder: true },
      { source: Track.Source.ScreenShare, withPlaceholder: false },
    ],
    { updateOnlyOn: [RoomEvent.ActiveSpeakersChanged], onlySubscribed: false },
  )

  const [copied, setCopied] = useState(false)

  async function copyLink() {
    const url = `${window.location.origin}/room/${roomId}`
    try {
      await navigator.clipboard.writeText(url)
    } catch {
      window.prompt('Скопируй ссылку', url)
    }
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1200)
  }

  return (
    <div className="callShell">
      <div className="callShell__stage" style={{ ['--chat-open' as any]: chatOpen ? 1 : 0 }}>
        <div className={`callShell__video ${chatOpen ? 'is-chat-open' : ''}`}>
          <GridLayout className="callGrid" tracks={tracks}>
            <ParticipantTile disableSpeakingIndicator />
          </GridLayout>

          {!chatOpen && (
            <button
              type="button"
              className="chatHandle"
              onClick={onToggleChat}
              aria-label="Открыть чат"
            >
              <IconChatHandle />
            </button>
          )}
        </div>
        <ChatDrawer open={chatOpen} onClose={onCloseChat} />
      </div>

      <div className={`copyLinkToast copyLinkToast--bar ${copied ? 'is-show' : ''}`} role="status" aria-live="polite">
        Ссылка скопирована
      </div>
      <InCallControlBar onCopyLink={copyLink} copied={copied} />
    </div>
  )
}

