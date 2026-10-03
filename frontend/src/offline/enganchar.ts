import { useEffect } from "react"
import { vaciarRegistros } from "./registros"

/**
 * Escucha `online` y vacía la cola que recibe (estados y altas) y, después, los registros de campo.
 * No toca la cola de evidencias: esa ya tiene su propio listener.
 */
export function escucharOnline(destino: EventTarget, vaciar: () => void | Promise<void>): () => void {
  const alVolver = () => {
    void Promise.resolve(vaciar()).then(() => vaciarRegistros())
  }
  destino.addEventListener("online", alVolver)
  return () => destino.removeEventListener("online", alVolver)
}

/** Vacía la cola local al montar y al volver la red, cuando ya hay sesión. */
export function useEngancharCola(sesion: object | null, vaciar: () => Promise<void>): void {
  useEffect(() => {
    if (!sesion) return
    let cancelled = false
    void Promise.resolve().then(() => {
      if (cancelled) return
      return vaciar().then(() => {
        if (!cancelled) return vaciarRegistros()
      })
    })
    const parar = escucharOnline(window, () => {
      if (cancelled) return
      return vaciar()
    })
    return () => {
      cancelled = true
      parar()
    }
  }, [vaciar, sesion])
}
