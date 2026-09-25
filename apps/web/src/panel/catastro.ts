import { erroresGeometria, medirGeom, type MultiPolygon, type Position } from "../map/draw.ts"

/** Valores presentes en areas_verdes.geojson. No se acepta texto libre nuevo. */
export const USOS_AREA = [
  "Áreas de uso administrativo",
  "Áreas de manejo sostenible y reducción de consumo de agua",
  "Áreas de uso recreativo/descanso",
  "Áreas deportivas y recreación activa",
  "Uso Institucional",
  "Áreas de conservación",
] as const

export const PROYECTOS_RIEGO = ["Por validar con unidad", "Goteo", "Falta aspersión", "Cuenta con aspersión"] as const

export const RIEGOS_ACTUALES = ["Sin riego tecnificado", "Riego por aspersión", "Riego por goteo"] as const

export const CODIGOS_ZONA = ["Z1", "Z2", "Z3", "Z4"] as const

/** Campos de MAPA-DATOS-Y-EDICION.md §4.1, en el orden del formulario. */
export const CAMPOS_AREA = [
  "feature_id",
  "codigo",
  "nombre",
  "uso",
  "proy_riego",
  "riego_act",
  "referencia",
  "perimetro_m",
  "area_m2",
  "geom",
  "zona_supervision_id",
] as const

/** Campos de MAPA-DATOS-Y-EDICION.md §4.2. */
export const CAMPOS_ZONA = ["codigo", "nombre", "area_m2", "geom"] as const

export type CampoArea = (typeof CAMPOS_AREA)[number]
export type CampoZona = (typeof CAMPOS_ZONA)[number]

export type AreaVerde = {
  feature_id: string
  codigo: string | null
  nombre: string | null
  uso: string | null
  proy_riego: string | null
  riego_act: string | null
  referencia: string | null
  perimetro_m: number | null
  area_m2: number | null
  geom: MultiPolygon | null
  zona_supervision_id: string | null
}

export type ZonaSupervision = {
  codigo: string
  nombre: string
  area_m2: number | null
  geom: MultiPolygon | null
}

export type ErrorCampo = { campo: string; motivo: string }

const REF_MAX = 500

function texto(value: unknown): string | null {
  if (value == null) return null
  const s = String(value).trim()
  return s === "" ? null : s
}

function numero(value: unknown): number | null {
  if (value == null || value === "") return null
  const n = typeof value === "number" ? value : Number(String(value).replace(",", "."))
  return Number.isFinite(n) ? n : null
}

function esMulti(value: unknown): value is MultiPolygon {
  if (!value || typeof value !== "object") return false
  const g = value as MultiPolygon
  return g.type === "MultiPolygon" && Array.isArray(g.coordinates)
}

export function avisoMedidas(declarado: number | null, medido: number): string | null {
  if (declarado == null || medido <= 0) return null
  const delta = Math.abs(declarado - medido) / medido
  if (delta <= 0.05) return null
  return `Difiere más del 5 % del cálculo sobre la geometría (${medido}).`
}

