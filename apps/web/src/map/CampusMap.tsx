import { useEffect, useRef, useState } from "react"
import type { FeatureCollection as GJFeatureCollection } from "geojson"
import {
  GeoJSONSource,
  Map,
  MapMouseEvent,
  Marker,
  NavigationControl,
  Popup,
  ScaleControl,
} from "maplibre-gl"
import "maplibre-gl/dist/maplibre-gl.css"
import { INVENTARIO } from "../inventario"
import { etiquetaEstado, etiquetaTipo } from "../operacion"
import { EMPTY, LAYERS, type FeatureCollection, type LayerId } from "../types"

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
  onSelectCatastro: (hit: { layer: string; props: Record<string, unknown> } | null) => void
  onSelectActividad: (id: string | null) => void
  onPin: (lon: number, lat: number) => void
}

const CAMPUS: [number, number] = [-77.0796, -12.0696]

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
}: Props) {
  const host = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<Map | null>(null)
  const popupRef = useRef<Popup | null>(null)
  const markerRef = useRef<Marker | null>(null)
  const [ready, setReady] = useState(false)
  const pinRef = useRef(pinMode)
  const onCatastro = useRef(onSelectCatastro)
  const onActividad = useRef(onSelectActividad)
  const onPinRef = useRef(onPin)
  pinRef.current = pinMode
  const inventoryRef = useRef(inventory)
  const inventoryOnRef = useRef(inventoryOn)
  inventoryRef.current = inventory
  inventoryOnRef.current = inventoryOn
  onCatastro.current = onSelectCatastro
  onActividad.current = onSelectActividad
  onPinRef.current = onPin

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
      for (const layer of LAYERS) {
        map.addSource(layer.id, { type: "geojson", data: { type: "FeatureCollection", features: [] } })
        map.addLayer({
          id: `${layer.id}-fill`,
          type: "fill",
          source: layer.id,
          paint: { "fill-color": layer.fill, "fill-opacity": layer.fillOpacity },
        })
        map.addLayer({
          id: `${layer.id}-line`,
          type: "line",
          source: layer.id,
          paint: { "line-color": layer.line, "line-width": layer.id === "zonas" ? 1.1 : 1.25 },
        })
        if (layer.id === "zonas") map.setPaintProperty(`${layer.id}-line`, "line-dasharray", [1.4, 1.1])
      }
      map.addLayer({
        id: "areas-extrusion",
        type: "fill-extrusion",
        source: "areas",
        layout: { visibility: "none" },
        paint: {
          "fill-extrusion-color": "#1a5c44",
          "fill-extrusion-opacity": 0.8,
          "fill-extrusion-height": [
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
          ],
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
            "#1a5c44",
            "bloqueada",
            "#8c3832",
            "cerrada",
            "#5c6a62",
            "cancelada",
            "#5c6a62",
            "#1a2420",
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
          ? `<img alt="" src="/api/v1/geo/inventario/fotos/${encodeURIComponent(foto)}" />`
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
      map.on("mousemove", (event: MapMouseEvent) => {
        if (pinRef.current) {
          map.getCanvas().style.cursor = "crosshair"
          return
        }
        const near = nearestInventoryPoint(map, event.point, inventoryRef.current, inventoryOnRef.current)
        const hits = map.queryRenderedFeatures(event.point, { layers: ["actividades-circle", ...invFills, ...fills] })
        map.getCanvas().style.cursor = near || hits.length ? "pointer" : ""
      })
      map.on("click", (event: MapMouseEvent) => {
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
               <p class="cv-popup-title">${escapeHtml(props.titulo || "Labor")}</p>
               <p class="cv-popup-meta">${escapeHtml(props.equipo || "Sin equipo")}</p>`,
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
          return
        }
        const hit = hits[0]
        const props = (hit.properties ?? {}) as Record<string, unknown>
        onActividad.current(null)
        onCatastro.current({ layer: labelOf(String(hit.source)), props })
        const crudo = String(props.nombre ?? "").trim()
        const nombre = crudo || String(props.feature_id ?? props.codigo ?? "Área sin nombre")
        const codigo = props.codigo ? String(props.codigo) : "sin código"
        const uso = props.uso ? String(props.uso) : ""
        const popup = new Popup({ closeButton: true, maxWidth: "280px", className: "cv-popup" })
          .setLngLat(event.lngLat)
          .setHTML(
            `<p class="cv-popup-kicker">${escapeHtml(labelOf(String(hit.source)))}</p>
             <p class="cv-popup-title">${escapeHtml(nombre)}</p>
             <p class="cv-popup-meta">${escapeHtml(codigo)}${uso ? ` · ${escapeHtml(uso)}` : ""}</p>`,
          )
          .addTo(map)
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
      const source = map.getSource(layer.id) as GeoJSONSource | undefined
      if (source) source.setData(asCollection(data[layer.id]))
      const showFill = visible[layer.id] && !(relieve && layer.id === "areas")
      const vis = showFill ? "visible" : "none"
      if (map.getLayer(`${layer.id}-fill`)) {
        map.setLayoutProperty(`${layer.id}-fill`, "visibility", vis)
        map.setLayoutProperty(`${layer.id}-line`, "visibility", visible[layer.id] ? "visible" : "none")
      }
    }
    if (map.getLayer("areas-extrusion")) {
      map.setLayoutProperty("areas-extrusion", "visibility", relieve && visible.areas ? "visible" : "none")
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
  }, [data, visible, activities, ready, relieve, edificios, showEdificios, inventory, inventoryOn])

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
    if (quiet) map.jumpTo({ center: [focus.lon, focus.lat], zoom })
    else map.flyTo({ center: [focus.lon, focus.lat], zoom, duration: 900, essential: true })
  }, [focus, ready])

  return <div ref={host} className="map-host" />
}
