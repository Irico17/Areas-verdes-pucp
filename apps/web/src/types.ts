export type Rol = "jefatura" | "coordinacion" | "capataz" | "admin"

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
    fill: "#1a5c44",
    line: "#0e3b2c",
    fillOpacity: 0.32,
  },
  {
    id: "zonas",
    label: "Zonas",
    hint: "Sectores operativos, sin nombres de personas",
    path: "/api/v1/geo/zonas",
    defaultOn: true,
    fill: "#5c6a62",
    line: "#3e5148",
    fillOpacity: 0.14,
  },
  {
    id: "jardines_reserva",
    label: "Jardines de reserva",
    hint: "Capa auxiliar",
    path: "/api/v1/geo/capas/jardines_reserva",
    defaultOn: false,
    fill: "#2f4a44",
    line: "#1a3330",
    fillOpacity: 0.28,
  },
  {
    id: "xerofitica",
    label: "Xerofítica",
    hint: "Capa auxiliar",
    path: "/api/v1/geo/capas/xerofitica",
    defaultOn: false,
    fill: "#6e5344",
    line: "#4a3428",
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
