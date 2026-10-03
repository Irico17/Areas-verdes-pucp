/** Corte del panel de capas. El resto de la app sigue en 820 px. */
export const MQ_PANEL_MAPA = "(max-width: 640px)"

const PREFIJO = "cv:panel-mapa"

export function clavePanelMapa(usuario: string, estrecho: boolean): string {
  const quien = usuario.trim() || "anonimo"
  return `${PREFIJO}:${quien}:${estrecho ? "estrecho" : "ancho"}`
}

function almacen(): Storage | null {
  try {
    return globalThis.localStorage
  } catch {
    return null
  }
}

/** `null` si no hay preferencia guardada o el almacenamiento no responde. */
export function leerPreferenciaPanel(usuario: string, estrecho: boolean): boolean | null {
  const caja = almacen()
  if (!caja) return null
  try {
    const valor = caja.getItem(clavePanelMapa(usuario, estrecho))
    if (valor === "1") return true
    if (valor === "0") return false
    return null
  } catch {
    return null
  }
}

export function guardarPreferenciaPanel(usuario: string, estrecho: boolean, abierto: boolean): void {
  const caja = almacen()
  if (!caja) return
  try {
    caja.setItem(clavePanelMapa(usuario, estrecho), abierto ? "1" : "0")
  } catch {
    /* el navegador puede rechazar el almacenamiento */
  }
}

/** En el teléfono arranca cerrado. En escritorio, abierto. */
export function panelAbiertoInicial(usuario: string, estrecho: boolean): boolean {
  const guardado = leerPreferenciaPanel(usuario, estrecho)
  if (guardado == null) return !estrecho
  return guardado
}
