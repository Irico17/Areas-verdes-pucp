import type { Rol } from "../types"

export type Modulo =
  | "mapa"
  | "labores"
  | "catastro"
  | "inventario"
  | "solicitudes"
  | "reportes"
  | "catalogos"
  | "importaciones"
  | "admin"

export const MODULOS: { id: Modulo; label: string }[] = [
  { id: "mapa", label: "Mapa" },
  { id: "labores", label: "Labores" },
  { id: "catastro", label: "Catastro" },
  { id: "inventario", label: "Inventario" },
  { id: "solicitudes", label: "Solicitudes" },
  { id: "reportes", label: "Reportes" },
  { id: "catalogos", label: "Catálogos" },
  { id: "importaciones", label: "Importar" },
  { id: "admin", label: "Admin" },
]

export function modulosDe(rol: Rol): Modulo[] {
  if (rol === "capataz") return ["mapa", "labores", "catastro", "inventario"]
  if (rol === "jefatura") return ["mapa", "labores", "catastro", "inventario", "solicitudes", "reportes", "importaciones", "admin"]
  if (rol === "admin") return MODULOS.map((item) => item.id)
  return ["mapa", "labores", "catastro", "inventario", "solicitudes", "reportes", "catalogos", "importaciones"]
}
