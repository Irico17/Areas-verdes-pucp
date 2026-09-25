/** Alturas respecto al espacio útil del móvil, de la vista mínima a la casi completa. */
export const ANCLAJES = [0.28, 0.55, 0.92] as const

export type Anclaje = (typeof ANCLAJES)[number]

export function anclajeCercano(ratio: number): Anclaje | null {
  if (ratio < 0.16) return null
  let mejor: Anclaje = ANCLAJES[0]
  let dist = Math.abs(ratio - mejor)
  for (const punto of ANCLAJES) {
    const d = Math.abs(ratio - punto)
    if (d < dist) {
      mejor = punto
      dist = d
    }
  }
  return mejor
}

export function siguienteAnclaje(actual: Anclaje, direccion: 1 | -1): Anclaje | null {
  const i = ANCLAJES.indexOf(actual)
  const siguiente = i + direccion
  if (siguiente < 0) return null
  if (siguiente >= ANCLAJES.length) return ANCLAJES[ANCLAJES.length - 1]
  return ANCLAJES[siguiente]
}

export function alturaPx(anclaje: Anclaje, altoUtil: number): number {
  return Math.round(altoUtil * anclaje)
}
