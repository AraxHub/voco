import { useEffect, useState } from 'react'

/**
 * Fancy FX only when the machine is clearly capable.
 * Default is lite — weak laptops and phones both struggle with
 * continuous blur + SVG dash animations.
 *
 * Gate: desktop pointer, wide viewport, ≥8 cores, ≥8GB (if reported),
 * no Save-Data, no prefers-reduced-motion.
 */
export function useFancyEffects() {
  const [fancy, setFancy] = useState(false)

  useEffect(() => {
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)')
    const desktop = window.matchMedia('(hover: hover) and (pointer: fine)')
    const wide = window.matchMedia('(min-width: 1100px)')

    const evaluate = () => {
      const nav = navigator as Navigator & { deviceMemory?: number; connection?: { saveData?: boolean } }
      const cores = navigator.hardwareConcurrency || 0
      const memory = nav.deviceMemory // Chrome; GiB, may be undefined
      const saveData = Boolean(nav.connection?.saveData)

      setFancy(
        !reduced.matches &&
          desktop.matches &&
          wide.matches &&
          !saveData &&
          cores >= 8 &&
          (memory === undefined || memory >= 8),
      )
    }

    evaluate()
    reduced.addEventListener('change', evaluate)
    desktop.addEventListener('change', evaluate)
    wide.addEventListener('change', evaluate)
    return () => {
      reduced.removeEventListener('change', evaluate)
      desktop.removeEventListener('change', evaluate)
      wide.removeEventListener('change', evaluate)
    }
  }, [])

  return fancy
}

/** Sync html[data-fx] for CSS progressive enhancement. Call once from App shell. */
export function useFancyEffectsAttr() {
  const fancy = useFancyEffects()

  useEffect(() => {
    document.documentElement.dataset.fx = fancy ? 'fancy' : 'lite'
    return () => {
      delete document.documentElement.dataset.fx
    }
  }, [fancy])

  return fancy
}
