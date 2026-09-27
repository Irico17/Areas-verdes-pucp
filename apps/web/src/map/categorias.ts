import type { ExpressionSpecification, FilterSpecification } from "@maplibre/maplibre-gl-style-spec"
import type { FeatureCollection, GeoFeature } from "../types"

export type ColorPor = "uso" | "sector"
export type UsoId = "institucional" | "administrativo" | "recreativo" | "sostenible" | "otros"
export type SectorId = "cua-valeria" | "cua-mateo" | "cua-renato" | "campo-deportivo" | "bosque-humedo" | "sin-sector"
export type Categoria<T extends string> = { id: T; label: string; fill: string; line: string }

// Mismo tono que el original (colores-original-*), reinterpretado con tonos
// apagados de luminosidad media y sin repetir marca ni semántica. Naranja
// (administrativo) y verde (recreativo) se separan por luminosidad para
// deuteranopía, igual que verde azulado (institucional) y azul (sostenible).
export const USOS: Categoria<UsoId>[] = [
  { id: "institucional", label: "Institucional", fill: "#1f5f58", line: "#143f3a" },
  { id: "administrativo", label: "Administrativo", fill: "#b86a24", line: "#7d4716" },
  { id: "recreativo", label: "Recreativo / descanso", fill: "#7fa94a", line: "#4f6e2c" },
  { id: "sostenible", label: "Manejo sostenible", fill: "#4f86bf", line: "#2f5a86" },
  { id: "otros", label: "Otros", fill: "#8d968f", line: "#5c6a62" },
]

export const SECTORES: Categoria<SectorId>[] = [
  { id: "cua-valeria", label: "Cuadrilla Valeria Quispe", fill: "#6b5596", line: "#463763" },
  { id: "cua-mateo", label: "Cuadrilla Mateo Salazar", fill: "#3f73b0", line: "#284c78" },
  { id: "cua-renato", label: "Cuadrilla Renato Cárdenas", fill: "#c27c2c", line: "#83521b" },
  { id: "campo-deportivo", label: "Campo deportivo", fill: "#9aab3e", line: "#66722a" },
  { id: "bosque-humedo", label: "Bosque húmedo", fill: "#3f9a82", line: "#286656" },
  { id: "sin-sector", label: "Sin sector", fill: "#8d968f", line: "#5c6a62" },
]

export const CAMPO: Record<ColorPor, "cat_uso" | "cat_sector"> = { uso: "cat_uso", sector: "cat_sector" }

export function normalizar(s: unknown): string {
  return String(s ?? "")
    .toLowerCase()
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .trim()
    .replace(/\s+/g, " ")
}

// Réplica de obtenerCategoriaUso (data/raw/legacy-app/script.js): "Áreas
// deportivas y recreación activa" y "Áreas de conservación" caen en otros,
// igual que en el mapa original (docs/MAPA-DATOS-Y-EDICION.md).
export function categoriaUso(uso: unknown): UsoId {
  const n = normalizar(uso)
  if (n.includes("institucional")) return "institucional"
  if (n.includes("administrativo")) return "administrativo"
  if (n.includes("recreativo") || n.includes("descanso")) return "recreativo"
  if (n.includes("sostenible") || n.includes("reduccion") || n.includes("consumo")) return "sostenible"
  return "otros"
}

const SECTOR_IDS = new Set<string>(SECTORES.map((s) => s.id).filter((id) => id !== "sin-sector"))

export function categoriaSector(sector: unknown): SectorId {
  const s = typeof sector === "string" ? sector : ""
  return SECTOR_IDS.has(s) ? (s as SectorId) : "sin-sector"
}

function clonarFeature(f: GeoFeature, campo: string, valor: string): GeoFeature {
  return { ...f, properties: { ...(f.properties ?? {}), [campo]: valor } }
}

// conCategorias no muta fc: agrega cat_uso o cat_sector a cada feature para
// que las expresiones data-driven de maplibre lean un campo ya resuelto.
export function conCategorias(fc: FeatureCollection | undefined, modo: ColorPor): FeatureCollection {
  if (!fc) return { type: "FeatureCollection", features: [] }
  const campo = CAMPO[modo]
  const features = fc.features.map((f) => {
    const valor = modo === "uso" ? categoriaUso(f.properties?.uso) : categoriaSector(f.properties?.sector)
    return clonarFeature(f, campo, valor)
  })
  return { ...fc, features }
}

export function conteoPorCategoria<T extends string>(
  fc: FeatureCollection | undefined,
  campo: string,
  cats: Categoria<T>[],
): Record<T, number> {
  const out = Object.fromEntries(cats.map((c) => [c.id, 0])) as Record<T, number>
  for (const f of fc?.features ?? []) {
    const valor = f.properties?.[campo]
    if (typeof valor === "string" && valor in out) out[valor as T]++
  }
  return out
}

export function expresionColor<T extends string>(campo: string, cats: Categoria<T>[], clave: "fill" | "line"): ExpressionSpecification {
  const expr: unknown[] = ["match", ["get", campo]]
  for (const cat of cats) {
    expr.push(cat.id, cat[clave])
  }
  expr.push(cats[cats.length - 1][clave])
  return expr as ExpressionSpecification
}

export const OPACIDAD_RELLENO: ExpressionSpecification = [
  "case",
  ["boolean", ["feature-state", "sel"], false],
  0.8,
  ["boolean", ["feature-state", "hover"], false],
  0.68,
  0.5,
] as ExpressionSpecification

export const ANCHO_REALCE: ExpressionSpecification = [
  "case",
  ["boolean", ["feature-state", "sel"], false],
  3,
  ["boolean", ["feature-state", "hover"], false],
  2,
  0,
] as ExpressionSpecification

export const COLOR_REALCE: ExpressionSpecification = [
  "case",
  ["boolean", ["feature-state", "sel"], false],
  "#083465",
  "#102033",
] as ExpressionSpecification

export function filtroCategorias(campo: string, activos: string[], todos: string[]): FilterSpecification | null {
  if (activos.length >= todos.length && todos.every((id) => activos.includes(id))) return null
  return ["in", ["get", campo], ["literal", activos]] as FilterSpecification
}

export function etiquetaCategoria(modo: ColorPor, id: string): string {
  const cats = modo === "uso" ? USOS : SECTORES
  return cats.find((c) => c.id === id)?.label ?? id
}
