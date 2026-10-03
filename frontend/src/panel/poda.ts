export type PodaItem = {
  id: string
  codigo: string
  codigo_externo: string
  tipo: string
  tipo_actividad: string
  fecha_reporte: string
  fecha_ejecucion: string
  personal: string
  ubicacion: string
  unidad: string
  cantidad_pedida: string
  cantidad_ejecutada: string
  prioridad: string
  comentario: string
  nombre_comun: string
  nombre_cientifico: string
}

export type Fallo = { campo: string; motivo: string }

export const PODA_VACIA: PodaItem = {
  id: "",
  codigo: "",
  codigo_externo: "",
  tipo: "",
  tipo_actividad: "",
  fecha_reporte: "",
  fecha_ejecucion: "",
  personal: "",
  ubicacion: "",
  unidad: "",
  cantidad_pedida: "0",
  cantidad_ejecutada: "0",
  prioridad: "media",
  comentario: "",
  nombre_comun: "",
  nombre_cientifico: "",
}

export function codigoExterno(valor: string): { codigo: string; fallo?: Fallo } {
  const v = valor.trim()
  if (!v || /^(aun no tiene codigo|no aplica|sin codigo)$/i.test(v)) return { codigo: "" }
  if (/^OSG-/i.test(v)) return { codigo: v.toUpperCase() }
  return { codigo: "", fallo: { campo: "codigo_externo", motivo: "El código OSG no se genera. Déjelo vacío o copie el que ya trae la fuente." } }
}

export function validarPoda(item: PodaItem): Fallo[] {
  const fallos: Fallo[] = []
  if (!/^PO-\d+$/.test(item.codigo.trim())) fallos.push({ campo: "codigo", motivo: "El código es PO-n." })
  const externo = codigoExterno(item.codigo_externo)
  if (externo.fallo) fallos.push(externo.fallo)
  if (item.fecha_reporte && item.fecha_ejecucion && item.fecha_ejecucion < item.fecha_reporte) {
    fallos.push({ campo: "fecha_ejecucion", motivo: "La ejecución no puede ser anterior al reporte." })
  }
  for (const campo of ["cantidad_pedida", "cantidad_ejecutada"] as const) {
    const n = Number(item[campo])
    if (!Number.isFinite(n) || n < 0) fallos.push({ campo, motivo: "La cantidad es cero o más." })
  }
  if (!["baja", "media", "alta"].includes(item.prioridad)) fallos.push({ campo: "prioridad", motivo: "Prioridad no catalogada." })
  return fallos
}
