import { apiUrl, fetchCollection } from "./api"
import type { FeatureCollection, GeoFeature, Rol } from "./types"
import { etiquetaHito, TIPO_RESPALDO } from "./ui/nomenclatura"

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
  { id: "incidencia", label: TIPO_RESPALDO.incidencia, marca: "I" },
  { id: "inspeccion", label: "Inspección", marca: "V" },
] as const

export type EstadoVista = { id: string; label: string; color: string }

const COLOR_ESTADO: Record<string, string> = {
  sin_estado: "#8d8478",
  pendiente: "#8a6410",
  en_proceso: "#1a5c44",
  ejecutado: "#2f6f4e",
  bloqueada: "#8c3832",
  cerrada: "#5c6a62",
  cancelada: "#5c6a62",
  archivada: "#5c6a62",
}

const ETIQUETA_RESPALDO: Record<string, string> = {
  sin_estado: "Sin estado",
  pendiente: "Por iniciar",
  en_proceso: "En proceso",
  ejecutado: "Ejecutado",
  bloqueada: "Bloqueada (provisional)",
  cerrada: "Cerrado",
  cancelada: "Cancelado",
  archivada: "Archivado",
}

const CIERRE_OFICINA = new Set(["cerrada", "cancelada", "archivada"])

export const ESTADOS: EstadoVista[] = [
  { id: "sin_estado", label: "Sin estado", color: COLOR_ESTADO.sin_estado },
  { id: "pendiente", label: "Por iniciar", color: COLOR_ESTADO.pendiente },
  { id: "en_proceso", label: "En proceso", color: COLOR_ESTADO.en_proceso },
  { id: "ejecutado", label: "Ejecutado", color: COLOR_ESTADO.ejecutado },
  { id: "cerrada", label: "Cerrado", color: COLOR_ESTADO.cerrada },
  { id: "cancelada", label: "Cancelado", color: COLOR_ESTADO.cancelada },
  { id: "archivada", label: "Archivado", color: COLOR_ESTADO.archivada },
]

export const ABIERTOS = ["pendiente", "en_proceso", "ejecutado", "bloqueada"] as const

export type EstadoCatalogo = {
  codigo: string
  nombre: string
  activo: boolean
  orden?: number
  provisional?: boolean
}

export function aplicarEstadosCatalogo(items: EstadoCatalogo[]) {
  const activos = items.filter((item) => item.activo && item.codigo && item.nombre.trim() !== "")
  if (activos.length === 0) return
  activos.sort((a, b) => (a.orden ?? 0) - (b.orden ?? 0) || a.codigo.localeCompare(b.codigo))
  const next = activos.map((item) => ({
    id: item.codigo,
    label: etiquetaVisible(item.nombre, item.provisional),
    color: COLOR_ESTADO[item.codigo] ?? "#5c6a62",
  }))
  ESTADOS.splice(0, ESTADOS.length, ...next)
}

function etiquetaVisible(nombre: string, provisional?: boolean) {
  const limpio = nombre.trim()
  if (provisional && !/provisional/i.test(limpio)) return `${limpio} (provisional)`
  return limpio
}

export async function sincronizarEstados(): Promise<void> {
  try {
    const res = await send(apiUrl("/catalogos?clase=estado&activos=1"), "GET")
    const body = (await res.json()) as { items?: EstadoCatalogo[] }
    const items = (body.items ?? []).map((item) => ({
      codigo: item.codigo,
      nombre: item.nombre,
      activo: item.activo,
      orden: item.orden,
      provisional: item.provisional,
    }))
    aplicarEstadosCatalogo(items)
  } catch {
    /* se conserva la etiqueta de respaldo; no se muestra el slug */
  }
}

export function estadosPermitidos(rol?: string) {
  if (rol === "capataz") {
    return ESTADOS.filter((estado) => !CIERRE_OFICINA.has(estado.id))
  }
  return ESTADOS
}

export function puedeEncolarEstado(rol: string | undefined, estado: string): boolean {
  if (rol === "capataz" && CIERRE_OFICINA.has(estado)) {
    return false
  }
  return true
}

export type Capataz = { id: string; equipo: string; turno: string }

export type EvidenciaEvento = {
  id: string
  nombre: string
  mime: string
}

export type Evento = {
  id: number
  tipo: string
  estado?: string
  capataz_id?: string
  equipo?: string
  capataz_anterior?: string
  cuadrilla_anterior?: string
  actor_rol: string
  usuario_nombre?: string
  nota?: string
  created_at: string
  evidencias?: EvidenciaEvento[]
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
  clase?: string
  subtipo?: string
  origen?: string
  codigo_externo?: string
  unidad_solicitante?: string
  nivel_riesgo?: string
  fecha_programada?: string
  cantidad?: number
  personal?: string[]
  lugar_id?: string
  zona_supervision_id?: string
}

const TIPO_DE_CLASE: Record<string, string> = {
  poda: "poda",
  riego: "riego",
  inspeccion_monitoreo: "inspeccion",
  fitosanitario: "inspeccion",
}

