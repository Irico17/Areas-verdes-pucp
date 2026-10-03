export function conteoCapas(data: { areas?: { features: unknown[] }; zonas?: { features: unknown[] } }) {
  return {
    areas: data.areas?.features.length ?? 0,
    zonas: data.zonas?.features.length ?? 0,
  }
}

export function catastroVisible(conteo: { areas: number; zonas: number }) {
  return conteo.areas > 0 && conteo.zonas > 0
}
