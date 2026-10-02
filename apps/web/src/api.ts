import type { FeatureCollection } from "./types"

/** Base de la API. VITE_API_BASE=/api/v1 conserva el prefijo de la API anterior. */
export const API_BASE =
  (import.meta as { env?: { VITE_API_BASE?: string } }).env?.VITE_API_BASE || "/areas-verdes/v1"

/** Une API_BASE con una ruta que antes empezaba en /api/v1. */
export function apiUrl(path: string): string {
  const base = API_BASE.replace(/\/$/, "")
  const suffix = path.startsWith("/") ? path : `/${path}`
  return `${base}${suffix}`
}

export async function fetchCollection(path: string): Promise<FeatureCollection> {
  const res = await fetch(path, { credentials: "include" })
  if (!res.ok) {
    throw new Error(`${path} respondió ${res.status}`)
  }
  const body = (await res.json()) as FeatureCollection
  if (body?.type !== "FeatureCollection" || !Array.isArray(body.features)) {
    throw new Error(`${path} no devolvió un FeatureCollection`)
  }
  return body
}

export function emptyCollection(): FeatureCollection {
  return { type: "FeatureCollection", features: [] }
}