export function validarArea(area: AreaVerde, otras: AreaVerde[] = []): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  if (!/^AV-\S+$/.test(area.feature_id)) {
    errores.push({ campo: "feature_id", motivo: "El identificador empieza por AV- y no lleva espacios." })
  }
  if (otras.some((row) => row.feature_id === area.feature_id)) {
    errores.push({ campo: "feature_id", motivo: "Ese identificador ya está en el catastro." })
  }
  if (area.codigo && otras.some((row) => row.codigo && row.codigo === area.codigo)) {
    errores.push({ campo: "codigo", motivo: "Ese código ya está usado." })
  }
  if (area.uso && !USOS_AREA.includes(area.uso as (typeof USOS_AREA)[number])) {
    errores.push({ campo: "uso", motivo: "El uso no está en el catálogo de la fuente." })
  }
  if (area.proy_riego && !PROYECTOS_RIEGO.includes(area.proy_riego as (typeof PROYECTOS_RIEGO)[number])) {
    errores.push({ campo: "proy_riego", motivo: "El proyecto de riego no está dado de alta." })
  }
  if (area.riego_act && !RIEGOS_ACTUALES.includes(area.riego_act as (typeof RIEGOS_ACTUALES)[number])) {
    errores.push({ campo: "riego_act", motivo: "El riego actual no está dado de alta." })
  }
  if (area.referencia && area.referencia.length > REF_MAX) {
    errores.push({ campo: "referencia", motivo: "La referencia admite hasta 500 caracteres." })
  }
  if (area.perimetro_m != null && area.perimetro_m < 0) {
    errores.push({ campo: "perimetro_m", motivo: "El perímetro no puede ser negativo." })
  }
  if (area.area_m2 != null && area.area_m2 < 0) {
    errores.push({ campo: "area_m2", motivo: "El área no puede ser negativa." })
  }
  if (area.zona_supervision_id && !CODIGOS_ZONA.includes(area.zona_supervision_id as (typeof CODIGOS_ZONA)[number])) {
    errores.push({ campo: "zona_supervision_id", motivo: "La zona de supervisión es Z1, Z2, Z3 o Z4." })
  }
  for (const motivo of erroresGeometria(area.geom, false)) {
    errores.push({ campo: "geom", motivo })
  }
  return errores
}

export function validarZona(zona: ZonaSupervision, otras: ZonaSupervision[] = []): ErrorCampo[] {
  const errores: ErrorCampo[] = []
  if (!CODIGOS_ZONA.includes(zona.codigo as (typeof CODIGOS_ZONA)[number])) {
    errores.push({ campo: "codigo", motivo: "El código es Z1, Z2, Z3 o Z4." })
  }
  if (otras.some((row) => row.codigo === zona.codigo)) {
    errores.push({ campo: "codigo", motivo: "Ese código ya está usado." })
  }
  if (!zona.nombre.trim()) {
    errores.push({ campo: "nombre", motivo: "El nombre es obligatorio." })
  }
  if (zona.area_m2 != null && zona.area_m2 < 0) {
    errores.push({ campo: "area_m2", motivo: "El área no puede ser negativa." })
  }
  for (const motivo of erroresGeometria(zona.geom, true)) {
    errores.push({ campo: "geom", motivo })
  }
  return errores
}

/** Payload de escritura. Los nombres son los del contrato, no los de la fuente. */
export function payloadArea(area: AreaVerde): AreaVerde {
  return {
    feature_id: area.feature_id.trim(),
    codigo: texto(area.codigo),
    nombre: texto(area.nombre),
    uso: texto(area.uso),
    proy_riego: texto(area.proy_riego),
    riego_act: texto(area.riego_act),
    referencia: texto(area.referencia),
    perimetro_m: area.perimetro_m,
    area_m2: area.area_m2,
    geom: area.geom,
    zona_supervision_id: texto(area.zona_supervision_id),
  }
}

export function payloadZona(zona: ZonaSupervision): ZonaSupervision {
  return {
    codigo: zona.codigo.trim(),
    nombre: zona.nombre.trim(),
    area_m2: zona.area_m2,
    geom: zona.geom,
  }
}

type PropsFuente = Record<string, unknown>

function geomDe(feature: { geometry?: unknown }): MultiPolygon | null {
  const g = feature.geometry
  if (!esMulti(g)) return null
  return { type: "MultiPolygon", coordinates: g.coordinates as Position[][][] }
}

