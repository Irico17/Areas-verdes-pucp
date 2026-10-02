import { apiUrl } from "./api"
import { CAPA } from "./ui/nomenclatura"

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
    label: CAPA.areas.label,
    hint: CAPA.areas.hint,
    path: apiUrl("/geo/areas"),
    defaultOn: true,
    fill: "#308046",
    line: "#083465",
    fillOpacity: 0.55,
  },
  {
    id: "zonas",
    label: CAPA.zonas.label,
    hint: CAPA.zonas.hint,
    path: apiUrl("/geo/zonas"),
    defaultOn: true,
    fill: "#5a6e80",
    line: "#083465",
    fillOpacity: 0.14,
  },
  {
    id: "jardines_reserva",
    label: CAPA.jardinesReserva.label,
    hint: CAPA.jardinesReserva.hint,
    path: apiUrl("/geo/capas/jardines_reserva"),
    defaultOn: false,
    fill: "#2f4a44",
    line: "#1a3330",
    fillOpacity: 0.28,
  },
  {
    id: "xerofitica",
    label: CAPA.xerofitica.label,
    hint: CAPA.xerofitica.hint,
    path: apiUrl("/geo/capas/xerofitica"),
    defaultOn: false,
    fill: "#6e5344",
    line: "#4a3428",
    fillOpacity: 0.4,
  },
]

export const ROLES: { id: Rol; label: string; note: string }[] = [
  {
    id: "jefatura",
    label: "Jefatura de sección",
    note: "Ve todas las actividades abiertas. Puede crear, reasignar y archivar.",
  },
  {
    id: "coordinacion",
    label: "Ingeniería / Coordinación",
    note: "Crea actividades con un pin, asigna la cuadrilla y sigue la bitácora.",
  },
  {
    id: "capataz",
    label: "Capataz",
    note: "Solo ve las actividades de su cuadrilla. Puede cambiar el estado, no reasignar.",
  },
  {
    id: "admin",
    label: "Administrador del sistema",
    note: "Gestión técnica y configuración de accesos.",
  },
]
