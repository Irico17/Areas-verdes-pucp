import type { Map as MapLibre, MapMouseEvent } from "maplibre-gl"

/** Vértice dentro de un MultiPolygon: polígono, anillo y posición. */
export type VerticeRef = {
  polygon: number
  ring: number
  index: number
}

export type Position = [number, number]

export type MultiPolygon = {
  type: "MultiPolygon"
  coordinates: Position[][][]
}

export type EntidadDibujo = "area" | "zona"

/**
 * Único puente con CampusMap. El control de vértices vive en este archivo.
 * `onGeom` recibe el MultiPolygon ya movido, listo para el payload del contrato.
 */
export type ModoDibujo = {
  entidad: EntidadDibujo
  id: string
  geom: MultiPolygon | null
  onGeom: (geom: MultiPolygon) => void
}

const FUENTE = "dibujo-catastro"
const CAPA_RELLENO = "dibujo-catastro-fill"
const CAPA_LINEA = "dibujo-catastro-line"
const CAPA_VERTICE = "dibujo-catastro-vertice"

const CAMPUS = {
  lonMin: -77.3,
  lonMax: -76.9,
  latMin: -12.2,
  latMax: -11.9,
}

export function clonGeom(geom: MultiPolygon): MultiPolygon {
  return {
    type: "MultiPolygon",
    coordinates: geom.coordinates.map((polygon) => polygon.map((ring) => ring.map((p) => [p[0], p[1]] as Position))),
  }
}

export function anilloCerrado(ring: Position[]): boolean {
  if (ring.length < 4) return false
  const a = ring[0]
  const b = ring[ring.length - 1]
  return a[0] === b[0] && a[1] === b[1]
}

function casiIgual(a: Position, b: Position): boolean {
  return Math.abs(a[0] - b[0]) < 1e-12 && Math.abs(a[1] - b[1]) < 1e-12
}

function segmentosCruzan(a: Position, b: Position, c: Position, d: Position): boolean {
  const cruza = (p: Position, q: Position, r: Position) => {
    const v = (q[0] - p[0]) * (r[1] - p[1]) - (q[1] - p[1]) * (r[0] - p[0])
    return v
  }
  const o1 = cruza(a, b, c)
  const o2 = cruza(a, b, d)
  const o3 = cruza(c, d, a)
  const o4 = cruza(c, d, b)
  return o1 * o2 < 0 && o3 * o4 < 0
}

function anilloSeCruza(ring: Position[]): boolean {
  const n = ring.length - 1
  if (n < 4) return false
  for (let i = 0; i < n; i++) {
    const a = ring[i]
    const b = ring[(i + 1) % n]
    for (let j = i + 1; j < n; j++) {
      if (Math.abs(i - j) <= 1 || (i === 0 && j === n - 1)) continue
      const c = ring[j]
      const d = ring[(j + 1) % n]
      if (segmentosCruzan(a, b, c, d)) return true
    }
  }
  return false
}

function areaAnillo(ring: Position[]): number {
  let sum = 0
  for (let i = 0; i < ring.length - 1; i++) {
    sum += ring[i][0] * ring[i + 1][1] - ring[i + 1][0] * ring[i][1]
  }
  return sum / 2
}

/** Metros aproximados en el plano local del campus. Sirve para avisar, no sustituye a PostGIS. */
export function medirGeom(geom: MultiPolygon): { perimetro_m: number; area_m2: number } {
  const origen = geom.coordinates[0]?.[0]?.[0]
  if (!origen) return { perimetro_m: 0, area_m2: 0 }
  const cos = Math.cos((origen[1] * Math.PI) / 180)
  const mx = 111_320 * cos
  const my = 110_540
  const aMetros = (p: Position): Position => [(p[0] - origen[0]) * mx, (p[1] - origen[1]) * my]
  let perimetro = 0
  let area = 0
  for (const polygon of geom.coordinates) {
    polygon.forEach((ring, index) => {
      const pts = ring.map(aMetros)
      let borde = 0
      for (let i = 0; i < pts.length - 1; i++) {
        borde += Math.hypot(pts[i + 1][0] - pts[i][0], pts[i + 1][1] - pts[i][1])
      }
      perimetro += borde
      const signo = index === 0 ? 1 : -1
      area += signo * Math.abs(areaAnillo(pts))
    })
  }
  return { perimetro_m: Math.round(perimetro * 1000) / 1000, area_m2: Math.round(area * 1000) / 1000 }
}

export function dentroDelCampus(p: Position): boolean {
  return p[0] >= CAMPUS.lonMin && p[0] <= CAMPUS.lonMax && p[1] >= CAMPUS.latMin && p[1] <= CAMPUS.latMax
}

