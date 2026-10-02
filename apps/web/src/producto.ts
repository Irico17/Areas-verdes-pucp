import { ApiError } from "./operacion"
import { ROLES } from "./types"
import { apiUrl } from "./api"

export type Usuario = {
  id: number
  usuario: string
  nombre: string
  rol: "capataz" | "coordinacion" | "jefatura" | "admin"
  capataz_id?: string
  rol_nombre?: string
}

export function etiquetaRol(rol: string, rolNombre?: string): string {
  if (rolNombre && rolNombre.trim() !== "") return rolNombre
  const encontrado = ROLES.find((r) => r.id === rol)
  if (encontrado) return encontrado.label
  return rol
}

export type CatalogoItem = {
  id: number
  clase: string
  codigo: string
  nombre: string
  activo: boolean
  orden: number
  provisional?: boolean
}

export type Ficha = {
  feature_id: string
  nombre: string
  uso: string
  riego_act: string
  referencia: string
  area_m2?: number
  con_geometria: boolean
}

export type Solicitud = {
  id: string
  codigo_externo?: string
  fuente: string
  titulo: string
  detalle?: string
  prioridad: string
  estado: string
  lugar?: string
  actividad_id?: string
  created_at: string
}

export type Orden = {
  id: string
  actividad_id: string
  empresa: string
  referencia: string
  frecuencia?: string
  estado: string
  created_at: string
}

export type Riego = {
  id: string
  sector: string
  turno: string
  capataz_id?: string
  equipo?: string
  fecha: string
  nota?: string
}

export type Evidencia = {
  id: string
  actividad_id?: string
  nombre: string
  mime: string
  bytes: number
  nota?: string
  created_at: string
}

export type FiltroReporte = {
  estado?: string
  desde?: string
  hasta?: string
  zona?: string
  cuadrilla?: string
  origen?: string
}

export type Hueco = {
  clave: string
  nombre: string
  estado: string
  nota: string
}

/** Huecos de HUID 16, 24 y 26. La frase es el estado: no hay fórmula ni meta. */
export const HUECOS: Hueco[] = [
  {
    clave: "cobertura",
    nombre: "Cobertura",
    estado: "definición pendiente",
    nota: "El registro de riego no produce un porcentaje.",
  },
  {
    clave: "rendimiento",
    nombre: "Rendimiento",
    estado: "definición pendiente",
    nota: "No hay horas, personal ni superficie para un reporte de proceso.",
  },
  {
    clave: "metricas_proveedor",
    nombre: "Métricas de proveedor",
    estado: "definición pendiente",
    nota: "Jefatura no ha acordado la fórmula. No hay portal del proveedor.",
  },
]

export type Reporte = {
  aviso: string
  por_estado: { estado: string; n: number }[]
  pendientes?: Hueco[]
  filas: {
    id: string
    titulo: string
    tipo: string
    estado: string
    ejecutor: string
    equipo?: string
    zona?: string
    codigo_externo?: string
    fuente?: string
    created_at: string
    clase?: string
    lugar?: string
    cuadrilla?: string
    fecha_solicitud?: string
    fecha_atencion?: string
  }[]
}

export function reporteQuery(filtro: FiltroReporte, formato?: "csv" | "xls"): string {
  const q = new URLSearchParams()
  if (formato) q.set("formato", formato)
  for (const [clave, valor] of [
    ["estado", filtro.estado],
    ["desde", filtro.desde],
    ["hasta", filtro.hasta],
    ["zona", filtro.zona],
    ["cuadrilla", filtro.cuadrilla],
    ["origen", filtro.origen],
  ] as const) {
    if (valor) q.set(clave, valor)
  }
  return q.toString()
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
  if (!res.ok) {
    let message = `La API respondió ${res.status}`
    try {
      const payload = (await res.json()) as { error?: string }
      if (payload.error) message = payload.error
    } catch {
      /* vacío */
    }
    throw new ApiError(res.status, message)
  }
  return res
}

export async function fetchSesion(): Promise<Usuario | null> {
  const res = await fetch(apiUrl("/sesion"), { credentials: "include" })
  if (res.status === 401) return null
  if (!res.ok) throw new ApiError(res.status, "No se pudo leer la sesión")
  const body = (await res.json()) as { usuario: Usuario }
  return body.usuario
}

export async function entrar(usuario: string, clave: string): Promise<Usuario> {
  const res = await send(apiUrl("/sesion"), "POST", { usuario, clave })
  const body = (await res.json()) as { usuario: Usuario }
  return body.usuario
}

export async function salir(): Promise<void> {
  await send(apiUrl("/sesion"), "DELETE")
}

