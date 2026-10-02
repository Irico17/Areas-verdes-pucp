/** En el teléfono caben cinco casillas: cuatro módulos y «Más» con el resto. */
export function repartirModulos<T>(modulos: readonly T[], casillas = 5): { barra: T[]; mas: T[] } {
  if (modulos.length <= casillas) return { barra: [...modulos], mas: [] }
  return { barra: modulos.slice(0, casillas - 1), mas: modulos.slice(casillas - 1) }
}

const BARRA_CAPATAZ = ["hoy", "mapa", "labores", "registros"]
const BARRA_JEFATURA = ["resumen", "mapa", "labores", "reportes"]
const BARRA_OFICINA = ["resumen", "mapa", "labores", "solicitudes"]

/** El capataz no usa «Más»: solo las cuatro tareas de campo. */
export function barraMovil<T extends { id: string }>(rol: string, modulos: readonly T[]): { barra: T[]; mas: T[] } {
  const ids = rol === "capataz" ? BARRA_CAPATAZ : rol === "jefatura" ? BARRA_JEFATURA : BARRA_OFICINA
  const barra = ids.flatMap((id) => {
    const item = modulos.find((modulo) => modulo.id === id)
    return item ? [item] : []
  })
  if (rol === "capataz") return { barra, mas: [] }
  const enBarra = new Set(barra.map((item) => item.id))
  return { barra, mas: modulos.filter((item) => !enBarra.has(item.id)) }
}
