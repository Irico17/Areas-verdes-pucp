import { useEffect, useRef } from "react"
import type { FeatureCollection as GJFeatureCollection } from "geojson"
import { GeoJSONSource, Map, MapMouseEvent, NavigationControl, Popup, ScaleControl } from "maplibre-gl"
import "maplibre-gl/dist/maplibre-gl.css"
import { LAYERS, type FeatureCollection, type LayerId } from "../types"

type Props = {
  data: Partial<Record<LayerId, FeatureCollection>>
  visible: Record<LayerId, boolean>
  onSelect: (hit: { layer: string; props: Record<string, unknown> } | null) => void
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

export function CampusMap({ data, visible, onSelect }: Props) {
  const host = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<Map | null>(null)
  const popupRef = useRef<Popup | null>(null)
  const ready = useRef(false)
  const onSelectRef = useRef(onSelect)
  onSelectRef.current = onSelect

  useEffect(() => {
    if (!host.current || mapRef.current) return
    const map = new Map({
      container: host.current,
      center: CAMPUS,
      zoom: 15.4,
      style: {
        version: 8,
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
        map.addSource(layer.id, {
          type: "geojson",
          data: { type: "FeatureCollection", features: [] },
        })
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
          layout: layer.id === "zonas" ? { "line-cap": "butt" } : undefined,
        })
        if (layer.id === "zonas") {
          map.setPaintProperty(`${layer.id}-line`, "line-dasharray", [1.4, 1.1])
        }
      }
      const fills = LAYERS.map((layer) => `${layer.id}-fill`)
      map.on("mousemove", (event: MapMouseEvent) => {
        const hits = map.queryRenderedFeatures(event.point, { layers: fills })
        map.getCanvas().style.cursor = hits.length ? "pointer" : ""
      })
      map.on("click", (event: MapMouseEvent) => {
        const hits = map.queryRenderedFeatures(event.point, { layers: fills })
        popupRef.current?.remove()
        if (!hits.length) {
          onSelectRef.current(null)
          return
        }
        const hit = hits[0]
        const source = String(hit.source)
        const props = (hit.properties ?? {}) as Record<string, unknown>
        onSelectRef.current({ layer: labelOf(source), props })
        const nombre = props.nombre || "Sin nombre"
        const codigo = props.codigo ? String(props.codigo) : "sin código"
        const uso = props.uso ? String(props.uso) : ""
        const popup = new Popup({ closeButton: true, maxWidth: "280px", className: "cv-popup" })
          .setLngLat(event.lngLat)
          .setHTML(
            `<p class="cv-popup-kicker">${escapeHtml(labelOf(source))}</p>
             <p class="cv-popup-title">${escapeHtml(nombre)}</p>
             <p class="cv-popup-meta">${escapeHtml(codigo)}${uso ? ` · ${escapeHtml(uso)}` : ""}</p>`,
          )
          .addTo(map)
        popupRef.current = popup
      })
      ready.current = true
      map.resize()
    })
    mapRef.current = map
    return () => {
      ready.current = false
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return
    const apply = () => {
      for (const layer of LAYERS) {
        const source = map.getSource(layer.id) as GeoJSONSource | undefined
        const fc = data[layer.id]
        if (source && fc) source.setData(fc as unknown as GJFeatureCollection)
        const vis = visible[layer.id] ? "visible" : "none"
        if (map.getLayer(`${layer.id}-fill`)) {
          map.setLayoutProperty(`${layer.id}-fill`, "visibility", vis)
          map.setLayoutProperty(`${layer.id}-line`, "visibility", vis)
        }
      }
    }
    if (ready.current && map.isStyleLoaded()) apply()
    else map.once("load", apply)
  }, [data, visible])

  return <div ref={host} className="map-host" />
}