export async function fetchCatalogoCompleto(clase: string, activos = false): Promise<{ items: CatalogoItem[]; clases: string[] }> {
  const q = new URLSearchParams()
  if (clase) q.set("clase", clase)
  if (activos) q.set("activos", "1")
  const res = await send(apiUrl(`/catalogos?${q.toString()}`), "GET")
  const body = (await res.json()) as { items?: CatalogoItem[]; clases?: string[] }
  return { items: body.items ?? [], clases: body.clases ?? [] }
}

export async function fetchCatalogo(clase: string, activos = false): Promise<CatalogoItem[]> {
  const body = await fetchCatalogoCompleto(clase, activos)
  return body.items
}

export async function renombrarCatalogo(id: number, nombre: string): Promise<void> {
  await send(apiUrl(`/catalogos/${id}`), "PATCH", { nombre })
}

export async function crearCatalogo(clase: string, codigo: string, nombre: string): Promise<void> {
  await send(apiUrl("/catalogos"), "POST", { clase, codigo, nombre })
}

export async function desactivarCatalogo(id: number): Promise<void> {
  await send(apiUrl(`/catalogos/${id}/desactivar`), "POST")
}

export async function fetchFichas(q: string): Promise<Ficha[]> {
  const res = await send(apiUrl(`/catastro/areas?q=${encodeURIComponent(q)}`), "GET")
  const body = (await res.json()) as { areas?: Ficha[] }
  return body.areas ?? []
}

export async function guardarFicha(ficha: Ficha): Promise<void> {
  await send(apiUrl(`/catastro/areas/${encodeURIComponent(ficha.feature_id)}`), "PATCH", {
    nombre: ficha.nombre,
    uso: ficha.uso,
    riego_act: ficha.riego_act,
    referencia: ficha.referencia,
  })
}

export async function crearAreaSinGeom(nombre: string, uso: string): Promise<void> {
  await send(apiUrl("/catastro/areas"), "POST", { nombre, uso })
}

export async function fetchSolicitudes(): Promise<Solicitud[]> {
  const res = await send(apiUrl("/solicitudes"), "GET")
  const body = (await res.json()) as { solicitudes?: Solicitud[] }
  return body.solicitudes ?? []
}

export async function crearSolicitud(body: {
  id: string
  titulo: string
  fuente: string
  codigo_externo: string
  prioridad: string
  lugar: string
  detalle: string
}): Promise<void> {
  await send(apiUrl("/solicitudes"), "POST", body)
}

export async function fetchOrdenes(): Promise<Orden[]> {
  const res = await send(apiUrl("/ordenes"), "GET")
  const body = (await res.json()) as { ordenes?: Orden[] }
  return body.ordenes ?? []
}

export async function crearOrden(body: {
  id: string
  actividad_id: string
  empresa: string
  referencia: string
  frecuencia: string
}): Promise<void> {
  await send(apiUrl("/ordenes"), "POST", body)
}

export async function fetchRiego(): Promise<{ aviso: string; registros: Riego[] }> {
  const res = await send(apiUrl("/riego"), "GET")
  const body = (await res.json()) as { aviso?: string; registros?: Riego[] }
  return { aviso: body.aviso ?? "", registros: body.registros ?? [] }
}

export async function crearRiego(body: {
  id: string
  sector: string
  turno: string
  capataz_id: string
  fecha: string
  nota: string
  zona_supervision_id: string
  ciclo?: string
  superficie_m2?: number
}): Promise<void> {
  await send(apiUrl("/riego"), "POST", body)
}

export async function fetchEvidencias(actividadId: string): Promise<Evidencia[]> {
  const res = await send(apiUrl(`/evidencias?actividad_id=${encodeURIComponent(actividadId)}`), "GET")
  const body = (await res.json()) as { evidencias?: Evidencia[] }
  return body.evidencias ?? []
}

export async function fetchReporte(filtro: FiltroReporte): Promise<Reporte> {
  const res = await send(apiUrl(`/reportes/labores?${reporteQuery(filtro)}`), "GET")
  return res.json() as Promise<Reporte>
}

export function reporteHref(formato: "csv" | "xls", filtro: FiltroReporte): string {
  return apiUrl(`/reportes/labores?${reporteQuery(filtro, formato)}`)
}

export type SugerenciaTipo = {
  codigo: string
  etiqueta: string
  explicacion: string
  requiere_humano: boolean
}

export async function sugerirTipo(titulo: string): Promise<SugerenciaTipo> {
  const res = await send(apiUrl("/ia/sugerir-tipo"), "POST", { titulo })
  return res.json() as Promise<SugerenciaTipo>
}

export async function fetchCuentas(): Promise<{
  aviso: string
  usuarios: Usuario[]
  permisos: { rol: string; accion: string }[]
}> {
  const res = await send(apiUrl("/accesos/usuarios"), "GET")
  const body = (await res.json()) as {
    aviso?: string
    usuarios?: Usuario[]
    permisos?: { rol: string; accion: string }[]
  }
  return { aviso: body.aviso ?? "", usuarios: body.usuarios ?? [], permisos: body.permisos ?? [] }
}
