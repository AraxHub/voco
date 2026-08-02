import { useFancyEffectsAttr } from '../hooks/useFancyEffects'

/** Cheap static wash — radial gradients, no filter:blur, no animation. */
function LiteBackground() {
  return (
    <div className="voco-bg fixed inset-0 overflow-hidden" style={{ zIndex: 0 }} aria-hidden>
      <div className="voco-bg__blob voco-bg__blob--1 voco-bg__blob--static absolute rounded-full pointer-events-none" />
      <div className="voco-bg__blob voco-bg__blob--2 voco-bg__blob--static absolute rounded-full pointer-events-none" />
      <div className="voco-bg__vignette absolute inset-0 pointer-events-none" />
      <svg
        className="absolute inset-0 h-full w-full pointer-events-none"
        viewBox="0 0 1440 900"
        preserveAspectRatio="xMidYMid slice"
        aria-hidden
      >
        <defs>
          <linearGradient id="railA" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-a-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-a)" />
            <stop offset="100%" stopColor="var(--voco-rail-a-dim)" />
          </linearGradient>
          <linearGradient id="railB" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-b-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-b)" />
            <stop offset="100%" stopColor="var(--voco-rail-b-dim)" />
          </linearGradient>
          <linearGradient id="railC" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-c-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-c)" />
            <stop offset="100%" stopColor="var(--voco-rail-c-dim)" />
          </linearGradient>
        </defs>
        <g strokeWidth="1" fill="none" opacity="0.28">
          <path d="M60 110 H380 L430 160 H620" stroke="url(#railA)" />
          <path d="M430 160 V280 L510 360 H740" stroke="url(#railA)" />
          <path d="M980 95 H1280 L1340 155 V290" stroke="url(#railB)" />
          <path d="M1120 155 H1380" stroke="url(#railB)" />
          <path d="M80 720 H320 L390 650 H560 L620 710 H860" stroke="url(#railC)" />
          <path d="M1050 780 H1320 L1380 720 V580" stroke="url(#railB)" />
          <path d="M700 40 V140 L780 220 H980" stroke="url(#railA)" />
          <path d="M40 420 H180 L240 480 H420" stroke="url(#railB)" />
        </g>
      </svg>
    </div>
  )
}

function FancyBackground() {
  return (
    <div className="voco-bg fixed inset-0 overflow-hidden" style={{ zIndex: 0 }}>
      <div className="aurora-1 voco-bg__blob voco-bg__blob--1 absolute rounded-full pointer-events-none" />
      <div className="aurora-2 voco-bg__blob voco-bg__blob--2 absolute rounded-full pointer-events-none" />
      <div className="aurora-3 voco-bg__blob voco-bg__blob--3 absolute rounded-full pointer-events-none" />
      <div className="aurora-4 voco-bg__blob voco-bg__blob--4 absolute rounded-full pointer-events-none" />

      <div className="surge-bloom surge-bloom--a pointer-events-none" />
      <div className="surge-bloom surge-bloom--b pointer-events-none" />
      <div className="surge-bloom surge-bloom--c pointer-events-none" />

      <div className="voco-bg__vignette absolute inset-0 pointer-events-none" />

      <svg
        className="voco-bg__fx absolute inset-0 h-full w-full pointer-events-none"
        viewBox="0 0 1440 900"
        preserveAspectRatio="xMidYMid slice"
        aria-hidden
      >
        <defs>
          <linearGradient id="railA" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-a-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-a)" />
            <stop offset="100%" stopColor="var(--voco-rail-a-dim)" />
          </linearGradient>
          <linearGradient id="railB" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-b-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-b)" />
            <stop offset="100%" stopColor="var(--voco-rail-b-dim)" />
          </linearGradient>
          <linearGradient id="railC" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--voco-rail-c-dim)" />
            <stop offset="50%" stopColor="var(--voco-rail-c)" />
            <stop offset="100%" stopColor="var(--voco-rail-c-dim)" />
          </linearGradient>
          <filter id="surgeGlow" x="-40%" y="-40%" width="180%" height="180%">
            <feGaussianBlur stdDeviation="2.2" result="b" />
            <feMerge>
              <feMergeNode in="b" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <g strokeWidth="1" fill="none" opacity="0.35">
          <path d="M60 110 H380 L430 160 H620" stroke="url(#railA)" />
          <path d="M430 160 V280 L510 360 H740" stroke="url(#railA)" />
          <path d="M980 95 H1280 L1340 155 V290" stroke="url(#railB)" />
          <path d="M1120 155 H1380" stroke="url(#railB)" />
          <path d="M80 720 H320 L390 650 H560 L620 710 H860" stroke="url(#railC)" />
          <path d="M1050 780 H1320 L1380 720 V580" stroke="url(#railB)" />
          <path d="M700 40 V140 L780 220 H980" stroke="url(#railA)" />
          <path d="M40 420 H180 L240 480 H420" stroke="url(#railB)" />
        </g>

        <g fill="none" strokeLinecap="round" filter="url(#surgeGlow)">
          <path className="bg-surge bg-surge--1" d="M60 110 H380 L430 160 H620" stroke="var(--voco-surge-a)" strokeWidth="1.6" />
          <path className="bg-surge bg-surge--2" d="M430 160 V280 L510 360 H740" stroke="var(--voco-surge-b)" strokeWidth="1.5" />
          <path className="bg-surge bg-surge--3" d="M980 95 H1280 L1340 155 V290" stroke="var(--voco-surge-a)" strokeWidth="1.5" />
          <path className="bg-surge bg-surge--4" d="M80 720 H320 L390 650 H560 L620 710 H860" stroke="var(--voco-surge-c)" strokeWidth="1.4" />
          <path className="bg-surge bg-surge--5" d="M1050 780 H1320 L1380 720 V580" stroke="var(--voco-surge-b)" strokeWidth="1.5" />
          <path className="bg-surge bg-surge--6" d="M700 40 V140 L780 220 H980" stroke="var(--voco-surge-a)" strokeWidth="1.35" />
          <path className="bg-surge bg-surge--7" d="M40 420 H180 L240 480 H420" stroke="var(--voco-surge-b)" strokeWidth="1.35" />
        </g>
      </svg>
    </div>
  )
}

export default function Background() {
  const fancy = useFancyEffectsAttr()
  return fancy ? <FancyBackground /> : <LiteBackground />
}
