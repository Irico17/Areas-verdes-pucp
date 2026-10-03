import { apiUrl } from "../api"
import { ApiError, fetchActividades, fetchTimeline, type Evento } from "../operacion"
import type { Rol } from "../types"
import { ACTIVIDAD, HITO } from "../ui/nomenclatura"

export const TIPOS_HITO = [
  { id: "inicio", label: HITO.inicio },
  { id: "supervision", label: HITO.supervision },
  { id: "derivacion", label: HITO.derivacion },
  { id: "observacion", label: HITO.observacion },
  { id: "conformidad", label: HITO.conformidad },
] as const

export type ActividadBitacora = { id: string; titulo: string }

async function leerError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    /* cuerpo vacío */
  }
  return `La API respondió ${res.status}`
}

export async function anotarHito(actividadId: string, tipo: string, texto: string): Promise<void> {
  const res = await fetch(apiUrl(`/operacion/actividades/${actividadId}/hitos`), {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tipo, texto }),
  })
  if (!res.ok) throw new ApiError(res.status, await leerError(res))
}

export async function adjuntarEvidencia(actividadId: string, eventoId: number, archivo: File): Promise<void> {
  const bytes = await archivo.arrayBuffer()
  const dig = await crypto.subtle.digest("SHA-256", bytes)
  const sha = [...new Uint8Array(dig)].map((b) => b.toString(16).padStart(2, "0")).join("")
  const data = new FormData()
  data.set("id", crypto.randomUUID())
  data.set("actividad_id", actividadId)
  data.set("evento_id", String(eventoId))
  data.set("sha256", sha)
  data.set("archivo", new Blob([bytes], { type: archivo.type || "application/octet-stream" }), archivo.name)
  const res = await fetch(apiUrl("/evidencias"), { method: "POST", credentials: "include", body: data })
  if (!res.ok) throw new ApiError(res.status, await leerError(res))
}

export async function listarActividadesBitacora(rol: Rol, capatazId: string): Promise<ActividadBitacora[]> {
  const coleccion = await fetchActividades(rol, capatazId)
  return coleccion.features.flatMap((feature) => {
    const props = feature.properties ?? {}
    const id = props.id ?? feature.id
    if (id == null || id === "") return []
    const titulo = props.titulo == null || props.titulo === "" ? ACTIVIDAD.sinTitulo : String(props.titulo)
    return [{ id: String(id), titulo }]
  })
}

export async function leerCadena(actividadId: string): Promise<Evento[]> {
  return fetchTimeline(actividadId)
}
