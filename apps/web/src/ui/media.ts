import { useCallback, useSyncExternalStore } from "react"

/** Único corte entre teléfono y escritorio. styles.css usa el mismo valor. */
export const MQ_MOVIL = "(max-width: 820px)"

export function useMedia(query: string): boolean {
  const suscribir = useCallback(
    (avisar: () => void) => {
      const lista = window.matchMedia(query)
      lista.addEventListener("change", avisar)
      return () => lista.removeEventListener("change", avisar)
    },
    [query],
  )
  return useSyncExternalStore(suscribir, () => window.matchMedia(query).matches, () => false)
}
