/**
 * Participant tile + кастомная кнопка «развернуть» вместо стандартной иконки LiveKit.
 * Pin / полно-блочное отображение — через useFocusToggle (CarouselLayout / FocusLayout).
 */
import type { ParticipantClickEvent, TrackReferenceOrPlaceholder } from '@livekit/components-core'
import { isTrackReference, isTrackReferencePinned } from '@livekit/components-core'
import type { ParticipantTileProps } from '@livekit/components-react'
import {
  AudioTrack,
  ConnectionQualityIndicator,
  LockLockedIcon,
  ParticipantContextIfNeeded,
  ParticipantName,
  ParticipantPlaceholder,
  ScreenShareIcon,
  TrackMutedIndicator,
  TrackRefContextIfNeeded,
  VideoTrack,
  useEnsureTrackRef,
  useFeatureContext,
  useIsEncrypted,
  useMaybeLayoutContext,
  useParticipantTile,
} from '@livekit/components-react'
import { useFocusToggle } from '@livekit/components-react/hooks'
import * as React from 'react'
import { Track } from 'livekit-client'

/** Unicode U+2921 (dec 10529) — «NORTH EAST AND SOUTH WEST ARROW». */
const GLYPH_EXPAND_U2921 = String.fromCodePoint(0x2921)

function ExpandToggleGlyph({ collapse }: { collapse: boolean }) {
  return (
    <span
      aria-hidden
      className={`voco-expand-glyph${collapse ? ' voco-expand-glyph--collapse' : ''}`}
    >
      {GLYPH_EXPAND_U2921}
    </span>
  )
}

function ParticipantExpandToggle({ trackRef }: { trackRef: TrackReferenceOrPlaceholder }) {
  const trackReference = useEnsureTrackRef(trackRef)
  const { mergedProps, inFocus } = useFocusToggle({
    trackRef: trackReference,
    props: {
      type: 'button',
      className: 'focus-toggle-button voco-expand-btn',
    },
  })

  return (
    <button
      {...mergedProps}
      aria-label={inFocus ? 'Вернуть в сетку' : 'Развернуть участника'}
      title={inFocus ? 'Вернуть в сетку' : 'Развернуть'}
    >
      <ExpandToggleGlyph collapse={inFocus} />
    </button>
  )
}

export const VocoParticipantTile = React.forwardRef<HTMLDivElement, ParticipantTileProps>(
  function VocoParticipantTile(
    {
      trackRef,
      children,
      onParticipantClick,
      disableSpeakingIndicator,
      ...htmlProps
    }: ParticipantTileProps,
    ref,
  ) {
    const trackReference = useEnsureTrackRef(trackRef)

    const { elementProps } = useParticipantTile<HTMLDivElement>({
      htmlProps,
      disableSpeakingIndicator,
      onParticipantClick,
      trackRef: trackReference,
    })
    const isEncrypted = useIsEncrypted(trackReference.participant)
    const layoutContext = useMaybeLayoutContext()

    const autoManageSubscription = useFeatureContext()?.autoSubscription

    const handleSubscribe = React.useCallback(
      (subscribed: boolean) => {
        if (
          trackReference.source &&
          !subscribed &&
          layoutContext &&
          layoutContext.pin.dispatch &&
          isTrackReferencePinned(trackReference, layoutContext.pin.state)
        ) {
          layoutContext.pin.dispatch({ msg: 'clear_pin' })
        }
      },
      [trackReference, layoutContext],
    )

    return (
      <div ref={ref} style={{ position: 'relative' }} {...elementProps}>
        <TrackRefContextIfNeeded trackRef={trackReference}>
          <ParticipantContextIfNeeded participant={trackReference.participant}>
            {children ?? (
              <>
                {isTrackReference(trackReference) &&
                (trackReference.publication?.kind === 'video' ||
                  trackReference.source === Track.Source.Camera ||
                  trackReference.source === Track.Source.ScreenShare) ? (
                  <VideoTrack
                    trackRef={trackReference}
                    onSubscriptionStatusChanged={handleSubscribe}
                    manageSubscription={autoManageSubscription}
                  />
                ) : (
                  isTrackReference(trackReference) && (
                    <AudioTrack
                      trackRef={trackReference}
                      onSubscriptionStatusChanged={handleSubscribe}
                    />
                  )
                )}
                <div className="lk-participant-placeholder">
                  <ParticipantPlaceholder />
                </div>
                <div className="lk-participant-metadata">
                  <div className="lk-participant-metadata-item">
                    {trackReference.source === Track.Source.Camera ? (
                      <>
                        {isEncrypted && <LockLockedIcon style={{ marginRight: '0.25rem' }} />}
                        <TrackMutedIndicator
                          trackRef={{
                            participant: trackReference.participant,
                            source: Track.Source.Microphone,
                          }}
                          show={'muted'}
                        />
                        <ParticipantName />
                      </>
                    ) : (
                      <>
                        <ScreenShareIcon style={{ marginRight: '0.25rem' }} />
                        <ParticipantName>&apos;s screen</ParticipantName>
                      </>
                    )}
                  </div>
                  <ConnectionQualityIndicator className="lk-participant-metadata-item" />
                </div>
              </>
            )}
            <ParticipantExpandToggle trackRef={trackReference} />
          </ParticipantContextIfNeeded>
        </TrackRefContextIfNeeded>
      </div>
    )
  },
)

export type VocoFocusLayoutProps = Omit<React.HTMLAttributes<HTMLElement>, 'onParticipantClick'> & {
  trackRef?: TrackReferenceOrPlaceholder
  onParticipantClick?: (evt: ParticipantClickEvent) => void
}

export function VocoFocusLayout({ trackRef, ...htmlProps }: VocoFocusLayoutProps) {
  return <VocoParticipantTile trackRef={trackRef} {...htmlProps} />
}
