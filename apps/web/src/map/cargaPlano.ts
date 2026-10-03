/** Cuántas teselas distintas tienen que fallar, sin ninguna buena, para dar el plano por no arrancado. */
export const UMBRAL_TESELAS = 4

export type CargaPlano = {
  estiloListo: boolean
  estiloRoto: boolean
  ok: number
  fallidas: Record<string, true>
}

export function cargaInicial(): CargaPlano {
  return { estiloListo: false, estiloRoto: false, ok: 0, fallidas: {} }
}

export function textoDeError(error: unknown): string {
  if (typeof error === "string") return error
  if (!error || typeof error !== "object") return ""
  const mensaje = error instanceof Error ? error.message : ""
  const url = "url" in error && error.url != null ? String(error.url) : ""
  if (url && !mensaje.includes(url)) return `${mensaje} ${url}`.trim()
  return mensaje || url
}

/** Una tesela suelta, un clic o una capa ausente no son «el plano no cargó». */
export function clasificarErrorMapa(texto: string, sourceId?: string): "tesela" | "estilo" | "ignorar" {
  const t = texto.toLowerCase()
  if (sourceId === "osm" || /tile\.openstreetmap\.org\/\d+\/\d+\/\d+/.test(t)) return "tesela"
  if (/feature|does not exist|queryrendered|style is not done loading|cannot be queried/.test(t)) return "ignorar"
  if (/failed to load style|style could not be loaded|cannot read style/.test(t)) return "estilo"
  return "ignorar"
}

export function claveTesela(texto: string): string | null {
  const coincidencia = texto.match(/tile\.openstreetmap\.org\/(\d+)\/(\d+)\/(\d+)/i)
  if (!coincidencia) return null
  return `${coincidencia[1]}/${coincidencia[2]}/${coincidencia[3]}`
}

export function registrarFalloTesela(estado: CargaPlano, texto: string): CargaPlano {
  const clave = claveTesela(texto)
  if (!clave || estado.fallidas[clave]) return estado
  return { ...estado, fallidas: { ...estado.fallidas, [clave]: true } }
}

/**
 * MapLibre avisa cada tesela ráster servida con `tile` en `sourcedata`,
 * sin `sourceDataType: "content"` (eso es de un GeoJSON).
 */
export function esTeselaServida(evento: { sourceId?: string; tile?: unknown }): boolean {
  return evento.sourceId === "osm" && evento.tile != null
}

export function registrarTeselaOk(estado: CargaPlano): CargaPlano {
  return { ...estado, ok: estado.ok + 1 }
}

export function marcarEstiloListo(estado: CargaPlano): CargaPlano {
  return { ...estado, estiloListo: true }
}

export function marcarEstiloRoto(estado: CargaPlano): CargaPlano {
  return { ...estado, estiloRoto: true }
}

/**
 * El plano no arrancó si el estilo no cargó, o si ninguna tesela llegó.
 * Con una tesela buena, los fallos posteriores no cambian la vista.
 */
export function planoNoArranco(estado: CargaPlano, enReposo = false): boolean {
  if (estado.ok > 0) return false
  if (estado.estiloRoto && !estado.estiloListo) return true
  const fallos = Object.keys(estado.fallidas).length
  if (fallos >= UMBRAL_TESELAS) return true
  return enReposo && estado.estiloListo && fallos >= 1
}