export function erroresGeometria(geom: MultiPolygon | null, obligatoria: boolean): string[] {
  if (!geom) return obligatoria ? ["La geometría es obligatoria."] : []
  if (geom.type !== "MultiPolygon" || !Array.isArray(geom.coordinates) || geom.coordinates.length === 0) {
    return ["La geometría tiene que ser un MultiPolygon."]
  }
  const errores: string[] = []
  geom.coordinates.forEach((polygon, pi) => {
    if (!Array.isArray(polygon) || polygon.length === 0) {
      errores.push(`El polígono ${pi + 1} no tiene anillos.`)
      return
    }
    polygon.forEach((ring, ri) => {
      const donde = `polígono ${pi + 1}, anillo ${ri + 1}`
      if (!Array.isArray(ring) || ring.length < 4) {
        errores.push(`El ${donde} necesita al menos cuatro posiciones.`)
        return
      }
      for (const p of ring) {
        if (!Array.isArray(p) || p.length < 2 || !Number.isFinite(p[0]) || !Number.isFinite(p[1])) {
          errores.push(`El ${donde} tiene una coordenada inválida.`)
          return
        }
        if (!dentroDelCampus(p)) {
          errores.push(`El ${donde} sale del campus (lon −77.30 a −76.90, lat −12.20 a −11.90).`)
          return
        }
      }
      if (!anilloCerrado(ring)) errores.push(`El ${donde} no está cerrado.`)
      if (Math.abs(areaAnillo(ring)) < 1e-14) errores.push(`El ${donde} no tiene área.`)
      if (anilloSeCruza(ring)) errores.push(`El ${donde} se cruza consigo mismo.`)
    })
  })
  return errores
}

export function listarVertices(geom: MultiPolygon): VerticeRef[] {
  const out: VerticeRef[] = []
  geom.coordinates.forEach((polygon, polygonIndex) => {
    polygon.forEach((ring, ringIndex) => {
      const ultimo = ring.length - 1
      ring.forEach((_, index) => {
        if (index === ultimo && anilloCerrado(ring)) return
        out.push({ polygon: polygonIndex, ring: ringIndex, index })
      })
    })
  })
  return out
}

export function moverVertice(geom: MultiPolygon, ref: VerticeRef, destino: Position): MultiPolygon {
  const next = clonGeom(geom)
  const ring = next.coordinates[ref.polygon]?.[ref.ring]
  if (!ring || !ring[ref.index]) return geom
  const previo = ring[ref.index]
  ring[ref.index] = [destino[0], destino[1]]
  const cierre = ring.length - 1
  if (anilloCerrado(ring) || casiIgual(ring[0], ring[cierre]) || casiIgual(previo, ring[cierre])) {
    if (ref.index === 0) ring[cierre] = [destino[0], destino[1]]
    if (ref.index === cierre) ring[0] = [destino[0], destino[1]]
  }
  return next
}

export function cerrarTrazo(puntos: Position[]): MultiPolygon | null {
  if (puntos.length < 3) return null
  const ring = puntos.map((p) => [p[0], p[1]] as Position)
  const primero = ring[0]
  const ultimo = ring[ring.length - 1]
  if (!casiIgual(primero, ultimo)) ring.push([primero[0], primero[1]])
  return { type: "MultiPolygon", coordinates: [[ring]] }
}

function puntosDe(geom: MultiPolygon | null): GeoJSON.FeatureCollection {
  const features: GeoJSON.Feature[] = []
  if (!geom) return { type: "FeatureCollection", features }
  for (const ref of listarVertices(geom)) {
    const p = geom.coordinates[ref.polygon][ref.ring][ref.index]
    features.push({
      type: "Feature",
      properties: { polygon: ref.polygon, ring: ref.ring, index: ref.index },
      geometry: { type: "Point", coordinates: p },
    })
  }
  return { type: "FeatureCollection", features }
}

function coleccion(geom: MultiPolygon | null): GeoJSON.FeatureCollection {
  if (!geom) return { type: "FeatureCollection", features: [] }
  return {
    type: "FeatureCollection",
    features: [{ type: "Feature", properties: {}, geometry: geom }],
  }
}

type Src = { setData: (data: GeoJSON.FeatureCollection) => void }

function quitar(map: MapLibre) {
  for (const id of [CAPA_VERTICE, CAPA_LINEA, CAPA_RELLENO]) {
    if (map.getLayer(id)) map.removeLayer(id)
  }
  if (map.getSource(FUENTE)) map.removeSource(FUENTE)
  if (map.getSource(`${FUENTE}-vertices`)) map.removeSource(`${FUENTE}-vertices`)
}

/**
 * Dibuja el polígono en edición y mueve vértices con el mouse.
 * Flechas nudgen el vértice activo; Tab pasa al siguiente; Escape no confirma.
 * Devuelve la función que desmonta capas y listeners.
 */
