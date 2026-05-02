import { isEqualTrackRef, isTrackReference } from '@livekit/components-core'
import {
  CarouselLayout,
  FocusLayoutContainer,
  GridLayout,
  LayoutContextProvider,
  useCreateLayoutContext,
  usePinnedTracks,
  useTracks,
} from '@livekit/components-react'
import { RoomEvent, Track } from 'livekit-client'
import { useEffect, useState } from 'react'
import { ChatDrawer } from './ChatDrawer'
import { InCallControlBar } from './InCallControlBar'
import { RoomDurationClock } from './RoomDurationClock'
import { VocoFocusLayout, VocoParticipantTile } from './VocoParticipantTile'
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
  const layoutContext = useCreateLayoutContext()

  const tracks = useTracks(
    [
      { source: Track.Source.Camera, withPlaceholder: true },
      { source: Track.Source.ScreenShare, withPlaceholder: false },
    ],
    { updateOnlyOn: [RoomEvent.ActiveSpeakersChanged], onlySubscribed: false },
  )

  const focusTrack = usePinnedTracks(layoutContext)?.[0]
  const carouselTracks = tracks.filter((t) => !isEqualTrackRef(t, focusTrack))

  useEffect(() => {
    if (focusTrack && !isTrackReference(focusTrack)) {
      const updated = tracks.find(
        (tr) =>
          tr.participant.identity === focusTrack.participant.identity && tr.source === focusTrack.source,
      )
      if (updated !== focusTrack && isTrackReference(updated)) {
        layoutContext.pin.dispatch?.({ msg: 'set_pin', trackReference: updated })
      }
    }
  }, [tracks, focusTrack, layoutContext.pin])

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
      <div className="callShell__stage" style={{ ['--chat-open' as never]: chatOpen ? 1 : 0 }}>
        <header className="callShell__masthead" aria-label="Voco · длительность звонка">
          <span className="callShell__brand">voco</span>
          <RoomDurationClock />
        </header>

        <div className={`callShell__video ${chatOpen ? 'is-chat-open' : ''}`}>
          <div className="callShell__videoMain">
            <LayoutContextProvider value={layoutContext}>
              {!focusTrack ? (
                <GridLayout className="callGrid" tracks={tracks}>
                  <VocoParticipantTile disableSpeakingIndicator />
                </GridLayout>
              ) : (
                <FocusLayoutContainer className="callShell__focusLayout">
                  <CarouselLayout tracks={carouselTracks}>
                    <VocoParticipantTile disableSpeakingIndicator />
                  </CarouselLayout>
                  {focusTrack && <VocoFocusLayout trackRef={focusTrack} />}
                </FocusLayoutContainer>
              )}
            </LayoutContextProvider>
          </div>

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
