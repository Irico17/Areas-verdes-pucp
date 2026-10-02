import { useEffect, useRef, useState } from "react"
import type { FeatureCollection as GJFeatureCollection } from "geojson"
import type { ExpressionSpecification } from "@maplibre/maplibre-gl-style-spec"
import {
  GeoJSONSource,
  Map,
  MapMouseEvent,
  Marker,
  NavigationControl,
  Popup,
  ScaleControl,
  setWorkerUrl,
  type LngLatBoundsLike,
} from "maplibre-gl"
import "maplibre-gl/dist/maplibre-gl.css"
import {
  ANCHO_REALCE,
  COLOR_REALCE,
  etiquetaCategoria,
  expresionColor,
  filtroCategorias,
  OPACIDAD_RELLENO,
  OPACIDAD_SECTOR,
  SIN_SECTOR,
  USOS,
  type Categoria,
} from "./categorias"
import { catastroVisible, conteoCapas } from "./coverage"
import { activarDibujo, type ModoDibujo } from "./draw"
import { INVENTARIO } from "../inventario"
import { etiquetaEstado, etiquetaTipo } from "../operacion"
import { EMPTY, LAYERS, type FeatureCollection, type LayerId } from "../types"
import { ACTIVIDAD } from "../ui/nomenclatura"
import { apiUrl } from "../api"

const CATASTRO_FILLS = ["areas-fill", "zonas-fill"]
const USO_IDS = USOS.map((c) => c.id)

type Props = {
  data: Partial<Record<LayerId, FeatureCollection>>
  visible: Record<LayerId, boolean>
  activities: FeatureCollection
  pinMode: boolean
  draft: { lon: number; lat: number } | null
  focus: { lon: number; lat: number; token: number } | null
  relieve: boolean
  edificios: FeatureCollection
  showEdificios: boolean
  inventory: Partial<Record<string, FeatureCollection>>
  inventoryOn: Record<string, boolean>
  ocultas: string[]
  sectores?: Categoria<string>[]
  onSelectCatastro: (hit: { layer: string; props: Record<string, unknown> } | null) => void
  onSelectActividad: (id: string | null) => void
  onPin: (lon: number, lat: number) => void
  modoDibujo?: ModoDibujo | null
}

const CAMPUS: [number, number] = [-77.0796, -12.0696]
const CAMPUS_BOUNDS: LngLatBoundsLike = [
  [-77.0832, -12.07415],
  [-77.07795, -12.0644],
]

// En producción el bundle busca ./maplibre-gl-worker.mjs junto al JS con hash
// y nginx devolvía index.html. El worker y su módulo compartido viven en la raíz.
setWorkerUrl("/maplibre-gl-worker.mjs")

function escapeHtml(value: unknown): string {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
}

function labelOf(id: string): string {
  return LAYERS.find((layer) => layer.id === id)?.label ?? id
}

function asCollection(fc: FeatureCollection | undefined): GJFeatureCollection {
  return (fc ?? EMPTY) as unknown as GJFeatureCollection
}

function nearestInventoryPoint(
  map: Map,
  point: { x: number; y: number },
  data: Partial<Record<string, FeatureCollection>>,
  on: Record<string, boolean>,
): { id: string; props: Record<string, unknown> } | null {
  let best: { id: string; props: Record<string, unknown>; d: number } | null = null
  for (const layer of INVENTARIO) {
    if (layer.kind !== "point" || !on[layer.id]) continue
    for (const feature of data[layer.id]?.features ?? []) {
      const geometry = feature.geometry
      if (!geometry || geometry.type !== "Point" || !Array.isArray(geometry.coordinates)) continue
      const [lon, lat] = geometry.coordinates as number[]
      if (typeof lon !== "number" || typeof lat !== "number") continue
      const projected = map.project([lon, lat])
      const distance = Math.hypot(projected.x - point.x, projected.y - point.y)
      if (distance <= 18 && (!best || distance < best.d)) {
        best = { id: layer.id, props: { ...(feature.properties ?? {}) }, d: distance }
      }
    }
  }
  return best ? { id: best.id, props: best.props } : null
}

