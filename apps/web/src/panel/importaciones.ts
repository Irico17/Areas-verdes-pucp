export const ENTIDADES: { id: string; etiqueta: string }[] = [
  { id: "areas_verdes", etiqueta: "Áreas verdes" },
  { id: "zonas_supervision", etiqueta: "Zonas de supervisión" },
  { id: "poligonos_cuadrilla", etiqueta: "Polígonos de cuadrilla" },
  { id: "cuadrillas", etiqueta: "Cuadrillas" },
  { id: "lugares", etiqueta: "Lugares" },
  { id: "ejemplares", etiqueta: "Ejemplares" },
  { id: "palmeras", etiqueta: "Palmeras" },
  { id: "cafetos", etiqueta: "Cafetos" },
  { id: "catalogo_actividades", etiqueta: "Catálogo de actividades" },
  { id: "labores", etiqueta: "Labores" },
  { id: "poda", etiqueta: "Poda" },
  { id: "vivero", etiqueta: "Vivero" },
  { id: "tachos", etiqueta: "Tachos" },
  { id: "bebederos", etiqueta: "Bebederos" },
  { id: "fauna", etiqueta: "Fauna" },
  { id: "puertas", etiqueta: "Puertas" },
  { id: "playas", etiqueta: "Playas" },
  { id: "vereda", etiqueta: "Vereda en riesgo" },
  { id: "xerofitica", etiqueta: "Xerofítica" },
  { id: "jardines_reserva", etiqueta: "Jardines de reserva" },
  { id: "reservas", etiqueta: "Reservas" },
  { id: "puntos_pucp", etiqueta: "Puntos PUCP" },
]

export type ErrorFila = { fila: number; campo: string; motivo: string }

export type VistaPrevia = {
  id: number
  lote_id: number
  entidad: string
  formato: string
  validas: number
  errores: ErrorFila[]
  filas: Record<string, unknown>[]
  columnas_omitidas?: string[]
  aviso_omitidas?: string
  avisos?: string[]
  escrito: boolean
}

export async function previsualizar(entidad: string, archivo: File): Promise<VistaPrevia> {
  const datos = new FormData()
  datos.set("archivo", archivo)
  const res = await fetch(`/api/v1/importaciones?entidad=${encodeURIComponent(entidad)}`, {
    method: "POST",
    body: datos,
    credentials: "include",
  })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json() as Promise<VistaPrevia>
}

export async function confirmarImportacion(id: number): Promise<{ lote_id: number; validas: number; escrito: boolean }> {
  const res = await fetch(`/api/v1/importaciones/${id}/confirmar`, { method: "POST", credentials: "include" })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json() as Promise<{ lote_id: number; validas: number; escrito: boolean }>
}

export async function revertirLote(id: number, confirmar: boolean): Promise<{ lote_id: number; revertidas: string[]; excluidas: { entidad_id: string; motivo: string }[] }> {
  const res = await fetch(`/api/v1/lotes/${id}/revertir`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ confirmar }),
  })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json() as Promise<{ lote_id: number; revertidas: string[]; excluidas: { entidad_id: string; motivo: string }[] }>
}

async function mensaje(res: Response): Promise<string> {
  const data = (await res.json().catch(() => null)) as { error?: string } | null
  return data?.error || "No se pudo completar la importación"
}
