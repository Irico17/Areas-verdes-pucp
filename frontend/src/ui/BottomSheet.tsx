import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type PointerEvent as ReactPointerEvent,
  type ReactNode,
} from "react"
import { ANCLAJES, alturaPx, anclajeCercano, siguienteAnclaje, type Anclaje } from "./bottomSheet"
import { mostrarEnPanel } from "./desplazar"
import { MQ_MOVIL, useMedia } from "./media"
import { useTeclado } from "./teclado"

const MEDIA: Anclaje = 0.55
const UMBRAL = 4

type Props = {
  open: boolean
  onClose: () => void
  /** Módulo activo: al cambiar, el cuerpo vuelve arriba. */
  vista: string
  /** Línea fija sobre el cuerpo (Lote B). */
  cabeza?: ReactNode
  /** Alto que tapa la hoja en el teléfono, para los controles y el encuadre del mapa. */
  onAltura?: (px: number) => void
  children: ReactNode
}

export function BottomSheet({ open, onClose, vista, cabeza, onAltura, children }: Props) {
  const movil = useMedia(MQ_MOVIL)
  const hoja = useRef<HTMLElement>(null)
  const [anclaje, setAnclaje] = useState<Anclaje>(MEDIA)
  const [arrastre, setArrastre] = useState<number | null>(null)
  const [vistoAbierto, setVistoAbierto] = useState(open)
  const inicio = useRef({ y: 0, alto: 0, tope: 0 })
  const arrastreRef = useRef<number | null>(null)
  const movio = useRef(false)
  const teclado = useTeclado(movil && open)
  if (open !== vistoAbierto) {
    setVistoAbierto(open)
    if (open) setAnclaje(MEDIA)
  }

  function tope(): number {
    const px = hoja.current ? parseFloat(getComputedStyle(hoja.current).maxHeight) : NaN
    return Number.isFinite(px) && px > 0 ? px : window.innerHeight * 0.8
  }

  useEffect(() => {
    if (!onAltura) return
    if (!movil || !open || teclado.abierto) {
      onAltura(0)
      return
    }
    const publicar = () => onAltura(arrastre ?? alturaPx(anclaje, tope()))
    publicar()
    window.addEventListener("resize", publicar)
    return () => window.removeEventListener("resize", publicar)
  }, [movil, open, anclaje, arrastre, teclado.abierto, onAltura])

  useEffect(() => {
    const raiz = document.documentElement
    raiz.toggleAttribute("data-teclado", teclado.abierto)
    if (teclado.abierto) requestAnimationFrame(() => mostrarEnPanel(document.activeElement, "cerca"))
    return () => raiz.removeAttribute("data-teclado")
  }, [teclado.abierto, teclado.alto])

  function alBajar(event: ReactPointerEvent<HTMLElement>) {
    if (!movil) return
    inicio.current = { y: event.clientY, alto: hoja.current?.getBoundingClientRect().height ?? 0, tope: tope() }
    movio.current = false
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  function alMover(event: ReactPointerEvent<HTMLElement>) {
    if (!movil || !event.currentTarget.hasPointerCapture(event.pointerId)) return
    const delta = inicio.current.y - event.clientY
    if (!movio.current && Math.abs(delta) < UMBRAL) return
    movio.current = true
    const alto = Math.min(inicio.current.tope, Math.max(48, inicio.current.alto + delta))
    arrastreRef.current = alto
    setArrastre(alto)
  }

  function alSoltar(event: ReactPointerEvent<HTMLElement>) {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    const alto = arrastreRef.current
    if (!movil || alto == null) return
    arrastreRef.current = null
    setArrastre(null)
    const destino = anclajeCercano(alto / inicio.current.tope)
    if (destino == null) onClose()
    else setAnclaje(destino)
  }

  function alTocarAsa() {
    if (movio.current) {
      movio.current = false
      return
    }
    const ultimo = ANCLAJES[ANCLAJES.length - 1]
    setAnclaje((actual) => (actual === ultimo ? MEDIA : (siguienteAnclaje(actual, 1) ?? MEDIA)))
  }

  function flechas(event: ReactKeyboardEvent) {
    if (!movil || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) return
    event.preventDefault()
    const siguiente = siguienteAnclaje(anclaje, event.key === "ArrowUp" ? 1 : -1)
    if (siguiente == null) onClose()
    else setAnclaje(siguiente)
  }

  function escape(event: ReactKeyboardEvent) {
    if (!movil || event.key !== "Escape" || event.defaultPrevented) return
    event.stopPropagation()
    onClose()
  }

  const estilo = !movil
    ? undefined
    : teclado.abierto
      ? { top: `${teclado.top}px`, bottom: "auto", height: `${teclado.alto}px` }
      : { height: arrastre != null ? `${arrastre}px` : `calc(var(--sheet-util) * ${anclaje})` }
  const clases = ["panel", movil && "sheet", arrastre != null && "arrastrando", teclado.abierto && "con-teclado"]
    .filter(Boolean)
    .join(" ")
  const indice = ANCLAJES.indexOf(anclaje)

  return (
    <aside ref={hoja} className={clases} id="panel" style={estilo} aria-label="Panel" onKeyDown={escape}>
      <div className="sheet-chrome">
        <button
          type="button"
          className="sheet-handle"
          role="slider"
          aria-label="Arrastrar el panel. Flechas para subir o bajar."
          aria-valuemin={0}
          aria-valuemax={ANCLAJES.length - 1}
          aria-valuenow={Math.max(indice, 0)}
          aria-valuetext={indice === 0 ? "Vista mínima" : indice === 1 ? "Media altura" : "Casi pantalla completa"}
          onKeyDown={flechas}
          onClick={alTocarAsa}
          onPointerDown={alBajar}
          onPointerMove={alMover}
          onPointerCancel={alSoltar}
          onLostPointerCapture={alSoltar}
        >
          <span className="sheet-grab" aria-hidden="true" />
        </button>
        <button type="button" className="sheet-close" onClick={onClose}>
          Cerrar
        </button>
      </div>
      {cabeza && <div className="panel-head">{cabeza}</div>}
      <div className="panel-view" key={vista}>
        {children}
      </div>
    </aside>
  )
}
