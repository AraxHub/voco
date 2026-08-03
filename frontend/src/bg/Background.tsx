import { useFancyEffectsAttr } from '../hooks/useFancyEffects'

/** Soft wash — no circuit noise, calm Telegram-like atmosphere. */
function LiteBackground() {
  return (
    <div className="voco-bg fixed inset-0 overflow-hidden" style={{ zIndex: 0 }} aria-hidden>
      <div className="voco-bg__blob voco-bg__blob--1 voco-bg__blob--static absolute rounded-full pointer-events-none" />
      <div className="voco-bg__blob voco-bg__blob--2 voco-bg__blob--static absolute rounded-full pointer-events-none" />
      <div className="voco-bg__vignette absolute inset-0 pointer-events-none" />
    </div>
  )
}

function FancyBackground() {
  return (
    <div className="voco-bg fixed inset-0 overflow-hidden" style={{ zIndex: 0 }} aria-hidden>
      <div className="aurora-1 voco-bg__blob voco-bg__blob--1 absolute rounded-full pointer-events-none" />
      <div className="aurora-2 voco-bg__blob voco-bg__blob--2 absolute rounded-full pointer-events-none" />
      <div className="aurora-3 voco-bg__blob voco-bg__blob--3 absolute rounded-full pointer-events-none" />
      <div className="voco-bg__vignette absolute inset-0 pointer-events-none" />
    </div>
  )
}

export default function Background() {
  const fancy = useFancyEffectsAttr()
  return fancy ? <FancyBackground /> : <LiteBackground />
}
