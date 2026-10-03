import { useEffect, useId, useRef, useState, type ReactNode } from "react"
import { MAPA } from "../ui/nomenclatura"
import { useMedia } from "../ui/media"
import { MQ_PANEL_MAPA, guardarPreferenciaPanel, panelAbiertoInicial } from "./panelMapa"

type Props = {
  usuario: string
  vista: ReactNode
  control: ReactNode
}

function IconoCapas() {
  return (
    <svg className="icono-capas" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 3.5 3.5 8 12 12.5 20.5 8 12 3.5z" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" />
      <path d="M3.5 12 12 16.5 20.5 12" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M3.5 16 12 20.5 20.5 16" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function IconoOcultar() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M6 14.5 12 8.5 18 14.5" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export function MarcoMapa({ usuario, vista, control }: Props) {
  const estrecho = useMedia(MQ_PANEL_MAPA)
  const [visto, setVisto] = useState({ usuario, estrecho })
  const [abierto, setAbierto] = useState(() => panelAbiertoInicial(usuario, window.matchMedia(MQ_PANEL_MAPA).matches))
  if (visto.usuario !== usuario || visto.estrecho !== estrecho) {
    setVisto({ usuario, estrecho })
    setAbierto(panelAbiertoInicial(usuario, estrecho))
  }
  const abrirBtn = useRef<HTMLButtonElement>(null)
  const cerrarBtn = useRef<HTMLButtonElement>(null)
  const idCapas = useId()

  function fijar(siguiente: boolean, enfocar: "abrir" | "cerrar" | "nada") {
    setAbierto(siguiente)
    guardarPreferenciaPanel(usuario, estrecho, siguiente)
    if (enfocar === "nada") return
    const ref = enfocar === "cerrar" ? cerrarBtn : abrirBtn
    requestAnimationFrame(() => ref.current?.focus({ preventScroll: true }))
  }

  useEffect(() => {
    if (!abierto) return
    function tecla(event: KeyboardEvent) {
      if (event.key !== "Escape" || event.defaultPrevented) return
      const activo = document.activeElement
      if (activo instanceof Element && activo.closest("dialog, .cuenta-menu, .guard-menu, .recorrido")) return
      event.preventDefault()
      event.stopPropagation()
      setAbierto(false)
      guardarPreferenciaPanel(usuario, estrecho, false)
      requestAnimationFrame(() => abrirBtn.current?.focus({ preventScroll: true }))
    }
    document.addEventListener("keydown", tecla)
    return () => document.removeEventListener("keydown", tecla)
  }, [abierto, usuario, estrecho])

  useEffect(() => {
    if (!abierto || !estrecho) return
    function puntero(event: PointerEvent) {
      const objetivo = event.target
      if (!(objetivo instanceof Element)) return
      if (objetivo.closest(".marco-mapa, .topbar, .guard, .guard-menu, .cuenta-menu, .recorrido, .maplibregl-ctrl, .cv-popup, .lista-sobre-mapa, .panel, .skip")) return
      event.preventDefault()
      event.stopPropagation()
      setAbierto(false)
      guardarPreferenciaPanel(usuario, estrecho, false)
    }
    document.addEventListener("pointerdown", puntero, true)
    return () => document.removeEventListener("pointerdown", puntero, true)
  }, [abierto, estrecho, usuario])

  return (
    <div className="marco-mapa" data-abierto={abierto ? "si" : "no"} data-ancho={estrecho ? "estrecho" : "ancho"}>
      {!abierto && (
        <button
          ref={abrirBtn}
          type="button"
          className="marco-mapa-abrir"
          aria-expanded={false}
          aria-controls={idCapas}
          aria-label={MAPA.mostrarCapas}
          onClick={() => fijar(true, "cerrar")}
        >
          <IconoCapas />
          <span>{MAPA.capas}</span>
        </button>
      )}
      <div className="marco-mapa-vista">{vista}</div>
      {abierto && (
        <div className="marco-mapa-capas" id={idCapas}>
          <div className="marco-mapa-cromo">
            <button
              ref={cerrarBtn}
              type="button"
              className="marco-mapa-cerrar"
              aria-expanded={true}
              aria-controls={idCapas}
              aria-label={estrecho ? MAPA.cerrarPanel : MAPA.ocultarPanel}
              onClick={() => fijar(false, "abrir")}
            >
              {estrecho ? MAPA.cerrarPanel : <IconoOcultar />}
            </button>
          </div>
          <div className="marco-mapa-cuerpo">{control}</div>
        </div>
      )}
    </div>
  )
}
