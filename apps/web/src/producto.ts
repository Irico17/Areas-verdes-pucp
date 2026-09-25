import { ApiError } from "./operacion"

export type Usuario = {
  id: number
  usuario: string
  nombre: string
  rol: "capataz" | "coordinacion" | "jefatura" | "admin"
  capataz_id?: string
}

export type CatalogoItem = {
  id: number
  clase: string
  codigo: string
  nombre: string
  activo: boolean
  orden: number
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

export type Reporte = {
  aviso: string
  por_estado: { estado: string; n: number }[]
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
  }[]
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
  const res = await fetch("/api/v1/sesion", { credentials: "include" })
  if (res.status === 401) return null
  if (!res.ok) throw new ApiError(res.status, "No se pudo leer la sesión")
  const body = (await res.json()) as { usuario: Usuario }
  return body.usuario
}

export async function entrar(usuario: string, clave: string): Promise<Usuario> {
  const res = await send("/api/v1/sesion", "POST", { usuario, clave })
  const body = (await res.json()) as { usuario: Usuario }
  return body.usuario
}

export async function salir(): Promise<void> {
  await send("/api/v1/sesion", "DELETE")
}

export async function fetchCatalogo(clase: string, activos = false): Promise<CatalogoItem[]> {
  const q = new URLSearchParams()
  if (clase) q.set("clase", clase)
  if (activos) q.set("activos", "1")
  const res = await send(`/api/v1/catalogos?${q.toString()}`, "GET")
  const body = (await res.json()) as { items?: CatalogoItem[] }
  return body.items ?? []
}

export async function crearCatalogo(clase: string, codigo: string, nombre: string): Promise<void> {
  await send("/api/v1/catalogos", "POST", { clase, codigo, nombre })
}

export async function desactivarCatalogo(id: number): Promise<void> {
  await send(`/api/v1/catalogos/${id}/desactivar`, "POST")
}

export async function fetchFichas(q: string): Promise<Ficha[]> {
  const res = await send(`/api/v1/catastro/areas?q=${encodeURIComponent(q)}`, "GET")
  const body = (await res.json()) as { areas?: Ficha[] }
  return body.areas ?? []
}

export async function guardarFicha(ficha: Ficha): Promise<void> {
  await send(`/api/v1/catastro/areas/${encodeURIComponent(ficha.feature_id)}`, "PATCH", {
    nombre: ficha.nombre,
    uso: ficha.uso,
    riego_act: ficha.riego_act,
    referencia: ficha.referencia,
  })
}

export async function crearAreaSinGeom(nombre: string, uso: string): Promise<void> {
  await send("/api/v1/catastro/areas", "POST", { nombre, uso })
}

export async function fetchSolicitudes(): Promise<Solicitud[]> {
  const res = await send("/api/v1/solicitudes", "GET")
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
  await send("/api/v1/solicitudes", "POST", body)
}

export async function fetchOrdenes(): Promise<Orden[]> {
  const res = await send("/api/v1/ordenes", "GET")
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
  await send("/api/v1/ordenes", "POST", body)
}

export async function fetchRiego(): Promise<{ aviso: string; registros: Riego[] }> {
  const res = await send("/api/v1/riego", "GET")
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
}): Promise<void> {
  await send("/api/v1/riego", "POST", body)
}

export async function fetchEvidencias(actividadId: string): Promise<Evidencia[]> {
  const res = await send(`/api/v1/evidencias?actividad_id=${encodeURIComponent(actividadId)}`, "GET")
  const body = (await res.json()) as { evidencias?: Evidencia[] }
  return body.evidencias ?? []
}

export async function subirEvidencia(actividadId: string, archivo: File, nota: string): Promise<void> {
  const data = new FormData()
  data.set("id", crypto.randomUUID())
  data.set("actividad_id", actividadId)
  data.set("nota", nota)
  data.set("archivo", archivo)
  let res: Response
  try {
    res = await fetch("/api/v1/evidencias", { method: "POST", body: data, credentials: "include" })
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
}

export async function fetchReporte(estado: string, desde: string, hasta: string): Promise<Reporte> {
  const q = new URLSearchParams()
  if (estado) q.set("estado", estado)
  if (desde) q.set("desde", desde)
  if (hasta) q.set("hasta", hasta)
  const res = await send(`/api/v1/reportes/labores?${q.toString()}`, "GET")
  return res.json() as Promise<Reporte>
}

export function reporteHref(formato: "csv" | "xls", estado: string, desde: string, hasta: string): string {
  const q = new URLSearchParams({ formato })
  if (estado) q.set("estado", estado)
  if (desde) q.set("desde", desde)
  if (hasta) q.set("hasta", hasta)
  return `/api/v1/reportes/labores?${q.toString()}`
}

export async function sugerirTipo(titulo: string): Promise<{ codigo: string; etiqueta: string; explicacion: string }> {
  const res = await send("/api/v1/ia/sugerir-tipo", "POST", { titulo })
  return res.json() as Promise<{ codigo: string; etiqueta: string; explicacion: string }>
}

export async function fetchCuentas(): Promise<{
  aviso: string
  usuarios: Usuario[]
  permisos: { rol: string; accion: string }[]
}> {
  const res = await send("/api/v1/accesos/usuarios", "GET")
  const body = (await res.json()) as {
    aviso?: string
    usuarios?: Usuario[]
    permisos?: { rol: string; accion: string }[]
  }
  return { aviso: body.aviso ?? "", usuarios: body.usuarios ?? [], permisos: body.permisos ?? [] }
}
