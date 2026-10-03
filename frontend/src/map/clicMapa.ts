export type DecisionClic = "ignorar" | "vacio" | "detalle"

function tieneGeometria(geometry: unknown): boolean {
  if (!geometry || typeof geometry !== "object") return false
  const figura = geometry as { type?: unknown; coordinates?: unknown; geometries?: unknown }
  if (figura.type === "GeometryCollection") return Array.isArray(figura.geometries) && figura.geometries.length > 0
  if (typeof figura.type !== "string" || !figura.type) return false
  return Array.isArray(figura.coordinates) && figura.coordinates.length > 0
}

function tieneDatos(props: Record<string, unknown> | null | undefined): boolean {
  if (!props) return false
  return Object.values(props).some((valor) => {
    if (valor == null) return false
    if (typeof valor === "string") return valor.trim() !== ""
    if (typeof valor === "number") return Number.isFinite(valor)
    if (typeof valor === "boolean") return true
    return false
  })
}

/** Un elemento sin geometría se ignora. Sin propiedades, detalle vacío. Nunca cambia la vista. */
export function decisionClic(geometry: unknown, props: Record<string, unknown> | null | undefined): DecisionClic {
  if (!tieneGeometria(geometry)) return "ignorar"
  if (!tieneDatos(props)) return "vacio"
  return "detalle"
}
