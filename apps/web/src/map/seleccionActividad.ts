export type AccionesSeleccion = {
  elegir: (id: string) => void
  irALabores: () => void
  abrirPanel: () => void
  limpiar: () => void
}

/** Elige la actividad, deja el resumen en el popup y abre el detalle en el panel. La pestaña no cambia. */
export function seleccionarActividad(id: string | null, acciones: AccionesSeleccion): void {
  if (id) {
    acciones.elegir(id)
    acciones.abrirPanel()
  } else {
    acciones.limpiar()
  }
}
