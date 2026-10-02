import { useEffect } from "react"

/** Vacía la cola local al montar, cuando ya hay sesión. */
export function useEngancharCola(sesion: object | null, vaciar: () => Promise<void>): void {
  useEffect(() => {
    if (!sesion) return
    let cancelled = false
    void Promise.resolve().then(() => {
      if (!cancelled) return vaciar()
    })
    return () => {
      cancelled = true
    }
  }, [vaciar, sesion])
}
