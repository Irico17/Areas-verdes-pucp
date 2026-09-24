export type Rol = "jefatura" | "coordinacion" | "capataz"

export type FeatureProps = Record<string, unknown>

export type GeoFeature = {
  type: "Feature"
  id?: string | number
  geometry: { type: string; coordinates: unknown } | null
  properties: FeatureProps | null
}

export type FeatureCollection = {
  type: "FeatureCollection"
  name?: string
  features: GeoFeature[]
}

export const EMPTY: FeatureCollection = { type: "FeatureCollection", features: [] }

export type LayerId = "areas" | "zonas" | "jardines_reserva" | "xerofitica"

export type LayerSpec = {
  id: LayerId
  label: string
  hint: string
  path: string
  defaultOn: boolean
  fill: string
  line: string
  fillOpacity: number
}

export const LAYERS: LayerSpec[] = [
  {
    id: "areas",
    label: "Áreas verdes",
    hint: "Catastro principal",
    path: "/api/v1/geo/areas",
    defaultOn: true,
    fill: "#1e4d3a",
    line: "#10281e",
    fillOpacity: 0.32,
  },
  {
    id: "zonas",
    label: "Zonas",
    hint: "Sectores operativos, sin nombres de personas",
    path: "/api/v1/geo/zonas",
    defaultOn: true,
    fill: "#a8843d",
    line: "#5c4618",
    fillOpacity: 0.14,
  },
  {
    id: "jardines_reserva",
    label: "Jardines de reserva",
    hint: "Capa auxiliar",
    path: "/api/v1/geo/capas/jardines_reserva",
    defaultOn: false,
    fill: "#3d5c78",
    line: "#24384a",
    fillOpacity: 0.28,
  },
  {
    id: "xerofitica",
    label: "Xerofítica",
    hint: "Capa auxiliar",
    path: "/api/v1/geo/capas/xerofitica",
    defaultOn: false,
    fill: "#8c4a32",
    line: "#5a2e1e",
    fillOpacity: 0.4,
  },
]

export const ROLES: { id: Rol; label: string; note: string }[] = [
  {
    id: "jefatura",
    label: "Jefatura",
    note: "Ve todas las labores abiertas. Puede crear, reasignar y archivar.",
  },
  {
    id: "coordinacion",
    label: "Coordinación",
    note: "Crea labores con un pin, asigna el equipo y sigue la bitácora.",
  },
  {
    id: "capataz",
    label: "Capataz",
    note: "Solo ve las labores de su equipo. Puede cambiar el estado, no reasignar.",
  },
]
