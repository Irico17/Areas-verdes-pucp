import { useEffect, useRef, useState, type PointerEvent as ReactPointerEvent, type ReactNode } from "react"
import { ANCLAJES, anclajeCercano, siguienteAnclaje, type Anclaje } from "./bottomSheet"

const MOVIL = "(max-width: 820px)"

function esMovil() {
  return window.matchMedia(MOVIL).matches
}

function reducido() {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches
}

export function BottomSheet({
  open,
  onClose,
  children,
}: {
  open: boolean
  onClose: () => void
  children: ReactNode
}) {
  const [movil, setMovil] = useState(esMovil)
  const [anclaje, setAnclaje] = useState<Anclaje>(0.55)
  const [arrastre, setArrastre] = useState<number | null>(null)
  const [vistoAbierto, setVistoAbierto] = useState(open)
  const inicio = useRef({ y: 0, alto: 0 })
  const arrastreRef = useRef<number | null>(null)
  if (open !== vistoAbierto) {
    setVistoAbierto(open)
    if (open) setAnclaje(0.55)
  }

  useEffect(() => {
    const media = window.matchMedia(MOVIL)
    const sync = () => setMovil(media.matches)
    media.addEventListener("change", sync)
    return () => media.removeEventListener("change", sync)
  }, [])

  useEffect(() => {
    if (!movil || !open) {
      document.documentElement.style.removeProperty("--sheet-h")
      return
    }
    const alto = arrastre ?? Math.round(window.innerHeight * anclaje)
    document.documentElement.style.setProperty("--sheet-h", `${alto}px`)
    return () => {
      document.documentElement.style.removeProperty("--sheet-h")
    }
  }, [movil, open, anclaje, arrastre])

  function alBajar(event: ReactPointerEvent<HTMLElement>) {
    if (!movil) return
    inicio.current = { y: event.clientY, alto: Math.round(window.innerHeight * anclaje) }
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  function alMover(event: ReactPointerEvent<HTMLElement>) {
    if (!movil || !event.currentTarget.hasPointerCapture(event.pointerId)) return
    const delta = inicio.current.y - event.clientY
    const alto = Math.max(48, inicio.current.alto + delta)
    arrastreRef.current = alto
    setArrastre(alto)
  }

  function alSoltar(event: ReactPointerEvent<HTMLElement>) {
    const alto = arrastreRef.current
    if (!movil || alto == null) return
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    arrastreRef.current = null
    const ratio = alto / window.innerHeight
    const destino = anclajeCercano(ratio)
    setArrastre(null)
    if (destino == null) onClose()
    else setAnclaje(destino)
  }

  function tecla(event: React.KeyboardEvent) {
    if (!movil) return
    if (event.key === "Escape") {
      onClose()
      return
    }
    if (event.key === "ArrowUp" || event.key === "ArrowDown") {
      event.preventDefault()
      const dir = event.key === "ArrowUp" ? 1 : -1
      const siguiente = siguienteAnclaje(anclaje, dir as 1 | -1)
      if (siguiente == null) onClose()
      else setAnclaje(siguiente)
    }
  }

  const altura = movil && open ? (arrastre ?? `calc(${anclaje * 100}dvh)`) : undefined
  const indice = ANCLAJES.indexOf(anclaje)

  return (
    <aside
      className={movil ? "panel sheet" : "panel"}
      id="panel"
      style={altura == null ? undefined : { height: typeof altura === "number" ? `${altura}px` : altura, transition: arrastre != null || reducido() ? "none" : undefined }}
      aria-label="Panel"
    >
      <div className="sheet-chrome">
        <button
          type="button"
          className="sheet-handle"
          aria-label="Arrastrar el panel. Flechas para subir o bajar."
          aria-valuemin={0}
          aria-valuemax={ANCLAJES.length - 1}
          aria-valuenow={Math.max(indice, 0)}
          aria-valuetext={indice === 0 ? "Vista mínima" : indice === 1 ? "Media altura" : "Casi pantalla completa"}
          role="slider"
          onKeyDown={tecla}
          onPointerDown={alBajar}
          onPointerMove={alMover}
          onPointerUp={alSoltar}
          onPointerCancel={alSoltar}
        >
          <span className="sheet-grab" aria-hidden="true" />
        </button>
        <button type="button" className="sheet-close" onClick={onClose}>
          Cerrar
        </button>
      </div>
      <div className="panel-view">{children}</div>
    </aside>
  )
}
