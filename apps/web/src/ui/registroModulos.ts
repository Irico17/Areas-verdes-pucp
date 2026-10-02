import type { Rol } from "../types"
import { MODULO } from "./nomenclatura"

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
  { id: "mapa", label: MODULO.mapa },
  { id: "labores", label: MODULO.labores },
  { id: "catastro", label: MODULO.catastro },
  { id: "inventario", label: MODULO.inventario },
  { id: "solicitudes", label: MODULO.solicitudes },
  { id: "reportes", label: MODULO.reportes },
  { id: "catalogos", label: MODULO.catalogos },
  { id: "importaciones", label: MODULO.importaciones },
  { id: "admin", label: MODULO.admin },
]

export function modulosDe(rol: Rol): Modulo[] {
  if (rol === "capataz") return ["mapa", "labores", "catastro", "inventario"]
  if (rol === "jefatura") return ["mapa", "labores", "catastro", "inventario", "solicitudes", "reportes", "importaciones", "admin"]
  if (rol === "admin") return MODULOS.map((item) => item.id)
  return ["mapa", "labores", "catastro", "inventario", "solicitudes", "reportes", "catalogos", "importaciones"]
}