/** Lee un feature del GeoJSON de áreas (propiedades de la fuente o ya normalizadas). */
export function areaDesdeFeature(feature: { geometry?: unknown; properties?: PropsFuente | null }, indice = 1): AreaVerde {
  const p = feature.properties ?? {}
  const featureId = texto(p.feature_id) ?? `AV-${String(indice).padStart(4, "0")}`
  return {
    feature_id: featureId,
    codigo: texto(p.codigo ?? p["código"]),
    nombre: texto(p.nombre ?? p.Nombre),
    uso: texto(p.uso ?? p.Uso),
    proy_riego: texto(p.proy_riego ?? p["Proy riego"]),
    riego_act: texto(p.riego_act ?? p["Riego act"]),
    referencia: texto(p.referencia ?? p.Referenc_1),
    perimetro_m: numero(p.perimetro_m ?? p.Perimetro),
    area_m2: numero(p.area_m2 ?? p["Área"]),
    geom: geomDe(feature),
    zona_supervision_id: texto(p.zona_supervision_id),
  }
}

/** Lee un feature del GeoJSON de supervisión (`id`, `ZONA`, `Area`). */
export function zonaDesdeFeature(feature: { geometry?: unknown; properties?: PropsFuente | null }): ZonaSupervision {
  const p = feature.properties ?? {}
  const codigo = texto(p.codigo ?? p.ZONA ?? p.id) ?? ""
  return {
    codigo,
    nombre: texto(p.nombre ?? p.ZONA) ?? "",
    area_m2: numero(p.area_m2 ?? p.Area),
    geom: geomDe(feature),
  }
}

export function medidasArea(area: AreaVerde): { perimetro_m: number; area_m2: number } | null {
  if (!area.geom) return null
  return medirGeom(area.geom)
}

const AREA_VACIA: AreaVerde = {
  feature_id: "",
  codigo: null,
  nombre: null,
  uso: null,
  proy_riego: null,
  riego_act: null,
  referencia: null,
  perimetro_m: null,
  area_m2: null,
  geom: null,
  zona_supervision_id: null,
}

export function areaNueva(usadas: string[]): AreaVerde {
  let n = usadas.length + 1
  let feature_id = `AV-${String(n).padStart(4, "0")}`
  while (usadas.includes(feature_id)) {
    n += 1
    feature_id = `AV-${String(n).padStart(4, "0")}`
  }
  return { ...AREA_VACIA, feature_id }
}

export function zonaNueva(usadas: string[]): ZonaSupervision {
  const codigo = CODIGOS_ZONA.find((item) => !usadas.includes(item)) ?? ""
  return { codigo, nombre: "", area_m2: null, geom: null }
}

async function leer(path: string): Promise<unknown> {
  const res = await fetch(path, { headers: { Accept: "application/json" } })
  if (!res.ok) throw new Error(`La API respondió ${res.status}`)
  return res.json()
}

async function enviar(path: string, method: string, body: unknown): Promise<void> {
  const res = await fetch(path, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`La API respondió ${res.status}`)
}

function lista(body: unknown, clave: string): unknown[] {
  if (Array.isArray(body)) return body
  if (body && typeof body === "object" && Array.isArray((body as Record<string, unknown>)[clave])) {
    return (body as Record<string, unknown>)[clave] as unknown[]
  }
  return []
}

export async function listarAreas(): Promise<AreaVerde[]> {
  const body = await leer("/api/v1/catastro/areas")
  return lista(body, "areas").map((row, index) => {
    if (row && typeof row === "object" && "geometry" in (row as object)) {
      return areaDesdeFeature(row as { geometry?: unknown; properties?: PropsFuente | null }, index + 1)
    }
    const item = (row ?? {}) as PropsFuente
    return areaDesdeFeature({ properties: item, geometry: item.geom }, index + 1)
  })
}

export async function guardarArea(area: AreaVerde, alta: boolean): Promise<void> {
  const payload = payloadArea(area)
  if (alta) {
    await enviar("/api/v1/catastro/areas", "POST", payload)
    return
  }
  await enviar(`/api/v1/catastro/areas/${encodeURIComponent(area.feature_id)}`, "PATCH", payload)
}

export async function bajaArea(featureId: string): Promise<void> {
  await enviar(`/api/v1/catastro/areas/${encodeURIComponent(featureId)}/baja`, "POST", {})
}