function desfaseHoja(contenedor: HTMLElement): [number, number] {
  const hoja = parseFloat(getComputedStyle(contenedor).getPropertyValue("--sheet-h")) || 0
  if (hoja <= 0) return [0, 0]
  const alto = contenedor.clientHeight
  const arriba = document.querySelector(".topbar")?.getBoundingClientRect().bottom ?? 0
  const nav = document.querySelector(".guard")?.getBoundingClientRect().height ?? 0
  const bordeHoja = alto - nav - hoja
  return [0, Math.round((arriba + bordeHoja) / 2 - alto / 2)]
}

export function CampusMap({
  data,
  visible,
  activities,
  pinMode,
  draft,
  focus,
  onSelectCatastro,
  onSelectActividad,
  onPin,
  relieve,
  edificios,
  showEdificios,
  inventory,
  inventoryOn,
  ocultas,
  sectores = [SIN_SECTOR],
  modoDibujo = null,
}: Props) {
  const host = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<Map | null>(null)
  const popupRef = useRef<Popup | null>(null)
  const markerRef = useRef<Marker | null>(null)
  const [ready, setReady] = useState(false)
  const encuadrado = useRef(false)
  const pinRef = useRef(pinMode)
  const onCatastro = useRef(onSelectCatastro)
  const onActividad = useRef(onSelectActividad)
  const onPinRef = useRef(onPin)
  const inventoryRef = useRef(inventory)
  const inventoryOnRef = useRef(inventoryOn)
  const dibujoRef = useRef(modoDibujo)
  const sectoresRef = useRef(sectores)
  const hoverRef = useRef<{ source: string; id: string | number } | null>(null)
  const selRef = useRef<{ source: string; id: string | number } | null>(null)

  useEffect(() => {
    pinRef.current = pinMode
    dibujoRef.current = modoDibujo
    inventoryRef.current = inventory
    inventoryOnRef.current = inventoryOn
    onCatastro.current = onSelectCatastro
    onActividad.current = onSelectActividad
    onPinRef.current = onPin
    sectoresRef.current = sectores
  }, [pinMode, modoDibujo, inventory, inventoryOn, onSelectCatastro, onSelectActividad, onPin, sectores])

  useEffect(() => {
    if (!host.current || mapRef.current) return
    const map = new Map({
      container: host.current,
      center: CAMPUS,
      zoom: 15.4,
      style: {
        version: 8,
        glyphs: "https://demotiles.maplibre.org/font/{fontstack}/{range}.pbf",
        sources: {
          osm: {
            type: "raster",
            tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"],
            tileSize: 256,
            attribution: "© OpenStreetMap",
          },
        },
        layers: [{ id: "osm", type: "raster", source: "osm" }],
      },
    })
    map.addControl(new NavigationControl({ showCompass: true, visualizePitch: false }), "bottom-right")
    map.addControl(new ScaleControl({ maxWidth: 120, unit: "metric" }), "bottom-left")
    map.on("load", () => {
      const alturaExtrusion = [
        "interpolate",
        ["linear"],
        ["coalesce", ["to-number", ["get", "area_m2"]], 0],
        0,
        2.5,
        500,
        4,
        2500,
        8,
        12000,
        14,
      ] as ExpressionSpecification
      for (const layer of LAYERS) {
        map.addSource(layer.id, { type: "geojson", data: { type: "FeatureCollection", features: [] }, promoteId: "id" })
        const esCatastro = layer.id === "areas" || layer.id === "zonas"
        const esLinea = layer.id === "vias"
        const campo = layer.id === "zonas" ? "cat_sector" : "cat_uso"
        const cats = layer.id === "zonas" ? sectoresRef.current : USOS
        if (!esLinea) {
          map.addLayer({
            id: `${layer.id}-fill`,
            type: "fill",
            source: layer.id,
            paint: {
              "fill-color": esCatastro ? expresionColor(campo, cats, "fill") : layer.fill,
              "fill-opacity": layer.id === "zonas" ? OPACIDAD_SECTOR : esCatastro ? OPACIDAD_RELLENO : layer.fillOpacity,
            },
          })
        }
        map.addLayer({
          id: `${layer.id}-line`,
          type: "line",
          source: layer.id,
          paint: {
            "line-color": esCatastro ? expresionColor(campo, cats, "line") : layer.line,
            "line-width": layer.id === "zonas" ? 1.6 : esCatastro ? 0.8 : 1.25,
          },
        })
      }
      map.addLayer({
        id: "areas-realce",
        type: "line",
        source: "areas",
        paint: { "line-color": COLOR_REALCE, "line-width": ANCHO_REALCE },
      })
      map.addLayer({
        id: "zonas-realce",
        type: "line",
        source: "zonas",
        paint: { "line-color": COLOR_REALCE, "line-width": ANCHO_REALCE },
      })
      map.addLayer({
        id: "areas-extrusion",
        type: "fill-extrusion",
        source: "areas",
        layout: { visibility: "none" },
        paint: {
          "fill-extrusion-color": ["case", ["boolean", ["feature-state", "sel"], false], "#083465", expresionColor("cat_uso", USOS, "fill")],
          "fill-extrusion-opacity": 0.8,
          "fill-extrusion-height": alturaExtrusion,
        },
      })
      map.addLayer({
        id: "zonas-extrusion",
        type: "fill-extrusion",
        source: "zonas",
        layout: { visibility: "none" },
        paint: {
          "fill-extrusion-color": ["case", ["boolean", ["feature-state", "sel"], false], "#083465", expresionColor("cat_sector", sectoresRef.current, "fill")],
          "fill-extrusion-opacity": 0.8,
          "fill-extrusion-height": alturaExtrusion,
        },
      })
      map.addSource("edificios", { type: "geojson", data: { type: "FeatureCollection", features: [] } })
      map.addLayer({
        id: "edificios-extrusion",
        type: "fill-extrusion",
        source: "edificios",
        layout: { visibility: "none" },
        paint: {
          "fill-extrusion-color": "#c8c0b2",
          "fill-extrusion-opacity": 0.55,
          "fill-extrusion-height": ["coalesce", ["to-number", ["get", "altura_m"]], 8],
        },
      })
      for (const layer of INVENTARIO) {
        map.addSource(`inv-${layer.id}`, { type: "geojson", data: { type: "FeatureCollection", features: [] } })
        if (layer.kind === "point") {
          map.addLayer({
            id: `inv-${layer.id}-circle`,
            type: "circle",
            source: `inv-${layer.id}`,
            paint: {
              "circle-radius": layer.id === "bebederos" ? 9 : 7,
              "circle-color": layer.color,
              "circle-stroke-width": 1.5,
              "circle-stroke-color": "#f4f7f4",
            },
          })
        } else {
          map.addLayer({
            id: `inv-${layer.id}-fill`,
            type: "fill",
            source: `inv-${layer.id}`,
            paint: { "fill-color": layer.color, "fill-opacity": 0.28 },
          })
          map.addLayer({
            id: `inv-${layer.id}-line`,
            type: "line",
            source: `inv-${layer.id}`,
            paint: { "line-color": layer.color, "line-width": 1.25 },
          })
        }
      }
      map.addSource("actividades", { type: "geojson", data: { type: "FeatureCollection", features: [] } })
      map.addLayer({
        id: "actividades-circle",
        type: "circle",
        source: "actividades",
        paint: {
          "circle-radius": 13,
          "circle-stroke-width": 2,
          "circle-stroke-color": "#f7faf6",
          "circle-color": [
            "match",
            ["get", "estado"],
            "pendiente",
            "#8a6410",
            "en_proceso",
            "#308046",
            "bloqueada",
            "#8c3832",
            "cerrada",
            "#5c6a62",
            "cancelada",
            "#5c6a62",
            "#083465",
          ],
        },
      })
      map.addLayer({
        id: "actividades-marca",
        type: "symbol",
        source: "actividades",
        layout: {
          "text-field": [
            "match",
            ["get", "tipo"],
            "riego",
            "R",
            "poda",
            "P",
            "limpieza",
            "L",
            "incidencia",
            "I",
            "inspeccion",
            "V",
            "·",
          ],
          "text-font": ["Open Sans Regular", "Arial Unicode MS Regular"],
          "text-size": 11,
        },
        paint: { "text-color": "#f6f3ec" },
      })
      const fills = LAYERS.map((layer) => `${layer.id}-fill`)
      const invFills = INVENTARIO.filter((layer) => layer.kind === "polygon").map((layer) => `inv-${layer.id}-fill`)
      const openInventory = (lngLat: { lng: number; lat: number }, layerId: string, props: Record<string, unknown>) => {
        popupRef.current?.remove()
        const spec = INVENTARIO.find((layer) => layer.id === layerId)
        onActividad.current(null)
        onCatastro.current({ layer: spec?.label ?? "Inventario", props })
        const foto = typeof props.foto === "string" ? props.foto : ""
        const fotoHtml = foto
          ? `<img alt="" src="${apiUrl(`/geo/inventario/fotos/${encodeURIComponent(foto)}`)}" />`
          : spec?.id === "bebederos"
            ? `<p class="cv-popup-meta">Sin fotografía recuperada</p>`
            : ""
        const popup = new Popup({ closeButton: true, maxWidth: "280px", className: "cv-popup" })
          .setLngLat(lngLat)
          .setHTML(
            `<p class="cv-popup-kicker">${escapeHtml(spec?.label ?? "Inventario")}${props.subtipo ? ` · ${escapeHtml(props.subtipo)}` : ""}</p>
             <p class="cv-popup-title">${escapeHtml(props.nombre || "Elemento")}</p>
             <p class="cv-popup-meta">${escapeHtml(props.lugar || props.detalle || "")}</p>
             ${fotoHtml}`,
          )
          .addTo(map)
        popupRef.current = popup
      }
      const limpiarHover = () => {
        if (!hoverRef.current) return
        map.setFeatureState(hoverRef.current, { hover: false })
        hoverRef.current = null
      }
      map.on("mousemove", (event: MapMouseEvent) => {
        if (pinRef.current || dibujoRef.current) {
          map.getCanvas().style.cursor = pinRef.current ? "crosshair" : ""
          limpiarHover()
          return
        }
        const near = nearestInventoryPoint(map, event.point, inventoryRef.current, inventoryOnRef.current)
        const hits = map.queryRenderedFeatures(event.point, { layers: ["actividades-circle", ...invFills, ...fills] })
        map.getCanvas().style.cursor = near || hits.length ? "pointer" : ""
        const catastroHit = map.queryRenderedFeatures(event.point, { layers: CATASTRO_FILLS })[0]
        if (catastroHit?.id != null) {
          const siguiente = { source: String(catastroHit.source), id: catastroHit.id }
          if (!hoverRef.current || hoverRef.current.source !== siguiente.source || hoverRef.current.id !== siguiente.id) {
            limpiarHover()
            map.setFeatureState(siguiente, { hover: true })
            hoverRef.current = siguiente
          }
        } else {
          limpiarHover()
        }
      })
      map.on("mouseout", limpiarHover)
      map.on("click", (event: MapMouseEvent) => {
        if (dibujoRef.current) return
        if (pinRef.current) {
          onPinRef.current(event.lngLat.lng, event.lngLat.lat)
          return
        }
        popupRef.current?.remove()
        const acts = map.queryRenderedFeatures(event.point, { layers: ["actividades-circle"] })
        if (acts.length) {
          const props = (acts[0].properties ?? {}) as Record<string, unknown>
          const id = String(props.id ?? acts[0].id ?? "")
          onActividad.current(id || null)
          onCatastro.current(null)
          const popup = new Popup({ closeButton: true, maxWidth: "280px", className: "cv-popup" })
            .setLngLat(event.lngLat)
            .setHTML(
              `<p class="cv-popup-kicker">${escapeHtml(etiquetaTipo(String(props.tipo ?? "")))} · ${escapeHtml(etiquetaEstado(String(props.estado ?? "")))}</p>
               <p class="cv-popup-title">${escapeHtml(props.titulo || ACTIVIDAD.sinTitulo)}</p>
               <p class="cv-popup-meta">${escapeHtml(props.equipo || ACTIVIDAD.sinCuadrilla)}</p>`,
            )
            .addTo(map)
          popupRef.current = popup
          return
        }
        const near = nearestInventoryPoint(map, event.point, inventoryRef.current, inventoryOnRef.current)
        if (near) {
          openInventory(event.lngLat, near.id, near.props)
          return
        }
        const polyHits = invFills.length ? map.queryRenderedFeatures(event.point, { layers: invFills }) : []
        if (polyHits.length) {
          const source = String(polyHits[0].source).replace(/^inv-/, "")
          openInventory(event.lngLat, source, (polyHits[0].properties ?? {}) as Record<string, unknown>)
          return
        }
        const hits = map.queryRenderedFeatures(event.point, { layers: fills })
        if (!hits.length) {
          onCatastro.current(null)
          onActividad.current(null)
          if (selRef.current) {
            map.setFeatureState(selRef.current, { sel: false })
            selRef.current = null
          }
          return
        }
        const hit = hits[0]
        const props = (hit.properties ?? {}) as Record<string, unknown>
        onActividad.current(null)
        onCatastro.current({ layer: labelOf(String(hit.source)), props })
        const esZona = String(hit.source) === "zonas"
        if (hit.id != null) {
          const siguiente = { source: String(hit.source), id: hit.id }
          if (selRef.current && (selRef.current.source !== siguiente.source || selRef.current.id !== siguiente.id)) {
            map.setFeatureState(selRef.current, { sel: false })
          }
          map.setFeatureState(siguiente, { sel: true })
          selRef.current = siguiente
        }
        const crudo = String(props.nombre ?? "").trim()
        const nombre = crudo || String(props.feature_id ?? props.codigo ?? "Área sin nombre")
        const codigo = props.codigo ? String(props.codigo) : "sin código"
        const uso = props.uso ? String(props.uso) : ""
        const categoria = esZona
          ? `Sector de capataz: ${props.sector_etiqueta ?? etiquetaCategoria("sector", String(props.cat_sector ?? ""), sectoresRef.current)}`
          : `Categoría: ${escapeHtml(etiquetaCategoria("uso", String(props.cat_uso ?? "")))}`
        const popup = new Popup({ closeButton: true, maxWidth: "280px", className: "cv-popup" })
          .setLngLat(event.lngLat)
          .setHTML(
            `<p class="cv-popup-kicker">${escapeHtml(labelOf(String(hit.source)))}</p>
             <p class="cv-popup-title">${escapeHtml(nombre)}</p>
             <p class="cv-popup-meta">${escapeHtml(codigo)}${uso ? ` · ${escapeHtml(uso)}` : ""}</p>
             <p class="cv-popup-meta">${categoria}</p>`,
          )
          .addTo(map)
        popup.on("close", () => {
          if (popupRef.current === popup && selRef.current) {
            map.setFeatureState(selRef.current, { sel: false })
            selRef.current = null
          }
        })
        popupRef.current = popup
      })
      setReady(true)
      map.resize()
    })
    mapRef.current = map
    return () => {
      setReady(false)
      markerRef.current?.remove()
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready) return
    for (const layer of LAYERS) {
      const esCatastro = layer.id === "areas" || layer.id === "zonas"
      const source = map.getSource(layer.id) as GeoJSONSource | undefined
      if (source) source.setData(asCollection(data[layer.id]))
      const showFill = visible[layer.id] && !(relieve && esCatastro)
      const vis = showFill ? "visible" : "none"
      if (map.getLayer(`${layer.id}-fill`)) {
        map.setLayoutProperty(`${layer.id}-fill`, "visibility", vis)
      }
      if (map.getLayer(`${layer.id}-line`)) {
        map.setLayoutProperty(`${layer.id}-line`, "visibility", visible[layer.id] ? "visible" : "none")
      }
      if (esCatastro && map.getLayer(`${layer.id}-realce`)) {
        map.setLayoutProperty(`${layer.id}-realce`, "visibility", visible[layer.id] ? "visible" : "none")
      }
      if (esCatastro) {
        const campo = layer.id === "zonas" ? "cat_sector" : "cat_uso"
        const todos = layer.id === "zonas" ? sectores.map((cat) => cat.id) : USO_IDS
        const filtro = filtroCategorias(campo, todos.filter((id) => !ocultas.includes(id)), todos)
        for (const capa of [`${layer.id}-fill`, `${layer.id}-line`, `${layer.id}-realce`, `${layer.id}-extrusion`]) {
          if (map.getLayer(capa)) map.setFilter(capa, filtro)
        }
      }
    }
    if (map.getLayer("areas-extrusion")) {
      map.setLayoutProperty("areas-extrusion", "visibility", relieve && visible.areas ? "visible" : "none")
    }
    if (map.getLayer("zonas-extrusion")) {
      map.setLayoutProperty("zonas-extrusion", "visibility", relieve && visible.zonas ? "visible" : "none")
    }
    if (map.getLayer("zonas-fill")) {
      map.setPaintProperty("zonas-fill", "fill-color", expresionColor("cat_sector", sectores, "fill"))
      map.setPaintProperty("zonas-line", "line-color", expresionColor("cat_sector", sectores, "line"))
      map.setPaintProperty("zonas-extrusion", "fill-extrusion-color", [
        "case",
        ["boolean", ["feature-state", "sel"], false],
        "#083465",
        expresionColor("cat_sector", sectores, "fill"),
      ])
    }
    const conteo = conteoCapas(data)
    if (!encuadrado.current && catastroVisible(conteo)) {
      encuadrado.current = true
      map.fitBounds(CAMPUS_BOUNDS, { padding: 36, duration: 0, maxZoom: 17 })
    }
    const acts = map.getSource("actividades") as GeoJSONSource | undefined
    acts?.setData(asCollection(activities))
    const buildings = map.getSource("edificios") as GeoJSONSource | undefined
    buildings?.setData(asCollection(edificios))
    if (map.getLayer("edificios-extrusion")) {
      map.setLayoutProperty("edificios-extrusion", "visibility", relieve && showEdificios ? "visible" : "none")
    }
    for (const layer of INVENTARIO) {
      const source = map.getSource(`inv-${layer.id}`) as GeoJSONSource | undefined
      source?.setData(asCollection(inventory[layer.id]))
      const vis = inventoryOn[layer.id] ? "visible" : "none"
      for (const suffix of ["circle", "fill", "line"]) {
        if (map.getLayer(`inv-${layer.id}-${suffix}`)) {
          map.setLayoutProperty(`inv-${layer.id}-${suffix}`, "visibility", vis)
        }
      }
    }
  }, [data, visible, activities, ready, relieve, edificios, showEdificios, inventory, inventoryOn, ocultas, sectores])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready) return
    const quiet = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (quiet) map.jumpTo({ pitch: relieve ? 52 : 0, bearing: relieve ? -18 : 0 })
    else map.easeTo({ pitch: relieve ? 52 : 0, bearing: relieve ? -18 : 0, duration: 650 })
  }, [relieve, ready])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready) return
    map.getCanvas().style.cursor = pinMode ? "crosshair" : ""
  }, [pinMode, ready])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready) return
    markerRef.current?.remove()
    markerRef.current = null
    if (!draft) return
    const el = document.createElement("div")
    el.className = "draft-pin"
    markerRef.current = new Marker({ element: el }).setLngLat([draft.lon, draft.lat]).addTo(map)
  }, [draft, ready])

  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready || !focus) return
    const quiet = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    const zoom = Math.max(map.getZoom(), 16.4)
    const offset = desfaseHoja(map.getContainer())
    if (quiet) map.easeTo({ center: [focus.lon, focus.lat], zoom, offset, duration: 0 })
    else map.flyTo({ center: [focus.lon, focus.lat], zoom, offset, duration: 900, essential: true })
  }, [focus, ready])

  const dibujoId = modoDibujo ? `${modoDibujo.entidad}:${modoDibujo.id}` : ""
  const sesionDibujo = useRef<ReturnType<typeof activarDibujo> | null>(null)
  useEffect(() => {
    const map = mapRef.current
    if (!map || !ready) return
    sesionDibujo.current?.cerrar()
    sesionDibujo.current = dibujoRef.current ? activarDibujo(map, dibujoRef.current) : null
    return () => {
      sesionDibujo.current?.cerrar()
      sesionDibujo.current = null
    }
  }, [dibujoId, ready])
  useEffect(() => {
    sesionDibujo.current?.actualizar(modoDibujo?.geom ?? null)
  }, [modoDibujo?.geom])

  return <div ref={host} className="map-host" />
}