/** El mapa y el filtro siguen usando el tipo grueso. La clase fina va en subtipo. */
export function tipoGrueso(clase: string, actual: string): string {
  if (TIPO_DE_CLASE[clase]) return TIPO_DE_CLASE[clase]
  if (clase) return "limpieza"
  return actual || "riego"
}

export type OpcionAlta = { codigo: string; nombre: string }

export type TaxonomiaActividad = {
  clases: { codigo: string; nombre: string; tipos: OpcionAlta[] }[]
  riesgos: OpcionAlta[]
  origenes: OpcionAlta[]
  personal: { id: string; nombre_ficticio: string }[]
}

export type AltaCampos = {
  tipo: string
  clase: string
  subtipo: string
  origen: string
  codigo_externo: string
  unidad_solicitante: string
  nivel_riesgo: string
  fecha_programada: string
  cantidad: string
  personal: string[]
  lugar_id: string
  zona_supervision_id: string
}

export async function fetchTaxonomiaActividad(): Promise<TaxonomiaActividad> {
  const res = await send(apiUrl("/operacion/taxonomia-actividad"), "GET")
  const body = (await res.json()) as Partial<TaxonomiaActividad>
  return {
    clases: body.clases ?? [],
    riesgos: body.riesgos ?? [],
    origenes: body.origenes ?? [],
    personal: body.personal ?? [],
  }
}

export function etiquetaEvento(tipo: string): string {
  return etiquetaHito(tipo)
}

export function etiquetaEstado(id: string): string {
  return ESTADOS.find((item) => item.id === id)?.label ?? ETIQUETA_RESPALDO[id] ?? id
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
  const init: RequestInit = { method, credentials: "include" }
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

export type FiltroActividades = {
  estado: string
  tipo: string
  cuadrillaId: string
  sector: string
  ejecutor: string
  origen: string
  nivelRiesgo: string
  desde: string
  hasta: string
  ejemplarId: string
  responsable: string
  historico: boolean
}

export const FILTRO_ACTIVIDADES: FiltroActividades = {
  estado: "",
  tipo: "",
  cuadrillaId: "",
  sector: "",
  ejecutor: "",
  origen: "",
  nivelRiesgo: "",
  desde: "",
  hasta: "",
  ejemplarId: "",
  responsable: "",
  historico: false,
}

export function actividadesPath(rol: Rol, capatazId: string, filtro: FiltroActividades = FILTRO_ACTIVIDADES): string {
  const q = new URLSearchParams({ rol, abiertas: filtro.historico ? "0" : "1" })
  if (rol === "capataz") q.set("capataz_id", capatazId)
  if (filtro.estado) q.set("estado", filtro.estado)
  if (filtro.tipo) q.set("tipo", filtro.tipo)
  if (filtro.cuadrillaId && rol !== "capataz") q.set("cuadrilla_id", filtro.cuadrillaId)
  if (filtro.sector) q.set("sector", filtro.sector)
  if (filtro.ejecutor) q.set("ejecutor", filtro.ejecutor)
  if (filtro.origen) q.set("origen", filtro.origen)
  if (filtro.nivelRiesgo) q.set("nivel_riesgo", filtro.nivelRiesgo)
  if (filtro.desde) q.set("desde", filtro.desde)
  if (filtro.hasta) q.set("hasta", filtro.hasta)
  if (filtro.ejemplarId) q.set("ejemplar_id", filtro.ejemplarId)
  if (filtro.responsable) q.set("responsable", filtro.responsable)
  return apiUrl(`/operacion/actividades?${q.toString()}`)
}

export async function fetchActividades(rol: Rol, capatazId: string, filtro: FiltroActividades = FILTRO_ACTIVIDADES): Promise<FeatureCollection> {
  await sincronizarEstados()
  return fetchCollection(actividadesPath(rol, capatazId, filtro))
}

export async function fetchCapataces(): Promise<Capataz[]> {
  const res = await send(apiUrl("/operacion/capataces"), "GET")
  return res.json().then((body: { capataces?: Capataz[] }) => body.capataces ?? [])
}

export async function crearActividad(body: CreateBody): Promise<void> {
  const { actor_rol: _rol, ...resto } = body
  await send(apiUrl("/operacion/actividades"), "POST", resto)
}

export async function asignar(id: string, capatazId: string, _rol: Rol): Promise<void> {
  await send(apiUrl(`/operacion/actividades/${id}/asignacion`), "PATCH", {
    capataz_id: capatazId,
  })
}

export async function cambiarEstado(id: string, estado: string, rol: Rol, capatazId: string): Promise<void> {
  await send(apiUrl(`/operacion/actividades/${id}/estado`), "PATCH", {
    estado,
    capataz_id: rol === "capataz" ? capatazId : "",
  })
}

export async function archivar(id: string, _rol: Rol, motivo: string): Promise<void> {
  await send(apiUrl(`/operacion/actividades/${id}/archivar`), "POST", { motivo })
}

export async function fetchTimeline(id: string): Promise<Evento[]> {
  const res = await send(apiUrl(`/operacion/actividades/${id}/timeline`), "GET")
  const body = (await res.json()) as { eventos?: Evento[] }
  return body.eventos ?? []
}
