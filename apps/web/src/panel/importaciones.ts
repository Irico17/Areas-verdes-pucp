import { apiUrl } from "../api"
import { IMPORTACION } from "../ui/nomenclatura"

export const ENTIDADES: { id: string; etiqueta: string }[] = [
  { id: "areas_verdes", etiqueta: IMPORTACION.areas_verdes },
  { id: "zonas_supervision", etiqueta: IMPORTACION.zonas_supervision },
  { id: "poligonos_cuadrilla", etiqueta: IMPORTACION.poligonos_cuadrilla },
  { id: "cuadrillas", etiqueta: IMPORTACION.cuadrillas },
  { id: "lugares", etiqueta: IMPORTACION.lugares },
  { id: "ejemplares", etiqueta: IMPORTACION.ejemplares },
  { id: "palmeras", etiqueta: IMPORTACION.palmeras },
  { id: "cafetos", etiqueta: IMPORTACION.cafetos },
  { id: "catalogo_actividades", etiqueta: IMPORTACION.catalogo_actividades },
  { id: "labores", etiqueta: IMPORTACION.labores },
  { id: "poda", etiqueta: IMPORTACION.poda },
  { id: "vivero", etiqueta: IMPORTACION.vivero },
  { id: "tachos", etiqueta: IMPORTACION.tachos },
  { id: "bebederos", etiqueta: IMPORTACION.bebederos },
  { id: "fauna", etiqueta: IMPORTACION.fauna },
  { id: "puertas", etiqueta: IMPORTACION.puertas },
  { id: "playas", etiqueta: IMPORTACION.playas },
  { id: "vereda", etiqueta: IMPORTACION.vereda },
  { id: "xerofitica", etiqueta: IMPORTACION.xerofitica },
  { id: "jardines_reserva", etiqueta: IMPORTACION.jardines_reserva },
  { id: "reservas", etiqueta: IMPORTACION.reservas },
  { id: "puntos_pucp", etiqueta: IMPORTACION.puntos_pucp },
  { id: "sectores_capataz", etiqueta: IMPORTACION.sectores_capataz },
  { id: "vias", etiqueta: IMPORTACION.vias },
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
  const res = await fetch(apiUrl(`/importaciones?entidad=${encodeURIComponent(entidad)}`), {
    method: "POST",
    body: datos,
    credentials: "include",
  })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json() as Promise<VistaPrevia>
}

export async function confirmarImportacion(id: number): Promise<{ lote_id: number; validas: number; escrito: boolean }> {
  const res = await fetch(apiUrl(`/importaciones/${id}/confirmar`), { method: "POST", credentials: "include" })
  if (!res.ok) throw new Error(await mensaje(res))
  return res.json() as Promise<{ lote_id: number; validas: number; escrito: boolean }>
}

export async function revertirLote(id: number, confirmar: boolean): Promise<{ lote_id: number; revertidas: string[]; excluidas: { entidad_id: string; motivo: string }[] }> {
  const res = await fetch(apiUrl(`/lotes/${id}/revertir`), {
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
