import { useEffect, useState } from "react"

export type Teclado = { abierto: boolean; alto: number; top: number }
const CERRADO: Teclado = { abierto: false, alto: 0, top: 0 }

/** Parte baja del layout viewport que tapa el teclado, en px. */
export function altoTapado(innerHeight: number, vv: { height: number; offsetTop: number }): number {
  return Math.max(0, Math.round(innerHeight - vv.height - vv.offsetTop))
}

export function tecladoAbierto(tapado: number, escala: number, editable: boolean): boolean {
  return editable && tapado > 120 && Math.abs(escala - 1) < 0.01
}

function esEditable(el: Element | null): boolean {
  if (el instanceof HTMLTextAreaElement || el instanceof HTMLSelectElement) return true
  if (el instanceof HTMLInputElement) {
    return !["checkbox", "radio", "button", "submit", "reset", "file", "range", "color"].includes(el.type)
  }
  return el instanceof HTMLElement && el.isContentEditable
}

/** Estado del teclado del teléfono a partir del visual viewport. */
export function useTeclado(activo: boolean): Teclado {
  const [estado, setEstado] = useState<Teclado>(CERRADO)
  useEffect(() => {
    const vv = window.visualViewport
    if (!activo || !vv) return
    let cuadro = 0
    const medir = () => {
      cancelAnimationFrame(cuadro)
      cuadro = requestAnimationFrame(() => {
        const abierto = tecladoAbierto(altoTapado(window.innerHeight, vv), vv.scale, esEditable(document.activeElement))
        const siguiente = abierto ? { abierto, alto: Math.round(vv.height), top: Math.round(vv.offsetTop) } : CERRADO
        setEstado((prev) =>
          prev.abierto === siguiente.abierto && prev.alto === siguiente.alto && prev.top === siguiente.top ? prev : siguiente,
        )
      })
    }
    vv.addEventListener("resize", medir)
    vv.addEventListener("scroll", medir)
    document.addEventListener("focusin", medir)
    document.addEventListener("focusout", medir)
    medir()
    return () => {
      cancelAnimationFrame(cuadro)
      vv.removeEventListener("resize", medir)
      vv.removeEventListener("scroll", medir)
      document.removeEventListener("focusin", medir)
      document.removeEventListener("focusout", medir)
    }
  }, [activo])
  return activo ? estado : CERRADO
}
