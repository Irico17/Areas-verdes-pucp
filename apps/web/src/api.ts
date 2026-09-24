import { EMPTY, type FeatureCollection } from "./types"

export async function fetchCollection(path: string): Promise<FeatureCollection> {
  const res = await fetch(path)
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
  return { ...EMPTY, features: [] }
}
