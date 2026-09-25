export const CAPAS_EDITABLES = [
  { id: "fauna", label: "Fauna", campos: ["nombre"] },
  { id: "puertas", label: "Puertas", campos: ["codigo"] },
  { id: "playas_estacionamiento", label: "Playas", campos: ["codigo"] },
  { id: "veredas_riesgo", label: "Vereda en riesgo", campos: ["nota"] },
  { id: "xerofiticas", label: "Xerofítica", campos: ["clase", "riego", "area_m2", "perimetro_m"] },
  { id: "jardines_reserva", label: "Jardines de reserva", campos: ["codigo", "nombre", "nota", "uso", "riego", "pertenecen", "area_m2", "perimetro_m"] },
] as const

export type CapaId = (typeof CAPAS_EDITABLES)[number]["id"]

export const CONTEOS_TACHO = [
  "no_aprovechables",
  "papel_carton",
  "plastico",
  "vidrio",
  "pilas",
  "peligrosos",
  "raee",
  "metales",
  "aniquem",
  "intermedios_plastico",
  "intermedios_metal",
] as const

export type Conteos = Record<(typeof CONTEOS_TACHO)[number], number>

export const SUMAS_PUBLICADAS: Conteos = {
  no_aprovechables: 294,
  papel_carton: 142,
  plastico: 260,
  vidrio: 242,
  pilas: 56,
  peligrosos: 0,
  raee: 12,
  metales: 2,
  aniquem: 38,
  intermedios_plastico: 26,
  intermedios_metal: 4,
}

export type ErrorCampo = { campo: string; motivo: string }

export function validarConteos(conteos: Conteos): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  for (const campo of CONTEOS_TACHO) {
    const n = conteos[campo]
    if (!Number.isInteger(n) || n < 0) errores.push({ campo, motivo: "El conteo es un entero mayor o igual que cero." })
  }
  return errores
}

export function sumarConteos(filas: Conteos[]): Conteos {
  const total = { ...SUMAS_PUBLICADAS }
  for (const campo of CONTEOS_TACHO) total[campo] = 0
  for (const fila of filas) {
    for (const campo of CONTEOS_TACHO) total[campo] += fila[campo]
  }
  return total
}

export function validarBebedero(row: { codigo: string; subtipo: string; estado: string; sede: string }): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  if (!row.codigo.startsWith("PT_")) errores.push({ campo: "codigo", motivo: "El código empieza por PT_." })
  if (!["fuente", "llenador", "nuevo", "deterioro", "baja"].includes(row.subtipo)) {
    errores.push({ campo: "subtipo", motivo: "Subtipo fuera del catálogo." })
  }
  if (!row.estado.trim()) errores.push({ campo: "estado", motivo: "Falta el estado del archivo." })
  if (!row.sede.trim()) errores.push({ campo: "sede", motivo: "Falta la sede." })
  return errores
}

export function validarPunto(row: { titulo: string; lat: number; lon: number; url?: string }): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  if (!row.titulo.trim()) errores.push({ campo: "titulo", motivo: "Falta el título." })
  if (row.lat < -12.2 || row.lat > -11.9 || row.lon < -77.3 || row.lon > -76.9) {
    errores.push({ campo: "lat", motivo: "El pin queda fuera del campus." })
  }
  const url = row.url ?? ""
  if (/phone|placeId|place_id|website/i.test(url)) {
    errores.push({ campo: "url", motivo: "No se guarda teléfono, placeId ni website." })
  }
  return errores
}

export function validarReserva(row: { origen: string; fecha: string; hora_inicio: string; hora_fin: string; estado: string; evento: string }): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  if (row.origen !== "ficticio") errores.push({ campo: "origen", motivo: "Mientras la hoja responda 401 solo se acepta origen ficticio." })
  if (!/^\d{4}-\d{2}-\d{2}$/.test(row.fecha)) errores.push({ campo: "fecha", motivo: "La fecha va en ISO." })
  if (!(row.hora_fin > row.hora_inicio)) errores.push({ campo: "hora_fin", motivo: "La hora fin es posterior al inicio." })
  if (!["reservado", "realizado", "cancelado"].includes(row.estado)) errores.push({ campo: "estado", motivo: "Estado fuera del catálogo." })
  if (!row.evento.trim()) errores.push({ campo: "evento", motivo: "Falta el evento." })
  return errores
}

export function csvFichas(filas: { feature_id: string; nombre: string; codigo: string }[]): string {
  const lineas = ["feature_id,nombre,codigo"]
  for (const fila of filas) lineas.push([fila.feature_id, fila.nombre, fila.codigo].join(","))
  return lineas.join("\n") + "\n"
}

export type EntidadInventario = "tachos" | "bebederos" | "puntos" | "reservas" | CapaId

const CAPA_IDS = new Set<string>(CAPAS_EDITABLES.map((capa) => capa.id))

export function esCapa(entidad: string): entidad is CapaId {
  return CAPA_IDS.has(entidad)
}

export type Solicitud = { method: "GET" | "POST" | "PATCH" | "DELETE"; path: string; body?: unknown }

export function rutaListar(entidad: EntidadInventario): string {
  if (esCapa(entidad)) return `/api/v1/inventario/capas/${entidad}`
  return `/api/v1/inventario/${entidad}`
}

export function solicitudGuardar(entidad: EntidadInventario, id: number | null, body: unknown): Solicitud {
  if (id == null) return { method: "POST", path: rutaListar(entidad), body }
  if (esCapa(entidad)) return { method: "PATCH", path: `/api/v1/inventario/capas/${entidad}/${id}`, body }
  return { method: "PATCH", path: `/api/v1/inventario/${entidad}/${id}`, body }
}

export function solicitudBaja(entidad: EntidadInventario, id: number): Solicitud {
  if (esCapa(entidad)) return { method: "DELETE", path: `/api/v1/inventario/capas/${entidad}/${id}` }
  return { method: "DELETE", path: `/api/v1/inventario/${entidad}/${id}` }
}

export function cuerpoPunto(row: { titulo: string; lat: number; lon: number; url: string }): { titulo: string; lat: number; lon: number; url: string } {
  return { titulo: row.titulo.trim(), lat: row.lat, lon: row.lon, url: row.url.trim() }
}

export async function enviarInventario(cliente: typeof fetch, solicitud: Solicitud): Promise<unknown> {
  const res = await cliente(solicitud.path, {
    method: solicitud.method,
    credentials: "include",
    headers: {
      Accept: "application/json",
      ...(solicitud.body === undefined ? {} : { "Content-Type": "application/json" }),
    },
    body: solicitud.body === undefined ? undefined : JSON.stringify(solicitud.body),
  })
  if (!res.ok) {
    let detalle = ""
    try {
      detalle = await res.text()
    } catch {
      detalle = ""
    }
    throw new Error(detalle || `La API respondió ${res.status}`)
  }
  const tipo = res.headers.get("content-type") ?? ""
  if (!tipo.includes("json")) return null
  return res.json()
}
