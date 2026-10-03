import { CAPA } from "./ui/nomenclatura"

export const INVENTARIO = [
  { id: "bebederos", label: CAPA.bebederos.label, hint: CAPA.bebederos.hint, color: "#2f6b4f", kind: "point" as const },
  { id: "fauna", label: CAPA.fauna.label, hint: CAPA.fauna.hint, color: "#3d5c78", kind: "point" as const },
  { id: "puertas", label: CAPA.puertas.label, hint: CAPA.puertas.hint, color: "#1c211e", kind: "point" as const },
  { id: "tachos", label: CAPA.tachos.label, hint: CAPA.tachos.hint, color: "#5c6560", kind: "point" as const },
  { id: "flora", label: CAPA.flora.label, hint: CAPA.flora.hint, color: "#1e4d3a", kind: "point" as const },
  { id: "cafetos", label: CAPA.cafetos.label, hint: CAPA.cafetos.hint, color: "#6b4a2f", kind: "point" as const },
  { id: "playas_estacionamiento", label: CAPA.playas.label, hint: CAPA.playas.hint, color: "#8a8478", kind: "polygon" as const },
  { id: "area_vereda_peligro", label: CAPA.vereda.label, hint: CAPA.vereda.hint, color: "#8c3a32", kind: "polygon" as const },
  { id: "xerofitica", label: CAPA.xerofitica.label, hint: "Clase, riego, área y perímetro", color: "#8a9a62", kind: "polygon" as const },
  { id: "jardines_reserva", label: CAPA.jardinesReserva.label, hint: "Pertenecen a una unidad, no a una persona", color: "#1e4d3a", kind: "polygon" as const },
  { id: "puntos_pucp", label: CAPA.puntos.label, hint: CAPA.puntos.hint, color: "#6b4a2f", kind: "point" as const },
]
