import { fetchCollection } from "./api"
import type { FeatureCollection, GeoFeature, Rol } from "./types"

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

export const TIPOS = [
  { id: "riego", label: "Riego", marca: "R" },
  { id: "poda", label: "Poda", marca: "P" },
  { id: "limpieza", label: "Limpieza", marca: "L" },
  { id: "incidencia", label: "Incidencia", marca: "I" },
  { id: "inspeccion", label: "Inspección", marca: "V" },
] as const

export const ESTADOS = [
  { id: "pendiente", label: "Pendiente", color: "#8a6410" },
  { id: "en_proceso", label: "En proceso", color: "#1a5c44" },
  { id: "bloqueada", label: "Bloqueada", color: "#8c3832" },
  { id: "cerrada", label: "Cerrada", color: "#5c6a62" },
  { id: "cancelada", label: "Cancelada", color: "#5c6a62" },
] as const

export const ABIERTOS = ["pendiente", "en_proceso", "bloqueada"] as const

export type Capataz = { id: string; equipo: string; turno: string }

export type Evento = {
  id: number
  tipo: string
  estado?: string
  capataz_id?: string
  equipo?: string
  actor_rol: string
  nota?: string
  created_at: string
}

export type CreateBody = {
  id: string
  tipo: string
  titulo: string
  detalle: string
  lon: number
  lat: number
  assigned_capataz_id: string
  actor_rol: Rol
  ejecutor?: string
}

const EVENTO_LABEL: Record<string, string> = {
  creada: "Creada",
  asignada: "Asignada",
  reasignada: "Reasignada",
  estado: "Estado",
  cancelada: "Cancelada",
  archivada: "Archivada",
  evidencia: "Evidencia",
}

export function etiquetaEvento(tipo: string): string {
  return EVENTO_LABEL[tipo] ?? tipo
}

export function etiquetaEstado(id: string): string {
  return ESTADOS.find((item) => item.id === id)?.label ?? id
}

export function etiquetaTipo(id: string): string {
  return TIPOS.find((item) => item.id === id)?.label ?? id
}

export function lonLat(feature: GeoFeature): { lon: number; lat: number } | null {
  const geometry = feature.geometry
  if (!geometry || geometry.type !== "Point" || !Array.isArray(geometry.coordinates)) return null
  const [lon, lat] = geometry.coordinates as number[]
  if (typeof lon !== "number" || typeof lat !== "number") return null
  return { lon, lat }
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    /* cuerpo vacío */
  }
  return `La API respondió ${res.status}`
}

async function send(path: string, method: string, body?: unknown): Promise<Response> {
  const init: RequestInit = { method }
  if (body !== undefined) {
    init.headers = { "Content-Type": "application/json" }
    init.body = JSON.stringify(body)
  }
  let res: Response
  try {
    res = await fetch(path, init)
  } catch {
    throw new ApiError(0, "Sin conexión con la API")
  }
  if (!res.ok) throw new ApiError(res.status, await readError(res))
  return res
}

export function actividadesPath(rol: Rol, capatazId: string): string {
  const q = new URLSearchParams({ rol, abiertas: "1" })
  if (rol === "capataz") q.set("capataz_id", capatazId)
  return `/api/v1/operacion/actividades?${q.toString()}`
}

export function fetchActividades(rol: Rol, capatazId: string): Promise<FeatureCollection> {
  return fetchCollection(actividadesPath(rol, capatazId))
}

export async function fetchCapataces(): Promise<Capataz[]> {
  const res = await send("/api/v1/operacion/capataces", "GET")
  return res.json().then((body: { capataces?: Capataz[] }) => body.capataces ?? [])
}

export async function crearActividad(body: CreateBody): Promise<void> {
  await send("/api/v1/operacion/actividades", "POST", body)
}

export async function asignar(id: string, capatazId: string, rol: Rol): Promise<void> {
  await send(`/api/v1/operacion/actividades/${id}/asignacion`, "PATCH", {
    capataz_id: capatazId,
    actor_rol: rol,
  })
}

export async function cambiarEstado(id: string, estado: string, rol: Rol, capatazId: string): Promise<void> {
  await send(`/api/v1/operacion/actividades/${id}/estado`, "PATCH", {
    estado,
    actor_rol: rol,
    capataz_id: rol === "capataz" ? capatazId : "",
  })
}

export async function archivar(id: string, rol: Rol, motivo: string): Promise<void> {
  await send(`/api/v1/operacion/actividades/${id}/archivar`, "POST", { actor_rol: rol, motivo })
}

export async function fetchTimeline(id: string): Promise<Evento[]> {
  const res = await send(`/api/v1/operacion/actividades/${id}/timeline`, "GET")
  const body = (await res.json()) as { eventos?: Evento[] }
  return body.eventos ?? []
}