export async function listarZonas(): Promise<ZonaSupervision[]> {
  const body = await leer("/api/v1/catastro/zonas-supervision")
  return lista(body, "zonas").map((row) => {
    if (row && typeof row === "object" && "geometry" in (row as object)) {
      return zonaDesdeFeature(row as { geometry?: unknown; properties?: PropsFuente | null })
    }
    const item = (row ?? {}) as PropsFuente
    return zonaDesdeFeature({ properties: item, geometry: item.geom })
  })
}

export async function guardarZona(zona: ZonaSupervision, alta: boolean): Promise<void> {
  const payload = payloadZona(zona)
  if (alta) {
    await enviar("/api/v1/catastro/zonas-supervision", "POST", payload)
    return
  }
  await enviar(`/api/v1/catastro/zonas-supervision/${encodeURIComponent(zona.codigo)}`, "PATCH", payload)
}

export async function bajaZona(codigo: string): Promise<void> {
  await enviar(`/api/v1/catastro/zonas-supervision/${encodeURIComponent(codigo)}/baja`, "POST", {})
}

function csvCelda(value: string | number | null): string {
  const text = value == null ? "" : String(value)
  if (/[",\n]/.test(text)) return `"${text.replaceAll('"', '""')}"`
  return text
}

export function csvAreas(rows: AreaVerde[]): string {
  const head = [...CAMPOS_AREA]
  const lines = [head.join(",")]
  for (const row of rows) {
    const payload = payloadArea(row)
    lines.push(
      head
        .map((campo) => {
          const value = payload[campo]
          if (campo === "geom") return csvCelda(value ? JSON.stringify(value) : null)
          return csvCelda(value as string | number | null)
        })
        .join(","),
    )
  }
  return lines.join("\n")
}

export function csvZonas(rows: ZonaSupervision[]): string {
  const head = [...CAMPOS_ZONA]
  const lines = [head.join(",")]
  for (const row of rows) {
    const payload = payloadZona(row)
    lines.push(
      head
        .map((campo) => {
          const value = payload[campo]
          if (campo === "geom") return csvCelda(value ? JSON.stringify(value) : null)
          return csvCelda(value as string | number | null)
        })
        .join(","),
    )
  }
  return lines.join("\n")
}

const CUADRO: Position[] = [
  [-77.08, -12.07],
  [-77.079, -12.07],
  [-77.079, -12.0692],
  [-77.08, -12.0692],
  [-77.08, -12.07],
]

export const FIXTURE_AREA_GEOJSON = {
  type: "FeatureCollection" as const,
  features: [
    {
      type: "Feature" as const,
      properties: {
        Nombre: "Bosque de prueba",
        código: "G 13",
        Uso: "Uso Institucional",
        "Proy riego": "Cuenta con aspersión",
        "Riego act": "Riego por aspersión",
        Referenc_1: "Junto al eje central",
        Perimetro: 180.4,
        Área: 980.2,
        feature_id: "AV-0004",
        zona_supervision_id: "Z1",
      },
      geometry: { type: "MultiPolygon" as const, coordinates: [[CUADRO]] },
    },
  ],
}

export const FIXTURE_ZONA_GEOJSON = {
  type: "FeatureCollection" as const,
  features: [
    {
      type: "Feature" as const,
      properties: { id: "Z1", ZONA: "Z1", nombre: "Zona 1", Area: 4200 },
      geometry: {
        type: "MultiPolygon" as const,
        coordinates: [
          [
            [
              [-77.082, -12.072],
              [-77.078, -12.072],
              [-77.078, -12.068],
              [-77.082, -12.068],
              [-77.082, -12.072],
            ],
          ],
        ],
      },
    },
  ],
}

export function areasDeFixture(): AreaVerde[] {
  return FIXTURE_AREA_GEOJSON.features.map((feature, index) => areaDesdeFeature(feature, index + 1))
}

export function zonasDeFixture(): ZonaSupervision[] {
  return FIXTURE_ZONA_GEOJSON.features.map((feature) => zonaDesdeFeature(feature))
}