export function activarDibujo(map: MapLibre, modo: ModoDibujo | null): { actualizar: (geom: MultiPolygon | null) => void; cerrar: () => void } {
  const vacio = { actualizar: () => undefined, cerrar: () => undefined }
  quitar(map)
  if (!modo) return vacio

  let geom = modo.geom ? clonGeom(modo.geom) : null
  let trazo: Position[] = []
  let activo: VerticeRef | null = geom ? (listarVertices(geom)[0] ?? null) : null
  let arrastrando = false

  map.addSource(FUENTE, { type: "geojson", data: coleccion(geom) })
  map.addSource(`${FUENTE}-vertices`, { type: "geojson", data: puntosDe(geom) })
  map.addLayer({
    id: CAPA_RELLENO,
    type: "fill",
    source: FUENTE,
    paint: { "fill-color": "#308046", "fill-opacity": 0.28 },
  })
  map.addLayer({
    id: CAPA_LINEA,
    type: "line",
    source: FUENTE,
    paint: { "line-color": "#083465", "line-width": 2 },
  })
  map.addLayer({
    id: CAPA_VERTICE,
    type: "circle",
    source: `${FUENTE}-vertices`,
    paint: {
      "circle-radius": 6,
      "circle-color": "#f7faf8",
      "circle-stroke-width": 2,
      "circle-stroke-color": "#083465",
    },
  })

  const pintar = () => {
    const src = map.getSource(FUENTE) as Src | undefined
    const verts = map.getSource(`${FUENTE}-vertices`) as Src | undefined
    src?.setData(coleccion(geom))
    if (geom) {
      verts?.setData(puntosDe(geom))
      return
    }
    verts?.setData({
      type: "FeatureCollection",
      features: trazo.map((p, index) => ({
        type: "Feature",
        properties: { index },
        geometry: { type: "Point", coordinates: p },
      })),
    })
  }

  const publicar = (next: MultiPolygon) => {
    geom = next
    trazo = []
    pintar()
    modo.onGeom(clonGeom(next))
  }

  const alPulsar = (event: MapMouseEvent) => {
    if (!geom) {
      trazo = [...trazo, [event.lngLat.lng, event.lngLat.lat]]
      pintar()
      return
    }
    const hits = map.queryRenderedFeatures(event.point, { layers: [CAPA_VERTICE] })
    const hit = hits[0]
    if (!hit) return
    const props = hit.properties ?? {}
    activo = {
      polygon: Number(props.polygon),
      ring: Number(props.ring),
      index: Number(props.index),
    }
    arrastrando = true
    map.dragPan.disable()
  }

  const alMover = (event: MapMouseEvent) => {
    if (!arrastrando || !geom || !activo) return
    const next = moverVertice(geom, activo, [event.lngLat.lng, event.lngLat.lat])
    geom = next
    const src = map.getSource(FUENTE) as Src | undefined
    const verts = map.getSource(`${FUENTE}-vertices`) as Src | undefined
    src?.setData(coleccion(geom))
    verts?.setData(puntosDe(geom))
  }

  const alSoltar = () => {
    if (!arrastrando || !geom) return
    arrastrando = false
    map.dragPan.enable()
    modo.onGeom(clonGeom(geom))
  }

  const alDoble = (event: MapMouseEvent) => {
    if (geom || trazo.length < 3) return
    event.preventDefault()
    const cerrado = cerrarTrazo(trazo)
    if (cerrado) publicar(cerrado)
  }

  const alTecla = (event: KeyboardEvent) => {
    if (!geom || !activo) return
    const paso = 0.00002 * Math.max(1, 16 - map.getZoom())
    let dx = 0
    let dy = 0
    if (event.key === "ArrowLeft") dx = -paso
    else if (event.key === "ArrowRight") dx = paso
    else if (event.key === "ArrowUp") dy = paso
    else if (event.key === "ArrowDown") dy = -paso
    else if (event.key === "Tab") {
      const todos = listarVertices(geom)
      const i = todos.findIndex((v) => v.polygon === activo?.polygon && v.ring === activo.ring && v.index === activo.index)
      const delta = event.shiftKey ? -1 : 1
      activo = todos[(i + delta + todos.length) % todos.length] ?? activo
      event.preventDefault()
      return
    } else return
    event.preventDefault()
    const actual = geom.coordinates[activo.polygon][activo.ring][activo.index]
    publicar(moverVertice(geom, activo, [actual[0] + dx, actual[1] + dy]))
  }

  map.on("mousedown", alPulsar)
  map.on("mousemove", alMover)
  map.on("mouseup", alSoltar)
  map.on("dblclick", alDoble)
  window.addEventListener("keydown", alTecla)
  map.doubleClickZoom.disable()
  pintar()

  return {
    actualizar: (next) => {
      if (arrastrando) return
      geom = next ? clonGeom(next) : null
      if (geom) trazo = []
      const sigue = activo && geom?.coordinates[activo.polygon]?.[activo.ring]?.[activo.index]
      if (!sigue) activo = geom ? (listarVertices(geom)[0] ?? null) : null
      pintar()
    },
    cerrar: () => {
      map.off("mousedown", alPulsar)
      map.off("mousemove", alMover)
      map.off("mouseup", alSoltar)
      map.off("dblclick", alDoble)
      window.removeEventListener("keydown", alTecla)
      map.doubleClickZoom.enable()
      map.dragPan.enable()
      quitar(map)
    },
  }
}
