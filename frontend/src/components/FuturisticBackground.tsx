import './futuristicBackground.css'

export function FuturisticBackground() {
  return (
    <div className="fxBg" aria-hidden>
      <div className="fxBg__base" />
      <div className="fxBg__ribbon" />
      <div className="fxBg__glow fxBg__glow--a" />
      <div className="fxBg__glow fxBg__glow--b" />
      <div className="fxBg__noise" />
    </div>
  )
}

