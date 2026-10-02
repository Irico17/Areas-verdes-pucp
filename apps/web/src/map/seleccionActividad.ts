export type AccionesSeleccion = {
  elegir: (id: string) => void
  irALabores: () => void
  abrirPanel: () => void
  limpiar: () => void
}

/** Elige la actividad en el mapa y abre la pestaña de labores. Sin id, limpia la selección. */
export function seleccionarActividad(id: string | null, acciones: AccionesSeleccion): void {
  if (id) {
    acciones.elegir(id)
    acciones.irALabores()
    acciones.abrirPanel()
  } else {
    acciones.limpiar()
  }
}
